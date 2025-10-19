// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/internal/core/ports/output"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// userService implementa el puerto de entrada UserService
type userService struct {
	userRepo  output.UserRepository
	cacheRepo output.CacheRepository
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(userRepo output.UserRepository, cacheRepo output.CacheRepository) input.UserService {
	return &userService{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
	}
}

func (s *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	// Verificar caché primero
	if cachedUser, err := s.cacheRepo.GetCachedUser(ctx, userID); err == nil {
		// Convertir User a UserProfile
		profile := &domain.UserProfile{
			ID:        cachedUser.ID,
			Email:     cachedUser.Email,
			FirstName: cachedUser.FirstName,
			LastName:  cachedUser.LastName,
			Role:      cachedUser.Role,
			Avatar:    cachedUser.Avatar,
			CreatedAt: cachedUser.CreatedAt,
		}
		return profile, nil
	}

	// Obtener de la base de datos
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convertir a perfil
	profile := &domain.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Cachear el usuario
	if err := s.cacheRepo.CacheUser(ctx, user); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache user: %v\n", err)
	}

	return profile, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*domain.UserProfile, error) {
	// Validar actualizaciones
	validatedUpdates := make(map[string]interface{})

	for key, value := range updates {
		switch key {
		case "firstName":
			if str, ok := value.(string); ok && len(strings.TrimSpace(str)) > 0 {
				validatedUpdates[key] = strings.TrimSpace(str)
			}
		case "lastName":
			if str, ok := value.(string); ok && len(strings.TrimSpace(str)) > 0 {
				validatedUpdates[key] = strings.TrimSpace(str)
			}
		case "avatar":
			if str, ok := value.(string); ok {
				validatedUpdates[key] = str
			}
		case "bio":
			if str, ok := value.(string); ok && len(str) <= 500 {
				validatedUpdates[key] = strings.TrimSpace(str)
			}
		}
	}

	if len(validatedUpdates) == 0 {
		return nil, fmt.Errorf("no valid fields to update")
	}

	// Actualizar en la base de datos
	if err := s.userRepo.UpdateProfile(ctx, userID, validatedUpdates); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// Invalidar caché
	if err := s.cacheRepo.InvalidateUserCache(ctx, userID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	// Retornar perfil actualizado
	return s.GetUserByID(ctx, userID)
}

func (s *userService) DeactivateAccount(ctx context.Context, userID uuid.UUID, password, reason string) error {
	// Obtener usuario para verificar contraseña
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.ErrInvalidCredentials
	}

	// Actualizar estado del usuario
	if err := s.userRepo.UpdateStatus(ctx, userID, false); err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	// TODO: Log deactivation reason for analytics
	if reason != "" {
		fmt.Printf("Account deactivated - User: %s, Reason: %s\n", userID, reason)
	}

	// Invalidar caché
	if err := s.cacheRepo.InvalidateUserCache(ctx, userID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (s *userService) GetSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	settings, err := s.userRepo.GetSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return settings, nil
}

func (s *userService) UpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error {
	// Validar configuración
	validatedSettings := make(map[string]interface{})

	for key, value := range settings {
		switch key {
		case "preferred_quality":
			if str, ok := value.(string); ok {
				validQuality := []string{"auto", "1080p", "720p", "480p", "360p"}
				for _, quality := range validQuality {
					if str == quality {
						validatedSettings[key] = str
						break
					}
				}
			}
		case "autoplay":
			if b, ok := value.(bool); ok {
				validatedSettings[key] = b
			}
		case "notifications":
			if b, ok := value.(bool); ok {
				validatedSettings[key] = b
			}
		case "language":
			if str, ok := value.(string); ok {
				validLanguages := []string{"es", "en", "fr", "de", "pt"}
				for _, lang := range validLanguages {
					if str == lang {
						validatedSettings[key] = str
						break
					}
				}
			}
		case "theme":
			if str, ok := value.(string); ok {
				validThemes := []string{"light", "dark", "auto"}
				for _, theme := range validThemes {
					if str == theme {
						validatedSettings[key] = str
						break
					}
				}
			}
		}
	}

	if err := s.userRepo.UpdateSettings(ctx, userID, validatedSettings); err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (s *userService) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	profile, err := s.userRepo.GetPublicProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public profile: %w", err)
	}

	return profile, nil
}

