// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/jwt"

	"github.com/google/uuid"
)

// streamingService implementa el puerto de entrada StreamingService
type streamingService struct {
	videoRepo  output.VideoRepository
	cacheRepo  output.CacheRepository
	cdnBaseURL string
	jwtSecret  string
}

// NewStreamingService crea una nueva instancia del servicio de streaming
func NewStreamingService(videoRepo output.VideoRepository, cacheRepo output.CacheRepository, cdnBaseURL, jwtSecret string) input.StreamingService {
	return &streamingService{
		videoRepo:  videoRepo,
		cacheRepo:  cacheRepo,
		cdnBaseURL: cdnBaseURL,
		jwtSecret:  jwtSecret,
	}
}

func (s *streamingService) CheckVideoAccess(ctx context.Context, userID, videoID uuid.UUID) (bool, error) {
	// Obtener video
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return false, fmt.Errorf("failed to get video: %w", err)
	}

	// Si el video es público, permitir acceso
	if video.IsPublic {
		return true, nil
	}

	// Si el usuario no está autenticado, denegar acceso
	if userID == uuid.Nil {
		return false, nil
	}

	// Si el usuario es el propietario del video, permitir acceso
	if video.InstructorID == userID {
		return true, nil
	}

	// TODO: Añadir más lógica de control de acceso aquí
	// - Verificar si el usuario ha comprado el video
	// - Verificar si el usuario tiene una suscripción
	// - Verificar si el usuario está inscrito en un curso

	return false, nil
}

func (s *streamingService) GetHLSPlaylist(ctx context.Context, videoID uuid.UUID) (*domain.HLSPlaylist, error) {
	// Verificar caché primero
	cacheKey := fmt.Sprintf("hls_playlist:%s", videoID.String())
	var cachedPlaylist domain.HLSPlaylist
	if err := s.cacheRepo.Get(ctx, cacheKey, &cachedPlaylist); err == nil {
		return &cachedPlaylist, nil
	}

	// Obtener video de la base de datos
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Verificar si el video está listo para streaming
	if video.Status != domain.VideoStatusReady {
		return nil, domain.ErrVideoNotReady
	}

	// Obtener archivos de video
	videoFiles, err := s.videoRepo.GetVideoFiles(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video files: %w", err)
	}

	if len(videoFiles) == 0 {
		return nil, domain.ErrVideoNotReady
	}

	// Construir playlist maestro
	masterPlaylist := s.buildMasterPlaylist(videoFiles)

	// Construir variantes
	variants := []domain.HLSVariant{}
	for _, file := range videoFiles {
		if file.Format == "hls" {
			variant := domain.HLSVariant{
				Quality:     file.Quality,
				Bandwidth:   file.Bitrate * 1000, // Convertir a bits por segundo
				Resolution:  s.getResolutionForQuality(file.Quality),
				PlaylistURL: fmt.Sprintf("%s/%s", s.cdnBaseURL, file.FilePath),
			}
			variants = append(variants, variant)
		}
	}

	playlist := &domain.HLSPlaylist{
		MasterPlaylist: masterPlaylist,
		Variants:       variants,
		BaseURL:        s.cdnBaseURL,
	}

	// Cachear el playlist por 30 minutos
	if err := s.cacheRepo.Set(ctx, cacheKey, playlist, 30*time.Minute); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache HLS playlist: %v\n", err)
	}

	return playlist, nil
}

