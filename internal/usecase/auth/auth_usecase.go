package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/domain/errors"
	"streaming-platform/internal/domain/interfaces"
	"streaming-platform/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo  interfaces.UserRepository
	cacheRepo interfaces.CacheRepository
	jwtSecret string
}

func NewAuthUsecase(userRepo interfaces.UserRepository, cacheRepo interfaces.CacheRepository, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		jwtSecret: jwtSecret,
	}
}

func (a *AuthUsecase) Login(ctx context.Context, email, password string) (string, string, *entities.UserProfile, error) {
	// Get user by email
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == errors.UserNotFound {
			return "", "", nil, errors.AuthInvalidCredentials
		}
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return "", "", nil, errors.UserInactive
	}

	// Verify password
	if !a.verifyPassword(password, user.PasswordHash) {
		return "", "", nil, errors.AuthInvalidCredentials
	}

	// Generate tokens
	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := a.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login
	if err := a.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to update last login for user %s: %v\n", user.ID, err)
	}

	// Create user profile
	profile := &entities.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Cache user profile
	if err := a.cacheRepo.CacheUser(ctx, user.ID.String(), profile); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to cache user profile: %v\n", err)
	}

	return accessToken, refreshToken, profile, nil
}

func (a *AuthUsecase) Register(ctx context.Context, user *entities.User, password string) (string, string, *entities.UserProfile, error) {
	// Validate password strength
	if err := a.validatePassword(password); err != nil {
		return "", "", nil, err
	}

	// Check if user already exists
	existingUser, err := a.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return "", "", nil, errors.UserAlreadyExists
	}

	// Hash password
	hashedPassword, err := a.hashPassword(password)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set user fields
	user.PasswordHash = hashedPassword
	user.IsActive = true
	if user.Role == "" {
		user.Role = "student" // Default role
	}

	// Create user
	if err := a.userRepo.Create(ctx, user); err != nil {
		return "", "", nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := a.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create user profile
	profile := &entities.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Cache user profile
	if err := a.cacheRepo.CacheUser(ctx, user.ID.String(), profile); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Warning: failed to cache user profile: %v\n", err)
	}

	return accessToken, refreshToken, profile, nil
}

func (a *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, *entities.UserProfile, error) {
	// Validate refresh token
	userID, err := a.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", nil, err
	}

	// Get user
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == errors.UserNotFound {
			return "", "", nil, errors.AuthTokenInvalid
		}
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return "", "", nil, errors.UserInactive
	}

	// Generate new tokens
	newAccessToken, err := a.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := a.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Invalidate old refresh token
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	if err := a.cacheRepo.Delete(ctx, refreshKey); err != nil {
		// Log error but don't fail refresh
		fmt.Printf("Warning: failed to invalidate old refresh token: %v\n", err)
	}

	// Create user profile
	profile := &entities.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	return newAccessToken, newRefreshToken, profile, nil
}

func (a *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	// Invalidate refresh token
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	if err := a.cacheRepo.Delete(ctx, refreshKey); err != nil {
		return fmt.Errorf("failed to invalidate refresh token: %w", err)
	}

	return nil
}

