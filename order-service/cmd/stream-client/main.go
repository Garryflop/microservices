package main

import (
	"context"
	"flag"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "github.com/Garryflop/microservices-gen-go/order"
)

func main() {
	addr := flag.String("addr", "localhost:50052", "order gRPC server address")
	orderID := flag.String("order", "", "order ID to subscribe to")
	flag.Parse()

	if *orderID == "" {
		log.Fatal("provide -order flag with order ID")
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := orderpb.NewOrderTrackingServiceClient(conn)

	stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderpb.OrderRequest{
		OrderId: *orderID,
	})
	if err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	log.Printf("subscribed to order %s updates...", *orderID)

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream ended")
			break
		}
		if err != nil {
			log.Fatalf("stream error: %v", err)
		}

		log.Printf("status update: %s (at %s)", update.GetStatus(), update.GetUpdatedAt().AsTime().Format("15:04:05"))
	}
}
