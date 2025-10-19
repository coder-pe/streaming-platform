// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// VideoRepository define el puerto de salida para persistencia de videos
type VideoRepository interface {
	// Create crea un nuevo video
	Create(ctx context.Context, video *domain.Video) error

	// GetByID obtiene un video por su ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Video, error)

	// Update actualiza un video
	Update(ctx context.Context, video *domain.Video) error

	// Delete elimina un video
	Delete(ctx context.Context, id uuid.UUID) error

	// List lista videos con filtros y paginación
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*domain.Video, int64, error)

	// Search busca videos
	Search(ctx context.Context, req domain.VideoSearchRequest) (*domain.VideoSearchResponse, error)

	// GetByCategory obtiene videos por categoría
	GetByCategory(ctx context.Context, category string, page, limit int) ([]*domain.Video, int64, error)

	// GetByInstructor obtiene videos de un instructor
	GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]*domain.Video, int64, error)

	// GetPublicVideos obtiene videos públicos
	GetPublicVideos(ctx context.Context, page, limit int) ([]*domain.Video, int64, error)

	// GetFeaturedVideos obtiene videos destacados
	GetFeaturedVideos(ctx context.Context, limit int) ([]*domain.Video, error)

	// GetPopularVideos obtiene videos populares
	GetPopularVideos(ctx context.Context, limit int) ([]*domain.Video, error)

	// IncrementViewCount incrementa el contador de visualizaciones
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// UpdateRating actualiza el rating del video
	UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error

	// UpdateStatus actualiza el estado del video
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.VideoStatus) error

	// GetVideosByStatus obtiene videos por estado
	GetVideosByStatus(ctx context.Context, status domain.VideoStatus) ([]*domain.Video, error)

	// GetVideoStats obtiene estadísticas del video
	GetVideoStats(ctx context.Context, id uuid.UUID) (*domain.VideoStats, error)

	// CreateVideoFile crea un archivo de video
	CreateVideoFile(ctx context.Context, videoFile *domain.VideoFile) error

	// GetVideoFiles obtiene los archivos de un video
	GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*domain.VideoFile, error)

	// UpdateVideoFile actualiza un archivo de video
	UpdateVideoFile(ctx context.Context, videoFile *domain.VideoFile) error

	// DeleteVideoFile elimina un archivo de video
	DeleteVideoFile(ctx context.Context, id uuid.UUID) error
}
