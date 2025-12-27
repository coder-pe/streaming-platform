// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/database"
	"streaming-platform/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DefaultQueueName = "transcoding_queue"
	MaxRetries       = 3
	RetryDelay       = 5 * time.Second
)

type queueRepository struct {
	client *database.RabbitMQClient
	logger logger.Logger
}

// NewQueueRepository crea una nueva instancia del repositorio de cola
func NewQueueRepository(client *database.RabbitMQClient, log logger.Logger) output.QueueRepository {
	return &queueRepository{
		client: client,
		logger: log,
	}
}

func (r *queueRepository) PublishJob(ctx context.Context, job output.JobMessage) error {
	// Serializar el mensaje
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Determinar el nombre de la cola basado en el tipo de trabajo
	queueName := r.getQueueNameByType(job.Type)

	// Publicar con reintentos
	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			// Esperar antes de reintentar
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(RetryDelay):
			}

			// Verificar si necesitamos reconectar
			if r.client.IsClosed() {
				r.logger.Info("Reconnecting to RabbitMQ...")
				if err := r.client.Reconnect(); err != nil {
					r.logger.Error("Failed to reconnect to RabbitMQ: %v", err)
					lastErr = err
					continue
				}
			}
		}

		// Intentar publicar
		err = r.client.PublishMessage(ctx, queueName, body)
		if err == nil {
			r.logger.Info("Job published successfully to queue %s: %s", queueName, job.Type)
			return nil
		}

		lastErr = err
		r.logger.Error("Failed to publish job (attempt %d/%d): %v", attempt+1, MaxRetries, err)
	}

	return fmt.Errorf("failed to publish job after %d attempts: %w", MaxRetries, lastErr)
}

func (r *queueRepository) ConsumeJobs(ctx context.Context, queueName string, handler func(output.JobMessage) error) error {
	if queueName == "" {
		queueName = DefaultQueueName
	}

	// Iniciar consumo de mensajes
	msgs, err := r.client.ConsumeMessages(queueName, "streaming-platform-consumer")
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	r.logger.Info("Started consuming messages from queue: %s", queueName)

	// Procesar mensajes
	go func() {
		for {
			select {
			case <-ctx.Done():
				r.logger.Info("Stopping message consumption")
				return

			case msg, ok := <-msgs:
				if !ok {
					r.logger.Info("Message channel closed")
					return
				}

				r.processMessage(msg, handler)
			}
		}
	}()

	return nil
}

func (r *queueRepository) processMessage(msg amqp.Delivery, handler func(output.JobMessage) error) {
	// Deserializar el mensaje
	var job output.JobMessage
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		r.logger.Error("Failed to unmarshal job message: %v", err)
		msg.Nack(false, false) // No requeue malformed messages
		return
	}

	r.logger.Info("Processing job: %s", job.Type)

	// Procesar el job con el handler
	err := handler(job)
	if err != nil {
		r.logger.Error("Failed to process job %s: %v", job.Type, err)

		// Requeue el mensaje si no ha sido procesado muchas veces
		retryCount := r.getRetryCount(msg)
		if retryCount < MaxRetries {
			r.logger.Info("Requeueing job %s (retry %d/%d)", job.Type, retryCount+1, MaxRetries)
			msg.Nack(false, true) // Requeue
		} else {
			r.logger.Error("Job %s exceeded max retries, discarding", job.Type)
			msg.Nack(false, false) // Don't requeue
		}
		return
	}

	// Acknowledge el mensaje
	if err := msg.Ack(false); err != nil {
		r.logger.Error("Failed to acknowledge message: %v", err)
	} else {
		r.logger.Info("Job %s processed successfully", job.Type)
	}
}

func (r *queueRepository) GetQueueInfo(ctx context.Context, queueName string) (int, int, error) {
	if queueName == "" {
		queueName = DefaultQueueName
	}

	messageCount, consumerCount, err := r.client.QueueInfo(queueName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get queue info: %w", err)
	}

	return messageCount, consumerCount, nil
}

func (r *queueRepository) PurgeQueue(ctx context.Context, queueName string) (int, error) {
	if queueName == "" {
		queueName = DefaultQueueName
	}

	count, err := r.client.PurgeQueue(queueName)
	if err != nil {
		return 0, fmt.Errorf("failed to purge queue: %w", err)
	}

	r.logger.Info("Purged %d messages from queue %s", count, queueName)
	return count, nil
}

func (r *queueRepository) Close() error {
	r.logger.Info("Closing queue repository")
	return r.client.Close()
}

func (r *queueRepository) HealthCheck() error {
	return r.client.HealthCheck()
}

// Helper functions

func (r *queueRepository) getQueueNameByType(jobType string) string {
	// Mapear tipos de trabajo a nombres de cola
	queueMap := map[string]string{
		"transcoding": "transcoding_queue",
		"thumbnail":   "thumbnail_queue",
		"analytics":   "analytics_queue",
		"notification": "notification_queue",
	}

	if queueName, ok := queueMap[jobType]; ok {
		return queueName
	}

	// Default queue
	return DefaultQueueName
}

func (r *queueRepository) getRetryCount(msg amqp.Delivery) int {
	// Obtener el conteo de reintentos desde los headers
	if msg.Headers == nil {
		return 0
	}

	if retryCount, ok := msg.Headers["x-retry-count"].(int); ok {
		return retryCount
	}

	return 0
}
