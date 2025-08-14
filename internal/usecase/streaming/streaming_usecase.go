// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package streaming

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/domain/errors"
	"streaming-platform/internal/domain/interfaces"
	"streaming-platform/pkg/jwt"

	"github.com/google/uuid"
)

type StreamingUsecase struct {
	videoRepo  interfaces.VideoRepository
	cacheRepo  interfaces.CacheRepository
	cdnBaseURL string
	jwtSecret  string
}

func NewStreamingUsecase(videoRepo interfaces.VideoRepository, cacheRepo interfaces.CacheRepository, cdnBaseURL, jwtSecret string) *StreamingUsecase {
	return &StreamingUsecase{
		videoRepo:  videoRepo,
		cacheRepo:  cacheRepo,
		cdnBaseURL: cdnBaseURL,
		jwtSecret:  jwtSecret,
	}
}

func (s *StreamingUsecase) GetHLSPlaylist(ctx context.Context, videoID uuid.UUID) (*entities.HLSPlaylist, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("hls_playlist:%s", videoID.String())
	var cachedPlaylist entities.HLSPlaylist
	if err := s.cacheRepo.Get(ctx, cacheKey, &cachedPlaylist); err == nil {
		return &cachedPlaylist, nil
	}

	// Get video from database
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Check if video is ready for streaming
	if video.Status != entities.VideoStatusReady {
		return nil, errors.VideoNotReady
	}

	// Get video files
	videoFiles, err := s.videoRepo.GetVideoFiles(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video files: %w", err)
	}

	if len(videoFiles) == 0 {
		return nil, errors.VideoNotReady
	}

	// Build master playlist
	masterPlaylist := s.buildMasterPlaylist(videoFiles)

	// Build variants
	variants := []entities.HLSVariant{}
	for _, file := range videoFiles {
		if file.Format == "hls" {
			variant := entities.HLSVariant{
				Quality:     file.Quality,
				Bandwidth:   file.Bitrate * 1000, // Convert to bits per second
				Resolution:  s.getResolutionForQuality(file.Quality),
				PlaylistURL: fmt.Sprintf("%s/%s", s.cdnBaseURL, file.FilePath),
			}
			variants = append(variants, variant)
		}
	}

	playlist := &entities.HLSPlaylist{
		MasterPlaylist: masterPlaylist,
		Variants:       variants,
		BaseURL:        s.cdnBaseURL,
	}

	// Cache the playlist for 30 minutes
	if err := s.cacheRepo.Set(ctx, cacheKey, playlist, 30*time.Minute); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache HLS playlist: %v\n", err)
	}

	return playlist, nil
}

func (s *StreamingUsecase) GetVariantPlaylist(ctx context.Context, videoID uuid.UUID, quality string) (string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("variant_playlist:%s:%s", videoID.String(), quality)
	var cachedPlaylist string
	if err := s.cacheRepo.Get(ctx, cacheKey, &cachedPlaylist); err == nil {
		return cachedPlaylist, nil
	}

	// Get video files
	videoFiles, err := s.videoRepo.GetVideoFiles(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("failed to get video files: %w", err)
	}

	// Find the specific quality file
	var targetFile *entities.VideoFile
	for _, file := range videoFiles {
		if file.Quality == quality && file.Format == "hls" {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return "", fmt.Errorf("video file not found for quality %s", quality)
	}

	// Read the playlist file
	playlistPath := filepath.Join("storage", targetFile.FilePath)
	playlistContent, err := os.ReadFile(playlistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read playlist file: %w", err)
	}

	playlist := string(playlistContent)

	// Cache the playlist for 30 minutes
	if err := s.cacheRepo.SetString(ctx, cacheKey, playlist, 30*time.Minute); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache variant playlist: %v\n", err)
	}

	return playlist, nil
}

func (s *StreamingUsecase) CheckVideoAccess(ctx context.Context, userID, videoID uuid.UUID) (bool, error) {
	// Get video
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return false, fmt.Errorf("failed to get video: %w", err)
	}

	// If video is public, allow access
	if video.IsPublic {
		return true, nil
	}

	// If user is not authenticated, deny access
	if userID == uuid.Nil {
		return false, nil
	}

	// If user is the video owner, allow access
	if video.InstructorID == userID {
		return true, nil
	}

	// TODO: Add more access control logic here
	// - Check if user has purchased the video
	// - Check if user has a subscription
	// - Check if user is enrolled in a course

	return false, nil
}

