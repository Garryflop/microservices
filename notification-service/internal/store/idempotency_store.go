package store

import "sync"

// IdempotencyStore tracks processed event IDs in memory to prevent
// duplicate processing. Uses sync.Map for thread-safe concurrent access.
//
// Trade-off: resets on service restart, but combined with manual ACKs
// this provides at-least-once delivery with best-effort deduplication.
type IdempotencyStore struct {
	processed sync.Map
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{}
}

// IsProcessed checks if an event has already been handled.
func (s *IdempotencyStore) IsProcessed(eventID string) bool {
	_, exists := s.processed.Load(eventID)
	return exists
}

// MarkProcessed records that an event has been successfully processed.
func (s *IdempotencyStore) MarkProcessed(eventID string) {
	s.processed.Store(eventID, true)
}
