// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package video

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/domain/errors"
	"streaming-platform/internal/domain/interfaces"

	"github.com/google/uuid"
)

type videoUsecaseImpl struct {
	videoRepo interfaces.VideoRepository
	cacheRepo interfaces.CacheRepository
	jobQueue  chan map[string]interface{} // Simple job queue for transcoding
}

func NewVideoUsecase(videoRepo interfaces.VideoRepository, cacheRepo interfaces.CacheRepository) *videoUsecaseImpl {
	return &videoUsecaseImpl{
		videoRepo: videoRepo,
		cacheRepo: cacheRepo,
		jobQueue:  make(chan map[string]interface{}, 100),
	}
}

func (v *videoUsecaseImpl) CreateVideo(ctx context.Context, video *entities.Video) error {
	// Validate video data
	if err := v.validateVideo(video); err != nil {
		return err
	}

	// Set default values
	if video.Status == "" {
		video.Status = entities.VideoStatusUploading
	}
	if video.ViewCount == 0 {
		video.ViewCount = 0
	}
	if video.Rating == 0 {
		video.Rating = 0.0
	}

	// Create video in database
	if err := v.videoRepo.Create(ctx, video); err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}

func (v *videoUsecaseImpl) GetVideoByID(ctx context.Context, videoID uuid.UUID) (*entities.Video, error) {
	// Check cache first
	if cachedVideo, err := v.cacheRepo.GetCachedVideo(ctx, videoID); err == nil {
		return cachedVideo, nil
	}

	// Get from database
	video, err := v.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Cache the video
	if err := v.cacheRepo.CacheVideo(ctx, video); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache video: %v\n", err)
	}

	return video, nil
}

func (v *videoUsecaseImpl) UpdateVideo(ctx context.Context, video *entities.Video) error {
	// Validate video data
	if err := v.validateVideo(video); err != nil {
		return err
	}

	// Update in database
	if err := v.videoRepo.Update(ctx, video); err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	// Invalidate cache
	if err := v.cacheRepo.InvalidateVideoCache(ctx, video.ID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	return nil
}

func (v *videoUsecaseImpl) DeleteVideo(ctx context.Context, videoID uuid.UUID) error {
	// Delete from database
	if err := v.videoRepo.Delete(ctx, videoID); err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	// Invalidate cache
	if err := v.cacheRepo.InvalidateVideoCache(ctx, videoID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	// TODO: Delete video files from storage
	// This should be handled by a background job

	return nil
}

func (v *videoUsecaseImpl) SearchPublicVideos(ctx context.Context, searchReq entities.VideoSearchRequest) (*entities.VideoSearchResponse, error) {
	// Validate search request
	if searchReq.Page <= 0 {
		searchReq.Page = 1
	}
	if searchReq.Limit <= 0 || searchReq.Limit > 50 {
		searchReq.Limit = 20
	}

	// Search videos
	response, err := v.videoRepo.Search(ctx, searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}

	return response, nil
}

func (v *videoUsecaseImpl) GetVideosByCategory(ctx context.Context, category string, page, limit int) ([]*entities.Video, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return v.videoRepo.GetByCategory(ctx, category, page, limit)
}

func (v *videoUsecaseImpl) GetVideosByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]*entities.Video, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return v.videoRepo.GetByInstructor(ctx, instructorID, page, limit)
}

func (v *videoUsecaseImpl) GetPublicVideos(ctx context.Context, page, limit int) ([]*entities.Video, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return v.videoRepo.GetPublicVideos(ctx, page, limit)
}

func (v *videoUsecaseImpl) GetFeaturedVideos(ctx context.Context, limit int) ([]*entities.Video, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return v.videoRepo.GetFeaturedVideos(ctx, limit)
}

