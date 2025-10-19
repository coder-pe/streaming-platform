// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/jwt"
	"streaming-platform/pkg/validator"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// authService implementa el puerto de entrada AuthService
type authService struct {
	userRepo  output.UserRepository
	cacheRepo output.CacheRepository
	jwtSecret string
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(userRepo output.UserRepository, cacheRepo output.CacheRepository, jwtSecret string) input.AuthService {
	return &authService{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, *domain.UserProfile, error) {
	// Validar email
	if err := validator.ValidateEmail(email); err != nil {
		return "", "", nil, domain.ErrInvalidCredentials
	}

	// Obtener usuario por email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", "", nil, domain.ErrInvalidCredentials
		}
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		return "", "", nil, domain.ErrUserInactive
	}

	// Verificar contraseña
	if !s.verifyPassword(password, user.PasswordHash) {
		return "", "", nil, domain.ErrInvalidCredentials
	}

	// Generar tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Actualizar último login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to update last login for user %s: %v\n", user.ID, err)
	}

	// Crear perfil de usuario
	profile := &domain.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Cachear usuario
	if err := s.cacheRepo.CacheUser(ctx, user); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: failed to cache user: %v\n", err)
	}

	return accessToken, refreshToken, profile, nil
}

func (s *authService) Register(ctx context.Context, email, password, firstName, lastName, role string) (string, string, *domain.UserProfile, error) {
	// Validar email
	if err := validator.ValidateEmail(email); err != nil {
		return "", "", nil, err
	}

	// Validar contraseña
	if err := validator.ValidatePassword(password); err != nil {
		return "", "", nil, err
	}

	// Validar nombre
	if err := validator.ValidateRequired("firstName", firstName); err != nil {
		return "", "", nil, err
	}
	if err := validator.ValidateRequired("lastName", lastName); err != nil {
		return "", "", nil, err
	}

	// Verificar si el usuario ya existe
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return "", "", nil, domain.ErrUserAlreadyExists
	}
	if err != nil && err != domain.ErrUserNotFound {
		return "", "", nil, fmt.Errorf("failed to check user: %w", err)
	}

	// Hash de la contraseña
	hashed, err := s.hashPassword(password)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Crear usuario
	user := &domain.User{
		Email:        strings.ToLower(email),
		PasswordHash: hashed,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		IsActive:     true,
	}

	// Rol por defecto: student
	if user.Role == "" {
		user.Role = "student"
	}

	// Persistir usuario
	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", "", nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generar tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Crear perfil de usuario
	profile := &domain.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Cachear usuario (best-effort)
	_ = s.cacheRepo.CacheUser(ctx, user)

	return accessToken, refreshToken, profile, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (string, string, *domain.UserProfile, error) {
	// Validar refresh token
	userID, err := s.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", nil, err
	}

	// Obtener usuario
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", "", nil, domain.ErrInvalidToken
		}
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		return "", "", nil, domain.ErrUserInactive
	}

	// Generar nuevos tokens
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Invalidar el refresh token antiguo
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	if err := s.cacheRepo.Delete(ctx, refreshKey); err != nil {
		// Log error but don't fail refresh
		fmt.Printf("Warning: failed to invalidate old refresh token: %v\n", err)
	}

	// Crear perfil de usuario
	profile := &domain.UserProfile{
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

func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// Invalidar refresh token
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	if err := s.cacheRepo.Delete(ctx, refreshKey); err != nil {
		return fmt.Errorf("failed to invalidate refresh token: %w", err)
	}

	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	// Validar email
	if err := validator.ValidateEmail(email); err != nil {
		// No revelar si el email existe por seguridad
		return nil
	}

	// Obtener usuario por email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// No revelar si el usuario existe por seguridad
			return nil
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Generar token de reset
	resetToken, err := s.generateRandomToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Almacenar token de reset en caché con expiración de 1 hora
	resetKey := fmt.Sprintf("password_reset:%s", resetToken)
	if err := s.cacheRepo.SetString(ctx, resetKey, user.ID.String(), time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// TODO: Enviar email con el enlace de reset
	// Esto se integraría con un servicio de email
	fmt.Printf("Password reset token for %s: %s\n", email, resetToken)

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Validar nueva contraseña
	if err := validator.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Obtener user ID del token de reset
	resetKey := fmt.Sprintf("password_reset:%s", token)
	userIDStr, err := s.cacheRepo.GetString(ctx, resetKey)
	if err != nil {
		return domain.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return domain.ErrInvalidToken
	}

	// Hash de la nueva contraseña
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Actualizar contraseña
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidar token de reset
	if err := s.cacheRepo.Delete(ctx, resetKey); err != nil {
		// Log error but don't fail reset
		fmt.Printf("Warning: failed to invalidate reset token: %v\n", err)
	}

	// Invalidar caché del usuario
	if err := s.cacheRepo.InvalidateUserCache(ctx, userID); err != nil {
		// Log error but don't fail reset
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Obtener usuario
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return domain.ErrUnauthorized
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verificar contraseña actual
	if !s.verifyPassword(currentPassword, user.PasswordHash) {
		return domain.ErrInvalidCredentials
	}

	// Validar nueva contraseña
	if err := validator.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash de la nueva contraseña
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Actualizar contraseña
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidar caché del usuario (best-effort)
	_ = s.cacheRepo.InvalidateUserCache(ctx, userID)

	return nil
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.TokenClaims, error) {
	// Validar token JWT
	claims, err := jwt.ValidateToken(token, s.jwtSecret)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Convertir claims JWT a claims del dominio
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	tokenClaims := &domain.TokenClaims{
		UserID: userID,
		Email:  claims.Email,
		Role:   claims.Role,
	}

	return tokenClaims, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string) error {
	// Añadir token a la lista negra
	blacklistKey := fmt.Sprintf("blacklist_token:%s", token)
	if err := s.cacheRepo.SetString(ctx, blacklistKey, "revoked", 24*time.Hour); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (s *authService) CreateSession(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generar ID de sesión
	sessionID, err := s.generateRandomToken(32)
	if err != nil {
		return "", err
	}

	// Almacenar sesión en caché
	if err := s.cacheRepo.SetSession(ctx, sessionID, userID, 24*time.Hour); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

func (s *authService) ValidateSession(ctx context.Context, sessionID string) (*domain.User, error) {
	// Obtener user ID de la sesión
	userID, err := s.cacheRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Obtener usuario
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	return user, nil
}

func (s *authService) InvalidateSession(ctx context.Context, sessionID string) error {
	return s.cacheRepo.InvalidateSession(ctx, sessionID)
}

// Métodos privados auxiliares

func (s *authService) generateAccessToken(user *domain.User) (string, error) {
	claims := &jwt.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
	}

	return jwt.GenerateToken(claims, s.jwtSecret, 15*time.Minute) // 15 minutos
}

func (s *authService) generateRefreshToken(userID uuid.UUID) (string, error) {
	// Generar token aleatorio
	token, err := s.generateRandomToken(32)
	if err != nil {
		return "", err
	}

	// Almacenar en caché con expiración de 7 días
	refreshKey := fmt.Sprintf("refresh_token:%s", token)
	ctx := context.Background()
	if err := s.cacheRepo.SetString(ctx, refreshKey, userID.String(), 7*24*time.Hour); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

func (s *authService) validateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	refreshKey := fmt.Sprintf("refresh_token:%s", token)
	userIDStr, err := s.cacheRepo.GetString(ctx, refreshKey)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidToken
	}

	return userID, nil
}

func (s *authService) generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *authService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (s *authService) verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
