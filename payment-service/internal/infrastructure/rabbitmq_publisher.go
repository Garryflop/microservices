package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"payment-service/internal/usecase"
)

// RabbitMQ event publisher
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
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

	// durable exchange
	if err := ch.ExchangeDeclare(
		"payment.events", "direct", true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// DLX exchange (must match consumer declaration)
	if err := ch.ExchangeDeclare(
		"payment.events.dlx", "direct", true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	// DLQ queue
	if _, err := ch.QueueDeclare(
		"payment.completed.dlq", true, false, false, false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	if err := ch.QueueBind(
		"payment.completed.dlq", "payment.completed.dlq", "payment.events.dlx", false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind DLQ: %w", err)
	}

	// durable queue with DLX args (identical to consumer declaration)
	q, err := ch.QueueDeclare(
		"payment.completed", true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":     "payment.events.dlx",
			"x-dead-letter-routing-key":  "payment.completed.dlq",
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// bind queue
	if err := ch.QueueBind(
		q.Name, "payment.completed", "payment.events", false, nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("[Publisher] RabbitMQ ready")

	return &RabbitMQPublisher{conn: conn, channel: ch}, nil
}

// PublishPaymentCompleted sends persistent event
func (p *RabbitMQPublisher) PublishPaymentCompleted(ctx context.Context, event usecase.PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.channel.PublishWithContext(ctx,
		"payment.events", "payment.completed", false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("[Publisher] Published event %s for order %s", event.EventID, event.OrderID)
	return nil
}

func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	log.Println("[Publisher] RabbitMQ connection closed")
}
