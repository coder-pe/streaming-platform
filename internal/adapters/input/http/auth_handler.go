// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package handlers

import (
	"net/http"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/pkg/httputil"
	"streaming-platform/pkg/logger"
	"streaming-platform/pkg/validator"

	"github.com/google/uuid"
)

type AuthHandler struct {
	authService input.AuthService
	logger      logger.Logger
}

func NewAuthHandler(authService input.AuthService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type AuthResponse struct {
	Token string             `json:"token"`
	User  domain.UserProfile `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Error("Invalid request body: %v", err)
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validar campos
	if err := validator.ValidateRequired("email", req.Email); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidateEmail(req.Email); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidateRequired("password", req.Password); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	// Normalizar email
	req.Email = validator.NormalizeEmail(req.Email)

	// Autenticar usuario
	token, refreshToken, user, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error("Login failed for email %s: %v", req.Email, err)
		h.handleAuthError(w, err)
		return
	}

	// Set refresh token cookie
	httputil.SetRefreshTokenCookie(w, refreshToken)

	// Responder
	h.respondWithAuth(w, http.StatusOK, token, user)
	h.logger.Info("User logged in successfully: %s", user.Email)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Error("Invalid request body: %v", err)
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validar todos los campos
	if err := h.validateRegisterRequest(&req); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	// Normalizar datos
	req.Email = validator.NormalizeEmail(req.Email)
	req.FirstName = validator.NormalizeString(req.FirstName)
	req.LastName = validator.NormalizeString(req.LastName)

	// Registrar usuario
	token, refreshToken, userProfile, err := h.authService.Register(
		r.Context(),
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		"student", // Default role
	)
	if err != nil {
		h.logger.Error("Registration failed for email %s: %v", req.Email, err)
		h.handleAuthError(w, err)
		return
	}

	// Set refresh token cookie
	httputil.SetRefreshTokenCookie(w, refreshToken)

	// Responder
	h.respondWithAuth(w, http.StatusCreated, token, userProfile)
	h.logger.Info("User registered successfully: %s", userProfile.Email)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Intentar obtener refresh token de cookie primero
	refreshToken, err := httputil.GetRefreshTokenFromCookie(r)
	if err != nil {
		// Fallback: intentar obtener del body
		var req RefreshRequest
		if err := httputil.DecodeJSON(r, &req); err != nil || req.RefreshToken == "" {
			httputil.WriteValidationError(w, "Refresh token required")
			return
		}
		refreshToken = req.RefreshToken
	}

	// Refrescar el token
	newToken, newRefreshToken, user, err := h.authService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		h.logger.Error("Token refresh failed: %v", err)

		// Limpiar cookie inválida
		httputil.ClearRefreshTokenCookie(w)
		h.handleAuthError(w, err)
		return
	}

	// Set nuevo refresh token cookie
	httputil.SetRefreshTokenCookie(w, newRefreshToken)

	// Responder
	h.respondWithAuth(w, http.StatusOK, newToken, user)
	h.logger.Info("Token refreshed successfully for user: %s", user.Email)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get userID del contexto
	userID := r.Context().Value("user_id")

	// Get refresh token de cookie
	refreshToken, err := httputil.GetRefreshTokenFromCookie(r)
	if err == nil && refreshToken != "" && userID != nil {
		// Invalidar refresh token
		if err := h.authService.Logout(r.Context(), userID.(uuid.UUID), refreshToken); err != nil {
			h.logger.Error("Logout error: %v", err)
		}
	}

	// Limpiar cookie
	httputil.ClearRefreshTokenCookie(w)

	// Responder
	httputil.WriteSuccessMessage(w, http.StatusOK, "Logged out successfully")
	h.logger.Info("User logged out successfully")
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	if err := validator.ValidateEmail(req.Email); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	req.Email = validator.NormalizeEmail(req.Email)

	err := h.authService.ForgotPassword(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("Forgot password failed for email %s: %v", req.Email, err)
		// Don't reveal if email exists or not for security
	}

	// Always return success to prevent email enumeration
	httputil.WriteSuccessMessage(w, http.StatusOK, "If the email exists, a password reset link has been sent")
	h.logger.Info("Password reset requested for email: %s", req.Email)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validar campos
	if err := validator.ValidateRequired("token", req.Token); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidatePasswordSimple(req.NewPassword); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		h.logger.Error("Password reset failed: %v", err)
		h.handleAuthError(w, err)
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Password reset successfully")
	h.logger.Info("Password reset completed successfully")
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		httputil.WriteUnauthorized(w, "")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteValidationError(w, "Invalid request body")
		return
	}

	// Validar campos
	if err := validator.ValidateRequired("current_password", req.CurrentPassword); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}
	if err := validator.ValidatePasswordSimple(req.NewPassword); err != nil {
		httputil.WriteValidationError(w, err.Error())
		return
	}

	err := h.authService.ChangePassword(r.Context(), userID.(uuid.UUID), req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.logger.Error("Password change failed for user %s: %v", userID, err)
		h.handleAuthError(w, err)
		return
	}

	httputil.WriteSuccessMessage(w, http.StatusOK, "Password changed successfully")
	h.logger.Info("Password changed successfully for user: %s", userID)
}

// Helper methods

func (h *AuthHandler) validateRegisterRequest(req *RegisterRequest) error {
	if err := validator.ValidateName("firstName", req.FirstName); err != nil {
		return err
	}
	if err := validator.ValidateName("lastName", req.LastName); err != nil {
		return err
	}
	if err := validator.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := validator.ValidatePasswordSimple(req.Password); err != nil {
		return err
	}
	return nil
}

func (h *AuthHandler) respondWithAuth(w http.ResponseWriter, status int, token string, user *domain.UserProfile) {
	response := AuthResponse{
		Token: token,
		User:  *user,
	}
	httputil.WriteJSON(w, status, response)
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	errMsg := err.Error()

	switch errMsg {
	case "invalid credentials", "user not found":
		httputil.WriteUnauthorized(w, "Invalid email or password")
	case "user inactive":
		httputil.WriteForbidden(w, "Account is inactive")
	case "email already exists":
		httputil.WriteError(w, http.StatusConflict, "Email already registered")
	case "weak password":
		httputil.WriteValidationError(w, "Password is too weak")
	case "invalid refresh token", "refresh token expired":
		httputil.WriteUnauthorized(w, "Invalid or expired refresh token")
	case "invalid token", "token expired":
		httputil.WriteValidationError(w, "Invalid or expired reset token")
	case "invalid current password":
		httputil.WriteValidationError(w, "Current password is incorrect")
	default:
		httputil.WriteInternalError(w, "Operation failed")
	}
}