func (s *StreamingUsecase) GenerateStreamingToken(ctx context.Context, userID, videoID uuid.UUID, duration time.Duration) (string, error) {
	// Check if user has access
	hasAccess, err := s.CheckVideoAccess(ctx, userID, videoID)
	if err != nil {
		return "", fmt.Errorf("failed to check video access: %w", err)
	}

	if !hasAccess {
		return "", errors.VideoAccessDenied
	}

	// Create streaming claims
	claims := &jwt.Claims{
		UserID: userID.String(),
		CustomClaims: map[string]interface{}{
			"video_id": videoID.String(),
			"type":     "streaming",
		},
	}

	// Generate token
	token, err := jwt.GenerateToken(claims, s.jwtSecret, duration)
	if err != nil {
		return "", fmt.Errorf("failed to generate streaming token: %w", err)
	}

	return token, nil
}

func (s *StreamingUsecase) ValidateStreamingToken(ctx context.Context, token string) (*entities.StreamingClaims, error) {
	// Validate JWT token
	claims, err := jwt.ValidateToken(token, s.jwtSecret)
	if err != nil {
		return nil, errors.AuthTokenInvalid
	}

	// Extract streaming claims
	videoID, ok := claims.CustomClaims["video_id"].(string)
	if !ok {
		return nil, errors.AuthTokenInvalid
	}

	tokenType, ok := claims.CustomClaims["type"].(string)
	if !ok || tokenType != "streaming" {
		return nil, errors.AuthTokenInvalid
	}

	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.AuthTokenInvalid
	}

	videoUUID, err := uuid.Parse(videoID)
	if err != nil {
		return nil, errors.AuthTokenInvalid
	}

	streamingClaims := &entities.StreamingClaims{
		UserID:  userUUID,
		VideoID: videoUUID,
	}

	return streamingClaims, nil
}

func (s *StreamingUsecase) CreateStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) (*entities.StreamingSession, error) {
	// Check video access
	hasAccess, err := s.CheckVideoAccess(ctx, userID, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to check video access: %w", err)
	}

	if !hasAccess {
		return nil, errors.VideoAccessDenied
	}

	// Create streaming session
	session := &entities.StreamingSession{
		UserID:    userID,
		VideoID:   videoID,
		Quality:   quality,
		Position:  0,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store session in cache
	sessionKey := fmt.Sprintf("streaming_session:%s:%s", userID.String(), videoID.String())
	if err := s.cacheRepo.Set(ctx, sessionKey, session, 4*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store streaming session: %w", err)
	}

	return session, nil
}

func (s *StreamingUsecase) UpdateStreamingSession(ctx context.Context, userID, videoID uuid.UUID, position int) error {
	sessionKey := fmt.Sprintf("streaming_session:%s:%s", userID.String(), videoID.String())

	// Get existing session
	var session entities.StreamingSession
	if err := s.cacheRepo.Get(ctx, sessionKey, &session); err != nil {
		// Session doesn't exist, create a new one
		_, err := s.CreateStreamingSession(ctx, userID, videoID, "720p")
		if err != nil {
			return fmt.Errorf("failed to create streaming session: %w", err)
		}

		// Get the newly created session
		if err := s.cacheRepo.Get(ctx, sessionKey, &session); err != nil {
			return fmt.Errorf("failed to get streaming session: %w", err)
		}
	}

	// Update session
	session.Position = position
	session.UpdatedAt = time.Now()

	// Store updated session
	if err := s.cacheRepo.Set(ctx, sessionKey, session, 4*time.Hour); err != nil {
		return fmt.Errorf("failed to update streaming session: %w", err)
	}

	return nil
}

func (s *StreamingUsecase) EndStreamingSession(ctx context.Context, userID, videoID uuid.UUID) error {
	sessionKey := fmt.Sprintf("streaming_session:%s:%s", userID.String(), videoID.String())

	// Get session to record final stats
	var session entities.StreamingSession
	if err := s.cacheRepo.Get(ctx, sessionKey, &session); err == nil {
		// Calculate watch duration
		watchDuration := int(time.Since(session.StartedAt).Seconds())

		// TODO: Record analytics
		// - Total watch time
		// - Completion percentage
		// - Quality used
		// - Bandwidth consumption

		fmt.Printf("Streaming session ended - User: %s, Video: %s, Duration: %d seconds\n",
			userID, videoID, watchDuration)
	}

	// Remove session from cache
	if err := s.cacheRepo.Delete(ctx, sessionKey); err != nil {
		return fmt.Errorf("failed to delete streaming session: %w", err)
	}

	return nil
}

func (s *StreamingUsecase) RecordStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) error {
	// Create or update streaming session
	return s.UpdateStreamingSession(ctx, userID, videoID, 0)
}

