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

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	handler *handler.NotificationHandler
}

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

	return &RabbitMQConsumer{conn: conn, channel: ch, handler: h}, nil
}

func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	// DLX exchange
	if err := c.channel.ExchangeDeclare(
		"payment.events.dlx", "direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	// DLQ queue
	if _, err := c.channel.QueueDeclare(
		"payment.completed.dlq", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// bind DLQ
	if err := c.channel.QueueBind(
		"payment.completed.dlq", "payment.completed.dlq", "payment.events.dlx", false, nil,
	); err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// main queue with DLX routing
	q, err := c.channel.QueueDeclare(
		"payment.completed", true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    "payment.events.dlx",
			"x-dead-letter-routing-key": "payment.completed.dlq",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// manual ACK
	msgs, err := c.channel.Consume(
		q.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("[Consumer] Waiting for messages on queue:", q.Name)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Consumer] Context cancelled, stopping")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[Consumer] Channel closed, stopping")
				return nil
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *RabbitMQConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var event domain.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // bad payload → DLQ
		return
	}

	// handler manages retries with exponential backoff internally
	if err := c.handler.Handle(ctx, event); err != nil {
		log.Printf("[Consumer] Handler failed for event %s: %v", event.EventID, err)
		msg.Nack(false, false) // max retries exhausted → DLQ
		return
	}

	msg.Ack(false)
}

func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("[Consumer] RabbitMQ connection closed")
}
