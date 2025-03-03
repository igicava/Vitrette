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
)

func main() {
	// run
	ctx, err := logger.NewContext(context.Background())
	if err != nil {
		panic(err)
	}

	conn, err := config.New()
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "config.New() error", zap.Error(err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", conn.GRPCPort))
	if err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to listen", zap.Error(err))
	}

	db := mapdb.NewMap()

	srv := service.NewService(db)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(logger.Interceptor))
	pb.RegisterOrderServiceServer(grpcServer, srv)

	logger.GetLogger(ctx).Info(ctx, "Server started")

	if err := grpcServer.Serve(listener); err != nil {
		logger.GetLogger(ctx).Fatal(ctx, "Failed to serve", zap.Error(err))
	}
}
