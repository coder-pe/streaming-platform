package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/domain/errors"
	"streaming-platform/internal/domain/interfaces"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo  interfaces.UserRepository
	cacheRepo interfaces.CacheRepository
}

func NewUserUsecase(userRepo interfaces.UserRepository, cacheRepo interfaces.CacheRepository) *UserUsecase {
	return &UserUsecase{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
	}
}

func (u *UserUsecase) GetUserByID(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error) {
	// Check cache first
	var cachedProfile entities.UserProfile
	if err := u.cacheRepo.GetCachedUser(ctx, userID.String(), &cachedProfile); err == nil {
		return &cachedProfile, nil
	}

	// Get from database
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert to profile
	profile := &entities.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt,
	}

	// Cache the profile
	if err := u.cacheRepo.CacheUser(ctx, userID.String(), profile); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache user profile: %v\n", err)
	}

	return profile, nil
}

func (u *UserUsecase) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entities.UserProfile, error) {
	// Validate updates
	validatedUpdates := make(map[string]interface{})

	for key, value := range updates {
		switch key {
		case "first_name":
			if str, ok := value.(string); ok && len(strings.TrimSpace(str)) > 0 {
				validatedUpdates[key] = strings.TrimSpace(str)
			}
		case "last_name":
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
		return nil, errors.ValidationFailed.WithContext("message", "no valid fields to update")
	}

	// Update in database
	if err := u.userRepo.UpdateProfile(ctx, userID, validatedUpdates); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// Invalidate cache
	if err := u.cacheRepo.InvalidateUserCache(ctx, userID.String()); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	// Return updated profile
	return u.GetUserByID(ctx, userID)
}

func (u *UserUsecase) DeactivateAccount(ctx context.Context, userID uuid.UUID, password, reason string) error {
	// Get user to verify password
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return errors.AuthInvalidCredentials.WithContext("field", "password")
	}

	// Update user status
	if err := u.userRepo.UpdateStatus(ctx, userID, false); err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}

	// TODO: Log deactivation reason for analytics
	if reason != "" {
		fmt.Printf("Account deactivated - User: %s, Reason: %s\n", userID, reason)
	}

	// Invalidate cache
	if err := u.cacheRepo.InvalidateUserCache(ctx, userID.String()); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (u *UserUsecase) GetSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	settings, err := u.userRepo.GetSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return settings, nil
}

func (u *UserUsecase) UpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error {
	// Validate settings
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

	if err := u.userRepo.UpdateSettings(ctx, userID, validatedSettings); err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (u *UserUsecase) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error) {
	profile, err := u.userRepo.GetPublicProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public profile: %w", err)
	}

	return profile, nil
}

func (u *UserUsecase) GetUserStats(ctx context.Context, userID uuid.UUID) (*entities.UserStats, error) {
	stats, err := u.userRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

// Favorites functionality
func (u *UserUsecase) AddToFavorites(ctx context.Context, userID, videoID uuid.UUID) error {
	// TODO: Implement with FavoriteRepository
	// For now, using cache as a temporary solution

	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())

	// Check if already in favorites
	favorites, err := u.getFavoritesFromCache(ctx, userID)
	if err == nil {
		for _, fav := range favorites {
			if fav == videoID.String() {
				return errors.NewDomainError("ALREADY_IN_FAVORITES", "Video already in favorites", 409)
			}
		}
	}

	// Add to favorites
	favorites = append(favorites, videoID.String())

	if err := u.cacheRepo.SetAdd(ctx, favoritesKey, videoID.String()); err != nil {
		return fmt.Errorf("failed to add to favorites: %w", err)
	}

	return nil
}

func (u *UserUsecase) RemoveFromFavorites(ctx context.Context, userID, videoID uuid.UUID) error {
	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())

	if err := u.cacheRepo.SetRemove(ctx, favoritesKey, videoID.String()); err != nil {
		return fmt.Errorf("failed to remove from favorites: %w", err)
	}

	return nil
}

func (u *UserUsecase) GetFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.Video, int64, error) {
	// TODO: Implement with proper repository and pagination
	// For now, returning empty list
	return []*entities.Video{}, 0, nil
}

// Watch history functionality
func (u *UserUsecase) GetWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entities.WatchHistory, int64, error) {
	// TODO: Implement with WatchHistoryRepository
	// For now, returning empty list
	return []*entities.WatchHistory{}, 0, nil
}

func (u *UserUsecase) UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error {
	// Store watch progress in cache
	progressKey := fmt.Sprintf("watch_progress:%s:%s", userID.String(), videoID.String())

	progress := map[string]interface{}{
		"position":   position,
		"quality":    quality,
		"updated_at": fmt.Sprintf("%d", time.Now().Unix()),
	}

	if err := u.cacheRepo.Set(ctx, progressKey, progress, 30*24*time.Hour); err != nil {
		return fmt.Errorf("failed to update watch progress: %w", err)
	}

	// TODO: Also store in database for persistence across devices

	return nil
}

// Search functionality
func (u *UserUsecase) SearchUsers(ctx context.Context, query string, page, limit int) ([]*entities.UserProfile, int64, error) {
	users, total, err := u.userRepo.Search(ctx, query, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert to profiles
	profiles := make([]*entities.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = &entities.UserProfile{
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

// Admin functionality
func (u *UserUsecase) GetAllUsers(ctx context.Context, page, limit int) ([]*entities.UserProfile, int64, error) {
	users, total, err := u.userRepo.List(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all users: %w", err)
	}

	// Convert to profiles
	profiles := make([]*entities.UserProfile, len(users))
	for i, user := range users {
		profiles[i] = &entities.UserProfile{
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

func (u *UserUsecase) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	// Validate role
	validRoles := []string{"student", "instructor", "admin"}
	isValidRole := false
	for _, validRole := range validRoles {
		if role == validRole {
			isValidRole = true
			break
		}
	}

	if !isValidRole {
		return errors.ValidationFailed.WithContext("field", "role").WithContext("message", "invalid role")
	}

	// Update role
	if err := u.userRepo.UpdateRole(ctx, userID, role); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	// Invalidate cache
	if err := u.cacheRepo.InvalidateUserCache(ctx, userID.String()); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

func (u *UserUsecase) UpdateUserStatus(ctx context.Context, userID uuid.UUID, isActive bool) error {
	if err := u.userRepo.UpdateStatus(ctx, userID, isActive); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	// Invalidate cache
	if err := u.cacheRepo.InvalidateUserCache(ctx, userID.String()); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to invalidate user cache: %v\n", err)
	}

	return nil
}

// Helper methods
func (u *UserUsecase) getFavoritesFromCache(ctx context.Context, userID uuid.UUID) ([]string, error) {
	favoritesKey := fmt.Sprintf("favorites:%s", userID.String())

	members, err := u.cacheRepo.SetMembers(ctx, favoritesKey)
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

func (u *UserUsecase) validateUserUpdate(updates map[string]interface{}) error {
	validationErrors := errors.NewValidationErrors()

	for field, value := range updates {
		switch field {
		case "first_name", "last_name":
			if str, ok := value.(string); ok {
				if len(strings.TrimSpace(str)) < 2 {
					validationErrors.Add(field, "must be at least 2 characters", value)
				}
				if len(str) > 50 {
					validationErrors.Add(field, "must be at most 50 characters", value)
				}
			} else {
				validationErrors.Add(field, "must be a string", value)
			}
		case "bio":
			if str, ok := value.(string); ok {
				if len(str) > 500 {
					validationErrors.Add(field, "must be at most 500 characters", value)
				}
			} else {
				validationErrors.Add(field, "must be a string", value)
			}
		}
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}
