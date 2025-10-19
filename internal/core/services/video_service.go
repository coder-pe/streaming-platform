// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package services

import (
	"context"
	"fmt"
	"time"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/validator"

	"github.com/google/uuid"
)

// videoService implementa el puerto de entrada VideoService
type videoService struct {
	videoRepo output.VideoRepository
	cacheRepo output.CacheRepository
	jobQueue  chan map[string]interface{} // Simple job queue for transcoding
}

// NewVideoService crea una nueva instancia del servicio de videos
func NewVideoService(videoRepo output.VideoRepository, cacheRepo output.CacheRepository) input.VideoService {
	return &videoService{
		videoRepo: videoRepo,
		cacheRepo: cacheRepo,
		jobQueue:  make(chan map[string]interface{}, 100),
	}
}

func (s *videoService) CreateVideo(ctx context.Context, video *domain.Video) error {
	// Validar título
	if err := validator.ValidateRequired("title", video.Title); err != nil {
		return err
	}
	if len(video.Title) > 255 {
		return fmt.Errorf("title must be at most 255 characters")
	}

	// Validar descripción
	if len(video.Description) > 2000 {
		return fmt.Errorf("description must be at most 2000 characters")
	}

	// Validar categoría
	if err := validator.ValidateRequired("category", video.Category); err != nil {
		return err
	}

	// Validar instructor ID
	if video.InstructorID == uuid.Nil {
		return fmt.Errorf("instructor ID is required")
	}

	// Establecer valores por defecto
	if video.Status == "" {
		video.Status = domain.VideoStatusUploading
	}
	if video.ViewCount == 0 {
		video.ViewCount = 0
	}
	if video.Rating == 0 {
		video.Rating = 0.0
	}

	// Crear video en la base de datos
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}

func (s *videoService) GetVideoByID(ctx context.Context, videoID uuid.UUID) (*domain.Video, error) {
	// Verificar caché primero
	if cachedVideo, err := s.cacheRepo.GetCachedVideo(ctx, videoID); err == nil {
		return cachedVideo, nil
	}

	// Obtener de la base de datos
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Cachear el video
	if err := s.cacheRepo.CacheVideo(ctx, video); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache video: %v\n", err)
	}

	return video, nil
}

func (s *videoService) UpdateVideo(ctx context.Context, video *domain.Video) error {
	// Validar título
	if err := validator.ValidateRequired("title", video.Title); err != nil {
		return err
	}

	// Actualizar en la base de datos
	if err := s.videoRepo.Update(ctx, video); err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	// Invalidar caché
	if err := s.cacheRepo.InvalidateVideoCache(ctx, video.ID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	return nil
}

func (s *videoService) DeleteVideo(ctx context.Context, videoID uuid.UUID) error {
	// Eliminar de la base de datos
	if err := s.videoRepo.Delete(ctx, videoID); err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	// Invalidar caché
	if err := s.cacheRepo.InvalidateVideoCache(ctx, videoID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	// TODO: Eliminar archivos de video del almacenamiento
	// Esto debería manejarse mediante un trabajo en segundo plano

	return nil
}

func (s *videoService) SearchPublicVideos(ctx context.Context, req domain.VideoSearchRequest) (*domain.VideoSearchResponse, error) {
	// Validar solicitud de búsqueda
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	// Buscar videos
	response, err := s.videoRepo.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}

	return response, nil
}

func (s *videoService) IncrementViewCount(ctx context.Context, videoID uuid.UUID) error {
	// Incrementar en la base de datos
	if err := s.videoRepo.IncrementViewCount(ctx, videoID); err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	// Invalidar caché para refrescar el contador de vistas
	if err := s.cacheRepo.InvalidateVideoCache(ctx, videoID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate video cache: %v\n", err)
	}

	return nil
}

func (s *videoService) QueueTranscodingJob(ctx context.Context, job map[string]interface{}) error {
	select {
	case s.jobQueue <- job:
		fmt.Printf("Transcoding job queued for video: %v\n", job["video_id"])
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout queuing transcoding job")
	}
}
