package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	pb "lyceum/pkg/api"
	"lyceum/pkg/logger"
	"os"
)

type DataBaseInterface interface {
	Create(ctx context.Context, id string, item string, quantity int32)
	Get(ctx context.Context, id string) (*pb.Order, error)
	Update(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) []*pb.Order
}

type Service struct {
	pb.UnimplementedOrderServiceServer
	DB          DataBaseInterface
	StreamStart chan os.Signal
}

func NewService(db DataBaseInterface) *Service {
	return &Service{DB: db, StreamStart: make(chan os.Signal, 1)}
}

func (s *Service) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	u := uuid.New()
	id := u.String()
	if req.Item == "" || req.Quantity == 0 {
		logger.GetLogger(ctx).Error(ctx, "Item and Quantity must not be empty")
		return nil, errors.New("item and quantity must not be empty")
	}
	s.DB.Create(ctx, id, req.Item, req.Quantity)
	return &pb.CreateOrderResponse{Id: id}, nil
}

func (s *Service) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	item, err := s.DB.Get(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "GetOrder failed: %v", zap.Error(err))
		return nil, err
	}
	order := &pb.Order{
		Id:       req.Id,
		Item:     item.Item,
		Quantity: item.Quantity,
	}
	return &pb.GetOrderResponse{Order: order}, nil
}

func (s *Service) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	_, err := s.DB.Get(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "GetOrder failed: %v", zap.Error(err))
		return nil, err
	}

	if req.Item == "" || req.Quantity == 0 {
		logger.GetLogger(ctx).Error(ctx, "Item and Quantity must not be empty")
		return nil, errors.New("item and quantity must not be empty")
	}

	item, err := s.DB.Update(ctx, req.Id, req.Item, req.Quantity)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "UpdateOrder failed: %v", zap.Error(err))
		return nil, err
	}

	order := &pb.Order{
		Id:       req.Id,
		Item:     item.Item,
		Quantity: item.Quantity,
	}

	return &pb.UpdateOrderResponse{Order: order}, nil
}

func (s *Service) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	_, err := s.DB.Get(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "GetOrder failed: %v", zap.Error(err))
		return nil, err
	}
	err = s.DB.Delete(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "DeleteOrder failed: %v", zap.Error(err))
		return nil, err
	}
	return &pb.DeleteOrderResponse{Success: true}, nil
}

func (s *Service) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders := s.DB.List(ctx)
	return &pb.ListOrdersResponse{Orders: orders}, nil
}