func (v *videoUsecaseImpl) IncrementViewCount(ctx context.Context, videoID uuid.UUID) error {
	// Increment in database
	if err := v.videoRepo.IncrementViewCount(ctx, videoID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	// Invalidate cache to refresh view count
	if err := v.cacheRepo.InvalidateVideoCache(ctx, videoID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	return nil
}

func (v *videoUsecaseImpl) GetVideoStats(ctx context.Context, videoID uuid.UUID) (*entities.VideoStats, error) {
	return v.videoRepo.GetVideoStats(ctx, videoID)
}

func (v *videoUsecaseImpl) UpdateVideoStatus(ctx context.Context, videoID uuid.UUID, status entities.VideoStatus) error {
	if err := v.videoRepo.UpdateStatus(ctx, videoID, status); err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	// Invalidate cache
	if err := v.cacheRepo.InvalidateVideoCache(ctx, videoID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	return nil
}

// Video file management
func (v *videoUsecaseImpl) AddVideoFile(ctx context.Context, videoFile *entities.VideoFile) error {
	if err := v.videoRepo.CreateVideoFile(ctx, videoFile); err != nil {
		return fmt.Errorf("failed to add video file: %w", err)
	}

	// Invalidate video cache to refresh file list
	if err := v.cacheRepo.InvalidateVideoCache(ctx, videoFile.VideoID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	return nil
}

func (v *videoUsecaseImpl) GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entities.VideoFile, error) {
	return v.videoRepo.GetVideoFiles(ctx, videoID)
}

func (v *videoUsecaseImpl) UpdateVideoWithFiles(ctx context.Context, video *entities.Video) error {
	// Update video
	if err := v.UpdateVideo(ctx, video); err != nil {
		return err
	}

	// Add video files
	for _, videoFile := range video.VideoFiles {
		if err := v.AddVideoFile(ctx, &videoFile); err != nil {
			return fmt.Errorf("failed to add video file: %w", err)
		}
	}

	return nil
}

// Transcoding job management
func (v *videoUsecaseImpl) QueueTranscodingJob(ctx context.Context, job map[string]interface{}) error {
	select {
	case v.jobQueue <- job:
		fmt.Printf("Transcoding job queued for video: %v\n", job["video_id"])
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout queuing transcoding job")
	}
}

func (v *videoUsecaseImpl) UpdateProcessingProgress(ctx context.Context, videoID uuid.UUID, progress int) error {
	// Store progress in cache
	progressKey := fmt.Sprintf("processing_progress:%s", videoID.String())
	progressData := map[string]interface{}{
		"progress":   progress,
		"updated_at": time.Now().Unix(),
	}

	if err := v.cacheRepo.Set(ctx, progressKey, progressData, time.Hour); err != nil {
		return fmt.Errorf("failed to update processing progress: %w", err)
	}

	return nil
}

func (v *videoUsecaseImpl) GetProcessingProgress(ctx context.Context, videoID uuid.UUID) (int, error) {
	progressKey := fmt.Sprintf("processing_progress:%s", videoID.String())

	var progressData map[string]interface{}
	if err := v.cacheRepo.Get(ctx, progressKey, &progressData); err != nil {
		return 0, fmt.Errorf("processing progress not found")
	}

	if progress, ok := progressData["progress"].(float64); ok {
		return int(progress), nil
	}

	return 0, fmt.Errorf("invalid progress data")
}

// Search and filtering
func (v *videoUsecaseImpl) SearchVideos(ctx context.Context, query string, category string, tags []string, page, limit int) (*entities.VideoSearchResponse, error) {
	searchReq := entities.VideoSearchRequest{
		Query:    query,
		Category: category,
		Tags:     tags,
		Page:     page,
		Limit:    limit,
	}

	return v.SearchPublicVideos(ctx, searchReq)
}

func (v *videoUsecaseImpl) GetPopularVideos(ctx context.Context, timeRange string, limit int) ([]*entities.Video, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return v.videoRepo.GetPopularVideos(ctx, limit)
}

func (v *videoUsecaseImpl) GetVideosByStatus(ctx context.Context, status entities.VideoStatus) ([]*entities.Video, error) {
	return v.videoRepo.GetVideosByStatus(ctx, status)
}

// Admin functionality
func (v *videoUsecaseImpl) GetAllVideos(ctx context.Context, page, limit int) ([]*entities.Video, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	filters := map[string]interface{}{} // No filters for admin view
	return v.videoRepo.List(ctx, filters, page, limit)
}

func (v *videoUsecaseImpl) ModerateVideo(ctx context.Context, videoID uuid.UUID, action string, reason string) error {
	video, err := v.GetVideoByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	switch action {
	case "approve":
		video.Status = entities.VideoStatusReady
		video.IsPublic = true
	case "reject":
		video.Status = entities.VideoStatusFailed
		video.IsPublic = false
	case "hide":
		video.IsPublic = false
	default:
		return errors.ValidationFailed.WithContext("field", "action").WithContext("message", "invalid action")
	}

	if err := v.UpdateVideo(ctx, video); err != nil {
		return fmt.Errorf("failed to moderate video: %w", err)
	}

	// Log moderation action
	fmt.Printf("Video moderated - ID: %s, Action: %s, Reason: %s\n", videoID, action, reason)

	return nil
}

// Analytics and reporting
func (v *videoUsecaseImpl) GetVideoAnalytics(ctx context.Context, videoID uuid.UUID, timeRange string) (map[string]interface{}, error) {
	// Get basic video stats
	stats, err := v.GetVideoStats(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video stats: %w", err)
	}

	// Get video info
	video, err := v.GetVideoByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	analytics := map[string]interface{}{
		"video_id":   videoID,
		"title":      video.Title,
		"views":      video.ViewCount,
		"rating":     video.Rating,
		"duration":   video.Duration,
		"created_at": video.CreatedAt,
		"status":     video.Status,
		"is_public":  video.IsPublic,
		"stats":      stats,
	}

	// TODO: Add more analytics data
	// - View trends over time
	// - Geographic distribution
	// - Device/browser statistics
	// - Retention rates
	// - Engagement metrics

	return analytics, nil
}

// Validation helpers
func (v *videoUsecaseImpl) validateVideo(video *entities.Video) error {
	validationErrors := errors.NewValidationErrors()

	// Validate title
	if strings.TrimSpace(video.Title) == "" {
		validationErrors.Add("title", "title is required", video.Title)
	} else if len(video.Title) > 255 {
		validationErrors.Add("title", "title must be at most 255 characters", video.Title)
	}

	// Validate description
	if len(video.Description) > 2000 {
		validationErrors.Add("description", "description must be at most 2000 characters", video.Description)
	}

	// Validate category
	if strings.TrimSpace(video.Category) == "" {
		validationErrors.Add("category", "category is required", video.Category)
	} else {
		validCategories := []string{"programming", "design", "marketing", "business", "other"}
		isValid := false
		for _, cat := range validCategories {
			if video.Category == cat {
				isValid = true
				break
			}
		}
		if !isValid {
			validationErrors.Add("category", "invalid category", video.Category)
		}
	}

	// Validate tags
	if len(video.Tags) > 10 {
		validationErrors.Add("tags", "maximum 10 tags allowed", video.Tags)
	}

	for _, tag := range video.Tags {
		if len(tag) > 50 {
			validationErrors.Add("tags", "tag must be at most 50 characters", tag)
		}
	}

	// Validate duration
	if video.Duration < 0 {
		validationErrors.Add("duration", "duration must be positive", video.Duration)
	}

	// Validate instructor ID
	if video.InstructorID == uuid.Nil {
		validationErrors.Add("instructor_id", "instructor ID is required", video.InstructorID)
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}

// Utility methods
func (v *videoUsecaseImpl) GetVideoQualityOptions() []string {
	return []string{"1080p", "720p", "480p", "360p"}
}

func (v *videoUsecaseImpl) GetVideoCategories() []string {
	return []string{"programming", "design", "marketing", "business", "other"}
}

func (v *videoUsecaseImpl) GetVideoFormats() []string {
	return []string{"mp4", "avi", "mov", "mkv", "webm"}
}

func (v *videoUsecaseImpl) IsValidVideoFormat(filename string) bool {
	validFormats := v.GetVideoFormats()
	ext := strings.ToLower(filepath.Ext(filename))

	for _, format := range validFormats {
		if ext == "."+format {
			return true
		}
	}

	return false
}
