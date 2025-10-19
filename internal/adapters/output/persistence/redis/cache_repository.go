// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/output"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type cacheRepository struct {
	client *redis.Client
}

// NewCacheRepository crea una nueva instancia del repositorio de caché
func NewCacheRepository(client *redis.Client) output.CacheRepository {
	return &cacheRepository{
		client: client,
	}
}

// Basic operations
func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
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

func (r *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
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

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}

	return nil
}

func (r *cacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence %s: %w", key, err)
	}

	return count > 0, nil
}

func (r *cacheRepository) SetTTL(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL for cache key %s: %w", key, err)
	}

	return nil
}

func (r *cacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for cache key %s: %w", key, err)
	}

	return ttl, nil
}

func (r *cacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
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
func (r *cacheRepository) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set string cache key %s: %w", key, err)
	}

	return nil
}

func (r *cacheRepository) GetString(ctx context.Context, key string) (string, error) {
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
func (r *cacheRepository) CacheUser(ctx context.Context, user *domain.User) error {
	key := fmt.Sprintf("user:%s", user.ID.String())
	return r.Set(ctx, key, user, 15*time.Minute)
}

func (r *cacheRepository) GetCachedUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	key := fmt.Sprintf("user:%s", userID.String())
	var user domain.User
	err := r.Get(ctx, key, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *cacheRepository) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	pattern := fmt.Sprintf("user:%s*", userID.String())
	return r.DeleteByPattern(ctx, pattern)
}

// Video cache operations - CORREGIDOS PARA COINCIDIR CON LA INTERFAZ
func (r *cacheRepository) CacheVideo(ctx context.Context, video *domain.Video) error {
	key := fmt.Sprintf("video:%s", video.ID.String())
	return r.Set(ctx, key, video, 30*time.Minute)
}

func (r *cacheRepository) GetCachedVideo(ctx context.Context, videoID uuid.UUID) (*domain.Video, error) {
	key := fmt.Sprintf("video:%s", videoID.String())
	var video domain.Video
	err := r.Get(ctx, key, &video)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (r *cacheRepository) InvalidateVideoCache(ctx context.Context, videoID uuid.UUID) error {
	pattern := fmt.Sprintf("video:%s*", videoID.String())
	return r.DeleteByPattern(ctx, pattern)
}

// Session operations - CORREGIDOS PARA COINCIDIR CON LA INTERFAZ
func (r *cacheRepository) SetSession(ctx context.Context, sessionID string, userID uuid.UUID, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.SetString(ctx, key, userID.String(), expiration)
}

func (r *cacheRepository) GetSession(ctx context.Context, sessionID string) (uuid.UUID, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	userIDStr, err := r.GetString(ctx, key)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(userIDStr)
}

func (r *cacheRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(ctx, key)
}

// Health check
func (r *cacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Counter operations - AGREGAR ESTOS MÉTODOS
func (r *cacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

func (r *cacheRepository) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return result, nil
}

func (r *cacheRepository) Decrement(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return result, nil
}

func (r *cacheRepository) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// Batch operations - AGREGAR ESTOS MÉTODOS
func (r *cacheRepository) SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
	pipe := r.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	results := make(map[string]interface{})
	for key, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == nil {
			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				results[key] = value
			}
		}
	}

	return results, nil
}

func (r *cacheRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete multiple keys: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys by pattern %s: %w", pattern, err)
	}
	return keys, nil
}

func (r *cacheRepository) ListPush(ctx context.Context, key string, values ...interface{}) error {
	for _, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for list push: %w", err)
		}
		if err := r.client.LPush(ctx, key, data).Err(); err != nil {
			return fmt.Errorf("failed to push to list %s: %w", key, err)
		}
	}
	return nil
}

func (r *cacheRepository) ListPop(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.LPop(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("list is empty: %s", key)
		}
		return nil, fmt.Errorf("failed to pop from list %s: %w", key, err)
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal list value: %w", err)
	}

	return value, nil
}

func (r *cacheRepository) ListLength(ctx context.Context, key string) (int64, error) {
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length for %s: %w", key, err)
	}
	return length, nil
}

// Set operations
func (r *cacheRepository) SetAdd(ctx context.Context, key string, member string) error {
	if err := r.client.SAdd(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to add to set %s: %w", key, err)
	}
	return nil
}

func (r *cacheRepository) SetRemove(ctx context.Context, key string, member string) error {
	if err := r.client.SRem(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to remove from set %s: %w", key, err)
	}
	return nil
}

func (r *cacheRepository) SetMembers(ctx context.Context, key string) ([]interface{}, error) {
	data, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members for %s: %w", key, err)
	}

	var members []interface{}
	for _, item := range data {
		var value interface{}
		if err := json.Unmarshal([]byte(item), &value); err == nil {
			members = append(members, value)
		}
	}

	return members, nil
}

func (r *cacheRepository) SetExists(ctx context.Context, key string, member interface{}) (bool, error) {
	data, err := json.Marshal(member)
	if err != nil {
		return false, fmt.Errorf("failed to marshal set member: %w", err)
	}

	exists, err := r.client.SIsMember(ctx, key, data).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership for %s: %w", key, err)
	}

	return exists, nil
}

// Hash operations - AGREGAR ESTOS MÉTODOS
func (r *cacheRepository) HashSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal hash value: %w", err)
	}

	if err := r.client.HSet(ctx, key, field, data).Err(); err != nil {
		return fmt.Errorf("failed to set hash field %s.%s: %w", key, field, err)
	}

	return nil
}

func (r *cacheRepository) HashGet(ctx context.Context, key string, field string) (interface{}, error) {
	data, err := r.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("hash field not found: %s.%s", key, field)
		}
		return nil, fmt.Errorf("failed to get hash field %s.%s: %w", key, field, err)
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hash value: %w", err)
	}

	return value, nil
}

func (r *cacheRepository) HashGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields for %s: %w", key, err)
	}

	result := make(map[string]interface{})
	for field, value := range data {
		var parsed interface{}
		if err := json.Unmarshal([]byte(value), &parsed); err == nil {
			result[field] = parsed
		}
	}

	return result, nil
}

func (r *cacheRepository) HashDelete(ctx context.Context, key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}

	if err := r.client.HDel(ctx, key, fields...).Err(); err != nil {
		return fmt.Errorf("failed to delete hash fields from %s: %w", key, err)
	}

	return nil
}
