// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/database"
	"streaming-platform/pkg/logger"
)

const (
	VideoIndex = "videos"
)

type searchRepository struct {
	client *database.ElasticsearchClient
	logger logger.Logger
}

// NewSearchRepository crea una nueva instancia del repositorio de búsqueda
func NewSearchRepository(client *database.ElasticsearchClient, log logger.Logger) output.SearchRepository {
	return &searchRepository{
		client: client,
		logger: log,
	}
}

func (r *searchRepository) InitializeIndex(ctx context.Context) error {
	// Verificar si el índice ya existe
	exists, err := r.client.IndexExists(ctx, VideoIndex)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		r.logger.Info("Index %s already exists", VideoIndex)
		return nil
	}

	// Definir mapping para el índice de videos
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "keyword",
				},
				"title": map[string]interface{}{
					"type": "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"description": map[string]interface{}{
					"type": "text",
					"analyzer": "standard",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"instructor_id": map[string]interface{}{
					"type": "keyword",
				},
				"instructor_name": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"status": map[string]interface{}{
					"type": "keyword",
				},
				"is_public": map[string]interface{}{
					"type": "boolean",
				},
				"duration": map[string]interface{}{
					"type": "integer",
				},
				"view_count": map[string]interface{}{
					"type": "integer",
				},
				"rating": map[string]interface{}{
					"type": "float",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	err = r.client.CreateIndex(ctx, VideoIndex, mapping)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	r.logger.Info("Index %s created successfully", VideoIndex)
	return nil
}

func (r *searchRepository) IndexVideo(ctx context.Context, video *domain.Video) error {
	// Convertir video a documento indexable
	doc := r.videoToDocument(video)

	err := r.client.IndexDocument(ctx, VideoIndex, video.ID.String(), doc)
	if err != nil {
		return fmt.Errorf("failed to index video: %w", err)
	}

	r.logger.Info("Video indexed successfully: %s", video.ID)
	return nil
}

func (r *searchRepository) UpdateVideoIndex(ctx context.Context, video *domain.Video) error {
	doc := r.videoToDocument(video)

	err := r.client.UpdateDocument(ctx, VideoIndex, video.ID.String(), doc)
	if err != nil {
		return fmt.Errorf("failed to update video index: %w", err)
	}

	r.logger.Info("Video index updated successfully: %s", video.ID)
	return nil
}

func (r *searchRepository) DeleteVideoIndex(ctx context.Context, videoID string) error {
	err := r.client.DeleteDocument(ctx, VideoIndex, videoID)
	if err != nil {
		return fmt.Errorf("failed to delete video index: %w", err)
	}

	r.logger.Info("Video index deleted successfully: %s", videoID)
	return nil
}

func (r *searchRepository) SearchVideos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*domain.Video, int64, error) {
	// Calcular offset
	offset := (page - 1) * limit

	// Construir query de Elasticsearch
	esQuery := r.buildSearchQuery(query, filters)

	// Agregar paginación
	esQuery["from"] = offset
	esQuery["size"] = limit

	// Ejecutar búsqueda
	response, err := r.client.SearchDocuments(ctx, VideoIndex, esQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search videos: %w", err)
	}

	// Parsear respuesta
	var searchResult struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source json.RawMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(response, &searchResult); err != nil {
		return nil, 0, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convertir documentos a videos
	videos := make([]*domain.Video, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		video, err := r.documentToVideo(hit.Source)
		if err != nil {
			r.logger.Error("Failed to parse video document: %v", err)
			continue
		}
		videos = append(videos, video)
	}

	return videos, searchResult.Hits.Total.Value, nil
}

func (r *searchRepository) SuggestVideos(ctx context.Context, partial string, limit int) ([]string, error) {
	// Construir query de sugerencias
	esQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"video-suggestion": map[string]interface{}{
				"prefix": partial,
				"completion": map[string]interface{}{
					"field": "title.keyword",
					"size":  limit,
				},
			},
		},
	}

	response, err := r.client.SearchDocuments(ctx, VideoIndex, esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	// Parsear respuesta de sugerencias
	var suggestResult struct {
		Suggest map[string][]struct {
			Options []struct {
				Text string `json:"text"`
			} `json:"options"`
		} `json:"suggest"`
	}

	if err := json.Unmarshal(response, &suggestResult); err != nil {
		return nil, fmt.Errorf("failed to parse suggest response: %w", err)
	}

	suggestions := make([]string, 0)
	for _, suggest := range suggestResult.Suggest {
		for _, s := range suggest {
			for _, option := range s.Options {
				suggestions = append(suggestions, option.Text)
			}
		}
	}

	return suggestions, nil
}

func (r *searchRepository) HealthCheck() error {
	return r.client.HealthCheck()
}

// Helper functions

func (r *searchRepository) videoToDocument(video *domain.Video) map[string]interface{} {
	return map[string]interface{}{
		"id":              video.ID.String(),
		"title":           video.Title,
		"description":     video.Description,
		"category":        video.Category,
		"tags":            video.Tags,
		"instructor_id":   video.InstructorID.String(),
		"instructor_name": "", // Se puede poblar con información del instructor
		"status":          video.Status,
		"is_public":       video.IsPublic,
		"duration":        video.Duration,
		"view_count":      video.ViewCount,
		"rating":          video.Rating,
		"created_at":      video.CreatedAt,
		"updated_at":      video.UpdatedAt,
	}
}

func (r *searchRepository) documentToVideo(doc json.RawMessage) (*domain.Video, error) {
	var video domain.Video
	if err := json.Unmarshal(doc, &video); err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *searchRepository) buildSearchQuery(query string, filters map[string]interface{}) map[string]interface{} {
	must := []interface{}{}

	// Agregar búsqueda de texto si hay query
	if query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title^3", "description", "tags"},
				"type":   "best_fields",
			},
		})
	}

	// Agregar filtro para videos públicos
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"is_public": true,
		},
	})

	// Agregar filtro para videos ready
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"status": domain.VideoStatusReady,
		},
	})

	// Aplicar filtros adicionales
	for key, value := range filters {
		if key == "category" && value != "" {
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"category": value,
				},
			})
		}
		if key == "tags" && value != nil {
			must = append(must, map[string]interface{}{
				"terms": map[string]interface{}{
					"tags": value,
				},
			})
		}
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
		"sort": []interface{}{
			map[string]interface{}{
				"_score": map[string]interface{}{
					"order": "desc",
				},
			},
			map[string]interface{}{
				"view_count": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}
}
