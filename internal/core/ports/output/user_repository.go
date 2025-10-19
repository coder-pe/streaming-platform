// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// UserRepository define el puerto de salida para persistencia de usuarios
// Este es el contrato que los adapters de salida (PostgreSQL, MongoDB, etc.) deben implementar
type UserRepository interface {
	// Create crea un nuevo usuario
	Create(ctx context.Context, user *domain.User) error

	// GetByID obtiene un usuario por su ID
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail obtiene un usuario por su email
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update actualiza un usuario
	Update(ctx context.Context, user *domain.User) error

	// Delete elimina un usuario
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateLastLogin actualiza la última fecha de login
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// UpdatePassword actualiza la contraseña del usuario
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error

	// GetPublicProfile obtiene el perfil público
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)

	// UpdateProfile actualiza el perfil del usuario
	UpdateProfile(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// GetUserStats obtiene estadísticas del usuario
	GetUserStats(ctx context.Context, id uuid.UUID) (*domain.UserStats, error)

	// GetSettings obtiene la configuración del usuario
	GetSettings(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)

	// UpdateSettings actualiza la configuración del usuario
	UpdateSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error

	// List lista usuarios con paginación
	List(ctx context.Context, page, limit int) ([]*domain.User, int64, error)

	// Search busca usuarios
	Search(ctx context.Context, query string, page, limit int) ([]*domain.User, int64, error)

	// UpdateStatus actualiza el estado del usuario
	UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error

	// UpdateRole actualiza el rol del usuario
	UpdateRole(ctx context.Context, id uuid.UUID, role string) error
}
