// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package interfaces

import (
	"context"
	"time"

	"streaming-platform/internal/domain/entities"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Authentication related
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error

	// Profile operations
	GetPublicProfile(ctx context.Context, id uuid.UUID) (*entities.UserProfile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

	// User stats
	GetUserStats(ctx context.Context, id uuid.UUID) (*entities.UserStats, error)

	// Settings
	GetSettings(ctx context.Context, id uuid.UUID) (map[string]interface{}, error)
	UpdateSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error

	// Admin operations
	List(ctx context.Context, page, limit int) ([]*entities.User, int64, error)
	Search(ctx context.Context, query string, page, limit int) ([]*entities.User, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error
	UpdateRole(ctx context.Context, id uuid.UUID, role string) error
}

// VideoRepository defines the interface for video data operations
type VideoRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, video *entities.Video) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Video, error)
	Update(ctx context.Context, video *entities.Video) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and searching
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*entities.Video, int64, error)
	Search(ctx context.Context, searchReq entities.VideoSearchRequest) (*entities.VideoSearchResponse, error)
	GetByCategory(ctx context.Context, category string, page, limit int) ([]*entities.Video, int64, error)
	GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]*entities.Video, int64, error)

	// Public videos
	GetPublicVideos(ctx context.Context, page, limit int) ([]*entities.Video, int64, error)
	GetFeaturedVideos(ctx context.Context, limit int) ([]*entities.Video, error)
	GetPopularVideos(ctx context.Context, limit int) ([]*entities.Video, error)

	// Video files
	CreateVideoFile(ctx context.Context, videoFile *entities.VideoFile) error
	GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entities.VideoFile, error)
	UpdateVideoFile(ctx context.Context, videoFile *entities.VideoFile) error
	DeleteVideoFile(ctx context.Context, id uuid.UUID) error

	// Statistics
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error
	GetVideoStats(ctx context.Context, id uuid.UUID) (*entities.VideoStats, error)

	// Status management
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus) error
	GetVideosByStatus(ctx context.Context, status entities.VideoStatus) ([]*entities.Video, error)
}

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	// Basic operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Batch operations
	SetMultiple(ctx context.Context, items map[string]interface{}, expiration time.Duration) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
	DeleteMultiple(ctx context.Context, keys []string) error

	// Pattern operations
	DeleteByPattern(ctx context.Context, pattern string) error
	GetKeys(ctx context.Context, pattern string) ([]string, error)

	// TTL operations
	SetTTL(ctx context.Context, key string, expiration time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)

	// Counter operations
	Increment(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	DecrementBy(ctx context.Context, key string, value int64) (int64, error)

	// List operations
	ListPush(ctx context.Context, key string, values ...interface{}) error
	ListPop(ctx context.Context, key string) (interface{}, error)
	ListLength(ctx context.Context, key string) (int64, error)

	// Set operations
	SetAdd(ctx context.Context, key string, members ...interface{}) error
	SetRemove(ctx context.Context, key string, members ...interface{}) error
	SetMembers(ctx context.Context, key string) ([]interface{}, error)
	SetExists(ctx context.Context, key string, member interface{}) (bool, error)

	// Hash operations
	HashSet(ctx context.Context, key string, field string, value interface{}) error
	HashGet(ctx context.Context, key string, field string) (interface{}, error)
	HashGetAll(ctx context.Context, key string) (map[string]interface{}, error)
	HashDelete(ctx context.Context, key string, fields ...string) error
}

// WatchHistoryRepository defines the interface for watch history operations
type WatchHistoryRepository interface {
	Create(ctx context.Context, history *entities.WatchHistory) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.WatchHistory, int64, error)
	GetByVideo(ctx context.Context, videoID uuid.UUID, page, limit int) ([]*entities.WatchHistory, int64, error)
	UpdateProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error
	GetProgress(ctx context.Context, userID, videoID uuid.UUID) (*entities.WatchHistory, error)
	Delete(ctx context.Context, userID, videoID uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}

// FavoriteRepository defines the interface for favorite operations
type FavoriteRepository interface {
	Add(ctx context.Context, userID, videoID uuid.UUID) error
	Remove(ctx context.Context, userID, videoID uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.Video, int64, error)
	IsInFavorites(ctx context.Context, userID, videoID uuid.UUID) (bool, error)
	GetFavoriteCount(ctx context.Context, videoID uuid.UUID) (int64, error)
}

// CommentRepository defines the interface for comment operations
type CommentRepository interface {
	Create(ctx context.Context, comment *entities.Comment) error
	GetByVideo(ctx context.Context, videoID uuid.UUID, page, limit int) ([]*entities.Comment, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Comment, error)
	Update(ctx context.Context, comment *entities.Comment) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.Comment, int64, error)
}

// RatingRepository defines the interface for rating operations
type RatingRepository interface {
	Create(ctx context.Context, rating *entities.Rating) error
	Update(ctx context.Context, rating *entities.Rating) error
	GetByVideoAndUser(ctx context.Context, videoID, userID uuid.UUID) (*entities.Rating, error)
	GetAverageRating(ctx context.Context, videoID uuid.UUID) (float64, int64, error)
	Delete(ctx context.Context, videoID, userID uuid.UUID) error
}

// AnalyticsRepository defines the interface for analytics operations
type AnalyticsRepository interface {
	RecordView(ctx context.Context, event *entities.ViewEvent) error
	RecordInteraction(ctx context.Context, event *entities.InteractionEvent) error
	GetVideoAnalytics(ctx context.Context, videoID uuid.UUID, timeRange string) (*entities.VideoAnalytics, error)
	GetUserAnalytics(ctx context.Context, userID uuid.UUID, timeRange string) (*entities.UserAnalytics, error)
	GetPlatformAnalytics(ctx context.Context, timeRange string) (*entities.PlatformAnalytics, error)
}
