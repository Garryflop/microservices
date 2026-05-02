package store

import "sync"

// IdempotencyStore tracks processed event IDs in memory
type IdempotencyStore struct {
	processed sync.Map
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{}
}

func (s *IdempotencyStore) IsProcessed(eventID string) bool {
	_, exists := s.processed.Load(eventID)
	return exists
}

func (s *IdempotencyStore) MarkProcessed(eventID string) {
	s.processed.Store(eventID, true)
}
