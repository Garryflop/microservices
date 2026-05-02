package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// dependency injection
	orderRepo := repository.NewPostgresOrderRepository(db)

	paymentClient, err := infrastructure.NewGRPCPaymentClient(cfg.PaymentGRPCAddr)
	if err != nil {
		log.Fatalf("failed to create payment client: %v", err)
	}

	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentClient)

	// gRPC server
	grpcServer := grpc.NewServer()
	grpcHandler := transportgrpc.NewServer(orderUseCase)
	transportgrpc.RegisterServer(grpcServer, grpcHandler)

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}
		log.Printf("Order gRPC Server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// REST server
	handler := transporthttp.NewHandler(orderUseCase)
	router := transporthttp.NewRouter(handler)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Order REST Server listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Order] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	log.Println("[Order] gRPC server stopped")

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("[Order] HTTP server shutdown error: %v", err)
	}
	log.Println("[Order] HTTP server stopped")

	paymentClient.Close()
	db.Close()

	log.Println("[Order] Graceful shutdown complete")
}

func runMigrations(db *sql.DB) error {
	migration, err := os.ReadFile("migrations/001_create_orders.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(migration))
	return err
}
