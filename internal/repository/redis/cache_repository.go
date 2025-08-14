// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"streaming-platform/internal/domain/entities"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{
		client: client,
	}
}

// Basic operations
func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cache value for key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence %s: %w", key, err)
	}

	return count > 0, nil
}

func (r *CacheRepository) SetTTL(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL for cache key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for cache key %s: %w", key, err)
	}

	return ttl, nil
}

func (r *CacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	err = r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys by pattern %s: %w", pattern, err)
	}

	return nil
}

// String operations
func (r *CacheRepository) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set string cache key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) GetString(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get string cache key %s: %w", key, err)
	}

	return value, nil
}

// User cache operations - CORREGIDOS PARA COINCIDIR CON LA INTERFAZ
func (r *CacheRepository) CacheUser(ctx context.Context, user *entities.User) error {
	key := fmt.Sprintf("user:%s", user.ID.String())
	return r.Set(ctx, key, user, 15*time.Minute)
}

func (r *CacheRepository) GetCachedUser(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	key := fmt.Sprintf("user:%s", userID.String())
	var user entities.User
	err := r.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *CacheRepository) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("user:%s*", userID.String())
	return r.DeleteByPattern(ctx, pattern)
}

// Video cache operations - CORREGIDOS PARA COINCIDIR CON LA INTERFAZ
func (r *CacheRepository) CacheVideo(ctx context.Context, video *entities.Video) error {
	key := fmt.Sprintf("video:%s", video.ID.String())
	return r.Set(ctx, key, video, 30*time.Minute)
}

func (r *CacheRepository) GetCachedVideo(ctx context.Context, videoID uuid.UUID) (*entities.Video, error) {
	key := fmt.Sprintf("video:%s", videoID.String())
	var video entities.Video
	err := r.Get(ctx, key, &video)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *CacheRepository) InvalidateVideoCache(ctx context.Context, videoID uuid.UUID) error {
	pattern := fmt.Sprintf("video:%s*", videoID.String())
	return r.DeleteByPattern(ctx, pattern)
}

// Session operations - CORREGIDOS PARA COINCIDIR CON LA INTERFAZ
func (r *CacheRepository) SetSession(ctx context.Context, sessionID string, userID uuid.UUID, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.SetString(ctx, key, userID.String(), expiration)
}

func (r *CacheRepository) GetSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	userIDStr, err := r.GetString(ctx, key)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(userIDStr)
}

func (r *CacheRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(ctx, key)
}

// Health check
func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
