package main

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"lyceum/internal/service"
	pb "lyceum/pkg/api/test/api"
	"lyceum/pkg/logger"
	"lyceum/pkg/mapdb"
	"net"
)

func main() {
	// run
	ctx := logger.NewContext()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.GetLogger(ctx).Fatal("Failed to listen", zap.Error(err))
	}

	db := mapdb.NewMap()

	srv := service.NewService(db)

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, srv)

	logger.GetLogger(ctx).Info("Server started")

	if err := grpcServer.Serve(listener); err != nil {
		logger.GetLogger(ctx).Fatal("Failed to serve", zap.Error(err))
	}
}
