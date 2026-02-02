// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package database

import (
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

// MeilisearchClient wraps the Meilisearch client
type MeilisearchClient struct {
	client meilisearch.ServiceManager
}

// NewMeilisearch creates a new Meilisearch client connection
func NewMeilisearch(host, apiKey string) (*MeilisearchClient, error) {
	var opts []meilisearch.Option
	if apiKey != "" {
		opts = append(opts, meilisearch.WithAPIKey(apiKey))
	}

	client := meilisearch.New(host, opts...)

	// Test the connection
	health, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Meilisearch: %w", err)
	}

	if health.Status != "available" {
		return nil, fmt.Errorf("meilisearch is not available, status: %s", health.Status)
	}

	return &MeilisearchClient{
		client: client,
	}, nil
}

// GetClient returns the underlying Meilisearch service manager
func (ms *MeilisearchClient) GetClient() meilisearch.ServiceManager {
	return ms.client
}

// Index returns an IndexManager for the given index UID
func (ms *MeilisearchClient) Index(uid string) meilisearch.IndexManager {
	return ms.client.Index(uid)
}

// CreateIndex creates a new index with the given UID and primary key
func (ms *MeilisearchClient) CreateIndex(uid, primaryKey string) (*meilisearch.TaskInfo, error) {
	taskInfo, err := ms.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        uid,
		PrimaryKey: primaryKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	return taskInfo, nil
}

// WaitForTask waits for a Meilisearch task to complete
func (ms *MeilisearchClient) WaitForTask(taskUID int64) (*meilisearch.Task, error) {
	task, err := ms.client.WaitForTask(taskUID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for task: %w", err)
	}
	return task, nil
}

// HealthCheck performs a health check on the Meilisearch connection
func (ms *MeilisearchClient) HealthCheck() error {
	health, err := ms.client.Health()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if health.Status != "available" {
		return fmt.Errorf("meilisearch status: %s", health.Status)
	}

	return nil
}
