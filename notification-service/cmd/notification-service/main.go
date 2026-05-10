package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"notification-service/internal/config"
	"notification-service/internal/handler"
	"notification-service/internal/infrastructure"
	"notification-service/internal/provider"
	"notification-service/internal/store"
)

func main() {
	cfg := config.Load()

	// Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	ctx := context.Background()
	for i := 0; i < 30; i++ {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			break
		}
		log.Println("waiting for Redis...")
		time.Sleep(1 * time.Second)
	}
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis not reachable: %v", err)
	}
	log.Println("Redis connected")

	// select email provider based on PROVIDER_MODE
	var emailSender provider.EmailSender
	if strings.ToUpper(cfg.ProviderMode) == "REAL" {
		emailSender = provider.NewSMTPEmailSender(
			cfg.SMTPHost, cfg.SMTPPort,
			cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom,
		)
		log.Println("[Notification] Using REAL SMTP provider")
	} else {
		emailSender = provider.NewMockEmailSender()
		log.Println("[Notification] Using SIMULATED provider")
	}

	// dependency injection
	idempotencyStore := store.NewRedisIdempotencyStore(redisClient)
	notificationHandler := handler.NewNotificationHandler(idempotencyStore, emailSender, cfg.MaxRetries)

	consumer, err := infrastructure.NewRabbitMQConsumer(cfg.RabbitMQURL, notificationHandler)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	consumerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("[Notification] Shutting down gracefully...")
		cancel()
	}()

	log.Println("[Notification] Service started")
	if err := consumer.Start(consumerCtx); err != nil {
		log.Fatalf("consumer error: %v", err)
	}

	redisClient.Close()
	log.Println("[Notification] Graceful shutdown complete")
}
