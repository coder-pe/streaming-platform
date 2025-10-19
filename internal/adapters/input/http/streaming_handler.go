// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"streaming-platform/internal/core/ports/input"
	"streaming-platform/pkg/httputil"
	"streaming-platform/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type StreamingHandler struct {
	streamingService input.StreamingService
	logger           logger.Logger
}

func NewStreamingHandler(streamingService input.StreamingService, logger logger.Logger) *StreamingHandler {
	return &StreamingHandler{
		streamingService: streamingService,
		logger:           logger,
	}
}

func (h *StreamingHandler) GetMasterPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	// Verificar acceso al video
	userID := h.getUserIDFromContext(r)
	if !h.checkVideoAccess(w, r, userID, videoID) {
		return
	}

	// Obtener playlist HLS
	playlist, err := h.streamingService.GetHLSPlaylist(r.Context(), videoID)
	if err != nil {
		h.logger.Error("Error getting HLS playlist: %v", err)
		httputil.WriteNotFound(w, "Playlist not found")
		return
	}

	// Devolver master playlist
	h.writePlaylist(w, playlist.MasterPlaylist)

	// Registrar sesión de streaming
	go h.streamingService.RecordStreamingSession(r.Context(), userID, videoID, "auto")
}

func (h *StreamingHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	quality := vars["quality"]
	if quality == "" {
		quality = "720p" // default quality
	}

	// Verificar acceso
	userID := h.getUserIDFromContext(r)
	if !h.checkVideoAccess(w, r, userID, videoID) {
		return
	}

	// Devolver variant playlist específico
	variantPlaylist, err := h.streamingService.GetVariantPlaylist(r.Context(), videoID, quality)
	if err != nil {
		h.logger.Error("Error getting variant playlist: %v", err)
		httputil.WriteNotFound(w, "Variant playlist not found")
		return
	}

	h.writePlaylist(w, variantPlaylist)

	// Registrar sesión de streaming
	go h.streamingService.RecordStreamingSession(r.Context(), userID, videoID, quality)
}

func (h *StreamingHandler) GetSegment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	segment := vars["segment"]
	quality := vars["quality"]

	if quality == "" {
		quality = "720p" // default quality
	}

	// Verificar acceso
	userID := h.getUserIDFromContext(r)
	if !h.checkVideoAccess(w, r, userID, videoID) {
		return
	}

	// Construir ruta del segmento
	segmentPath := filepath.Join("storage", "videos", videoID.String(), "hls", quality, segment)

	// Verificar que el archivo existe
	if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
		httputil.WriteNotFound(w, "Segment not found")
		return
	}

	// Servir segmento
	file, err := os.Open(segmentPath)
	if err != nil {
		h.logger.Error("Error opening segment file: %v", err)
		httputil.WriteInternalError(w, "Error serving segment")
		return
	}
	defer file.Close()

	// Configurar headers para streaming
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "public, max-age=86400") // 24 horas
	w.Header().Set("Accept-Ranges", "bytes")

	// Soporte para range requests
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		h.serveRange(w, r, file, rangeHeader)
		return
	}

	// Servir archivo completo
	_, err = io.Copy(w, file)
	if err != nil {
		h.logger.Error("Error serving segment: %v", err)
	}

	// Actualizar estadísticas de streaming
	go h.streamingService.UpdateStreamingStats(r.Context(), userID, videoID, segment)
}

func (h *StreamingHandler) GetStreamingInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	userID := h.getUserIDFromContext(r)

	streamingInfo, err := h.streamingService.GetStreamingInfo(r.Context(), userID, videoID)
	if err != nil {
		h.logger.Error("Error getting streaming info: %v", err)
		httputil.WriteInternalError(w, "Error getting streaming info")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, streamingInfo)
}

func (h *StreamingHandler) UpdateWatchProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	userID := h.getUserIDFromContext(r)

	var progressData struct {
		Position int    `json:"position"` // en segundos
		Quality  string `json:"quality"`
	}

	if err := httputil.DecodeJSON(r, &progressData); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	err = h.streamingService.UpdateWatchProgress(r.Context(), userID, videoID, progressData.Position, progressData.Quality)
	if err != nil {
		h.logger.Error("Error updating watch progress: %v", err)
		httputil.WriteInternalError(w, "Error updating progress")
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Progress updated successfully")
}

// Helper methods

func (h *StreamingHandler) checkVideoAccess(w http.ResponseWriter, r *http.Request, userID, videoID uuid.UUID) bool {
	hasAccess, err := h.streamingService.CheckVideoAccess(r.Context(), userID, videoID)
	if err != nil {
		h.logger.Error("Error checking video access: %v", err)
		httputil.WriteInternalError(w, "Error checking access")
		return false
	}

	if !hasAccess {
		httputil.WriteForbidden(w, "Access denied")
		return false
	}

	return true
}

func (h *StreamingHandler) writePlaylist(w http.ResponseWriter, playlist string) {
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutos
	w.Write([]byte(playlist))
}

func (h *StreamingHandler) serveRange(w http.ResponseWriter, r *http.Request, file *os.File, rangeHeader string) {
	// Obtener información del archivo
	stat, err := file.Stat()
	if err != nil {
		httputil.WriteInternalError(w, "Error getting file info")
		return
	}

	fileSize := stat.Size()

	// Parsear range header
	ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(ranges) != 2 {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	var start, end int64

	if ranges[0] != "" {
		start = parseInt64(ranges[0])
	}

	if ranges[1] != "" {
		end = parseInt64(ranges[1])
	} else {
		end = fileSize - 1
	}

	if start > end || start >= fileSize {
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Configurar headers para partial content
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	w.Header().Set("Content-Type", "video/mp2t")
	w.WriteHeader(http.StatusPartialContent)

	// Buscar posición y copiar rango
	file.Seek(start, 0)
	io.CopyN(w, file, end-start+1)
}

func (h *StreamingHandler) getUserIDFromContext(r *http.Request) uuid.UUID {
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			return uid
		}
		// Try to parse if it's a string
		if uidStr, ok := userID.(string); ok {
			if parsed, err := uuid.Parse(uidStr); err == nil {
				return parsed
			}
		}
	}
	return uuid.Nil
}

func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}
