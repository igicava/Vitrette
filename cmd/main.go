package main

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"lyceum/internal/config"
	"lyceum/internal/service"
	pb "lyceum/pkg/api"
	"lyceum/pkg/logger"
	"lyceum/pkg/mapdb"
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

	db := mapdb.NewMap()

	srv := service.NewService(db)

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

	logger.GetLogger(ctx).Info(ctx, "Starting server...")
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
	logger.GetLogger(ctx).Info(ctx, "Starting server grpc gateway on port 8081...")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to serve gateway", zap.Error(err))
	}
}
