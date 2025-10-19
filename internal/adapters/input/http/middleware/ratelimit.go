// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter interface for different rate limiting implementations
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Reset(ctx context.Context, key string) error
	GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error)
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

// Allow checks if request is allowed based on rate limit
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	pipe := r.client.Pipeline()

	// Use sliding window log approach
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current requests in window
	pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// Set expiration
	pipe.Expire(ctx, key, window)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Get count result
	count := results[1].(*redis.IntCmd).Val()

	return count < int64(limit), nil
}

// Reset removes all entries for a key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// GetRemaining returns remaining requests in current window
func (r *RedisRateLimiter) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Count current requests in window
	count, err := r.client.ZCount(ctx, key, strconv.FormatInt(windowStart, 10), "+inf").Result()
	if err != nil {
		return 0, err
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// MemoryRateLimiter implements in-memory rate limiting (for development)
type MemoryRateLimiter struct {
	requests map[string][]time.Time
}

// NewMemoryRateLimiter creates a new in-memory rate limiter
func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		requests: make(map[string][]time.Time),
	}
}

// Allow checks if request is allowed (simplified implementation)
func (m *MemoryRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	// Clean old entries
	if times, exists := m.requests[key]; exists {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		m.requests[key] = validTimes
	}

	// Check if limit exceeded
	if len(m.requests[key]) >= limit {
		return false, nil
	}

	// Add current request
	m.requests[key] = append(m.requests[key], now)

	return true, nil
}

// Reset removes all entries for a key
func (m *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	delete(m.requests, key)
	return nil
}

// GetRemaining returns remaining requests
func (m *MemoryRateLimiter) GetRemaining(ctx context.Context, key string, limit int, window time.Duration) (int, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	count := 0
	if times, exists := m.requests[key]; exists {
		for _, t := range times {
			if t.After(windowStart) {
				count++
			}
		}
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Limit      int                        // Requests per window
	Window     time.Duration              // Time window
	SkipPaths  []string                   // Paths to skip rate limiting
	KeyFunc    func(*http.Request) string // Function to generate rate limit key
	SkipFunc   func(*http.Request) bool   // Function to skip rate limiting
	OnLimitHit func(*http.Request)        // Callback when limit is hit
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:  100,
		Window: time.Minute,
		KeyFunc: func(r *http.Request) string {
			return GetClientIP(r)
		},
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
	}
}

// RateLimit middleware with Redis backend
func RateLimit(client *redis.Client) func(http.Handler) http.Handler {
	limiter := NewRedisRateLimiter(client)
	config := DefaultRateLimitConfig()
	return RateLimitWithConfig(limiter, config)
}

// RateLimitWithConfig creates rate limit middleware with custom config
func RateLimitWithConfig(limiter RateLimiter, config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if configured to skip
			if config.SkipFunc != nil && config.SkipFunc(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip certain paths
			for _, skipPath := range config.SkipPaths {
				if r.URL.Path == skipPath {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Generate rate limit key
			key := "rate_limit:" + config.KeyFunc(r)

			// Check if request is allowed
			allowed, err := limiter.Allow(r.Context(), key, config.Limit, config.Window)
			if err != nil {
				// Log error but don't block request
				http.Error(w, "Rate limiting error", http.StatusInternalServerError)
				return
			}

			// Get remaining requests
			remaining, _ := limiter.GetRemaining(r.Context(), key, config.Limit, config.Window)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Window", config.Window.String())

			if !allowed {
				// Call limit hit callback
				if config.OnLimitHit != nil {
					config.OnLimitHit(r)
				}

				// Set retry after header
				w.Header().Set("Retry-After", strconv.Itoa(int(config.Window.Seconds())))

				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthenticatedRateLimit applies different limits for authenticated users
func AuthenticatedRateLimit(client *redis.Client) func(http.Handler) http.Handler {
	limiter := NewRedisRateLimiter(client)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var key string
			var limit int
			var window time.Duration

			// Check if user is authenticated
			userID := r.Context().Value("user_id")
			if userID != nil {
				// Higher limits for authenticated users
				key = "rate_limit:user:" + userID.(string)
				limit = 1000 // 1000 requests per hour
				window = time.Hour
			} else {
				// Lower limits for anonymous users
				key = "rate_limit:ip:" + GetClientIP(r)
				limit = 100 // 100 requests per hour
				window = time.Hour
			}

			// Check if request is allowed
			allowed, err := limiter.Allow(r.Context(), key, limit, window)
			if err != nil {
				http.Error(w, "Rate limiting error", http.StatusInternalServerError)
				return
			}

			// Get remaining requests
			remaining, _ := limiter.GetRemaining(r.Context(), key, limit, window)

			// Set headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Window", window.String())

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// APIRateLimit specialized rate limiting for API endpoints
func APIRateLimit(client *redis.Client) func(http.Handler) http.Handler {
	limiter := NewRedisRateLimiter(client)

	config := RateLimitConfig{
		Limit:  300, // 300 requests per 5 minutes
		Window: 5 * time.Minute,
		KeyFunc: func(r *http.Request) string {
			// Use API key if available, otherwise IP
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" {
				return "api:" + apiKey
			}
			return "ip:" + GetClientIP(r)
		},
		SkipPaths: []string{
			"/api/health",
			"/api/status",
		},
	}

	return RateLimitWithConfig(limiter, config)
}

// UploadRateLimit specialized rate limiting for file uploads
func UploadRateLimit(client *redis.Client) func(http.Handler) http.Handler {
	limiter := NewRedisRateLimiter(client)

	config := RateLimitConfig{
		Limit:  5, // 5 uploads per hour
		Window: time.Hour,
		KeyFunc: func(r *http.Request) string {
			userID := r.Context().Value("user_id")
			if userID != nil {
				return "upload:user:" + userID.(string)
			}
			return "upload:ip:" + GetClientIP(r)
		},
	}

	return RateLimitWithConfig(limiter, config)
}

// LoginRateLimit specialized rate limiting for login attempts
func LoginRateLimit(client *redis.Client) func(http.Handler) http.Handler {
	limiter := NewRedisRateLimiter(client)

	config := RateLimitConfig{
		Limit:  5, // 5 login attempts per 15 minutes
		Window: 15 * time.Minute,
		KeyFunc: func(r *http.Request) string {
			return "login:ip:" + GetClientIP(r)
		},
		OnLimitHit: func(r *http.Request) {
			// Log suspicious login attempts
			fmt.Printf("Rate limit hit for login from IP: %s\n", GetClientIP(r))
		},
	}

	return RateLimitWithConfig(limiter, config)
}

// GetClientIP extracts the real client IP from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in case of multiple proxies
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// DynamicRateLimit applies different limits based on endpoint
func DynamicRateLimit(client *redis.Client, limits map[string]RateLimitConfig) func(http.Handler) http.Handler {
	limiter := NewRedisRateLimiter(client)
	defaultConfig := DefaultRateLimitConfig()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Find matching config for the endpoint
			config := defaultConfig
			for pattern, cfg := range limits {
				if strings.HasPrefix(r.URL.Path, pattern) {
					config = cfg
					break
				}
			}

			// Apply rate limiting
			middleware := RateLimitWithConfig(limiter, config)
			middleware(next).ServeHTTP(w, r)
		})
	}
}
