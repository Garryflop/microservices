package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"order-service/internal/domain"
)

type RedisOrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOrderCache(client *redis.Client, ttlSeconds int) *RedisOrderCache {
	return &RedisOrderCache{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

func (c *RedisOrderCache) cacheKey(id string) string {
	return fmt.Sprintf("order:%s", id)
}

func (c *RedisOrderCache) Get(ctx context.Context, id string) (*domain.Order, error) {
	data, err := c.client.Get(ctx, c.cacheKey(id)).Bytes()
	if err == redis.Nil {
		log.Printf("[Cache] MISS for order %s, querying DB", id)
		return nil, fmt.Errorf("cache miss")
	}
	if err != nil {
		log.Printf("[Cache] Redis error for order %s: %v", id, err)
		return nil, err
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("unmarshal cached order: %w", err)
	}

	log.Printf("[Cache] HIT for order %s", id)
	return &order, nil
}

func (c *RedisOrderCache) Set(ctx context.Context, id string, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order for cache: %w", err)
	}

	if err := c.client.Set(ctx, c.cacheKey(id), data, c.ttl).Err(); err != nil {
		log.Printf("[Cache] Failed to set order %s: %v", id, err)
		return err
	}

	log.Printf("[Cache] SET order %s (TTL: %s)", id, c.ttl)
	return nil
}

func (c *RedisOrderCache) Delete(ctx context.Context, id string) error {
	if err := c.client.Del(ctx, c.cacheKey(id)).Err(); err != nil {
		log.Printf("[Cache] Failed to delete order %s: %v", id, err)
		return err
	}

	log.Printf("[Cache] INVALIDATED order %s", id)
	return nil
}
