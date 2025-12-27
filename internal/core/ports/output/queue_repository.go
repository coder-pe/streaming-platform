// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"
)

// JobMessage represents a job message in the queue
type JobMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// QueueRepository define el puerto de salida para gestión de colas
type QueueRepository interface {
	// PublishJob publica un trabajo en la cola
	PublishJob(ctx context.Context, job JobMessage) error

	// ConsumeJobs inicia el consumo de trabajos desde la cola
	ConsumeJobs(ctx context.Context, queueName string, handler func(JobMessage) error) error

	// GetQueueInfo obtiene información sobre una cola
	GetQueueInfo(ctx context.Context, queueName string) (messageCount int, consumerCount int, err error)

	// PurgeQueue elimina todos los mensajes de una cola
	PurgeQueue(ctx context.Context, queueName string) (int, error)

	// Close cierra la conexión a la cola
	Close() error

	// HealthCheck verifica la salud de la conexión
	HealthCheck() error
}
