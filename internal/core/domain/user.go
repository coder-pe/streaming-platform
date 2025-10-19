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
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Role         string    `json:"role" db:"role"`
	Avatar       string    `json:"avatar" db:"avatar"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

type UserStats struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	VideosWatched   int64     `json:"videos_watched" db:"videos_watched"`
	TotalWatchTime  int64     `json:"total_watch_time" db:"total_watch_time"`
	VideosUploaded  int64     `json:"videos_uploaded" db:"videos_uploaded"`
	CommentsPosted  int64     `json:"comments_posted" db:"comments_posted"`
	FavoritesCount  int64     `json:"favorites_count" db:"favorites_count"`
	TotalVideos     int64     `json:"total_videos" db:"total_videos"`
	TotalViews      int64     `json:"total_views" db:"total_views"`
	LastActiveAt    time.Time `json:"last_active_at" db:"last_active_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type WatchHistory struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	Position  int       `json:"position" db:"position"`
	Quality   string    `json:"quality" db:"quality"`
	WatchedAt time.Time `json:"watched_at" db:"watched_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Comment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Rating struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	Rating    int       `json:"rating" db:"rating"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

type StreamingClaims struct {
	UserID  uuid.UUID `json:"user_id"`
	VideoID uuid.UUID `json:"video_id"`
	Quality string    `json:"quality"`
}

type StreamingStats struct {
	VideoID        uuid.UUID `json:"video_id"`
	ActiveViewers  int64     `json:"active_viewers"`
	TotalViews     int64     `json:"total_views"`
	AverageQuality string    `json:"average_quality"`
	Bandwidth      int64     `json:"bandwidth"`
}

type Notification struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Type      string    `json:"type" db:"type"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type AnalyticsEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	EventType string    `json:"event_type" db:"event_type"`
	Data      string    `json:"data" db:"data"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type VideoReport struct {
	VideoID      uuid.UUID `json:"video_id"`
	Title        string    `json:"title"`
	Views        int64     `json:"views"`
	WatchTime    int64     `json:"watch_time"`
	Engagement   float64   `json:"engagement"`
	Revenue      float64   `json:"revenue"`
}

type UserReport struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	VideosPosted int64     `json:"videos_posted"`
	TotalViews   int64     `json:"total_views"`
	Engagement   float64   `json:"engagement"`
}

type PlatformReport struct {
	TotalUsers      int64   `json:"total_users"`
	TotalVideos     int64   `json:"total_videos"`
	TotalViews      int64   `json:"total_views"`
	TotalWatchTime  int64   `json:"total_watch_time"`
	AverageRating   float64 `json:"average_rating"`
	TopCategories   []string `json:"top_categories"`
}
