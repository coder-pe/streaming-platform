// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package input

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// StreamingService define el puerto de entrada para streaming de videos
type StreamingService interface {
	// CheckVideoAccess verifica si el usuario tiene acceso al video
	CheckVideoAccess(ctx context.Context, userID, videoID uuid.UUID) (bool, error)

	// GetHLSPlaylist obtiene el playlist HLS del video
	GetHLSPlaylist(ctx context.Context, videoID uuid.UUID) (*domain.HLSPlaylist, error)

	// GetVariantPlaylist obtiene el playlist de una calidad específica
	GetVariantPlaylist(ctx context.Context, videoID uuid.UUID, quality string) (string, error)

	// RecordStreamingSession registra una sesión de streaming
	RecordStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) error

	// UpdateStreamingStats actualiza las estadísticas de streaming
	UpdateStreamingStats(ctx context.Context, userID, videoID uuid.UUID, segment string) error

	// GetStreamingInfo obtiene información de streaming del video
	GetStreamingInfo(ctx context.Context, userID, videoID uuid.UUID) (*domain.StreamingInfo, error)

	// UpdateWatchProgress actualiza el progreso de visualización
	UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error
}
