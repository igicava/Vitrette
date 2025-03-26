package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"github.com/IBM/sarama"

	"lyceum/internal/config"
	"lyceum/internal/service"
	pb "lyceum/pkg/api"
	"lyceum/pkg/logger"
	"lyceum/pkg/postgres"

	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// run
	ctx, err := logger.NewContext(context.Background())
	if err != nil {
		panic(err)
	}

	conn, err := config.New()
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "config.New() error", zap.Error(err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", conn.GRPCPort))
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to listen", zap.Error(err))
	}

	pool, err := postgres.NewPool(ctx, conn.Postgres)
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to connect to postgres", zap.Error(err))
	}

	db := postgres.NewPG(pool)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("redis:%s", conn.Redis.Port),
		Password: conn.Redis.Password,
		DB:       conn.Redis.DB,
	})

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(
		[]string{fmt.Sprintf("kafka:%s", conn.Kafka.Port)},
		kafkaConfig,
	)
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to connect to kafka", zap.Error(err))
	}
	defer producer.Close()

	consumer, err := sarama.NewConsumer([]string{fmt.Sprintf("kafka:%s", conn.Kafka.Port)}, kafkaConfig)
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to connect to kafka", zap.Error(err))
	}
	defer consumer.Close()

	partConsumer, err := consumer.ConsumePartition("orders", 0, sarama.OffsetNewest)
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to connect to kafka", zap.Error(err))
	}
	defer partConsumer.Close()

	go func() {
		for {
			select {
			case msg, ok := <-partConsumer.Messages():
				if !ok {
					log.Println("Channel closed, exiting goroutine")
					return
				}
				fmt.Println(fmt.Sprintf("Key: %s, Value: %s, Topic: %s, Partition: %d, Offset: %d\n", msg.Key, msg.Value, msg.Topic, msg.Partition, msg.Offset))
			}
		}
	}()

	go func() {
		var (
			r  map[string]interface{}
			wg sync.WaitGroup
		)

		es, err := elasticsearch.NewDefaultClient()
		if err != nil {
			log.Fatalf("Error creating the client: %s", err)
		}

		// 1. Get cluster info
		//
		res, err := es.Info()
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		// Check response status
		if res.IsError() {
			log.Fatalf("Error: %s", res.String())
		}
		// Deserialize the response into a map.
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
		// Print client and server version numbers.
		log.Printf("Client: %s", elasticsearch.Version)
		log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
		log.Println(strings.Repeat("~", 37))

		// 2. Index documents concurrently
		//
		for i, title := range []string{"Test One", "Test Two"} {
			wg.Add(1)

			go func(i int, title string) {
				defer wg.Done()

				// Build the request body.
				data, err := json.Marshal(struct {
					Title string `json:"title"`
				}{Title: title})
				if err != nil {
					log.Fatalf("Error marshaling document: %s", err)
				}

				// Set up the request object.
				req := esapi.IndexRequest{
					Index:      "test",
					DocumentID: strconv.Itoa(i + 1),
					Body:       bytes.NewReader(data),
					Refresh:    "true",
				}

				// Perform the request with the client.
				res, err := req.Do(context.Background(), es)
				if err != nil {
					log.Fatalf("Error getting response: %s", err)
				}
				defer res.Body.Close()

				if res.IsError() {
					log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
				} else {
					// Deserialize the response into a map.
					var r map[string]interface{}
					if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
						log.Printf("Error parsing the response body: %s", err)
					} else {
						// Print the response status and indexed document version.
						log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
					}
				}
			}(i, title)
		}
		wg.Wait()

		log.Println(strings.Repeat("-", 37))

		// 3. Search for the indexed documents
		//
		// Build the request body.
		var buf bytes.Buffer
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"title": "test",
				},
			},
		}
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.Fatalf("Error encoding query: %s", err)
		}

		// Perform the search request.
		res, err = es.Search(
			es.Search.WithContext(context.Background()),
			es.Search.WithIndex("test"),
			es.Search.WithBody(&buf),
			es.Search.WithTrackTotalHits(true),
			es.Search.WithPretty(),
		)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				log.Fatalf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				log.Fatalf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		}

		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}
		// Print the response status, number of results, and request duration.
		log.Printf(
			"[%s] %d hits; took: %dms",
			res.Status(),
			int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(r["took"].(float64)),
		)
		// Print the ID and document source for each hit.
		for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
			log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
		}

		log.Println(strings.Repeat("=", 37))
	}()

	srv := service.NewService(db, *redisClient, producer)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor))
	pb.RegisterOrderServiceServer(grpcServer, srv)

	go func() {
		signal.Notify(srv.StreamStart, syscall.SIGINT, syscall.SIGTERM)
		<-srv.StreamStart
		time.Sleep(1 * time.Second)
		logger.GetLogger(ctx).Info(ctx, "Initiating graceful shutdown...")
		timer := time.AfterFunc(10*time.Second, func() {
			logger.GetLogger(ctx).Info(ctx, "Server couldn't stop gracefully in time. Doing force stop.")
			grpcServer.Stop()
		})
		defer timer.Stop()
		grpcServer.GracefulStop()
		fmt.Println("Server Stopped")
	}()

	go runRest(conn, ctx)

	logger.GetLogger(ctx).Info(ctx, "Starting server grpc...", zap.Int("port", conn.GRPCPort))
	if err := grpcServer.Serve(listener); err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to serve", zap.Error(err))
	}
}

func runRest(cfg *config.Config, ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf("0.0.0.0:%d", cfg.GRPCPort), opts)
	if err != nil {
		panic(err)
	}
	logger.GetLogger(ctx).Info(ctx, "Starting server grpc gateway...", zap.String("port", cfg.GATEWAYPort))
	conn := ":" + cfg.GATEWAYPort
	if err := http.ListenAndServe(conn, mux); err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to serve gateway", zap.Error(err))
	}
}
