package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/logger"

	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultQueueName = "transcoding_queue"
	consumerWait     = 5 * time.Second
)

type queueRepository struct {
	client *goredis.Client
	logger logger.Logger
}

// NewQueueRepository crea una implementación de QueueRepository usando Redis Lists.
func NewQueueRepository(client *goredis.Client, log logger.Logger) output.QueueRepository {
	return &queueRepository{
		client: client,
		logger: log,
	}
}

func (r *queueRepository) PublishJob(ctx context.Context, job output.JobMessage) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	queueName := r.getQueueNameByType(job.Type)
	if err := r.client.LPush(ctx, queueName, body).Err(); err != nil {
		return fmt.Errorf("failed to publish job to redis queue %s: %w", queueName, err)
	}

	r.logger.Info("Job published to Redis queue %s: %s", queueName, job.Type)
	return nil
}

func (r *queueRepository) ConsumeJobs(ctx context.Context, queueName string, handler func(output.JobMessage) error) error {
	if queueName == "" {
		queueName = defaultQueueName
	}

	r.logger.Info("Started consuming jobs from Redis queue: %s", queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				r.logger.Info("Stopping Redis queue consumer for %s", queueName)
				return
			default:
			}

			// BRPop bloquea hasta recibir mensaje o timeout; usamos timeout corto para respetar ctx.
			result, err := r.client.BRPop(ctx, consumerWait, queueName).Result()
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				if err == goredis.Nil {
					continue
				}
				r.logger.Error("Redis queue consume error on %s: %v", queueName, err)
				time.Sleep(time.Second)
				continue
			}

			// BRPop devuelve [queueName, payload]
			if len(result) != 2 {
				r.logger.Error("Unexpected BRPop payload format on %s", queueName)
				continue
			}

			var job output.JobMessage
			if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
				r.logger.Error("Failed to unmarshal Redis queue job: %v", err)
				continue
			}

			if err := handler(job); err != nil {
				r.logger.Error("Failed to process Redis queue job %s: %v", job.Type, err)

				// Requeue simple (best effort). Se puede mejorar con DLQ y retry counter.
				requeueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = r.PublishJob(requeueCtx, job)
				cancel()
				time.Sleep(time.Second)
				continue
			}

			r.logger.Info("Redis queue job processed successfully: %s", job.Type)
		}
	}()

	return nil
}

func (r *queueRepository) GetQueueInfo(ctx context.Context, queueName string) (int, int, error) {
	if queueName == "" {
		queueName = defaultQueueName
	}

	length, err := r.client.LLen(ctx, queueName).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get redis queue length: %w", err)
	}

	// Redis lists no exponen consumer count; devolvemos 0.
	return int(length), 0, nil
}

func (r *queueRepository) PurgeQueue(ctx context.Context, queueName string) (int, error) {
	if queueName == "" {
		queueName = defaultQueueName
	}

	length, err := r.client.LLen(ctx, queueName).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get redis queue length before purge: %w", err)
	}

	if err := r.client.Del(ctx, queueName).Err(); err != nil {
		return 0, fmt.Errorf("failed to purge redis queue %s: %w", queueName, err)
	}

	r.logger.Info("Purged %d jobs from Redis queue %s", length, queueName)
	return int(length), nil
}

func (r *queueRepository) Close() error {
	return nil // Redis client lifecycle is managed outside this repository.
}

func (r *queueRepository) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return r.client.Ping(ctx).Err()
}

func (r *queueRepository) getQueueNameByType(jobType string) string {
	queueMap := map[string]string{
		"transcoding":  "transcoding_queue",
		"thumbnail":    "thumbnail_queue",
		"analytics":    "analytics_queue",
		"notification": "notification_queue",
	}

	if queueName, ok := queueMap[jobType]; ok {
		return queueName
	}
	return defaultQueueName
}
