// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/pkg/httputil"
	"streaming-platform/pkg/logger"
	"streaming-platform/pkg/validator"
)

type VideoHandler struct {
	videoService input.VideoService
	logger       logger.Logger
}

func NewVideoHandler(videoService input.VideoService, logger logger.Logger) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
		logger:       logger,
	}
}

func (h *VideoHandler) GetPublicVideos(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	searchReq := domain.VideoSearchRequest{
		Query:    r.URL.Query().Get("query"),
		Category: r.URL.Query().Get("category"),
		Page:     page,
		Limit:    limit,
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		searchReq.Tags = strings.Split(tags, ",")
	}

	videos, err := h.videoService.SearchPublicVideos(r.Context(), searchReq)
	if err != nil {
		h.logger.Error("Error searching videos: %v", err)
		httputil.WriteInternalError(w, "Error searching videos")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, videos)
}

func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	video, err := h.videoService.GetVideoByID(r.Context(), videoID)
	if err != nil {
		h.logger.Error("Error getting video: %v", err)
		httputil.WriteNotFound(w, "Video not found")
		return
	}

	// Incrementar contador de visualizaciones
	go h.videoService.IncrementViewCount(r.Context(), videoID)

	httputil.WriteJSON(w, http.StatusOK, video)
}

func (h *VideoHandler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	// Obtener user ID del contexto (del middleware JWT)
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB max
	if err != nil {
		httputil.WriteValidationError(w, "Error parsing form")
		return
	}

	// Obtener metadatos del video
	var uploadReq domain.VideoUploadRequest
	metadataJSON := r.FormValue("metadata")
	if metadataJSON == "" {
		httputil.WriteValidationError(w, "Metadata is required")
		return
	}

	// Decodificar metadatos
	if err := httputil.DecodeJSON(r, &uploadReq); err != nil {
		httputil.WriteValidationError(w, "Invalid metadata")
		return
	}

	// Validar metadatos
	if err := validator.ValidateVideoTitle(uploadReq.Title); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidateVideoDescription(uploadReq.Description); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	// Obtener archivo de video
	file, header, err := r.FormFile("video")
	if err != nil {
		httputil.WriteValidationError(w, "Error getting video file")
		return
	}
	defer file.Close()

	// Validar tipo de archivo
	if !isValidVideoFormat(header.Filename) {
		httputil.WriteValidationError(w, "Invalid video format")
		return
	}

	// Crear registro de video en BD
	video := &domain.Video{
		ID:           uuid.New(),
		Title:        uploadReq.Title,
		Description:  uploadReq.Description,
		InstructorID: userID,
		Category:     uploadReq.Category,
		Tags:         uploadReq.Tags,
		Status:       domain.VideoStatusUploading,
		IsPublic:     uploadReq.IsPublic,
	}

	if err := h.videoService.CreateVideo(r.Context(), video); err != nil {
		h.logger.Error("Error creating video: %v", err)
		httputil.WriteInternalError(w, "Error creating video")
		return
	}

	// Guardar archivo temporal
	uploadDir := filepath.Join("uploads", "temp")
	os.MkdirAll(uploadDir, 0755)
	
	tempFilePath := filepath.Join(uploadDir, video.ID.String()+"_"+header.Filename)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		h.logger.Error("Error creating temp file: %v", err)
		httputil.WriteInternalError(w, "Error uploading file")
		return
	}
	defer tempFile.Close()

	// Copiar archivo
	_, err = io.Copy(tempFile, file)
	if err != nil {
		h.logger.Error("Error copying file: %v", err)
		httputil.WriteInternalError(w, "Error uploading file")
		return
	}

	// Enviar trabajo de transcodificación
	job := map[string]interface{}{
		"type":      "transcoding",
		"video_id":  video.ID.String(),
		"file_path": tempFilePath,
	}

	if err := h.videoService.QueueTranscodingJob(r.Context(), job); err != nil {
		h.logger.Error("Error queuing transcoding job: %v", err)
		httputil.WriteInternalError(w, "Error processing video")
		return
	}

	// Actualizar estado
	video.Status = domain.VideoStatusProcessing
	h.videoService.UpdateVideo(r.Context(), video)

	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message":  "Video uploaded successfully",
		"video_id": video.ID,
		"status":   video.Status,
	})
}

func (h *VideoHandler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Obtener video existente
	video, err := h.videoService.GetVideoByID(r.Context(), videoID)
	if err != nil {
		httputil.WriteNotFound(w, "Video not found")
		return
	}

	// Verificar permisos
	if video.InstructorID != userID {
		httputil.WriteForbidden(w, "")
		return
	}

	// Parse request body
	var updateReq domain.VideoUploadRequest
	if err := httputil.DecodeJSON(r, &updateReq); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Actualizar campos
	video.Title = updateReq.Title
	video.Description = updateReq.Description
	video.Category = updateReq.Category
	video.Tags = updateReq.Tags
	video.IsPublic = updateReq.IsPublic

	if err := h.videoService.UpdateVideo(r.Context(), video); err != nil {
		h.logger.Error("Error updating video: %v", err)
		httputil.WriteInternalError(w, "Error updating video")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, video)
}

func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Verificar permisos
	video, err := h.videoService.GetVideoByID(r.Context(), videoID)
	if err != nil {
		httputil.WriteNotFound(w, "Video not found")
		return
	}

	if video.InstructorID != userID {
		httputil.WriteForbidden(w, "")
		return
	}

	if err := h.videoService.DeleteVideo(r.Context(), videoID); err != nil {
		h.logger.Error("Error deleting video: %v", err)
		httputil.WriteInternalError(w, "Error deleting video")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidVideoFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validFormats := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}
	
	for _, format := range validFormats {
		if ext == format {
			return true
		}
	}
	return false
}
