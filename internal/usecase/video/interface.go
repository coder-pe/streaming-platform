// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package video

import (
	"context"
	"streaming-platform/internal/domain/entities"

	"github.com/google/uuid"
)

// VideoUsecase defines the interface for video use case
type VideoUsecase interface {
	CreateVideo(ctx context.Context, video *entities.Video) error
	GetVideoByID(ctx context.Context, videoID uuid.UUID) (*entities.Video, error)
	UpdateVideo(ctx context.Context, video *entities.Video) error
	DeleteVideo(ctx context.Context, videoID uuid.UUID) error
	SearchPublicVideos(ctx context.Context, searchReq entities.VideoSearchRequest) (*entities.VideoSearchResponse, error)
	GetVideosByCategory(ctx context.Context, category string, page, limit int) ([]*entities.Video, int64, error)
	GetVideosByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]*entities.Video, int64, error)
	GetPublicVideos(ctx context.Context, page, limit int) ([]*entities.Video, int64, error)
	GetFeaturedVideos(ctx context.Context, limit int) ([]*entities.Video, error)
	AddVideoFile(ctx context.Context, videoFile *entities.VideoFile) error
	GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entities.VideoFile, error)
	UpdateVideoStatus(ctx context.Context, videoID uuid.UUID, status entities.VideoStatus) error
	IncrementViewCount(ctx context.Context, videoID uuid.UUID) error
	GetVideoStats(ctx context.Context, videoID uuid.UUID) (*entities.VideoStats, error)
	QueueTranscodingJob(ctx context.Context, job map[string]interface{}) error
	UpdateProcessingProgress(ctx context.Context, videoID uuid.UUID, progress int) error
	GetProcessingProgress(ctx context.Context, videoID uuid.UUID) (int, error)
	UpdateVideoWithFiles(ctx context.Context, video *entities.Video) error
}