func (s *streamingService) GetVariantPlaylist(ctx context.Context, videoID uuid.UUID, quality string) (string, error) {
	// Verificar caché primero
	cacheKey := fmt.Sprintf("variant_playlist:%s:%s", videoID.String(), quality)
	var cachedPlaylist string
	if err := s.cacheRepo.Get(ctx, cacheKey, &cachedPlaylist); err == nil {
		return cachedPlaylist, nil
	}

	// Obtener archivos de video
	videoFiles, err := s.videoRepo.GetVideoFiles(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("failed to get video files: %w", err)
	}

	// Encontrar el archivo de la calidad específica
	var targetFile *domain.VideoFile
	for _, file := range videoFiles {
		if file.Quality == quality && file.Format == "hls" {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return "", fmt.Errorf("video file not found for quality %s", quality)
	}

	// Leer el archivo de playlist
	playlistPath := filepath.Join("storage", targetFile.FilePath)
	playlistContent, err := os.ReadFile(playlistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read playlist file: %w", err)
	}

	playlist := string(playlistContent)

	// Cachear el playlist por 30 minutos
	if err := s.cacheRepo.SetString(ctx, cacheKey, playlist, 30*time.Minute); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache variant playlist: %v\n", err)
	}

	return playlist, nil
}

func (s *streamingService) RecordStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) error {
	// Crear o actualizar sesión de streaming
	return s.updateStreamingSession(ctx, userID, videoID, 0)
}

func (s *streamingService) UpdateStreamingStats(ctx context.Context, userID, videoID uuid.UUID, segment string) error {
	// Registrar acceso a segmento para analítica
	statsKey := fmt.Sprintf("streaming_stats:%s:%s", videoID.String(), time.Now().Format("2006-01-02"))

	// Incrementar contador de acceso a segmento
	if _, err := s.cacheRepo.Increment(ctx, statsKey); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to update streaming stats: %v\n", err)
	}

	// Establecer expiración de 30 días
	if err := s.cacheRepo.SetTTL(ctx, statsKey, 30*24*time.Hour); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to set TTL for streaming stats: %v\n", err)
	}

	return nil
}

func (s *streamingService) GetStreamingInfo(ctx context.Context, userID, videoID uuid.UUID) (*domain.StreamingInfo, error) {
	// Obtener información del video
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Verificar acceso
	hasAccess, err := s.CheckVideoAccess(ctx, userID, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to check video access: %w", err)
	}

	// Obtener calidades disponibles
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

	// Obtener progreso de visualización del usuario (si está autenticado)
	var watchProgress *domain.WatchProgress
	if userID != uuid.Nil {
		indexKey := fmt.Sprintf("streaming_session_idx:%s:%s", userID.String(), videoID.String())
		sessionIDStr, err := s.cacheRepo.GetString(ctx, indexKey)
		if err == nil && sessionIDStr != "" {
			sessionKey := fmt.Sprintf("streaming_session:%s", sessionIDStr)
			var session domain.StreamingSession
			if err := s.cacheRepo.Get(ctx, sessionKey, &session); err == nil {
				percentage := float64(session.Position) / float64(video.Duration) * 100
				watchProgress = &domain.WatchProgress{
					Position:    session.Position,
					Percentage:  percentage,
					LastWatched: session.UpdatedAt,
				}
			}
		}
	}

	info := &domain.StreamingInfo{
		VideoID:       videoID,
		Title:         video.Title,
		Duration:      video.Duration,
		HasAccess:     hasAccess,
		Qualities:     qualities,
		WatchProgress: watchProgress,
	}

	if hasAccess {
		// Generar token de streaming
		token, err := s.generateStreamingToken(ctx, userID, videoID, 2*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate streaming token: %w", err)
		}
		info.StreamingToken = token

		// Añadir URL del playlist HLS
		info.PlaylistURL = fmt.Sprintf("/api/stream/%s/playlist", videoID)
	}

	return info, nil
}

func (s *streamingService) UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error {
	// Actualizar sesión de streaming
	if err := s.updateStreamingSession(ctx, userID, videoID, position); err != nil {
		return fmt.Errorf("failed to update streaming session: %w", err)
	}

	// TODO: Almacenar progreso de visualización persistente en la base de datos
	// Esto permitiría a los usuarios reanudar la visualización en diferentes dispositivos

	return nil
}

// Métodos privados auxiliares

