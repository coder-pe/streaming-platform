// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package domain

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	Title        string       `json:"title" db:"title"`
	Description  string       `json:"description" db:"description"`
	InstructorID uuid.UUID    `json:"instructor_id" db:"instructor_id"`
	Instructor   *UserProfile `json:"instructor,omitempty"`
	Category     string       `json:"category" db:"category"`
	Tags         []string     `json:"tags" db:"tags"`
	Duration     int          `json:"duration" db:"duration"` // en segundos
	Thumbnail    string       `json:"thumbnail" db:"thumbnail"`
	Status       VideoStatus  `json:"status" db:"status"`
	IsPublic     bool         `json:"is_public" db:"is_public"`
	ViewCount    int64        `json:"view_count" db:"view_count"`
	Rating       float64      `json:"rating" db:"rating"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	VideoFiles   []VideoFile  `json:"video_files,omitempty"`
}

type VideoStatus string

const (
	VideoStatusUploading  VideoStatus = "uploading"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusReady      VideoStatus = "ready"
	VideoStatusFailed     VideoStatus = "failed"
)

type VideoFile struct {
	ID       uuid.UUID `json:"id" db:"id"`
	VideoID  uuid.UUID `json:"video_id" db:"video_id"`
	Quality  string    `json:"quality" db:"quality"` // 1080p, 720p, 480p, 360p
	FilePath string    `json:"file_path" db:"file_path"`
	FileSize int64     `json:"file_size" db:"file_size"`
	Bitrate  int       `json:"bitrate" db:"bitrate"`
	Format   string    `json:"format" db:"format"` // hls, mp4
}

type VideoUploadRequest struct {
	Title       string   `json:"title" validate:"required,min=3,max=255"`
	Description string   `json:"description" validate:"max=2000"`
	Category    string   `json:"category" validate:"required"`
	Tags        []string `json:"tags"`
	IsPublic    bool     `json:"is_public"`
}

type VideoSearchRequest struct {
	Query    string   `json:"query"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Page     int      `json:"page"`
	Limit    int      `json:"limit"`
}

type VideoSearchResponse struct {
	Videos []Video `json:"videos"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

type VideoStats struct {
	ID            uuid.UUID `json:"id" db:"id"`
	VideoID       uuid.UUID `json:"video_id" db:"video_id"`
	ViewCount     int64     `json:"view_count" db:"view_count"`
	Views         int64     `json:"views" db:"views"`
	LikeCount     int64     `json:"like_count" db:"like_count"`
	Likes         int64     `json:"likes" db:"likes"`
	CommentCount  int64     `json:"comment_count" db:"comment_count"`
	Comments      int64     `json:"comments" db:"comments"`
	ShareCount    int64     `json:"share_count" db:"share_count"`
	Shares        int64     `json:"shares" db:"shares"`
	AverageRating float64   `json:"average_rating" db:"average_rating"`
	RatingCount   int64     `json:"rating_count" db:"rating_count"`
	WatchTime     int64     `json:"watch_time" db:"watch_time"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type ViewEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	Duration  int       `json:"duration" db:"duration"`
	Quality   string    `json:"quality" db:"quality"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type InteractionEvent struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	VideoID      uuid.UUID `json:"video_id" db:"video_id"`
	Action       string    `json:"action" db:"action"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	Metadata     string    `json:"metadata" db:"metadata"`
}

type VideoAnalytics struct {
	VideoID      uuid.UUID `json:"video_id"`
	Views        int64     `json:"views"`
	UniqueViews  int64     `json:"unique_views"`
	WatchTime    int64     `json:"watch_time"`
	Retention    float64   `json:"retention"`
	Engagement   float64   `json:"engagement"`
}

type UserAnalytics struct {
	UserID       uuid.UUID `json:"user_id"`
	WatchTime    int64     `json:"watch_time"`
	VideosViewed int64     `json:"videos_viewed"`
	Interactions int64     `json:"interactions"`
}

type PlatformAnalytics struct {
	TotalViews    int64   `json:"total_views"`
	TotalUsers    int64   `json:"total_users"`
	TotalVideos   int64   `json:"total_videos"`
	AvgWatchTime  float64 `json:"avg_watch_time"`
	PopularTags   []string `json:"popular_tags"`
}

// Favorite representa un video marcado como favorito por un usuario
type Favorite struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	Video     *Video    `json:"video,omitempty"` // Puede incluir el video completo
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
