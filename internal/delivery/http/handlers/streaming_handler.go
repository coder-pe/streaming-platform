// internal/delivery/http/handlers/streaming_handler.go
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"streaming-platform/internal/usecase/streaming"
	"streaming-platform/pkg/logger"
)

type StreamingHandler struct {
	streamingUsecase streaming.StreamingUsecase
	logger           logger.Logger
}

func NewStreamingHandler(streamingUsecase streaming.StreamingUsecase, logger logger.Logger) *StreamingHandler {
	return &StreamingHandler{
		streamingUsecase: streamingUsecase,
		logger:           logger,
	}
}

func (h *StreamingHandler) GetHLSPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	// Verificar acceso al video
	userID := h.getUserIDFromContext(r)
	hasAccess, err := h.streamingUsecase.CheckVideoAccess(r.Context(), userID, videoID)
	if err != nil {
		h.logger.Error("Error checking video access: %v", err)
		http.Error(w, "Error checking access", http.StatusInternalServerError)
		return
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Obtener playlist HLS
	playlist, err := h.streamingUsecase.GetHLSPlaylist(r.Context(), videoID)
	if err != nil {
		h.logger.Error("Error getting HLS playlist: %v", err)
		http.Error(w, "Playlist not found", http.StatusNotFound)
		return
	}

	// Determinar si es master playlist o variant playlist
	quality := r.URL.Query().Get("quality")
	if quality == "" {
		// Devolver master playlist
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutos
		w.Write([]byte(playlist.MasterPlaylist))
		return
	}

	// Devolver variant playlist específico
	variantPlaylist, err := h.streamingUsecase.GetVariantPlaylist(r.Context(), videoID, quality)
	if err != nil {
		h.logger.Error("Error getting variant playlist: %v", err)
		http.Error(w, "Variant playlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Write([]byte(variantPlaylist))

	// Registrar sesión de streaming
	go h.streamingUsecase.RecordStreamingSession(r.Context(), userID, videoID, quality)
}

func (h *StreamingHandler) GetHLSSegment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	segment := vars["segment"]
	quality := r.URL.Query().Get("quality")
	
	if quality == "" {
		quality = "720p" // default quality
	}

	// Verificar acceso
	userID := h.getUserIDFromContext(r)
	hasAccess, err := h.streamingUsecase.CheckVideoAccess(r.Context(), userID, videoID)
	if err != nil {
		h.logger.Error("Error checking video access: %v", err)
		http.Error(w, "Error checking access", http.StatusInternalServerError)
		return
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Construir ruta del segmento
	segmentPath := filepath.Join("storage", "videos", videoID.String(), "hls", quality, segment)
	
	// Verificar que el archivo existe
	if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}

	// Servir segmento
	file, err := os.Open(segmentPath)
	if err != nil {
		h.logger.Error("Error opening segment file: %v", err)
		http.Error(w, "Error serving segment", http.StatusInternalServerError)
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
	go h.streamingUsecase.UpdateStreamingStats(r.Context(), userID, videoID, segment)
}

func (h *StreamingHandler) GetStreamingInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserIDFromContext(r)
	
	streamingInfo, err := h.streamingUsecase.GetStreamingInfo(r.Context(), userID, videoID)
	if err != nil {
		h.logger.Error("Error getting streaming info: %v", err)
		http.Error(w, "Error getting streaming info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamingInfo)
}

func (h *StreamingHandler) UpdateWatchProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	userID := h.getUserIDFromContext(r)

	var progressData struct {
		Position int    `json:"position"` // en segundos
		Quality  string `json:"quality"`
	}

	if err := json.NewDecoder(r.Body).Decode(&progressData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.streamingUsecase.UpdateWatchProgress(r.Context(), userID, videoID, progressData.Position, progressData.Quality)
	if err != nil {
		h.logger.Error("Error updating watch progress: %v", err)
		http.Error(w, "Error updating progress", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *StreamingHandler) serveRange(w http.ResponseWriter, r *http.Request, file *os.File, rangeHeader string) {
	// Obtener información del archivo
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}

	fileSize := stat.Size()
	
	// Parsear range header
	ranges := strings.Split(strings.TrimPrefix(rangeHeader, "bytes="), "-")
	if len(ranges) != 2 {
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
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
		http.Error(w, "Invalid range", http.StatusRequestedRangeNotSatisfiable)
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
		return userID.(uuid.UUID)
	}
	return uuid.Nil
}

func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

