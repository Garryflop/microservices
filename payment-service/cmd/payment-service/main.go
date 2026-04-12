package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"payment-service/internal/config"
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

	// Manual Dependency Injection
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo)

	// Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port: %v", err)
		}

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(middleware.LoggingInterceptor()),
		)
		grpcHandler := transportgrpc.NewServer(paymentUseCase)
		transportgrpc.RegisterServer(grpcServer, grpcHandler)

		log.Printf("Payment gRPC Server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start REST Server
	handler := transporthttp.NewHandler(paymentUseCase)
	router := transporthttp.NewRouter(handler)

	log.Printf("Payment REST Server listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(db *sql.DB) error {
	migration, err := os.ReadFile("migrations/001_create_payments.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(migration))
	return err
}
