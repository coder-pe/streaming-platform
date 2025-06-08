// internal/domain/interfaces/services.go
package interfaces

import (
	"context"
	"time"

	"streaming-platform/internal/domain/entities"

	"github.com/google/uuid"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Authentication
	Login(ctx context.Context, email, password string) (string, string, *entities.UserProfile, error)
	Register(ctx context.Context, user *entities.User, password string) (string, string, *entities.UserProfile, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, string, *entities.UserProfile, error)
	Logout(ctx context.Context, refreshToken string) error

	// Password management
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error

	// Token operations
	ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error)
	RevokeToken(ctx context.Context, token string) error

	// Session management
	CreateSession(ctx context.Context, userID uuid.UUID) (string, error)
	ValidateSession(ctx context.Context, sessionID string) (*entities.User, error)
	InvalidateSession(ctx context.Context, sessionID string) error
}

// UserService defines the interface for user management operations
type UserService interface {
	// Profile management
	GetProfile(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entities.UserProfile, error)
	DeactivateAccount(ctx context.Context, userID uuid.UUID, password, reason string) error

	// Settings
	GetSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	UpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error

	// Favorites
	AddToFavorites(ctx context.Context, userID, videoID uuid.UUID) error
	RemoveFromFavorites(ctx context.Context, userID, videoID uuid.UUID) error
	GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.Video, int64, error)

	// Watch history
	GetWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.WatchHistory, int64, error)
	UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error

	// Statistics
	GetUserStats(ctx context.Context, userID uuid.UUID) (*entities.UserStats, error)

	// Public operations
	GetPublicProfile(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error)
	SearchUsers(ctx context.Context, query string, page, limit int) ([]*entities.UserProfile, int64, error)
}

// VideoService defines the interface for video management operations
type VideoService interface {
	// Basic operations
	CreateVideo(ctx context.Context, video *entities.Video) error
	GetVideo(ctx context.Context, videoID uuid.UUID) (*entities.Video, error)
	UpdateVideo(ctx context.Context, video *entities.Video) error
	DeleteVideo(ctx context.Context, videoID uuid.UUID) error

	// Search and listing
	SearchVideos(ctx context.Context, searchReq entities.VideoSearchRequest) (*entities.VideoSearchResponse, error)
	GetVideosByCategory(ctx context.Context, category string, page, limit int) ([]*entities.Video, int64, error)
	GetVideosByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]*entities.Video, int64, error)
	GetPublicVideos(ctx context.Context, page, limit int) ([]*entities.Video, int64, error)
	GetFeaturedVideos(ctx context.Context, limit int) ([]*entities.Video, error)

	// File management
	AddVideoFile(ctx context.Context, videoFile *entities.VideoFile) error
	GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entities.VideoFile, error)
	UpdateVideoStatus(ctx context.Context, videoID uuid.UUID, status entities.VideoStatus) error

	// Statistics
	IncrementViewCount(ctx context.Context, videoID uuid.UUID) error
	GetVideoStats(ctx context.Context, videoID uuid.UUID) (*entities.VideoStats, error)

	// Processing
	QueueVideoForProcessing(ctx context.Context, videoID uuid.UUID, filePath string) error
	UpdateProcessingProgress(ctx context.Context, videoID uuid.UUID, progress int) error
}

// StreamingService defines the interface for streaming operations
type StreamingService interface {
	// HLS operations
	GetHLSPlaylist(ctx context.Context, videoID uuid.UUID) (*entities.HLSPlaylist, error)
	GetVariantPlaylist(ctx context.Context, videoID uuid.UUID, quality string) (string, error)
	GetSegment(ctx context.Context, videoID uuid.UUID, quality, segment string) ([]byte, error)

	// Access control
	CheckVideoAccess(ctx context.Context, userID, videoID uuid.UUID) (bool, error)
	GenerateStreamingToken(ctx context.Context, userID, videoID uuid.UUID, duration time.Duration) (string, error)
	ValidateStreamingToken(ctx context.Context, token string) (*entities.StreamingClaims, error)

	// Session management
	CreateStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) (*entities.StreamingSession, error)
	UpdateStreamingSession(ctx context.Context, sessionID uuid.UUID, position int) error
	EndStreamingSession(ctx context.Context, sessionID uuid.UUID) error

	// Analytics
	RecordView(ctx context.Context, userID, videoID uuid.UUID, duration int) error
	RecordBandwidthUsage(ctx context.Context, userID, videoID uuid.UUID, bytes int64) error
	GetStreamingStats(ctx context.Context, videoID uuid.UUID) (*entities.StreamingStats, error)
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Email notifications
	SendEmail(ctx context.Context, to, subject, body string) error
	SendTemplatedEmail(ctx context.Context, to, template string, data map[string]interface{}) error

	// Push notifications
	SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error
	SendBroadcastNotification(ctx context.Context, title, body string, data map[string]interface{}) error

	// In-app notifications
	CreateNotification(ctx context.Context, notification *entities.Notification) error
	GetUserNotifications(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.Notification, int64, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	// Subscription management
	Subscribe(ctx context.Context, userID uuid.UUID, instructorID uuid.UUID) error
	Unsubscribe(ctx context.Context, userID uuid.UUID, instructorID uuid.UUID) error
	GetSubscriptions(ctx context.Context, userID uuid.UUID) ([]*entities.UserProfile, error)
	GetSubscribers(ctx context.Context, instructorID uuid.UUID) ([]*entities.UserProfile, error)
}

// AnalyticsService defines the interface for analytics operations
type AnalyticsService interface {
	// Event tracking
	TrackEvent(ctx context.Context, event *entities.AnalyticsEvent) error
	TrackVideoView(ctx context.Context, userID, videoID uuid.UUID, duration int, quality string) error
	TrackUserAction(ctx context.Context, userID uuid.UUID, action, resource string, metadata map[string]interface{}) error

	// Reports
	GenerateVideoReport(ctx context.Context, videoID uuid.UUID, timeRange string) (*entities.VideoReport, error)
	GenerateUserReport(ctx context.Context, userID uuid.UUID, timeRange string) (*entities.UserReport, error)
	GeneratePlatformReport(ctx context.Context, timeRange string) (*entities.PlatformReport, error)

	// Real-time metrics
	GetRealTimeViewers(ctx context.Context, videoID uuid.UUID) (int64, error)
	GetPopularContent(ctx context.Context, timeRange string, limit int) ([]*entities.Video, error)
	GetTrendingTopics(ctx context.Context, timeRange string, limit int) ([]string, error)
}

// SearchService defines the interface for search operations
type SearchService interface {
	// Video search
	SearchVideos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*entities.Video, int64, error)
	SearchSuggestions(ctx context.Context, query string, limit int) ([]string, error)

	// User search
	SearchUsers(ctx context.Context, query string, page, limit int) ([]*entities.UserProfile, int64, error)

	// Index management
	IndexVideo(ctx context.Context, video *entities.Video) error
	UpdateVideoIndex(ctx context.Context, videoID uuid.UUID, updates map[string]interface{}) error
	RemoveFromIndex(ctx context.Context, videoID uuid.UUID) error

	// Autocomplete
	GetAutocompleteSuggestions(ctx context.Context, query string, limit int) ([]string, error)

	// Popular searches
	GetPopularSearches(ctx context.Context, limit int) ([]string, error)
	TrackSearch(ctx context.Context, query string, userID *uuid.UUID) error
}
