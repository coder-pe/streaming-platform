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

	// MinIO (S3-compatible storage)
	MinIOEnabled   bool
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	// Meilisearch
	MeilisearchURL    string
	MeilisearchAPIKey string
}

func Load() *Config {
	return &Config{
		// Server
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgres://streaming_platform_app:m4m4n1@localhost:5432/streaming_platform_db?sslmode=disable"),
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

		// MinIO
		MinIOEnabled:   getEnvAsBool("MINIO_ENABLED", false),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "videos"),
		MinIOUseSSL:    getEnvAsBool("MINIO_USE_SSL", false),

		// Meilisearch
		MeilisearchURL:    getEnv("MEILISEARCH_URL", "http://localhost:7700"),
		MeilisearchAPIKey: getEnv("MEILISEARCH_API_KEY", "local-dev-key-123"),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
