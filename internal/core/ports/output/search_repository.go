// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"

	"streaming-platform/internal/core/domain"
)

// SearchRepository define el puerto de salida para búsqueda avanzada
type SearchRepository interface {
	// IndexVideo indexa un video para búsqueda
	IndexVideo(ctx context.Context, video *domain.Video) error

	// UpdateVideoIndex actualiza el índice de un video
	UpdateVideoIndex(ctx context.Context, video *domain.Video) error

	// DeleteVideoIndex elimina un video del índice
	DeleteVideoIndex(ctx context.Context, videoID string) error

	// SearchVideos busca videos usando búsqueda full-text
	SearchVideos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*domain.Video, int64, error)

	// SuggestVideos sugiere videos basados en el input del usuario
	SuggestVideos(ctx context.Context, partial string, limit int) ([]string, error)

	// HealthCheck verifica la salud de la conexión
	HealthCheck() error

	// InitializeIndex inicializa el índice de búsqueda
	InitializeIndex(ctx context.Context) error
}
