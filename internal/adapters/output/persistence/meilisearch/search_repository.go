// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package meilisearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/database"
	"streaming-platform/pkg/logger"

	meili "github.com/meilisearch/meilisearch-go"
)

const (
	VideoIndex = "videos"
)

// videoDocument es la estructura del documento indexado en Meilisearch
type videoDocument struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
	InstructorID   string   `json:"instructor_id"`
	InstructorName string   `json:"instructor_name"`
	Status         string   `json:"status"`
	IsPublic       bool     `json:"is_public"`
	Duration       int      `json:"duration"`
	ViewCount      int64    `json:"view_count"`
	Rating         float64  `json:"rating"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
}

type searchRepository struct {
	client *database.MeilisearchClient
	logger logger.Logger
}

// NewSearchRepository crea una nueva instancia del repositorio de búsqueda
func NewSearchRepository(client *database.MeilisearchClient, log logger.Logger) output.SearchRepository {
	return &searchRepository{
		client: client,
		logger: log,
	}
}

func (r *searchRepository) InitializeIndex(ctx context.Context) error {
	// Crear el índice con "id" como clave primaria
	taskInfo, err := r.client.CreateIndex(VideoIndex, "id")
	if err != nil {
		// Si el índice ya existe, Meilisearch devuelve un error que podemos ignorar
		r.logger.Info("Index %s may already exist: %v", VideoIndex, err)
	} else {
		// Esperar a que la tarea de creación se complete
		task, err := r.client.WaitForTask(taskInfo.TaskUID)
		if err != nil {
			return fmt.Errorf("failed waiting for index creation: %w", err)
		}
		if task.Status == meili.TaskStatusFailed {
			// index_already_exists no es un error real
			if task.Error.Code != "index_already_exists" {
				return fmt.Errorf("failed to create index: %s", task.Error.Message)
			}
		}
	}

	// Configurar los atributos del índice
	index := r.client.Index(VideoIndex)

	// Atributos buscables (en orden de importancia)
	searchableAttrs := []string{"title", "description", "tags", "category", "instructor_name"}
	if _, err := index.UpdateSearchableAttributes(&searchableAttrs); err != nil {
		return fmt.Errorf("failed to update searchable attributes: %w", err)
	}

	// Atributos filtrables
	filterableAttrs := []interface{}{"category", "tags", "is_public", "status", "instructor_id"}
	if _, err := index.UpdateFilterableAttributes(&filterableAttrs); err != nil {
		return fmt.Errorf("failed to update filterable attributes: %w", err)
	}

	// Atributos ordenables
	sortableAttrs := []string{"view_count", "rating", "created_at", "updated_at"}
	if _, err := index.UpdateSortableAttributes(&sortableAttrs); err != nil {
		return fmt.Errorf("failed to update sortable attributes: %w", err)
	}

	r.logger.Info("Index %s initialized successfully", VideoIndex)
	return nil
}

func (r *searchRepository) IndexVideo(ctx context.Context, video *domain.Video) error {
	doc := r.videoToDocument(video)
	index := r.client.Index(VideoIndex)

	_, err := index.AddDocumentsWithContext(ctx, []videoDocument{doc}, nil)
	if err != nil {
		return fmt.Errorf("failed to index video: %w", err)
	}

	r.logger.Info("Video indexed successfully: %s", video.ID)
	return nil
}

func (r *searchRepository) UpdateVideoIndex(ctx context.Context, video *domain.Video) error {
	doc := r.videoToDocument(video)
	index := r.client.Index(VideoIndex)

	_, err := index.UpdateDocumentsWithContext(ctx, []videoDocument{doc}, nil)
	if err != nil {
		return fmt.Errorf("failed to update video index: %w", err)
	}

	r.logger.Info("Video index updated successfully: %s", video.ID)
	return nil
}

func (r *searchRepository) DeleteVideoIndex(ctx context.Context, videoID string) error {
	index := r.client.Index(VideoIndex)

	_, err := index.DeleteDocumentWithContext(ctx, videoID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete video index: %w", err)
	}

	r.logger.Info("Video index deleted successfully: %s", videoID)
	return nil
}

func (r *searchRepository) SearchVideos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*domain.Video, int64, error) {
	offset := int64((page - 1) * limit)
	index := r.client.Index(VideoIndex)

	// Construir filtros de Meilisearch
	filterStr := r.buildFilter(filters)

	searchReq := &meili.SearchRequest{
		Offset: offset,
		Limit:  int64(limit),
		Filter: filterStr,
		Sort:   []string{"view_count:desc"},
	}

	resp, err := index.SearchWithContext(ctx, query, searchReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search videos: %w", err)
	}

	// Parsear hits a videos
	videos := make([]*domain.Video, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var doc videoDocument
		if err := hit.DecodeInto(&doc); err != nil {
			r.logger.Error("Failed to parse video document: %v", err)
			continue
		}
		video := r.documentToVideo(&doc)
		videos = append(videos, video)
	}

	return videos, resp.EstimatedTotalHits, nil
}

func (r *searchRepository) SuggestVideos(ctx context.Context, partial string, limit int) ([]string, error) {
	index := r.client.Index(VideoIndex)

	// Meilisearch tiene búsqueda typo-tolerant nativa, usamos search con límite bajo
	filterStr := "is_public = true AND status = \"ready\""
	resp, err := index.SearchWithContext(ctx, partial, &meili.SearchRequest{
		Limit:  int64(limit),
		Filter: filterStr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	suggestions := make([]string, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		var doc videoDocument
		if err := hit.DecodeInto(&doc); err != nil {
			continue
		}
		if doc.Title != "" {
			suggestions = append(suggestions, doc.Title)
		}
	}

	return suggestions, nil
}

func (r *searchRepository) HealthCheck() error {
	return r.client.HealthCheck()
}

// Helper functions

func (r *searchRepository) videoToDocument(video *domain.Video) videoDocument {
	return videoDocument{
		ID:             video.ID.String(),
		Title:          video.Title,
		Description:    video.Description,
		Category:       video.Category,
		Tags:           video.Tags,
		InstructorID:   video.InstructorID.String(),
		InstructorName: "",
		Status:         string(video.Status),
		IsPublic:       video.IsPublic,
		Duration:       video.Duration,
		ViewCount:      video.ViewCount,
		Rating:         video.Rating,
		CreatedAt:      video.CreatedAt.Unix(),
		UpdatedAt:      video.UpdatedAt.Unix(),
	}
}

func (r *searchRepository) documentToVideo(doc *videoDocument) *domain.Video {
	video := &domain.Video{
		Title:       doc.Title,
		Description: doc.Description,
		Category:    doc.Category,
		Tags:        doc.Tags,
		Status:      domain.VideoStatus(doc.Status),
		IsPublic:    doc.IsPublic,
		Duration:    doc.Duration,
		ViewCount:   doc.ViewCount,
		Rating:      doc.Rating,
		CreatedAt:   time.Unix(doc.CreatedAt, 0),
		UpdatedAt:   time.Unix(doc.UpdatedAt, 0),
	}

	// Parsear UUIDs (ignorar errores, quedarán como uuid.Nil)
	_ = video.ID.UnmarshalText([]byte(doc.ID))
	_ = video.InstructorID.UnmarshalText([]byte(doc.InstructorID))

	return video
}

func (r *searchRepository) buildFilter(filters map[string]interface{}) string {
	parts := []string{
		"is_public = true",
		fmt.Sprintf("status = \"%s\"", domain.VideoStatusReady),
	}

	for key, value := range filters {
		if key == "category" {
			if cat, ok := value.(string); ok && cat != "" {
				parts = append(parts, fmt.Sprintf("category = \"%s\"", cat))
			}
		}
		if key == "tags" {
			if tags, ok := value.([]string); ok && len(tags) > 0 {
				tagFilters := make([]string, 0, len(tags))
				for _, tag := range tags {
					tagFilters = append(tagFilters, fmt.Sprintf("tags = \"%s\"", tag))
				}
				parts = append(parts, "("+strings.Join(tagFilters, " OR ")+")")
			}
		}
	}

	return strings.Join(parts, " AND ")
}
