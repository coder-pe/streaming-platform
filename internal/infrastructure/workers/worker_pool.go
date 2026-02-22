// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/logger"
)

type Job struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type Worker interface {
	ProcessJob(ctx context.Context, job map[string]interface{}) error
}

type WorkerPool struct {
	workers     map[string]Worker
	jobQueue    chan Job
	workerCount int
	quit        chan bool
	wg          sync.WaitGroup
	logger      logger.Logger
	mu          sync.RWMutex
}

func NewWorkerPool(workerCount int, logger logger.Logger) *WorkerPool {
	return &WorkerPool{
		workers:     make(map[string]Worker),
		jobQueue:    make(chan Job, workerCount*2),
		workerCount: workerCount,
		quit:        make(chan bool),
		logger:      logger,
	}
}

func (wp *WorkerPool) RegisterWorker(jobType string, worker Worker) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.workers[jobType] = worker
	wp.logger.Info("Registered worker for job type: %s", jobType)
}

func (wp *WorkerPool) Start() {
	wp.logger.Info("Starting worker pool with %d workers", wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool...")
	close(wp.quit)
	wp.wg.Wait()
	wp.logger.Info("Worker pool stopped")
}

func (wp *WorkerPool) SubmitJob(job Job) error {
	select {
	case wp.jobQueue <- job:
		wp.logger.Info("Job submitted: %s", job.Type)
		return nil
	case <-time.After(5 * time.Second):
		wp.logger.Error("Timeout submitting job: %s", job.Type)
		return ErrJobSubmissionTimeout
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	wp.logger.Info("Worker %d started", id)

	for {
		select {
		case job := <-wp.jobQueue:
			wp.processJob(id, job)
		case <-wp.quit:
			wp.logger.Info("Worker %d stopping", id)
			return
		}
	}
}

func (wp *WorkerPool) processJob(workerID int, job Job) {
	start := time.Now()
	wp.logger.Info("Worker %d processing job: %s", workerID, job.Type)

	wp.mu.RLock()
	worker, exists := wp.workers[job.Type]
	wp.mu.RUnlock()

	if !exists {
		wp.logger.Error("No worker registered for job type: %s", job.Type)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := worker.ProcessJob(ctx, job.Data); err != nil {
		wp.logger.Error("Worker %d failed to process job %s: %v", workerID, job.Type, err)
		// Aquí podrías implementar retry logic o dead letter queue
		return
	}

	duration := time.Since(start)
	wp.logger.Info("Worker %d completed job %s in %v", workerID, job.Type, duration)
}

// ProcessJobFromQueue procesa un trabajo que viene de una cola externa (Redis u otra)
func (wp *WorkerPool) ProcessJobFromQueue(jobMessage interface{}) error {
	if typedJob, ok := jobMessage.(output.JobMessage); ok {
		return wp.SubmitJob(Job{
			Type: typedJob.Type,
			Data: typedJob.Data,
		})
	}

	// Convertir el mensaje de la cola en un Job del WorkerPool
	jobMap, ok := jobMessage.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid job message format")
	}

	jobType, ok := jobMap["type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid job type")
	}

	data, ok := jobMap["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing or invalid job data")
	}

	job := Job{
		Type: jobType,
		Data: data,
	}

	return wp.SubmitJob(job)
}

// Errores personalizados
var (
	ErrJobSubmissionTimeout = fmt.Errorf("job submission timeout")
)
