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
	"streaming-platform/pkg/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService input.UserService
	logger      logger.Logger
}

func NewUserHandler(userService input.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

type UpdateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Bio       string `json:"bio"`
}

type UserStatsResponse struct {
	TotalVideos      int64 `json:"total_videos"`
	TotalViews       int64 `json:"total_views"`
	TotalSubscribers int64 `json:"total_subscribers"`
	TotalWatchTime   int64 `json:"total_watch_time"` // in seconds
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting user profile: %v", err)
		httputil.WriteNotFound(w, "User not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	var req UpdateProfileRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validar campos
	if err := validator.ValidateName("firstName", req.FirstName); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidateName("lastName", req.LastName); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidateMaxLength("bio", req.Bio, 500); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	// Normalizar datos
	req.FirstName = validator.NormalizeString(req.FirstName)
	req.LastName = validator.NormalizeString(req.LastName)
	req.Bio = validator.NormalizeString(req.Bio)

	// Get current user
	currentUser, err := h.userService.GetUserByID(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting current user: %v", err)
		httputil.WriteNotFound(w, "User not found")
		return
	}

	// Update user data
	updateData := map[string]interface{}{
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"bio":       req.Bio,
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), userUUID, updateData)
	if err != nil {
		h.logger.Error("Error updating user profile: %v", err)
		httputil.WriteInternalError(w, "Failed to update profile")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, updatedUser)
	h.logger.Info("User profile updated successfully: %s", currentUser.Email)
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	// Parse multipart form (max 5MB)
	err = r.ParseMultipartForm(5 << 20)
	if err != nil {
		httputil.WriteValidationError(w, "File too large")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		httputil.WriteValidationError(w, "No file provided")
		return
	}
	defer file.Close()

	// Validate file type
	if !isValidImageFormat(header.Filename) {
		httputil.WriteValidationError(w, "Invalid image format. Allowed: jpg, jpeg, png, gif")
		return
	}

	// Validate file size (max 5MB)
	if header.Size > 5<<20 {
		httputil.WriteValidationError(w, "File too large. Maximum 5MB allowed")
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
		httputil.WriteInternalError(w, "Failed to save avatar")
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		h.logger.Error("Error saving avatar file: %v", err)
		httputil.WriteInternalError(w, "Failed to save avatar")
		return
	}

	// Update user avatar path
	avatarURL := "/static/avatars/" + filename
	updateData := map[string]interface{}{
		"avatar": avatarURL,
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), userUUID, updateData)
	if err != nil {
		h.logger.Error("Error updating user avatar: %v", err)
		httputil.WriteInternalError(w, "Failed to update avatar")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Avatar uploaded successfully",
		"avatar_url": avatarURL,
		"user":       updatedUser,
	})

	h.logger.Info("Avatar uploaded successfully for user: %s", userUUID)
}

