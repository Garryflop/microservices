package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/config"
	"notification-service/internal/handler"
	"notification-service/internal/infrastructure"
)

func main() {
	cfg := config.Load()

	// Manual Dependency Injection
	notificationHandler := handler.NewNotificationHandler()

	consumer, err := infrastructure.NewRabbitMQConsumer(cfg.RabbitMQURL, notificationHandler)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("[Notification] Shutting down gracefully...")
		cancel()
	}()

	log.Println("[Notification] Service started")
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("consumer error: %v", err)
	}

	log.Println("[Notification] Graceful shutdown complete")
}
