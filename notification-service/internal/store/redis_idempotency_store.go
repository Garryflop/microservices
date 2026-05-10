package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"

	idempotencyTTL = 24 * time.Hour
)

// RedisIdempotencyStore uses Redis to track processed payment IDs.
// Unlike the in-memory sync.Map, this survives service restarts.
type RedisIdempotencyStore struct {
	client *redis.Client
}

func NewRedisIdempotencyStore(client *redis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{client: client}
}

func (s *RedisIdempotencyStore) idempotencyKey(paymentID string) string {
	return fmt.Sprintf("notification:processed:%s", paymentID)
}

// IsProcessed checks if a payment ID has already been successfully processed.
func (s *RedisIdempotencyStore) IsProcessed(ctx context.Context, paymentID string) bool {
	status, err := s.client.Get(ctx, s.idempotencyKey(paymentID)).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		log.Printf("[Idempotency] Redis error checking %s: %v", paymentID, err)
		return false
	}
	return status == StatusSent
}

// GetStatus returns the current processing status of a payment ID.
func (s *RedisIdempotencyStore) GetStatus(ctx context.Context, paymentID string) string {
	status, err := s.client.Get(ctx, s.idempotencyKey(paymentID)).Result()
	if err != nil {
		return ""
	}
	return status
}

// MarkStatus sets the processing status for a payment ID with a 24h TTL.
func (s *RedisIdempotencyStore) MarkStatus(ctx context.Context, paymentID, status string) {
	if err := s.client.Set(ctx, s.idempotencyKey(paymentID), status, idempotencyTTL).Err(); err != nil {
		log.Printf("[Idempotency] Failed to mark %s as %s: %v", paymentID, status, err)
	}
	log.Printf("[Idempotency] Payment %s marked as %s", paymentID, status)
}