func (h *UserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	// Intentar obtener ID desde la URL primero (para rutas públicas)
	vars := mux.Vars(r)
	var userUUID uuid.UUID
	var err error

	if idStr, ok := vars["id"]; ok {
		// Si hay ID en la URL, usarlo (ruta pública: /api/users/{id}/stats)
		userUUID, err = uuid.Parse(idStr)
		if err != nil {
			httputil.WriteValidationError(w, "Invalid user ID")
			return
		}
	} else {
		// Si no, obtener del contexto (ruta protegida: /api/user/stats)
		userUUID, err = h.getUserIDFromContext(r)
		if err != nil {
			httputil.WriteUnauthorized(w, "")
			return
		}
	}

	stats, err := h.userService.GetUserStats(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting user stats: %v", err)
		httputil.WriteInternalError(w, "Failed to get user stats")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, stats)
}

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["id"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid user ID")
		return
	}

	profile, err := h.userService.GetPublicProfile(r.Context(), userID)
	if err != nil {
		h.logger.Error("Error getting public profile: %v", err)
		httputil.WriteNotFound(w, "User not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, profile)
}

func (h *UserHandler) GetWatchHistory(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	// Parse pagination parameters
	page, limit := h.getPaginationParams(r)

	history, totalCount, err := h.userService.GetWatchHistory(r.Context(), userUUID, page, limit)
	if err != nil {
		h.logger.Error("Error getting watch history: %v", err)
		httputil.WriteInternalError(w, "Failed to get watch history")
		return
	}

	response := map[string]interface{}{
		"data":        history,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	// Parse pagination parameters
	page, limit := h.getPaginationParams(r)

	favorites, totalCount, err := h.userService.GetFavorites(r.Context(), userUUID, page, limit)
	if err != nil {
		h.logger.Error("Error getting favorites: %v", err)
		httputil.WriteInternalError(w, "Failed to get favorites")
		return
	}

	response := map[string]interface{}{
		"data":        favorites,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	}

	httputil.WriteJSON(w, http.StatusOK, response)
}

func (h *UserHandler) AddToFavorites(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["videoId"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	err = h.userService.AddToFavorites(r.Context(), userUUID, videoID)
	if err != nil {
		h.logger.Error("Error adding to favorites: %v", err)
		h.handleFavoritesError(w, err)
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Added to favorites successfully")
	h.logger.Info("Video %s added to favorites for user %s", videoID, userUUID)
}

func (h *UserHandler) RemoveFromFavorites(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["videoId"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	err = h.userService.RemoveFromFavorites(r.Context(), userUUID, videoID)
	if err != nil {
		h.logger.Error("Error removing from favorites: %v", err)
		httputil.WriteInternalError(w, "Failed to remove from favorites")
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Removed from favorites successfully")
	h.logger.Info("Video %s removed from favorites for user %s", videoID, userUUID)
}

func (h *UserHandler) UpdateWatchProgress(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	vars := mux.Vars(r)
	videoID, err := uuid.Parse(vars["videoId"])
	if err != nil {
		httputil.WriteValidationError(w, "Invalid video ID")
		return
	}

	var req struct {
		Position int    `json:"position"` // Posición en segundos
		Quality  string `json:"quality"`
	}

	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validar posición
	if req.Position < 0 {
		httputil.WriteValidationError(w, "Invalid position")
		return
	}

	err = h.userService.UpdateWatchProgress(r.Context(), userUUID, videoID, req.Position, req.Quality)
	if err != nil {
		h.logger.Error("Error updating watch progress: %v", err)
		httputil.WriteInternalError(w, "Failed to update watch progress")
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Watch progress updated successfully")
}

func (h *UserHandler) GetContinueWatching(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	// Obtener límite (por defecto 10)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Obtener videos para continuar viendo
	history, err := h.userService.GetContinueWatching(r.Context(), userUUID, limit)
	if err != nil {
		h.logger.Error("Error getting continue watching: %v", err)
		httputil.WriteInternalError(w, "Failed to get continue watching")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": history,
	})
}

func (h *UserHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	var settings map[string]interface{}
	if err := httputil.DecodeJSON(r, &settings); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	err = h.userService.UpdateSettings(r.Context(), userUUID, settings)
	if err != nil {
		h.logger.Error("Error updating user settings: %v", err)
		httputil.WriteInternalError(w, "Failed to update settings")
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Settings updated successfully")
	h.logger.Info("Settings updated for user: %s", userUUID)
}

func (h *UserHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	settings, err := h.userService.GetSettings(r.Context(), userUUID)
	if err != nil {
		h.logger.Error("Error getting user settings: %v", err)
		httputil.WriteInternalError(w, "Failed to get settings")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, settings)
}

func (h *UserHandler) DeactivateAccount(w http.ResponseWriter, r *http.Request) {
	userUUID, err := h.getUserIDFromContext(r)
	if err != nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	var req struct {
		Password string `json:"password"`
		Reason   string `json:"reason"`
	}

	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	if err := validator.ValidateRequired("password", req.Password); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	err = h.userService.DeactivateAccount(r.Context(), userUUID, req.Password, req.Reason)
	if err != nil {
		h.logger.Error("Error deactivating account: %v", err)

		switch err.Error() {
		case "invalid password":
			httputil.WriteValidationError(w, "Invalid password")
		default:
			httputil.WriteInternalError(w, "Failed to deactivate account")
		}
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Account deactivated successfully")
	h.logger.Info("Account deactivated for user: %s", userUUID)
}

// Helper methods

func (h *UserHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		return uuid.Nil, fmt.Errorf("unauthorized")
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		return uuid.Nil, err
	}

	return userUUID, nil
}

func (h *UserHandler) getPaginationParams(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	return page, limit
}

func (h *UserHandler) handleFavoritesError(w http.ResponseWriter, err error) {
	errMsg := err.Error()

	switch errMsg {
	case "video not found":
		httputil.WriteNotFound(w, "Video not found")
	case "already in favorites":
		httputil.WriteError(w, http.StatusConflict, "Video already in favorites")
	default:
		httputil.WriteInternalError(w, "Failed to add to favorites")
	}
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
