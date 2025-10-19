// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package input

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// AuthService define el puerto de entrada para autenticación
// Este es el contrato que los adapters de entrada (HTTP, CLI, etc.) usarán
type AuthService interface {
	// Login autentica un usuario y retorna tokens
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user *domain.UserProfile, err error)

	// Register registra un nuevo usuario
	Register(ctx context.Context, email, password, firstName, lastName, role string) (accessToken, refreshToken string, user *domain.UserProfile, err error)

	// RefreshToken genera nuevos tokens usando un refresh token
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, user *domain.UserProfile, err error)

	// Logout invalida la sesión del usuario
	Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error

	// ForgotPassword inicia el proceso de recuperación de contraseña
	ForgotPassword(ctx context.Context, email string) error

	// ResetPassword resetea la contraseña usando un token
	ResetPassword(ctx context.Context, token, newPassword string) error

	// ChangePassword cambia la contraseña del usuario
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error

	// ValidateToken valida un token JWT
	ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error)
}