func (s *streamingService) generateStreamingToken(ctx context.Context, userID, videoID uuid.UUID, duration time.Duration) (string, error) {
	// Verificar si el usuario tiene acceso
	hasAccess, err := s.CheckVideoAccess(ctx, userID, videoID)
	if err != nil {
		return "", fmt.Errorf("failed to check video access: %w", err)
	}

	if !hasAccess {
		return "", domain.ErrVideoAccess
	}

	// Crear claims de streaming
	claims := &jwt.Claims{
		UserID: userID.String(),
		CustomClaims: map[string]interface{}{
			"video_id": videoID.String(),
			"type":     "streaming",
		},
	}

	// Generar token
	token, err := jwt.GenerateToken(claims, s.jwtSecret, duration)
	if err != nil {
		return "", fmt.Errorf("failed to generate streaming token: %w", err)
	}

	return token, nil
}

func (s *streamingService) updateStreamingSession(ctx context.Context, userID, videoID uuid.UUID, position int) error {
	indexKey := fmt.Sprintf("streaming_session_idx:%s:%s", userID.String(), videoID.String())

	// Obtener sessionID; si no existe, crear sesión
	sessionIDStr, err := s.cacheRepo.GetString(ctx, indexKey)
	if err != nil || sessionIDStr == "" {
		if err := s.createStreamingSession(ctx, userID, videoID, "720p"); err != nil {
			return fmt.Errorf("failed to create streaming session: %w", err)
		}
		// Reintentar obtener el sessionID
		sessionIDStr, err = s.cacheRepo.GetString(ctx, indexKey)
		if err != nil || sessionIDStr == "" {
			return fmt.Errorf("failed to get session id after creation")
		}
	}

	sessionKey := fmt.Sprintf("streaming_session:%s", sessionIDStr)

	var session domain.StreamingSession
	if err := s.cacheRepo.Get(ctx, sessionKey, &session); err != nil {
		return fmt.Errorf("failed to get streaming session: %w", err)
	}

	session.Position = position
	session.UpdatedAt = time.Now()

	if err := s.cacheRepo.Set(ctx, sessionKey, &session, 4*time.Hour); err != nil {
		return fmt.Errorf("failed to update streaming session: %w", err)
	}

	return nil
}

func (s *streamingService) createStreamingSession(ctx context.Context, userID, videoID uuid.UUID, quality string) error {
	// Verificar acceso al video
	hasAccess, err := s.CheckVideoAccess(ctx, userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to check video access: %w", err)
	}

	if !hasAccess {
		return domain.ErrVideoAccess
	}

	// Crear sesión de streaming
	session := &domain.StreamingSession{
		ID:        uuid.New(),
		UserID:    userID,
		VideoID:   videoID,
		Quality:   quality,
		Position:  0,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guardar por sessionID
	sessionKey := fmt.Sprintf("streaming_session:%s", session.ID.String())
	if err := s.cacheRepo.Set(ctx, sessionKey, session, 4*time.Hour); err != nil {
		return fmt.Errorf("failed to store streaming session: %w", err)
	}

	// Índice userID+videoID -> sessionID (para updateStreamingSession)
	indexKey := fmt.Sprintf("streaming_session_idx:%s:%s", userID.String(), videoID.String())
	if err := s.cacheRepo.SetString(ctx, indexKey, session.ID.String(), 4*time.Hour); err != nil {
		// No fallamos la creación por el índice
		fmt.Printf("Warning: failed to index streaming session: %v\n", err)
	}

	return nil
}

func (s *streamingService) buildMasterPlaylist(videoFiles []*domain.VideoFile) string {
	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n")
	playlist.WriteString("#EXT-X-VERSION:3\n\n")

	for _, file := range videoFiles {
		if file.Format == "hls" {
			bandwidth := file.Bitrate * 1000 // Convertir a bps
			resolution := s.getResolutionForQuality(file.Quality)

			playlist.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n",
				bandwidth, resolution))
			playlist.WriteString(fmt.Sprintf("hls/%s/playlist.m3u8\n", file.Quality))
		}
	}

	return playlist.String()
}

func (s *streamingService) getResolutionForQuality(quality string) string {
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
