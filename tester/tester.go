package tester

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	pb "lyceum/pkg/api/test/api"
)

func Tester() {
	// установим соединение
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("ERROR CONNECT")
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	grpcClient := pb.NewOrderServiceClient(conn)
	create, err := grpcClient.CreateOrder(context.Background(), &pb.CreateOrderRequest{
		Item:     "item1",
		Quantity: 100,
	})
	fmt.Println(create, err)

	get, err := grpcClient.GetOrder(context.Background(), &pb.GetOrderRequest{Id: create.Id})
	fmt.Println(get, err)

	update, err := grpcClient.UpdateOrder(context.Background(), &pb.UpdateOrderRequest{
		Id:       get.Order.Id,
		Quantity: 200,
		Item:     "item2",
	})
	fmt.Println(update, err)

	deleter, err := grpcClient.DeleteOrder(context.Background(), &pb.DeleteOrderRequest{
		Id: get.Order.Id,
	})
	fmt.Println(deleter, err)

	create2, err := grpcClient.CreateOrder(context.Background(), &pb.CreateOrderRequest{
		Item:     "item3",
		Quantity: 300,
	})
	fmt.Println(create2, err)

	create3, err := grpcClient.CreateOrder(context.Background(), &pb.CreateOrderRequest{
		Item:     "item4",
		Quantity: 400,
	})
	fmt.Println(create3, err)

	update, err = grpcClient.UpdateOrder(context.Background(), &pb.UpdateOrderRequest{
		Id:       create3.Id,
		Quantity: 500,
		Item:     "item5",
	})
	fmt.Println(update, err)

	get, err = grpcClient.GetOrder(context.Background(), &pb.GetOrderRequest{Id: create3.Id})
	fmt.Println("Pops", get, err)

	list, err := grpcClient.ListOrders(context.Background(), &pb.ListOrdersRequest{})
	fmt.Println(list, err)

}
