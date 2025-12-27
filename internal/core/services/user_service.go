// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package services

import (
	"context"
	"fmt"
	"strings"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/input"
	"streaming-platform/internal/core/ports/output"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// userService implementa el puerto de entrada UserService
type userService struct {
	userRepo         output.UserRepository
	cacheRepo        output.CacheRepository
	favoriteRepo     output.FavoriteRepository
	watchHistoryRepo output.WatchHistoryRepository
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(
	userRepo output.UserRepository,
	cacheRepo output.CacheRepository,
	favoriteRepo output.FavoriteRepository,
	watchHistoryRepo output.WatchHistoryRepository,
) input.UserService {
	return &userService{
		userRepo:         userRepo,
		cacheRepo:        cacheRepo,
		favoriteRepo:     favoriteRepo,
		watchHistoryRepo: watchHistoryRepo,
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
	if err := s.favoriteRepo.AddFavorite(ctx, userID, videoID); err != nil {
		return fmt.Errorf("failed to add to favorites: %w", err)
	}

	// Invalidar caché de favoritos si existe
	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())
	_ = s.cacheRepo.Delete(ctx, favoritesKey) // best-effort

	return nil
}

func (s *userService) RemoveFromFavorites(ctx context.Context, userID, videoID uuid.UUID) error {
	if err := s.favoriteRepo.RemoveFavorite(ctx, userID, videoID); err != nil {
		return fmt.Errorf("failed to remove from favorites: %w", err)
	}

	// Invalidar caché de favoritos
	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())
	_ = s.cacheRepo.Delete(ctx, favoritesKey) // best-effort

	return nil
}

func (s *userService) GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Video, int64, error) {
	// Validar paginación
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Obtener favoritos con información de videos
	favorites, total, err := s.favoriteRepo.GetUserFavorites(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get favorites: %w", err)
	}

	// Convertir []*Favorite a []Video
	videos := make([]domain.Video, len(favorites))
	for i, fav := range favorites {
		if fav.Video != nil {
			videos[i] = *fav.Video
		}
	}

	return videos, total, nil
}

// Funcionalidad de historial de visualización
func (s *userService) GetWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.WatchHistory, int64, error) {
	// Validar paginación
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Obtener historial del repositorio
	histories, total, err := s.watchHistoryRepo.GetUserWatchHistory(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get watch history: %w", err)
	}

	// Convertir []*WatchHistory a []WatchHistory
	result := make([]domain.WatchHistory, len(histories))
	for i, h := range histories {
		result[i] = *h
	}

	return result, total, nil
}

func (s *userService) UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error {
	// Crear registro de historial
	history := &domain.WatchHistory{
		UserID:   userID,
		VideoID:  videoID,
		Position: position,
		Quality:  quality,
	}

	// Guardar en base de datos (persiste entre dispositivos)
	if err := s.watchHistoryRepo.SaveWatchHistory(ctx, history); err != nil {
		return fmt.Errorf("failed to save watch history: %w", err)
	}

	return nil
}

func (s *userService) GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]domain.WatchHistory, error) {
	// Validar límite
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Obtener videos para continuar viendo (>10s y <90% completo)
	histories, err := s.watchHistoryRepo.GetContinueWatching(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get continue watching: %w", err)
	}

	// Convertir []*WatchHistory a []WatchHistory
	result := make([]domain.WatchHistory, len(histories))
	for i, h := range histories {
		result[i] = *h
	}

	return result, nil
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
