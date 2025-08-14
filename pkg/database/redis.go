// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedis creates a new Redis client connection
func NewRedis(redisURL string) (*redis.Client, error) {
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Set default timeouts
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	opt.PoolSize = 10
	opt.MinIdleConns = 5
	opt.MaxRetries = 3

	client := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// NewRedisWithConfig creates a new Redis client with custom configuration
func NewRedisWithConfig(config RedisConfig) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	// Set defaults if not provided
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 5 * time.Second
	}
	if opt.ReadTimeout == 0 {
		opt.ReadTimeout = 3 * time.Second
	}
	if opt.WriteTimeout == 0 {
		opt.WriteTimeout = 3 * time.Second
	}
	if opt.PoolSize == 0 {
		opt.PoolSize = 10
	}
	if opt.MinIdleConns == 0 {
		opt.MinIdleConns = 5
	}
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 3
	}

	client := redis.NewClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// NewRedisCluster creates a new Redis cluster client
func NewRedisCluster(addrs []string, password string) (*redis.ClusterClient, error) {
	opt := &redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	}

	client := redis.NewClusterClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis cluster: %w", err)
	}

	return client, nil
}

// RedisSentinelConfig holds Redis Sentinel configuration
type RedisSentinelConfig struct {
	MasterName    string
	SentinelAddrs []string
	Password      string
	DB            int
	PoolSize      int
	MinIdleConns  int
	MaxRetries    int
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}

// NewRedisSentinel creates a new Redis client with Sentinel support
func NewRedisSentinel(config RedisSentinelConfig) (*redis.Client, error) {
	opt := &redis.FailoverOptions{
		MasterName:    config.MasterName,
		SentinelAddrs: config.SentinelAddrs,
		Password:      config.Password,
		DB:            config.DB,
		PoolSize:      config.PoolSize,
		MinIdleConns:  config.MinIdleConns,
		MaxRetries:    config.MaxRetries,
		DialTimeout:   config.DialTimeout,
		ReadTimeout:   config.ReadTimeout,
		WriteTimeout:  config.WriteTimeout,
	}

	// Set defaults if not provided
	if opt.DialTimeout == 0 {
		opt.DialTimeout = 5 * time.Second
	}
	if opt.ReadTimeout == 0 {
		opt.ReadTimeout = 3 * time.Second
	}
	if opt.WriteTimeout == 0 {
		opt.WriteTimeout = 3 * time.Second
	}
	if opt.PoolSize == 0 {
		opt.PoolSize = 10
	}
	if opt.MinIdleConns == 0 {
		opt.MinIdleConns = 5
	}
	if opt.MaxRetries == 0 {
		opt.MaxRetries = 3
	}

	client := redis.NewFailoverClient(opt)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis with Sentinel: %w", err)
	}

	return client, nil
}

// RedisHealthCheck performs a health check on Redis connection
func RedisHealthCheck(client redis.Cmdable) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test basic operations
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test set/get operation
	testKey := "health_check_" + fmt.Sprintf("%d", time.Now().Unix())
	testValue := "ok"

	if err := client.Set(ctx, testKey, testValue, time.Minute).Err(); err != nil {
		return fmt.Errorf("set operation failed: %w", err)
	}

	result, err := client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("get operation failed: %w", err)
	}

	if result != testValue {
		return fmt.Errorf("get operation returned unexpected value: got %s, want %s", result, testValue)
	}

	// Clean up test key
	if err := client.Del(ctx, testKey).Err(); err != nil {
		// Log warning but don't fail health check
		fmt.Printf("Warning: failed to clean up health check key: %v\n", err)
	}

	return nil
}

// GetRedisStats returns Redis connection statistics
func GetRedisStats(client *redis.Client) map[string]interface{} {
	stats := client.PoolStats()

	return map[string]interface{}{
		"hits":        stats.Hits,
		"misses":      stats.Misses,
		"timeouts":    stats.Timeouts,
		"total_conns": stats.TotalConns,
		"idle_conns":  stats.IdleConns,
		"stale_conns": stats.StaleConns,
	}
}

// GetRedisInfo returns Redis server information
func GetRedisInfo(client redis.Cmdable) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse info string into map
	infoMap := make(map[string]string)
	lines := strings.Split(info, "\r\n")

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			infoMap[parts[0]] = parts[1]
		}
	}

	return infoMap, nil
}

// FlushRedisDB clears all data from the current Redis database
func FlushRedisDB(client redis.Cmdable) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush Redis database: %w", err)
	}

	return nil
}

// FlushAllRedis clears all data from all Redis databases
func FlushAllRedis(client redis.Cmdable) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.FlushAll(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush all Redis databases: %w", err)
	}

	return nil
}

// CloseRedis closes the Redis client connection gracefully
func CloseRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}

	return client.Close()
}

// RedisKeyExists checks if a key exists in Redis
func RedisKeyExists(client redis.Cmdable, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	count, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count > 0, nil
}

// RedisSetWithExpiry sets a key-value pair with expiration
func RedisSetWithExpiry(client redis.Cmdable, key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key with expiry: %w", err)
	}

	return nil
}

// RedisGetTTL gets the time-to-live of a key
func RedisGetTTL(client redis.Cmdable, key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// RedisDeletePattern deletes all keys matching a pattern
func RedisDeletePattern(client redis.Cmdable, pattern string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete in batches to avoid blocking Redis
	batchSize := 1000
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := client.Del(ctx, batch...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys batch: %w", err)
		}
	}

	return nil
}

// RedisPipeline helper for executing multiple commands atomically
func RedisPipeline(client redis.Cmdable, commands func(redis.Pipeliner) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipe := client.Pipeline()

	if err := commands(pipe); err != nil {
		return fmt.Errorf("failed to prepare pipeline commands: %w", err)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	return nil
}
