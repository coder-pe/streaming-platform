// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// WatchHistoryRepository define el puerto de salida para historial de visualización
type WatchHistoryRepository interface {
	// SaveWatchHistory guarda o actualiza el progreso de visualización
	SaveWatchHistory(ctx context.Context, history *domain.WatchHistory) error

	// GetWatchHistory obtiene el progreso de un video específico
	GetWatchHistory(ctx context.Context, userID, videoID uuid.UUID) (*domain.WatchHistory, error)

	// GetUserWatchHistory obtiene el historial completo de un usuario
	GetUserWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.WatchHistory, int64, error)

	// DeleteWatchHistory elimina un registro del historial
	DeleteWatchHistory(ctx context.Context, userID, videoID uuid.UUID) error

	// GetContinueWatching obtiene videos para continuar viendo (con progreso > 0 y < 90%)
	GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.WatchHistory, error)
}
