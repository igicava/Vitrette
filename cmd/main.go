package main

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	pb "lyceum/pkg/api/test/api"
	"lyceum/pkg/logger"
	"lyceum/pkg/mapdb"
	"lyceum/service"
	"net"
)

func main() {
	// run
	ctx := logger.NewContext()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		ctx.Value("logger").(*zap.Logger).Fatal("Failed to listen", zap.Error(err))
	}

	db := mapdb.NewMap()

	srv := service.NewService(db)

	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, srv)

	ctx.Value("logger").(*zap.Logger).Info("Server started")

	if err := grpcServer.Serve(listener); err != nil {
		ctx.Value("logger").(*zap.Logger).Fatal("Failed to serve", zap.Error(err))
	}
}
