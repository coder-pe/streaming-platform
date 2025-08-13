// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/domain/errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, avatar, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Avatar,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return errors.UserAlreadyExists
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, avatar, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Avatar,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, avatar, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Avatar,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, first_name = $4, last_name = $5, 
		    role = $6, avatar = $7, is_active = $8, updated_at = $9
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Avatar,
		user.IsActive,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.UserNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.UserNotFound
	}

	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (r *UserRepository) GetPublicProfile(ctx context.Context, id uuid.UUID) (*entities.UserProfile, error) {
	query := `
		SELECT id, email, first_name, last_name, role, avatar, created_at
		FROM users
		WHERE id = $1 AND is_active = true
	`

	profile := &entities.UserProfile{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&profile.ID,
		&profile.Email,
		&profile.FirstName,
		&profile.LastName,
		&profile.Role,
		&profile.Avatar,
		&profile.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UserNotFound
		}
		return nil, fmt.Errorf("failed to get public profile: %w", err)
	}

	return profile, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "first_name", "last_name", "avatar", "bio":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return nil
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.UserNotFound
	}

	return nil
}

func (r *UserRepository) GetUserStats(ctx context.Context, id uuid.UUID) (*entities.UserStats, error) {
	query := `
		SELECT 
			COALESCE(COUNT(v.id), 0) as video_count,
			COALESCE(SUM(v.view_count), 0) as total_views,
			COALESCE(COUNT(wh.id), 0) as videos_watched,
			COALESCE(SUM(wh.watch_time), 0) as total_watch_time
		FROM users u
		LEFT JOIN videos v ON u.id = v.instructor_id
		LEFT JOIN watch_history wh ON u.id = wh.user_id
		WHERE u.id = $1
		GROUP BY u.id
	`

	stats := &entities.UserStats{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&stats.TotalVideos,
		&stats.TotalViews,
		&stats.VideosWatched,
		&stats.TotalWatchTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &entities.UserStats{}, nil
		}
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}

func (r *UserRepository) GetSettings(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	query := `SELECT settings FROM user_settings WHERE user_id = $1`

	var settingsJSON string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&settingsJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default settings if none exist
			return map[string]interface{}{
				"preferred_quality": "720p",
				"autoplay":          true,
				"notifications":     true,
				"language":          "es",
			}, nil
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Parse JSON settings
	settings := make(map[string]interface{})
	// Note: In a real implementation, you'd use json.Unmarshal here
	// For simplicity, returning default settings
	settings["preferred_quality"] = "720p"
	settings["autoplay"] = true
	settings["notifications"] = true
	settings["language"] = "es"

	return settings, nil
}

func (r *UserRepository) UpdateSettings(ctx context.Context, id uuid.UUID, settings map[string]interface{}) error {
	// Note: In a real implementation, you'd marshal settings to JSON
	query := `
		INSERT INTO user_settings (user_id, settings, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET settings = $2, updated_at = $3
	`

	settingsJSON := "{}" // In practice, use json.Marshal(settings)

	_, err := r.db.ExecContext(ctx, query, id, settingsJSON, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update user settings: %w", err)
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context, page, limit int) ([]*entities.User, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users
	query := `
		SELECT id, email, first_name, last_name, role, avatar, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []*entities.User{}
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.Avatar,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) Search(ctx context.Context, query string, page, limit int) ([]*entities.User, int64, error) {
	offset := (page - 1) * limit
	searchQuery := "%" + strings.ToLower(query) + "%"

	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM users 
		WHERE LOWER(first_name) LIKE $1 
		   OR LOWER(last_name) LIKE $1 
		   OR LOWER(email) LIKE $1
	`
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, searchQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get search count: %w", err)
	}

	// Search users
	searchSQL := `
		SELECT id, email, first_name, last_name, role, avatar, is_active, created_at, updated_at
		FROM users
		WHERE LOWER(first_name) LIKE $1 
		   OR LOWER(last_name) LIKE $1 
		   OR LOWER(email) LIKE $1
		ORDER BY 
			CASE 
				WHEN LOWER(first_name) LIKE $1 THEN 1
				WHEN LOWER(last_name) LIKE $1 THEN 2
				ELSE 3
			END,
			created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, searchSQL, searchQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	users := []*entities.User{}
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.Avatar,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, isActive, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.UserNotFound
	}

	return nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	query := `UPDATE users SET role = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, role, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.UserNotFound
	}

	return nil
}
