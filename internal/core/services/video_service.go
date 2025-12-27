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
	videoRepo  output.VideoRepository
	cacheRepo  output.CacheRepository
	searchRepo output.SearchRepository
}

// NewVideoService crea una nueva instancia del servicio de videos
func NewVideoService(videoRepo output.VideoRepository, cacheRepo output.CacheRepository, searchRepo output.SearchRepository) input.VideoService {
	return &videoService{
		videoRepo:  videoRepo,
		cacheRepo:  cacheRepo,
		searchRepo: searchRepo,
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

	// Indexar en Elasticsearch (no bloquear si falla)
	// Solo indexar si está público y listo
	if video.IsPublic && video.Status == domain.VideoStatusReady {
		go func() {
			indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.searchRepo.IndexVideo(indexCtx, video); err != nil {
				fmt.Printf("Warning: failed to index video in Elasticsearch: %v\n", err)
			}
		}()
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

	// Actualizar índice de Elasticsearch
	if video.IsPublic && video.Status == domain.VideoStatusReady {
		go func() {
			indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.searchRepo.UpdateVideoIndex(indexCtx, video); err != nil {
				fmt.Printf("Warning: failed to update video index in Elasticsearch: %v\n", err)
			}
		}()
	} else {
		// Si el video ya no es público o no está listo, eliminarlo del índice
		go func() {
			indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.searchRepo.DeleteVideoIndex(indexCtx, video.ID.String()); err != nil {
				fmt.Printf("Warning: failed to delete video index in Elasticsearch: %v\n", err)
			}
		}()
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

	// Eliminar del índice de Elasticsearch
	go func() {
		indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.searchRepo.DeleteVideoIndex(indexCtx, videoID.String()); err != nil {
			fmt.Printf("Warning: failed to delete video index in Elasticsearch: %v\n", err)
		}
	}()

	// TODO: Eliminar archivos de video del almacenamiento (MinIO)
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

	// Si hay query de búsqueda de texto, usar Elasticsearch para búsqueda avanzada
	if req.Query != "" {
		filters := make(map[string]interface{})
		if req.Category != "" {
			filters["category"] = req.Category
		}
		if len(req.Tags) > 0 {
			filters["tags"] = req.Tags
		}

		videoPtrs, total, err := s.searchRepo.SearchVideos(ctx, req.Query, filters, req.Page, req.Limit)
		if err != nil {
			// Fallback a PostgreSQL si Elasticsearch falla
			fmt.Printf("Warning: Elasticsearch search failed, falling back to PostgreSQL: %v\n", err)
			response, err := s.videoRepo.Search(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to search videos: %w", err)
			}
			return response, nil
		}

		// Convertir []*domain.Video a []domain.Video
		videos := make([]domain.Video, len(videoPtrs))
		for i, v := range videoPtrs {
			videos[i] = *v
		}

		return &domain.VideoSearchResponse{
			Videos: videos,
			Total:  total,
			Page:   req.Page,
			Limit:  req.Limit,
		}, nil
	}

	// Sin query de texto, usar PostgreSQL para filtros simples
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

	// Actualizar contador de vistas en Elasticsearch
	go func() {
		indexCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Obtener el video actualizado
		video, err := s.videoRepo.GetByID(indexCtx, videoID)
		if err != nil {
			fmt.Printf("Warning: failed to get video for index update: %v\n", err)
			return
		}

		// Actualizar el índice si el video es público y está listo
		if video.IsPublic && video.Status == domain.VideoStatusReady {
			if err := s.searchRepo.UpdateVideoIndex(indexCtx, video); err != nil {
				fmt.Printf("Warning: failed to update view count in Elasticsearch: %v\n", err)
			}
		}
	}()

	return nil
}

// QueueTranscodingJob ya no se usa - los jobs se envían directamente a RabbitMQ desde VideoHandler
// Mantenido para compatibilidad, pero deprecated
func (s *videoService) QueueTranscodingJob(ctx context.Context, job map[string]interface{}) error {
	fmt.Printf("Warning: QueueTranscodingJob is deprecated, jobs should be sent directly to RabbitMQ\n")
	return fmt.Errorf("deprecated method - use RabbitMQ queue directly")
}
