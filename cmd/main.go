package main

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/redis/go-redis/v9"

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

	srv := service.NewService(db, *redisClient)

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
