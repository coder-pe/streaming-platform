// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/usecase/video"
	"streaming-platform/pkg/logger"
)

type VideoHandler struct {
	videoUsecase video.VideoUsecase
	logger       logger.Logger
}

func NewVideoHandler(videoUsecase video.VideoUsecase, logger logger.Logger) *VideoHandler {
	return &VideoHandler{
		videoUsecase: videoUsecase,
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

	searchReq := entities.VideoSearchRequest{
		Query:    r.URL.Query().Get("query"),
		Category: r.URL.Query().Get("category"),
		Page:     page,
		Limit:    limit,
	}

	if tags := r.URL.Query().Get("tags"); tags != "" {
		searchReq.Tags = strings.Split(tags, ",")
	}

	videos, err := h.videoUsecase.SearchPublicVideos(r.Context(), searchReq)
	if err != nil {
		h.logger.Error("Error searching videos: %v", err)
		http.Error(w, "Error searching videos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	video, err := h.videoUsecase.GetVideoByID(r.Context(), videoID)
	if err != nil {
		h.logger.Error("Error getting video: %v", err)
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// Incrementar contador de visualizaciones
	go h.videoUsecase.IncrementViewCount(r.Context(), videoID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(video)
}

func (h *VideoHandler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	// Obtener user ID del contexto (del middleware JWT)
	userID := r.Context().Value("user_id").(uuid.UUID)

	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB max
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Obtener metadatos del video
	var uploadReq entities.VideoUploadRequest
	if err := json.Unmarshal([]byte(r.FormValue("metadata")), &uploadReq); err != nil {
		http.Error(w, "Invalid metadata", http.StatusBadRequest)
		return
	}

	// Validar metadatos
	if uploadReq.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Obtener archivo de video
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Error getting video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validar tipo de archivo
	if !isValidVideoFormat(header.Filename) {
		http.Error(w, "Invalid video format", http.StatusBadRequest)
		return
	}

	// Crear registro de video en BD
	video := &entities.Video{
		ID:           uuid.New(),
		Title:        uploadReq.Title,
		Description:  uploadReq.Description,
		InstructorID: userID,
		Category:     uploadReq.Category,
		Tags:         uploadReq.Tags,
		Status:       entities.VideoStatusUploading,
		IsPublic:     uploadReq.IsPublic,
	}

	if err := h.videoUsecase.CreateVideo(r.Context(), video); err != nil {
		h.logger.Error("Error creating video: %v", err)
		http.Error(w, "Error creating video", http.StatusInternalServerError)
		return
	}

	// Guardar archivo temporal
	uploadDir := filepath.Join("uploads", "temp")
	os.MkdirAll(uploadDir, 0755)
	
	tempFilePath := filepath.Join(uploadDir, video.ID.String()+"_"+header.Filename)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		h.logger.Error("Error creating temp file: %v", err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Copiar archivo
	_, err = io.Copy(tempFile, file)
	if err != nil {
		h.logger.Error("Error copying file: %v", err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	// Enviar trabajo de transcodificación
	job := map[string]interface{}{
		"type":      "transcoding",
		"video_id":  video.ID.String(),
		"file_path": tempFilePath,
	}

	if err := h.videoUsecase.QueueTranscodingJob(r.Context(), job); err != nil {
		h.logger.Error("Error queuing transcoding job: %v", err)
		http.Error(w, "Error processing video", http.StatusInternalServerError)
		return
	}

	// Actualizar estado
	video.Status = entities.VideoStatusProcessing
	h.videoUsecase.UpdateVideo(r.Context(), video)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Video uploaded successfully",
		"video_id": video.ID,
		"status":   video.Status,
	})
}

func (h *VideoHandler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Obtener video existente
	video, err := h.videoUsecase.GetVideoByID(r.Context(), videoID)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// Verificar permisos
	if video.InstructorID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Parse request body
	var updateReq entities.VideoUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Actualizar campos
	video.Title = updateReq.Title
	video.Description = updateReq.Description
	video.Category = updateReq.Category
	video.Tags = updateReq.Tags
	video.IsPublic = updateReq.IsPublic

	if err := h.videoUsecase.UpdateVideo(r.Context(), video); err != nil {
		h.logger.Error("Error updating video: %v", err)
		http.Error(w, "Error updating video", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(video)
}

func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)

	// Verificar permisos
	video, err := h.videoUsecase.GetVideoByID(r.Context(), videoID)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	if video.InstructorID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if err := h.videoUsecase.DeleteVideo(r.Context(), videoID); err != nil {
		h.logger.Error("Error deleting video: %v", err)
		http.Error(w, "Error deleting video", http.StatusInternalServerError)
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
