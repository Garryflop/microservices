package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"notification-service/internal/domain"
	"notification-service/internal/handler"
)

// RabbitMQConsumer connects to RabbitMQ and consumes messages from the
// payment.completed queue. It delegates processing to the NotificationHandler.
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	handler *handler.NotificationHandler
}

// NewRabbitMQConsumer creates a new consumer with a connection to RabbitMQ.
// It retries the connection up to 30 times (waiting for RabbitMQ to start in Docker).
func NewRabbitMQConsumer(url string, h *handler.NotificationHandler) (*RabbitMQConsumer, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < 30; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("waiting for RabbitMQ... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		handler: h,
	}, nil
}

// Start begins consuming messages from the payment.completed queue.
// It blocks until the context is cancelled.
func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	// Declare durable queue — survives broker restarts
	q, err := c.channel.QueueDeclare(
		"payment.completed", // name
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Manual ACK: autoAck is false — we acknowledge only after successful processing
	msgs, err := c.channel.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // autoAck = false (manual ACK)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("[Consumer] Waiting for messages on queue:", q.Name)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Consumer] Context cancelled, stopping consumer")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[Consumer] Channel closed, stopping consumer")
				return nil
			}
			c.processMessage(msg)
		}
	}
}

// processMessage unmarshals and processes a single message.
func (c *RabbitMQConsumer) processMessage(msg amqp.Delivery) {
	var event domain.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] Failed to unmarshal message: %v", err)
		// Bad message format — reject without requeue (won't fix on retry)
		msg.Nack(false, false)
		return
	}

	if err := c.handler.Handle(event); err != nil {
		log.Printf("[Consumer] Failed to handle event %s: %v", event.EventID, err)
		// Processing error — requeue for retry
		msg.Nack(false, true)
		return
	}

	// Successfully processed — acknowledge the message
	msg.Ack(false)
}

// Close cleanly shuts down the RabbitMQ channel and connection.
func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("[Consumer] RabbitMQ connection closed")
}
