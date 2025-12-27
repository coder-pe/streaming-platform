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

type watchHistoryRepository struct {
	db *sql.DB
}

// NewWatchHistoryRepository crea una nueva instancia del repositorio de historial
func NewWatchHistoryRepository(db *sql.DB) output.WatchHistoryRepository {
	return &watchHistoryRepository{
		db: db,
	}
}

func (r *watchHistoryRepository) SaveWatchHistory(ctx context.Context, history *domain.WatchHistory) error {
	query := `
		INSERT INTO watch_history (id, user_id, video_id, position, quality, watched_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id, video_id)
		DO UPDATE SET
			position = EXCLUDED.position,
			quality = EXCLUDED.quality,
			watched_at = NOW(),
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	// Si no tiene ID, generar uno nuevo
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		history.ID,
		history.UserID,
		history.VideoID,
		history.Position,
		history.Quality,
	).Scan(&history.ID, &history.CreatedAt, &history.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save watch history: %w", err)
	}

	return nil
}

func (r *watchHistoryRepository) GetWatchHistory(ctx context.Context, userID, videoID uuid.UUID) (*domain.WatchHistory, error) {
	query := `
		SELECT
			id,
			user_id,
			video_id,
			position,
			quality,
			watched_at,
			created_at,
			updated_at
		FROM watch_history
		WHERE user_id = $1 AND video_id = $2
	`

	var history domain.WatchHistory
	err := r.db.QueryRowContext(ctx, query, userID, videoID).Scan(
		&history.ID,
		&history.UserID,
		&history.VideoID,
		&history.Position,
		&history.Quality,
		&history.WatchedAt,
		&history.CreatedAt,
		&history.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("watch history not found")
		}
		return nil, fmt.Errorf("failed to get watch history: %w", err)
	}

	return &history, nil
}

func (r *watchHistoryRepository) GetUserWatchHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]*domain.WatchHistory, int64, error) {
	offset := (page - 1) * limit

	// Obtener total
	var total int64
	countQuery := `SELECT COUNT(*) FROM watch_history WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count watch history: %w", err)
	}

	// Obtener historial con información del video
	query := `
		SELECT
			wh.id,
			wh.user_id,
			wh.video_id,
			wh.position,
			wh.quality,
			wh.watched_at,
			wh.created_at,
			wh.updated_at,
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
		FROM watch_history wh
		JOIN videos v ON wh.video_id = v.id
		WHERE wh.user_id = $1
		ORDER BY wh.watched_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query watch history: %w", err)
	}
	defer rows.Close()

	var histories []*domain.WatchHistory
	for rows.Next() {
		var history domain.WatchHistory
		history.Video = &domain.Video{}

		err := rows.Scan(
			&history.ID,
			&history.UserID,
			&history.VideoID,
			&history.Position,
			&history.Quality,
			&history.WatchedAt,
			&history.CreatedAt,
			&history.UpdatedAt,
			&history.Video.ID,
			&history.Video.Title,
			&history.Video.Description,
			&history.Video.InstructorID,
			&history.Video.Category,
			&history.Video.Tags,
			&history.Video.Duration,
			&history.Video.Thumbnail,
			&history.Video.Status,
			&history.Video.IsPublic,
			&history.Video.ViewCount,
			&history.Video.Rating,
			&history.Video.CreatedAt,
			&history.Video.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan watch history: %w", err)
		}

		histories = append(histories, &history)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating watch history: %w", err)
	}

	return histories, total, nil
}

func (r *watchHistoryRepository) DeleteWatchHistory(ctx context.Context, userID, videoID uuid.UUID) error {
	query := `
		DELETE FROM watch_history
		WHERE user_id = $1 AND video_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, videoID)
	if err != nil {
		return fmt.Errorf("failed to delete watch history: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("watch history not found")
	}

	return nil
}

func (r *watchHistoryRepository) GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.WatchHistory, error) {
	// Videos para continuar viendo: progreso > 10 segundos y < 90% del video
	query := `
		SELECT
			wh.id,
			wh.user_id,
			wh.video_id,
			wh.position,
			wh.quality,
			wh.watched_at,
			wh.created_at,
			wh.updated_at,
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
		FROM watch_history wh
		JOIN videos v ON wh.video_id = v.id
		WHERE wh.user_id = $1
		  AND wh.position > 10
		  AND v.duration > 0
		  AND (CAST(wh.position AS FLOAT) / CAST(v.duration AS FLOAT)) < 0.9
		  AND v.status = 'ready'
		  AND v.is_public = true
		ORDER BY wh.watched_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query continue watching: %w", err)
	}
	defer rows.Close()

	var histories []*domain.WatchHistory
	for rows.Next() {
		var history domain.WatchHistory
		history.Video = &domain.Video{}

		err := rows.Scan(
			&history.ID,
			&history.UserID,
			&history.VideoID,
			&history.Position,
			&history.Quality,
			&history.WatchedAt,
			&history.CreatedAt,
			&history.UpdatedAt,
			&history.Video.ID,
			&history.Video.Title,
			&history.Video.Description,
			&history.Video.InstructorID,
			&history.Video.Category,
			&history.Video.Tags,
			&history.Video.Duration,
			&history.Video.Thumbnail,
			&history.Video.Status,
			&history.Video.IsPublic,
			&history.Video.ViewCount,
			&history.Video.Rating,
			&history.Video.CreatedAt,
			&history.Video.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan continue watching: %w", err)
		}

		histories = append(histories, &history)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating continue watching: %w", err)
	}

	return histories, nil
}
