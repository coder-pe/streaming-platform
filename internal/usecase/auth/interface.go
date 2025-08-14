// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package auth

import (
	"context"
	"streaming-platform/internal/domain/entities"

	"github.com/google/uuid"
)

// AuthUsecase defines the interface for authentication use case
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (string, string, *entities.UserProfile, error)
	Register(ctx context.Context, email, password string, firstName, lastName, role string) (string, string, *entities.UserProfile, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, *entities.UserProfile, error)
	Logout(ctx context.Context, userID uuid.UUID, token string) error
	ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
}
