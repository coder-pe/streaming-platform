// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port     string
	LogLevel string

	// Database
	DatabaseURL string
	RedisURL    string

	// JWT
	JWTSecret string

	// File Storage
	StoragePath string
	CDNBaseURL  string

	// External Services
	FFmpegPath string

	// Worker Pool
	WorkerPoolSize int

	// Email
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
}

func Load() *Config {
	return &Config{
		// Server
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/streaming_platform?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),

		// File Storage
		StoragePath: getEnv("STORAGE_PATH", "./storage"),
		CDNBaseURL:  getEnv("CDN_BASE_URL", "http://localhost:8080/static"),

		// External Services
		FFmpegPath: getEnv("FFMPEG_PATH", "ffmpeg"),

		// Worker Pool
		WorkerPoolSize: getEnvAsInt("WORKER_POOL_SIZE", 4),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
