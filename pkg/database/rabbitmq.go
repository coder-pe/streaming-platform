// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package database

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient wraps the RabbitMQ connection and channel
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQ creates a new RabbitMQ client connection
func NewRabbitMQ(rabbitMQURL string) (*RabbitMQClient, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Set QoS to limit prefetch
	err = ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	client := &RabbitMQClient{
		conn:    conn,
		channel: ch,
		url:     rabbitMQURL,
	}

	return client, nil
}

// DeclareQueue declares a queue and returns the queue details
func (r *RabbitMQClient) DeclareQueue(queueName string) (amqp.Queue, error) {
	q, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}

	return q, nil
}

// PublishMessage publishes a message to a queue
func (r *RabbitMQClient) PublishMessage(ctx context.Context, queueName string, body []byte) error {
	// Ensure queue exists
	_, err := r.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	err = r.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// ConsumeMessages starts consuming messages from a queue
func (r *RabbitMQClient) ConsumeMessages(queueName string, consumerTag string) (<-chan amqp.Delivery, error) {
	// Ensure queue exists
	_, err := r.DeclareQueue(queueName)
	if err != nil {
		return nil, err
	}

	msgs, err := r.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}

	return msgs, nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQClient) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}

// IsClosed checks if the connection is closed
func (r *RabbitMQClient) IsClosed() bool {
	return r.conn.IsClosed()
}

// GetChannel returns the RabbitMQ channel
func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	return r.channel
}

// GetConnection returns the RabbitMQ connection
func (r *RabbitMQClient) GetConnection() *amqp.Connection {
	return r.conn
}

// Reconnect attempts to reconnect to RabbitMQ
func (r *RabbitMQClient) Reconnect() error {
	if err := r.Close(); err != nil {
		// Log error but continue with reconnection
		fmt.Printf("Warning: error during close before reconnect: %v\n", err)
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Set QoS
	err = ch.Qos(10, 0, false)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	r.conn = conn
	r.channel = ch

	return nil
}

// HealthCheck performs a health check on the RabbitMQ connection
func (r *RabbitMQClient) HealthCheck() error {
	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}

	if r.channel == nil {
		return fmt.Errorf("channel is nil")
	}

	// Try to declare a test queue
	testQueueName := fmt.Sprintf("health_check_%d", time.Now().Unix())
	q, err := r.channel.QueueDeclare(
		testQueueName,
		false, // not durable
		true,  // delete when unused
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Delete the test queue
	_, err = r.channel.QueueDelete(q.Name, false, false, false)
	if err != nil {
		// Log warning but don't fail health check
		fmt.Printf("Warning: failed to delete health check queue: %v\n", err)
	}

	return nil
}

// PurgeQueue removes all messages from a queue
func (r *RabbitMQClient) PurgeQueue(queueName string) (int, error) {
	count, err := r.channel.QueuePurge(queueName, false)
	if err != nil {
		return 0, fmt.Errorf("failed to purge queue: %w", err)
	}
	return count, nil
}

// DeleteQueue deletes a queue
func (r *RabbitMQClient) DeleteQueue(queueName string) (int, error) {
	count, err := r.channel.QueueDelete(queueName, false, false, false)
	if err != nil {
		return 0, fmt.Errorf("failed to delete queue: %w", err)
	}
	return count, nil
}

// QueueInfo returns information about a queue
func (r *RabbitMQClient) QueueInfo(queueName string) (int, int, error) {
	q, err := r.channel.QueueInspect(queueName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to inspect queue: %w", err)
	}
	return q.Messages, q.Consumers, nil
}
