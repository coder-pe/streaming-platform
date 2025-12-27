// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// ElasticsearchClient wraps the Elasticsearch client
type ElasticsearchClient struct {
	client *elasticsearch7.Client
}

// NewElasticsearch creates a new Elasticsearch client connection
func NewElasticsearch(urls []string) (*ElasticsearchClient, error) {
	cfg := elasticsearch7.Config{
		Addresses: urls,
	}

	client, err := elasticsearch7.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test the connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch returned error: %s", res.String())
	}

	return &ElasticsearchClient{
		client: client,
	}, nil
}

// IndexDocument indexes a document in Elasticsearch
func (es *ElasticsearchClient) IndexDocument(ctx context.Context, index string, docID string, document interface{}) error {
	data, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

// UpdateDocument updates a document in Elasticsearch
func (es *ElasticsearchClient) UpdateDocument(ctx context.Context, index string, docID string, document interface{}) error {
	doc := map[string]interface{}{
		"doc": document,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating document: %s", res.String())
	}

	return nil
}

// DeleteDocument deletes a document from Elasticsearch
func (es *ElasticsearchClient) DeleteDocument(ctx context.Context, index string, docID string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting document: %s", res.String())
	}

	return nil
}

// SearchDocuments searches for documents in Elasticsearch
func (es *ElasticsearchClient) SearchDocuments(ctx context.Context, index string, query map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  &buf,
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching documents: %s", res.String())
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// BulkIndex performs a bulk indexing operation
func (es *ElasticsearchClient) BulkIndex(ctx context.Context, index string, documents []interface{}) error {
	var buf bytes.Buffer

	for _, doc := range documents {
		// Action line
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}
		actionJSON, err := json.Marshal(action)
		if err != nil {
			return fmt.Errorf("failed to marshal action: %w", err)
		}
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		// Document line
		docJSON, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to execute bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error in bulk index: %s", res.String())
	}

	return nil
}

// CreateIndex creates an index with mappings
func (es *ElasticsearchClient) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	data, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(data),
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		// Check if error is "index already exists"
		if !strings.Contains(res.String(), "resource_already_exists_exception") {
			return fmt.Errorf("error creating index: %s", res.String())
		}
	}

	return nil
}

// DeleteIndex deletes an index
func (es *ElasticsearchClient) DeleteIndex(ctx context.Context, index string) error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting index: %s", res.String())
	}

	return nil
}

// IndexExists checks if an index exists
func (es *ElasticsearchClient) IndexExists(ctx context.Context, index string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// HealthCheck performs a health check on the Elasticsearch connection
func (es *ElasticsearchClient) HealthCheck() error {
	res, err := es.client.Cluster.Health()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch health check error: %s", res.String())
	}

	var health map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return fmt.Errorf("failed to parse health response: %w", err)
	}

	status, ok := health["status"].(string)
	if !ok {
		return fmt.Errorf("invalid health status response")
	}

	if status == "red" {
		return fmt.Errorf("elasticsearch cluster status is red")
	}

	return nil
}

// GetClient returns the underlying Elasticsearch client
func (es *ElasticsearchClient) GetClient() *elasticsearch7.Client {
	return es.client
}

// RefreshIndex refreshes an index
func (es *ElasticsearchClient) RefreshIndex(ctx context.Context, index string) error {
	req := esapi.IndicesRefreshRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error refreshing index: %s", res.String())
	}

	return nil
}
