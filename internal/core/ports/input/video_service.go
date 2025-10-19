// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package input

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// VideoService define el puerto de entrada para gestión de videos
type VideoService interface {
	// CreateVideo crea un nuevo video
	CreateVideo(ctx context.Context, video *domain.Video) error

	// GetVideoByID obtiene un video por su ID
	GetVideoByID(ctx context.Context, videoID uuid.UUID) (*domain.Video, error)

	// UpdateVideo actualiza un video
	UpdateVideo(ctx context.Context, video *domain.Video) error

	// DeleteVideo elimina un video
	DeleteVideo(ctx context.Context, videoID uuid.UUID) error

	// SearchPublicVideos busca videos públicos
	SearchPublicVideos(ctx context.Context, req domain.VideoSearchRequest) (*domain.VideoSearchResponse, error)

	// IncrementViewCount incrementa el contador de visualizaciones
	IncrementViewCount(ctx context.Context, videoID uuid.UUID) error

	// QueueTranscodingJob encola un trabajo de transcodificación
	QueueTranscodingJob(ctx context.Context, job map[string]interface{}) error
}