func (s *userService) GetUserStats(ctx context.Context, userID uuid.UUID) (*domain.UserStats, error) {
	stats, err := s.userRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// Funcionalidad de favoritos
func (s *userService) AddToFavorites(ctx context.Context, userID, videoID uuid.UUID) error {
	// TODO: Implementar con FavoriteRepository
	// Por ahora, usando caché como solución temporal

	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())

	// Verificar si ya está en favoritos
	favorites, err := s.getFavoritesFromCache(ctx, userID)
	if err == nil {
		for _, fav := range favorites {
			if fav == videoID.String() {
				return fmt.Errorf("video already in favorites")
			}
		}
	}

	// Añadir a favoritos
	if err := s.cacheRepo.SetAdd(ctx, favoritesKey, videoID.String()); err != nil {
		return fmt.Errorf("failed to add to favorites: %w", err)
	}

	return nil
}

func (s *userService) RemoveFromFavorites(ctx context.Context, userID, videoID uuid.UUID) error {
	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())

	if err := s.cacheRepo.SetRemove(ctx, favoritesKey, videoID.String()); err != nil {
		return fmt.Errorf("failed to remove from favorites: %w", err)
	}

	return nil
}

func (s *userService) GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Video, int64, error) {
	// TODO: Implementar con repositorio adecuado y paginación
	// Por ahora, retornando lista vacía
	return []domain.Video{}, 0, nil
}

// Funcionalidad de historial de visualización
func (s *userService) GetWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.WatchHistory, int64, error) {
	// TODO: Implementar con WatchHistoryRepository
	// Por ahora, retornando lista vacía
	return []domain.WatchHistory{}, 0, nil
}

func (s *userService) UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error {
	// Almacenar progreso de visualización en caché
	progressKey := fmt.Sprintf("watch_progress:%s:%s", userID.String(), videoID.String())

	progress := map[string]interface{}{
		"position":   position,
		"quality":    quality,
		"updated_at": fmt.Sprintf("%d", time.Now().Unix()),
	}

	if err := s.cacheRepo.Set(ctx, progressKey, progress, 30*24*time.Hour); err != nil {
		return fmt.Errorf("failed to update watch progress: %w", err)
	}

	// TODO: También almacenar en base de datos para persistencia entre dispositivos

	return nil
}

// Funcionalidad de búsqueda
func (s *userService) SearchUsers(ctx context.Context, query string, page, limit int) ([]*domain.UserProfile, int64, error) {
	users, total, err := s.userRepo.Search(ctx, query, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	// Convertir a perfiles
	profiles := make([]*domain.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = &domain.UserProfile{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Avatar:    user.Avatar,
			CreatedAt: user.CreatedAt,
		}
	}

	return profiles, total, nil
}

// Funcionalidad de administración
func (s *userService) GetAllUsers(ctx context.Context, page, limit int) ([]*domain.UserProfile, int64, error) {
	users, total, err := s.userRepo.List(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all users: %w", err)
	}

	// Convertir a perfiles
	profiles := make([]*domain.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = &domain.UserProfile{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
			Avatar:    user.Avatar,
			CreatedAt: user.CreatedAt,
		}
	}

	return profiles, total, nil
}

func (s *userService) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	// Validar rol
	validRoles := []string{"student", "instructor", "admin"}
	isValidRole := false
	for _, validRole := range validRoles {
		if role == validRole {
			isValidRole = true
			break
		}
	}

	if !isValidRole {
		return fmt.Errorf("invalid role")
	}

	// Actualizar rol
	if err := s.userRepo.UpdateRole(ctx, userID, role); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	// Invalidar caché
	if err := s.cacheRepo.InvalidateUserCache(ctx, userID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (s *userService) UpdateUserStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	if err := s.userRepo.UpdateStatus(ctx, userID, isActive); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Invalidar caché
	if err := s.cacheRepo.InvalidateUserCache(ctx, userID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	_ = s.cacheRepo.InvalidateUserCache(ctx, userID) // best-effort
	return nil
}

// Métodos privados auxiliares
func (s *userService) getFavoritesFromCache(ctx context.Context, userID uuid.UUID) ([]string, error) {
	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())

	members, err := s.cacheRepo.SetMembers(ctx, favoritesKey)
	if err != nil {
		return nil, err
	}

	favorites := make([]string, len(members))
	for i, member := range members {
		if str, ok := member.(string); ok {
			favorites[i] = str
		}
	}

	return favorites, nil
}
