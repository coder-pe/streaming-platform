// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"firstName" db:"first_name"`
	LastName     string    `json:"lastName" db:"last_name"`
	Role         string    `json:"role" db:"role"`
	Avatar       string    `json:"avatar" db:"avatar"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Role      string    `json:"role"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserStats struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"userId" db:"user_id"`
	VideosWatched  int64     `json:"videosWatched" db:"videos_watched"`
	TotalWatchTime int64     `json:"totalWatchTime" db:"total_watch_time"`
	VideosUploaded int64     `json:"videosUploaded" db:"videos_uploaded"`
	CommentsPosted int64     `json:"commentsPosted" db:"comments_posted"`
	FavoritesCount int64     `json:"favoritesCount" db:"favorites_count"`
	TotalVideos    int64     `json:"totalVideos" db:"total_videos"`
	TotalViews     int64     `json:"totalViews" db:"total_views"`
	LastActiveAt   time.Time `json:"lastActiveAt" db:"last_active_at"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type WatchHistory struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	VideoID   uuid.UUID `json:"videoId" db:"video_id"`
	Video     *Video    `json:"video,omitempty"` // Puede incluir el video completo
	Position  int       `json:"position" db:"position"`
	Quality   string    `json:"quality" db:"quality"`
	WatchedAt time.Time `json:"watchedAt" db:"watched_at"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	VideoID   uuid.UUID `json:"videoId" db:"video_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type Rating struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	VideoID   uuid.UUID `json:"videoId" db:"video_id"`
	Rating    int       `json:"rating" db:"rating"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type TokenClaims struct {
	UserID uuid.UUID `json:"userId"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

type StreamingClaims struct {
	UserID  uuid.UUID `json:"userId"`
	VideoID uuid.UUID `json:"videoId"`
	Quality string    `json:"quality"`
}

type StreamingStats struct {
	VideoID        uuid.UUID `json:"videoId"`
	ActiveViewers  int64     `json:"activeViewers"`
	TotalViews     int64     `json:"totalViews"`
	AverageQuality string    `json:"averageQuality"`
	Bandwidth      int64     `json:"bandwidth"`
}

type Notification struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Type      string    `json:"type" db:"type"`
	IsRead    bool      `json:"isRead" db:"is_read"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type AnalyticsEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	VideoID   uuid.UUID `json:"videoId" db:"video_id"`
	EventType string    `json:"eventType" db:"event_type"`
	Data      string    `json:"data" db:"data"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type VideoReport struct {
	VideoID    uuid.UUID `json:"videoId"`
	Title      string    `json:"title"`
	Views      int64     `json:"views"`
	WatchTime  int64     `json:"watchTime"`
	Engagement float64   `json:"engagement"`
	Revenue    float64   `json:"revenue"`
}

type UserReport struct {
	UserID       uuid.UUID `json:"userId"`
	Email        string    `json:"email"`
	VideosPosted int64     `json:"videosPosted"`
	TotalViews   int64     `json:"totalViews"`
	Engagement   float64   `json:"engagement"`
}

type PlatformReport struct {
	TotalUsers     int64    `json:"totalUsers"`
	TotalVideos    int64    `json:"totalVideos"`
	TotalViews     int64    `json:"totalViews"`
	TotalWatchTime int64    `json:"totalWatchTime"`
	AverageRating  float64  `json:"averageRating"`
	TopCategories  []string `json:"topCategories"`
}
