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


// RabbitMQPublisher publishes payment events to RabbitMQ.
// It implements the EventPublisher port from the usecase layer.
type RabbitMQPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQPublisher connects to RabbitMQ and declares the exchange and queue.
// Retries connection up to 30 times for Docker startup ordering.
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

	// Declare durable exchange
	if err := ch.ExchangeDeclare(
		"payment.events", // name
		"direct",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare durable queue
	q, err := ch.QueueDeclare(
		"payment.completed", // name
		true,                // durable
		false,               // auto-delete
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	if err := ch.QueueBind(
		q.Name,              // queue name
		"payment.completed", // routing key
		"payment.events",    // exchange
		false,               // no-wait
		nil,                 // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("[Publisher] Connected to RabbitMQ, exchange and queue ready")

	return &RabbitMQPublisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// PublishPaymentCompleted publishes a payment completed event to RabbitMQ.
// Messages are persistent (DeliveryMode=2) so they survive broker restarts.
func (p *RabbitMQPublisher) PublishPaymentCompleted(ctx context.Context, event usecase.PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.channel.PublishWithContext(ctx,
		"payment.events",    // exchange
		"payment.completed", // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // message survives broker restart
			ContentType:  "application/json",
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("[Publisher] Published event %s for order %s", event.EventID, event.OrderID)
	return nil
}

// Close cleanly shuts down the RabbitMQ channel and connection.
func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	log.Println("[Publisher] RabbitMQ connection closed")
}
