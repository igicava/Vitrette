package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	pb "lyceum/pkg/api"
	"lyceum/pkg/logger"
	"os"
	"time"
)

type DataBaseInterface interface {
	Create(ctx context.Context, id string, item string, quantity int32) error
	Get(ctx context.Context, id string) (*pb.Order, error)
	Update(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*pb.Order, error)
}

type Service struct {
	pb.UnimplementedOrderServiceServer
	DB            DataBaseInterface
	Redis         *redis.Client
	KafkaProducer sarama.SyncProducer
	StreamStart   chan os.Signal
}

func NewService(db DataBaseInterface, client redis.Client, producer sarama.SyncProducer) *Service {
	return &Service{
		DB:            db,
		Redis:         &client,
		KafkaProducer: producer,
		StreamStart:   make(chan os.Signal, 1),
	}
}

func (s *Service) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	u := uuid.New()
	id := u.String()
	if req.Item == "" || req.Quantity == 0 {
		logger.GetLogger(ctx).Error(ctx, "Item and Quantity must not be empty")
		return nil, errors.New("item and quantity must not be empty")
	}
	err := s.DB.Create(ctx, id, req.Item, req.Quantity)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "CreateOrder failed", zap.Error(err))
		return nil, err
	}
	_, _, err = s.KafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: "orders",
		Key:   sarama.StringEncoder(id),
		Value: sarama.StringEncoder(req.Item),
	})
	return &pb.CreateOrderResponse{Id: id}, nil
}

func (s *Service) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	str, err := s.Redis.Get(ctx, req.Id).Result()

	if errors.Is(err, redis.Nil) {
		item, err := s.DB.Get(ctx, req.Id)
		if err != nil {
			logger.GetLogger(ctx).Error(ctx, "GetOrder failed", zap.Error(err))
			return nil, err
		}

		order := &pb.Order{
			Id:       req.Id,
			Item:     item.Item,
			Quantity: item.Quantity,
		}

		str, _ := json.Marshal(order)
		err = s.Redis.Set(ctx, req.Id, string(str), time.Minute).Err()
		if err != nil {
			logger.GetLogger(ctx).Error(ctx, "SetCacheOrder failed", zap.Error(err))
		}

		return &pb.GetOrderResponse{Order: order}, nil
	}

	fmt.Println("Value from cache!")
	orderFromCache := pb.Order{}
	err = json.Unmarshal([]byte(str), &orderFromCache)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "Unmarshal order failed", zap.Error(err))
		return nil, err
	}
	return &pb.GetOrderResponse{Order: &orderFromCache}, nil
}

func (s *Service) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	_, err := s.DB.Get(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "GetOrder failed", zap.Error(err))
		return nil, err
	}

	if req.Item == "" || req.Quantity == 0 {
		logger.GetLogger(ctx).Error(ctx, "Item and Quantity must not be empty")
		return nil, errors.New("item and quantity must not be empty")
	}

	item, err := s.DB.Update(ctx, req.Id, req.Item, req.Quantity)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "UpdateOrder failed", zap.Error(err))
		return nil, err
	}

	_, _, err = s.KafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: "orders",
		Key:   sarama.StringEncoder(req.Id),
		Value: sarama.StringEncoder(req.Item),
	})

	order := &pb.Order{
		Id:       req.Id,
		Item:     item.Item,
		Quantity: item.Quantity,
	}

	str, _ := json.Marshal(order)
	err = s.Redis.Set(ctx, req.Id, string(str), time.Minute).Err()
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "SetOrder failed", zap.Error(err))
		return nil, err
	}

	return &pb.UpdateOrderResponse{Order: order}, nil
}

func (s *Service) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	_, err := s.DB.Get(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "GetOrder failed", zap.Error(err))
		return nil, err
	}
	err = s.DB.Delete(ctx, req.Id)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "DeleteOrder failed", zap.Error(err))
		return nil, err
	}
	return &pb.DeleteOrderResponse{Success: true}, nil
}

func (s *Service) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, err := s.DB.List(ctx)
	if err != nil {
		logger.GetLogger(ctx).Error(ctx, "ListOrders failed", zap.Error(err))
		return nil, err
	}
	return &pb.ListOrdersResponse{Orders: orders}, nil
}
