// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package input

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// UserService define el puerto de entrada para gestión de usuarios
type UserService interface {
	// GetUserByID obtiene un usuario por su ID
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error)

	// GetPublicProfile obtiene el perfil público de un usuario
	GetPublicProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error)

	// UpdateUser actualiza los datos del usuario
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*domain.UserProfile, error)

	// GetUserStats obtiene las estadísticas del usuario
	GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserStats, error)

	// GetWatchHistory obtiene el historial de visualización
	GetWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.WatchHistory, int64, error)

	// GetFavorites obtiene los videos favoritos del usuario
	GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Video, int64, error)

	// AddToFavorites agrega un video a favoritos
	AddToFavorites(ctx context.Context, userID, videoID uuid.UUID) error

	// RemoveFromFavorites remueve un video de favoritos
	RemoveFromFavorites(ctx context.Context, userID, videoID uuid.UUID) error

	// GetSettings obtiene la configuración del usuario
	GetSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)

	// UpdateSettings actualiza la configuración del usuario
	UpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error

	// DeactivateAccount desactiva la cuenta del usuario
	DeactivateAccount(ctx context.Context, userID uuid.UUID, password, reason string) error
}
