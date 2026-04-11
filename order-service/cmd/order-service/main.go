package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"order-service/internal/config"
	"order-service/internal/infrastructure"
	"order-service/internal/repository"
	transportgrpc "order-service/internal/transport/grpc"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("waiting for database...")
		time.Sleep(1 * time.Second)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("database not reachable: %v", err)
	}
	log.Println("database connected")

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	orderRepo := repository.NewPostgresOrderRepository(db)

	paymentClient, err := infrastructure.NewGRPCPaymentClient(cfg.PaymentGRPCAddr)
	if err != nil {
		log.Fatalf("failed to create payment client: %v", err)
	}
	defer paymentClient.Close()

	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentClient)

	//Start gRPC streaming server
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}

		grpcServer := grpc.NewServer()
		grpcHandler := transportgrpc.NewServer(orderUseCase)
		transportgrpc.RegisterServer(grpcServer, grpcHandler)

		log.Printf("Order gRPC Server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start REST server
	handler := transporthttp.NewHandler(orderUseCase)
	router := transporthttp.NewRouter(handler)

	log.Printf("Order REST Server listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	migration, err := os.ReadFile("migrations/001_create_orders.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(migration))
	return err
}
