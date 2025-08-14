// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package streaming

import (
	"context"
	"streaming-platform/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

// StreamingUsecase defines the interface for streaming use case
type StreamingUsecase interface {
	GetHLSPlaylist(ctx context.Context, videoID uuid.UUID) (*entities.HLSPlaylist, error)
	GetVariantPlaylist(ctx context.Context, videoID uuid.UUID, quality string) (string, error)
	CheckVideoAccess(ctx context.Context, userID, videoID uuid.UUID) (bool, error)
	GenerateStreamingToken(ctx context.Context, userID, videoID uuid.UUID, duration time.Duration) (string, error)
	ValidateStreamingToken(ctx context.Context, token string) (*entities.StreamingClaims, error)
	CreateStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) (*entities.StreamingSession, error)
	UpdateStreamingSession(ctx context.Context, userID, videoID uuid.UUID, position int) error
	EndStreamingSession(ctx context.Context, sessionID uuid.UUID) error
	RecordStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) error
	UpdateStreamingStats(ctx context.Context, userID, videoID uuid.UUID, segment string) error
	GetStreamingInfo(ctx context.Context, userID, videoID uuid.UUID) (map[string]interface{}, error)
	UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error
}
