// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// FavoriteRepository define el puerto de salida para gestión de favoritos
type FavoriteRepository interface {
	// AddFavorite agrega un video a favoritos
	AddFavorite(ctx context.Context, userID, videoID uuid.UUID) error

	// RemoveFavorite elimina un video de favoritos
	RemoveFavorite(ctx context.Context, userID, videoID uuid.UUID) error

	// IsFavorite verifica si un video está en favoritos
	IsFavorite(ctx context.Context, userID, videoID uuid.UUID) (bool, error)

	// GetUserFavorites obtiene todos los favoritos de un usuario
	GetUserFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.Favorite, int64, error)

	// GetFavoriteCount obtiene el número de favoritos de un usuario
	GetFavoriteCount(ctx context.Context, userID uuid.UUID) (int64, error)
}
