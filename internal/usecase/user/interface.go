// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package user

import (
	"context"
	"streaming-platform/internal/domain/entities"

	"github.com/google/uuid"
)

// UserUsecase defines the interface for user use case
type UserUsecase interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entities.UserProfile, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUserStats(ctx context.Context, userID uuid.UUID) (*entities.UserStats, error)
	GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.Video, int64, error)
	AddToFavorites(ctx context.Context, userID, videoID uuid.UUID) error
	RemoveFromFavorites(ctx context.Context, userID, videoID uuid.UUID) error
	GetWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.WatchHistory, int64, error)
	SearchUsers(ctx context.Context, query string, page, limit int) ([]*entities.UserProfile, int64, error)
	GetPublicProfile(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error)
	UpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error
	GetSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	DeactivateAccount(ctx context.Context, userID uuid.UUID, password, reason string) error
}
