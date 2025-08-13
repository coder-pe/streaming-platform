// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache key not found: %s", key)
		}
		return fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
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
		return false, fmt.Errorf("failed to check if key exists %s: %w", key, err)
	}

	return count > 0, nil
}

// Batch operations
func (r *CacheRepository) SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error {
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
		return fmt.Errorf("failed to set multiple cache keys: %w", err)
	}

	return nil
}

func (r *CacheRepository) GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get multiple cache keys: %w", err)
	}

	result := make(map[string]interface{})
	for key, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			if err != redis.Nil {
				return nil, fmt.Errorf("failed to get cache key %s: %w", key, err)
			}
			continue // Skip missing keys
		}

		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cached value for key %s: %w", key, err)
		}

		result[key] = value
	}

	return result, nil
}

func (r *CacheRepository) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete multiple cache keys: %w", err)
	}

	return nil
}

// Pattern operations
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

func (r *CacheRepository) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys by pattern %s: %w", pattern, err)
	}

	return keys, nil
}

// TTL operations
func (r *CacheRepository) SetTTL(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL for key %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}

	return ttl, nil
}

// Counter operations
func (r *CacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}

	return value, nil
}

func (r *CacheRepository) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}

	return result, nil
}

func (r *CacheRepository) Decrement(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}

	return value, nil
}

func (r *CacheRepository) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := r.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s by %d: %w", key, value, err)
	}

	return result, nil
}

// List operations
func (r *CacheRepository) ListPush(ctx context.Context, key string, values ...interface{}) error {
	// Convert values to strings
	stringValues := make([]interface{}, len(values))
	for i, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for list push: %w", err)
		}
		stringValues[i] = string(data)
	}

	err := r.client.LPush(ctx, key, stringValues...).Err()
	if err != nil {
		return fmt.Errorf("failed to push to list %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) ListPop(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("list is empty: %s", key)
		}
		return nil, fmt.Errorf("failed to pop from list %s: %w", key, err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal popped value: %w", err)
	}

	return value, nil
}

func (r *CacheRepository) ListLength(ctx context.Context, key string) (int64, error) {
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length %s: %w", key, err)
	}

	return length, nil
}

// Set operations
func (r *CacheRepository) SetAdd(ctx context.Context, key string, members ...interface{}) error {
	// Convert members to strings
	stringMembers := make([]interface{}, len(members))
	for i, member := range members {
		data, err := json.Marshal(member)
		if err != nil {
			return fmt.Errorf("failed to marshal member for set add: %w", err)
		}
		stringMembers[i] = string(data)
	}

	err := r.client.SAdd(ctx, key, stringMembers...).Err()
	if err != nil {
		return fmt.Errorf("failed to add to set %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) SetRemove(ctx context.Context, key string, members ...interface{}) error {
	// Convert members to strings
	stringMembers := make([]interface{}, len(members))
	for i, member := range members {
		data, err := json.Marshal(member)
		if err != nil {
			return fmt.Errorf("failed to marshal member for set remove: %w", err)
		}
		stringMembers[i] = string(data)
	}

	err := r.client.SRem(ctx, key, stringMembers...).Err()
	if err != nil {
		return fmt.Errorf("failed to remove from set %s: %w", key, err)
	}

	return nil
}

func (r *CacheRepository) SetMembers(ctx context.Context, key string) ([]interface{}, error) {
	stringMembers, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members %s: %w", key, err)
	}

	members := make([]interface{}, len(stringMembers))
	for i, stringMember := range stringMembers {
		var member interface{}
		if err := json.Unmarshal([]byte(stringMember), &member); err != nil {
			return nil, fmt.Errorf("failed to unmarshal set member: %w", err)
		}
		members[i] = member
	}

	return members, nil
}

func (r *CacheRepository) SetExists(ctx context.Context, key string, member interface{}) (bool, error) {
	data, err := json.Marshal(member)
	if err != nil {
		return false, fmt.Errorf("failed to marshal member for set exists: %w", err)
	}

	exists, err := r.client.SIsMember(ctx, key, string(data)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check set membership %s: %w", key, err)
	}

	return exists, nil
}

// Hash operations
func (r *CacheRepository) HashSet(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for hash set: %w", err)
	}

	err = r.client.HSet(ctx, key, field, string(data)).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field %s.%s: %w", key, field, err)
	}

	return nil
}

func (r *CacheRepository) HashGet(ctx context.Context, key string, field string) (interface{}, error) {
	data, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("hash field not found: %s.%s", key, field)
		}
		return nil, fmt.Errorf("failed to get hash field %s.%s: %w", key, field, err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hash value: %w", err)
	}

	return value, nil
}

func (r *CacheRepository) HashGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	stringMap, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields %s: %w", key, err)
	}

	result := make(map[string]interface{})
	for field, data := range stringMap {
		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal hash value for field %s: %w", field, err)
		}
		result[field] = value
	}

	return result, nil
}

func (r *CacheRepository) HashDelete(ctx context.Context, key string, fields ...string) error {
	if len(fields) == 0 {
		return nil
	}

	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete hash fields %s: %w", key, err)
	}

	return nil
}

// Utility methods for common cache patterns

// SetWithTimeout sets a key with a default timeout
func (r *CacheRepository) SetWithTimeout(ctx context.Context, key string, value interface{}) error {
	return r.Set(ctx, key, value, 1*time.Hour) // Default 1 hour
}

// GetString gets a string value from cache
func (r *CacheRepository) GetString(ctx context.Context, key string) (string, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("cache key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	return data, nil
}

// SetString sets a string value in cache
func (r *CacheRepository) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set string cache key %s: %w", key, err)
	}

	return nil
}

// CacheVideo caches video data with appropriate TTL
func (r *CacheRepository) CacheVideo(ctx context.Context, videoID string, video interface{}) error {
	key := fmt.Sprintf("video:%s", videoID)
	return r.Set(ctx, key, video, 30*time.Minute)
}

// GetCachedVideo retrieves cached video data
func (r *CacheRepository) GetCachedVideo(ctx context.Context, videoID string, dest interface{}) error {
	key := fmt.Sprintf("video:%s", videoID)
	return r.Get(ctx, key, dest)
}

// CacheUser caches user profile data
func (r *CacheRepository) CacheUser(ctx context.Context, userID string, user interface{}) error {
	key := fmt.Sprintf("user:%s", userID)
	return r.Set(ctx, key, user, 15*time.Minute)
}

// GetCachedUser retrieves cached user data
func (r *CacheRepository) GetCachedUser(ctx context.Context, userID string, dest interface{}) error {
	key := fmt.Sprintf("user:%s", userID)
	return r.Get(ctx, key, dest)
}

// InvalidateUserCache removes user cache
func (r *CacheRepository) InvalidateUserCache(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("user:%s*", userID)
	return r.DeleteByPattern(ctx, pattern)
}

// InvalidateVideoCache removes video cache
func (r *CacheRepository) InvalidateVideoCache(ctx context.Context, videoID string) error {
	pattern := fmt.Sprintf("video:%s*", videoID)
	return r.DeleteByPattern(ctx, pattern)
}

// SetSession sets a user session
func (r *CacheRepository) SetSession(ctx context.Context, sessionID string, userID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.SetString(ctx, key, userID, 24*time.Hour)
}

// GetSession gets a user session
func (r *CacheRepository) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.GetString(ctx, key)
}

// InvalidateSession removes a user session
func (r *CacheRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(ctx, key)
}

// Health check
func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
