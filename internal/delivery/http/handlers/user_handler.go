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

	"streaming-platform/internal/usecase/user"
	"streaming-platform/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userUsecase user.UserUsecase
	logger      logger.Logger
}

func NewUserHandler(userUsecase user.UserUsecase, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
		logger:      logger,
	}
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Bio       string `json:"bio" validate:"max=500"`
}

type UserStatsResponse struct {
	TotalVideos      int64 `json:"total_videos"`
	TotalViews       int64 `json:"total_views"`
	TotalSubscribers int64 `json:"total_subscribers"`
	TotalWatchTime   int64 `json:"total_watch_time"` // in seconds
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userUsecase.GetUserByID(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting user profile: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "First name and last name are required", http.StatusBadRequest)
		return
	}

	// Trim whitespace
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Bio = strings.TrimSpace(req.Bio)

	// Get current user
	currentUser, err := h.userUsecase.GetUserByID(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting current user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Update user data
	updateData := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"bio":        req.Bio,
	}

	updatedUser, err := h.userUsecase.UpdateUser(r.Context(), userUUID, updateData)
	if err != nil {
		h.logger.Error("Error updating user profile: %v", err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)

	h.logger.Info("User profile updated successfully: %s", currentUser.Email)
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form (max 5MB)
	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageFormat(header.Filename) {
		http.Error(w, "Invalid image format. Allowed: jpg, jpeg, png, gif", http.StatusBadRequest)
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5<<20 {
		http.Error(w, "File too large. Maximum 5MB allowed", http.StatusBadRequest)
		return
	}

	// Create avatars directory
	avatarDir := filepath.Join("storage", "avatars")
	os.MkdirAll(avatarDir, 0755)

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := userUUID.String() + ext
	filePath := filepath.Join(avatarDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("Error creating avatar file: %v", err)
		http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		h.logger.Error("Error saving avatar file: %v", err)
		http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}

	// Update user avatar path
	avatarURL := "/static/avatars/" + filename
	updateData := map[string]interface{}{
		"avatar": avatarURL,
	}

	updatedUser, err := h.userUsecase.UpdateUser(r.Context(), userUUID, updateData)
	if err != nil {
		h.logger.Error("Error updating user avatar: %v", err)
		http.Error(w, "Failed to update avatar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Avatar uploaded successfully",
		"avatar_url": avatarURL,
		"user":       updatedUser,
	})

	h.logger.Info("Avatar uploaded successfully for user: %s", userUUID)
}

func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	stats, err := h.userUsecase.GetUserStats(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting user stats: %v", err)
		http.Error(w, "Failed to get user stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	profile, err := h.userUsecase.GetPublicProfile(r.Context(), userID)
	if err != nil {
		h.logger.Error("Error getting public profile: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *UserHandler) GetWatchHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	history, err := h.userUsecase.GetWatchHistory(r.Context(), userUUID, page, limit)
	if err != nil {
		h.logger.Error("Error getting watch history: %v", err)
		http.Error(w, "Failed to get watch history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *UserHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	favorites, err := h.userUsecase.GetFavorites(r.Context(), userUUID, page, limit)
	if err != nil {
		h.logger.Error("Error getting favorites: %v", err)
		http.Error(w, "Failed to get favorites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

func (h *UserHandler) AddToFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["videoId"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	err = h.userUsecase.AddToFavorites(r.Context(), userUUID, videoID)
	if err != nil {
		h.logger.Error("Error adding to favorites: %v", err)

		switch err.Error() {
		case "video not found":
			http.Error(w, "Video not found", http.StatusNotFound)
		case "already in favorites":
			http.Error(w, "Video already in favorites", http.StatusConflict)
		default:
			http.Error(w, "Failed to add to favorites", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Added to favorites successfully",
	})

	h.logger.Info("Video %s added to favorites for user %s", videoID, userUUID)
}

func (h *UserHandler) RemoveFromFavorites(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["videoId"])
	if err != nil {
		http.Error(w, "Invalid video ID", http.StatusBadRequest)
		return
	}

	err = h.userUsecase.RemoveFromFavorites(r.Context(), userUUID, videoID)
	if err != nil {
		h.logger.Error("Error removing from favorites: %v", err)
		http.Error(w, "Failed to remove from favorites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Removed from favorites successfully",
	})

	h.logger.Info("Video %s removed from favorites for user %s", videoID, userUUID)
}

func (h *UserHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var settings map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.userUsecase.UpdateSettings(r.Context(), userUUID, settings)
	if err != nil {
		h.logger.Error("Error updating user settings: %v", err)
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Settings updated successfully",
	})

	h.logger.Info("Settings updated for user: %s", userUUID)
}

func (h *UserHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	settings, err := h.userUsecase.GetSettings(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting user settings: %v", err)
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *UserHandler) DeactivateAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Password string `json:"password" validate:"required"`
		Reason   string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password confirmation required", http.StatusBadRequest)
		return
	}

	err = h.userUsecase.DeactivateAccount(r.Context(), userUUID, req.Password, req.Reason)
	if err != nil {
		h.logger.Error("Error deactivating account: %v", err)

		switch err.Error() {
		case "invalid password":
			http.Error(w, "Invalid password", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to deactivate account", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account deactivated successfully",
	})

	h.logger.Info("Account deactivated for user: %s", userUUID)
}

// Helper function to validate image format
func isValidImageFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validFormats := []string{".jpg", ".jpeg", ".png", ".gif"}

	for _, format := range validFormats {
		if ext == format {
			return true
		}
	}
	return false
}