func (s *StreamingUsecase) UpdateStreamingStats(ctx context.Context, userID, videoID uuid.UUID, segment string) error {
	// Record segment access for analytics
	statsKey := fmt.Sprintf("streaming_stats:%s:%s", videoID.String(), time.Now().Format("2006-01-02"))

	// Increment segment access count
	if _, err := s.cacheRepo.Increment(ctx, statsKey); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to update streaming stats: %v\n", err)
	}

	// Set expiration to 30 days
	if err := s.cacheRepo.SetTTL(ctx, statsKey, 30*24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to set TTL for streaming stats: %v\n", err)
	}

	return nil
}

func (s *StreamingUsecase) GetStreamingInfo(ctx context.Context, userID, videoID uuid.UUID) (map[string]interface{}, error) {
	// Get video info
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Check access
	hasAccess, err := s.CheckVideoAccess(ctx, userID, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to check video access: %w", err)
	}

	// Get available qualities
	videoFiles, err := s.videoRepo.GetVideoFiles(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video files: %w", err)
	}

	qualities := []string{}
	for _, file := range videoFiles {
		if file.Format == "hls" {
			qualities = append(qualities, file.Quality)
		}
	}

	// Get user's watch progress (if authenticated)
	var watchProgress map[string]interface{}
	if userID != uuid.Nil {
		sessionKey := fmt.Sprintf("streaming_session:%s:%s", userID.String(), videoID.String())
		var session entities.StreamingSession
		if err := s.cacheRepo.Get(ctx, sessionKey, &session); err == nil {
			watchProgress = map[string]interface{}{
				"position":     session.Position,
				"percentage":   float64(session.Position) / float64(video.Duration) * 100,
				"last_watched": session.UpdatedAt,
			}
		}
	}

	info := map[string]interface{}{
		"video_id":       videoID,
		"title":          video.Title,
		"duration":       video.Duration,
		"has_access":     hasAccess,
		"qualities":      qualities,
		"watch_progress": watchProgress,
	}

	if hasAccess {
		// Generate streaming token
		token, err := s.GenerateStreamingToken(ctx, userID, videoID, 2*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate streaming token: %w", err)
		}
		info["streaming_token"] = token

		// Add HLS playlist URL
		info["playlist_url"] = fmt.Sprintf("/api/stream/%s/playlist", videoID)
	}

	return info, nil
}

func (s *StreamingUsecase) UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error {
	// Update streaming session
	if err := s.UpdateStreamingSession(ctx, userID, videoID, position); err != nil {
		return fmt.Errorf("failed to update streaming session: %w", err)
	}

	// TODO: Store persistent watch progress in database
	// This would allow users to resume watching across devices

	return nil
}

// Helper methods

func (s *StreamingUsecase) buildMasterPlaylist(videoFiles []*entities.VideoFile) string {
	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n")
	playlist.WriteString("#EXT-X-VERSION:3\n\n")

	// Sort files by quality (highest first)
	// qualityOrder could be used for sorting if needed
	_ = map[string]int{
		"1080p": 1,
		"720p":  2,
		"480p":  3,
		"360p":  4,
	}

	for _, file := range videoFiles {
		if file.Format == "hls" {
			bandwidth := file.Bitrate * 1000 // Convert to bps
			resolution := s.getResolutionForQuality(file.Quality)

			playlist.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n",
				bandwidth, resolution))
			playlist.WriteString(fmt.Sprintf("hls/%s/playlist.m3u8\n", file.Quality))
		}
	}

	return playlist.String()
}

func (s *StreamingUsecase) getResolutionForQuality(quality string) string {
	switch quality {
	case "1080p":
		return "1920x1080"
	case "720p":
		return "1280x720"
	case "480p":
		return "854x480"
	case "360p":
		return "640x360"
	default:
		return "1280x720"
	}
}

func (s *StreamingUsecase) getBitrateForQuality(quality string) int {
	switch quality {
	case "1080p":
		return 5000
	case "720p":
		return 2500
	case "480p":
		return 1000
	case "360p":
		return 500
	default:
		return 2500
	}
}
