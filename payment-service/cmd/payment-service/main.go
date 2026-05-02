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

	"payment-service/internal/config"
	"payment-service/internal/infrastructure"
	"payment-service/internal/middleware"
	"payment-service/internal/repository"
	transportgrpc "payment-service/internal/transport/grpc"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
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

	// RabbitMQ Publisher
	publisher, err := infrastructure.NewRabbitMQPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to create RabbitMQ publisher: %v", err)
	}

	// Manual Dependency Injection
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo, publisher)

	// Start gRPC Server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggingInterceptor()),
	)
	grpcHandler := transportgrpc.NewServer(paymentUseCase)
	transportgrpc.RegisterServer(grpcServer, grpcHandler)

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}
		log.Printf("Payment gRPC Server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start REST Server
	handler := transporthttp.NewHandler(paymentUseCase)
	router := transporthttp.NewRouter(handler)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Payment REST Server listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Payment] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	log.Println("[Payment] gRPC server stopped")

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("[Payment] HTTP server shutdown error: %v", err)
	}
	log.Println("[Payment] HTTP server stopped")

	publisher.Close()
	db.Close()

	log.Println("[Payment] Graceful shutdown complete")
}

func runMigrations(db *sql.DB) error {
	migration, err := os.ReadFile("migrations/001_create_payments.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(migration))
	return err
}

