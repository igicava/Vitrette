package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"lyceum/internal/config"
	"lyceum/internal/service"

	pb "lyceum/pkg/api/test/api"
	"lyceum/pkg/logger"
	"lyceum/pkg/mapdb"
	"net"
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

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", conn.GRPCPort))
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

	logger.GetLogger(ctx).Info(ctx, "Starting server...")
	if err := grpcServer.Serve(listener); err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to serve", zap.Error(err))
	}
}
