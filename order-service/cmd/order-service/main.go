package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"order-service/internal/config"
	"order-service/internal/infrastructure"
	"order-service/internal/repository"
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

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}
	paymentClient := infrastructure.NewHTTPPaymentClient(cfg.PaymentServiceURL, httpClient)

	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentClient)
	handler := transporthttp.NewHandler(orderUseCase)
	router := transporthttp.NewRouter(handler)

	log.Printf("Order Service listening on :%s", cfg.Port)
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