func (a *AuthUsecase) ForgotPassword(ctx context.Context, email string) error {
	// Get user by email
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == errors.UserNotFound {
			// Don't reveal if user exists for security
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Generate reset token
	resetToken, err := a.generateResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Store reset token in cache with 1 hour expiration
	resetKey := fmt.Sprintf("password_reset:%s", resetToken)
	if err := a.cacheRepo.SetString(ctx, resetKey, user.ID.String(), time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// TODO: Send email with reset link
	// This would integrate with your email service
	fmt.Printf("Password reset token for %s: %s\n", email, resetToken)

	return nil
}

func (a *AuthUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validate new password
	if err := a.validatePassword(newPassword); err != nil {
		return err
	}

	// Get user ID from reset token
	resetKey := fmt.Sprintf("password_reset:%s", token)
	userIDStr, err := a.cacheRepo.GetString(ctx, resetKey)
	if err != nil {
		return errors.AuthTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.AuthTokenInvalid
	}

	// Hash new password
	hashedPassword, err := a.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := a.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate reset token
	if err := a.cacheRepo.Delete(ctx, resetKey); err != nil {
		// Log error but don't fail reset
		fmt.Printf("Warning: failed to invalidate reset token: %v\n", err)
	}

	// Invalidate user cache
	if err := a.cacheRepo.InvalidateUserCache(ctx, userID.String()); err != nil {
		// Log error but don't fail reset
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (a *AuthUsecase) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.AuthTokenInvalid
	}

	// Get user
	user, err := a.userRepo.GetByID(ctx, uid)
	if err != nil {
		if err == errors.UserNotFound {
			return errors.AuthUnauthorized
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if !a.verifyPassword(currentPassword, user.PasswordHash) {
		return errors.AuthInvalidCredentials.WithContext("field", "current_password")
	}

	// Validate new password
	if err := a.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := a.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := a.userRepo.UpdatePassword(ctx, uid, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate user cache
	if err := a.cacheRepo.InvalidateUserCache(ctx, uid.String()); err != nil {
		// Log error but don't fail change
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

// Helper methods

func (a *AuthUsecase) generateAccessToken(user *entities.User) (string, error) {
	claims := &jwt.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
	}

	return jwt.GenerateToken(claims, a.jwtSecret, 15*time.Minute) // 15 minutes
}

func (a *AuthUsecase) generateRefreshToken(userID uuid.UUID) (string, error) {
	// Generate random token
	token, err := a.generateRandomToken(32)
	if err != nil {
		return "", err
	}

	// Store in cache with 7 days expiration
	refreshKey := fmt.Sprintf("refresh_token:%s", token)
	ctx := context.Background()
	if err := a.cacheRepo.SetString(ctx, refreshKey, userID.String(), 7*24*time.Hour); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

func (a *AuthUsecase) validateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	refreshKey := fmt.Sprintf("refresh_token:%s", token)
	userIDStr, err := a.cacheRepo.GetString(ctx, refreshKey)
	if err != nil {
		return uuid.Nil, errors.AuthTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.AuthTokenInvalid
	}

	return userID, nil
}

func (a *AuthUsecase) generateResetToken() (string, error) {
	return a.generateRandomToken(32)
}

func (a *AuthUsecase) generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (a *AuthUsecase) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (a *AuthUsecase) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (a *AuthUsecase) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.ValidationFailed.WithContext("field", "password").WithContext("message", "password must be at least 8 characters")
	}

	// Add more password validation rules as needed
	// - Contains uppercase letter
	// - Contains lowercase letter
	// - Contains number
	// - Contains special character

	return nil
}

// Additional utility methods

func (a *AuthUsecase) ValidateToken(ctx context.Context, token string) (*entities.TokenClaims, error) {
	claims, err := jwt.ValidateToken(token, a.jwtSecret)
	if err != nil {
		return nil, errors.AuthTokenInvalid
	}

	// Convert JWT claims to domain claims
	tokenClaims := &entities.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}

	return tokenClaims, nil
}

func (a *AuthUsecase) RevokeToken(ctx context.Context, token string) error {
	// Add token to blacklist
	blacklistKey := fmt.Sprintf("blacklist_token:%s", token)
	if err := a.cacheRepo.SetString(ctx, blacklistKey, "revoked", 24*time.Hour); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (a *AuthUsecase) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	sessionID, err := a.generateRandomToken(32)
	if err != nil {
		return "", err
	}

	if err := a.cacheRepo.SetSession(ctx, sessionID, userID.String()); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

func (a *AuthUsecase) ValidateSession(ctx context.Context, sessionID string) (*entities.User, error) {
	userIDStr, err := a.cacheRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, errors.AuthTokenInvalid
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.AuthTokenInvalid
	}

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		return nil, errors.UserInactive
	}

	return user, nil
}

func (a *AuthUsecase) InvalidateSession(ctx context.Context, sessionID string) error {
	return a.cacheRepo.InvalidateSession(ctx, sessionID)
}
