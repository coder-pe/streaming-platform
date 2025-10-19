// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"
	"time"

	"streaming-platform/internal/core/domain"

	"github.com/google/uuid"
)

// CacheRepository define el puerto de salida para cache (Redis, Memcached, etc.)
type CacheRepository interface {
	// Get obtiene un valor del cache y lo deserializa en dest
	Get(ctx context.Context, key string, dest interface{}) error

	// Set establece un valor en el cache
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete elimina un valor del cache
	Delete(ctx context.Context, key string) error

	// GetString obtiene un string del cache
	GetString(ctx context.Context, key string) (string, error)

	// SetString establece un string en el cache
	SetString(ctx context.Context, key string, value string, expiration time.Duration) error

	// CacheUser cachea un usuario
	CacheUser(ctx context.Context, user *domain.User) error

	// GetCachedUser obtiene un usuario del cache
	GetCachedUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)

	// InvalidateUserCache invalida el cache de un usuario
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error

	// SetSession establece una sesión
	SetSession(ctx context.Context, sessionID string, userID uuid.UUID, expiration time.Duration) error

	// GetSession obtiene una sesión
	GetSession(ctx context.Context, sessionID string) (uuid.UUID, error)

	// InvalidateSession invalida una sesión
	InvalidateSession(ctx context.Context, sessionID string) error

	// CacheVideo cachea un video
	CacheVideo(ctx context.Context, video *domain.Video) error

	// GetCachedVideo obtiene un video del cache
	GetCachedVideo(ctx context.Context, videoID uuid.UUID) (*domain.Video, error)

	// InvalidateVideoCache invalida el cache de un video
	InvalidateVideoCache(ctx context.Context, videoID uuid.UUID) error

	// SetAdd añade un elemento a un conjunto (set)
	SetAdd(ctx context.Context, key string, member string) error

	// SetRemove elimina un elemento de un conjunto (set)
	SetRemove(ctx context.Context, key string, member string) error

	// SetMembers obtiene todos los miembros de un conjunto (set)
	SetMembers(ctx context.Context, key string) ([]interface{}, error)

	// Increment incrementa un contador
	Increment(ctx context.Context, key string) (int64, error)

	// SetTTL establece el tiempo de vida de una clave
	SetTTL(ctx context.Context, key string, expiration time.Duration) error
}
