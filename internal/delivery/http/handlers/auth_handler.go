package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/usecase/auth"
	"streaming-platform/pkg/logger"
)

type AuthHandler struct {
	authUsecase auth.AuthUsecase
	logger      logger.Logger
}

func NewAuthHandler(authUsecase auth.AuthUsecase, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		logger:      logger,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	Token string               `json:"token"`
	User  entities.UserProfile `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar campos requeridos
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Normalizar email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Autenticar usuario
	token, refreshToken, user, err := h.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Login failed for email %s: %v", req.Email, err)

		switch err.Error() {
		case "invalid credentials":
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		case "user not found":
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		case "user inactive":
			http.Error(w, "Account is inactive", http.StatusForbidden)
		default:
			http.Error(w, "Authentication failed", http.StatusInternalServerError)
		}
		return
	}

	// Set refresh token as httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	response := AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("User logged in successfully: %s", user.Email)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validar campos requeridos
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Validar formato de email
	if !isValidEmail(req.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validar contraseña
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Normalizar datos
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	// Crear usuario
	user := &entities.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "student", // Default role
		IsActive:  true,
	}

	token, refreshToken, userProfile, err := h.authUsecase.Register(r.Context(), user, req.Password)
	if err != nil {
		h.logger.Error("Registration failed for email %s: %v", req.Email, err)

		switch err.Error() {
		case "email already exists":
			http.Error(w, "Email already registered", http.StatusConflict)
		case "weak password":
			http.Error(w, "Password is too weak", http.StatusBadRequest)
		default:
			http.Error(w, "Registration failed", http.StatusInternalServerError)
		}
		return
	}

	// Set refresh token as httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth",
		MaxAge:   7 * 24 * 60 * 60,
	})

	response := AuthResponse{
		Token: token,
		User:  *userProfile,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("User registered successfully: %s", userProfile.Email)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		// Try to get from request body as fallback
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Refresh token required", http.StatusBadRequest)
			return
		}

		if req.RefreshToken == "" {
			http.Error(w, "Refresh token required", http.StatusBadRequest)
			return
		}

		cookie = &http.Cookie{Value: req.RefreshToken}
	}

	// Refresh the token
	newToken, newRefreshToken, user, err := h.authUsecase.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		h.logger.Error("Token refresh failed: %v", err)

		// Clear the invalid cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/api/auth",
			MaxAge:   -1,
		})

		switch err.Error() {
		case "invalid refresh token":
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		case "refresh token expired":
			http.Error(w, "Refresh token expired", http.StatusUnauthorized)
		default:
			http.Error(w, "Token refresh failed", http.StatusInternalServerError)
		}
		return
	}

	// Set new refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth",
		MaxAge:   7 * 24 * 60 * 60,
	})

	response := AuthResponse{
		Token: newToken,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Token refreshed successfully for user: %s", user.Email)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err == nil && cookie.Value != "" {
		// Invalidate refresh token
		if err := h.authUsecase.Logout(r.Context(), cookie.Value); err != nil {
			h.logger.Error("Logout error: %v", err)
		}
	}

	// Clear the refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})

	h.logger.Info("User logged out successfully")
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	err := h.authUsecase.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("Forgot password failed for email %s: %v", req.Email, err)
		// Don't reveal if email exists or not for security
	}

	// Always return success to prevent email enumeration
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})

	h.logger.Info("Password reset requested for email: %s", req.Email)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		http.Error(w, "Token and new password are required", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	err := h.authUsecase.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		h.logger.Error("Password reset failed: %v", err)

		switch err.Error() {
		case "invalid token":
			http.Error(w, "Invalid or expired reset token", http.StatusBadRequest)
		case "token expired":
			http.Error(w, "Reset token has expired", http.StatusBadRequest)
		default:
			http.Error(w, "Password reset failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully",
	})

	h.logger.Info("Password reset completed successfully")
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Current password and new password are required", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "New password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	err := h.authUsecase.ChangePassword(r.Context(), userID.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.logger.Error("Password change failed for user %s: %v", userID, err)

		switch err.Error() {
		case "invalid current password":
			http.Error(w, "Current password is incorrect", http.StatusBadRequest)
		case "weak password":
			http.Error(w, "New password is too weak", http.StatusBadRequest)
		default:
			http.Error(w, "Password change failed", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})

	h.logger.Info("Password changed successfully for user: %s", userID)
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Simple email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) >= 5
}
