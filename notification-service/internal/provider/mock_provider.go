package provider

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
)


type MockEmailSender struct{}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{}
}

func (m *MockEmailSender) SendNotification(ctx context.Context, to, subject, body string) error {
	// simulate network latency (200-800ms)
	latency := time.Duration(200+rand.Intn(600)) * time.Millisecond
	time.Sleep(latency)

	// simulate ~20% failure rate
	if rand.Float64() < 0.2 {
		log.Printf("[MockProvider] Simulated failure sending to %s (latency: %s)", to, latency)
		return fmt.Errorf("simulated provider error: connection timeout")
	}

	log.Printf("[MockProvider] Email sent to %s | Subject: %s (latency: %s)", to, subject, latency)
	return nil
}
