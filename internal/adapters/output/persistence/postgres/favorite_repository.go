// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"streaming-platform/internal/core/domain"
	"streaming-platform/internal/core/ports/output"

	"github.com/google/uuid"
)

type favoriteRepository struct {
	db *sql.DB
}

// NewFavoriteRepository crea una nueva instancia del repositorio de favoritos
func NewFavoriteRepository(db *sql.DB) output.FavoriteRepository {
	return &favoriteRepository{
		db: db,
	}
}

func (r *favoriteRepository) AddFavorite(ctx context.Context, userID, videoID uuid.UUID) error {
	query := `
		INSERT INTO favorites (user_id, video_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, video_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to add favorite: %w", err)
	}

	return nil
}

func (r *favoriteRepository) RemoveFavorite(ctx context.Context, userID, videoID uuid.UUID) error {
	query := `
		DELETE FROM favorites
		WHERE user_id = $1 AND video_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to remove favorite: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("favorite not found")
	}

	return nil
}

func (r *favoriteRepository) IsFavorite(ctx context.Context, userID, videoID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM favorites
			WHERE user_id = $1 AND video_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, videoID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check favorite: %w", err)
	}

	return exists, nil
}

func (r *favoriteRepository) GetUserFavorites(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.Favorite, int64, error) {
	offset := (page - 1) * limit

	// Obtener total
	var total int64
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count favorites: %w", err)
	}

	// Obtener favoritos con información del video
	query := `
		SELECT
			f.id,
			f.user_id,
			f.video_id,
			f.created_at,
			v.id,
			v.title,
			v.description,
			v.instructor_id,
			v.category,
			v.tags,
			v.duration,
			v.thumbnail,
			v.status,
			v.is_public,
			v.view_count,
			v.rating,
			v.created_at,
			v.updated_at
		FROM favorites f
		JOIN videos v ON f.video_id = v.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []*domain.Favorite
	for rows.Next() {
		var favorite domain.Favorite
		favorite.Video = &domain.Video{}

		err := rows.Scan(
			&favorite.ID,
			&favorite.UserID,
			&favorite.VideoID,
			&favorite.CreatedAt,
			&favorite.Video.ID,
			&favorite.Video.Title,
			&favorite.Video.Description,
			&favorite.Video.InstructorID,
			&favorite.Video.Category,
			&favorite.Video.Tags,
			&favorite.Video.Duration,
			&favorite.Video.Thumbnail,
			&favorite.Video.Status,
			&favorite.Video.IsPublic,
			&favorite.Video.ViewCount,
			&favorite.Video.Rating,
			&favorite.Video.CreatedAt,
			&favorite.Video.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan favorite: %w", err)
		}

		favorites = append(favorites, &favorite)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating favorites: %w", err)
	}

	return favorites, total, nil
}

func (r *favoriteRepository) GetFavoriteCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get favorite count: %w", err)
	}

	return count, nil
}
