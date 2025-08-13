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

type VideoRepository struct {
	db *sql.DB
}

func NewVideoRepository(db *sql.DB) *VideoRepository {
	return &VideoRepository{
		db: db,
	}
}

func (r *VideoRepository) Create(ctx context.Context, video *entities.Video) error {
	query := `
		INSERT INTO videos (id, title, description, instructor_id, category, tags, duration, 
		                   thumbnail, status, is_public, view_count, rating, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	video.ID = uuid.New()
	video.CreatedAt = time.Now()
	video.UpdatedAt = time.Now()
	if video.ViewCount == 0 {
		video.ViewCount = 0
	}
	if video.Rating == 0 {
		video.Rating = 0.0
	}

	_, err := r.db.ExecContext(ctx, query,
		video.ID,
		video.Title,
		video.Description,
		video.InstructorID,
		video.Category,
		pq.Array(video.Tags),
		video.Duration,
		video.Thumbnail,
		video.Status,
		video.IsPublic,
		video.ViewCount,
		video.Rating,
		video.CreatedAt,
		video.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	return nil
}

func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.instructor_id, v.category, v.tags, 
		       v.duration, v.thumbnail, v.status, v.is_public, v.view_count, v.rating,
		       v.created_at, v.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.avatar, u.created_at
		FROM videos v
		LEFT JOIN users u ON v.instructor_id = u.id
		WHERE v.id = $1
	`

	video := &entities.Video{}
	var instructor entities.UserProfile
	var instructorID sql.NullString
	var instructorEmail sql.NullString
	var instructorFirstName sql.NullString
	var instructorLastName sql.NullString
	var instructorRole sql.NullString
	var instructorAvatar sql.NullString
	var instructorCreatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&video.ID,
		&video.Title,
		&video.Description,
		&video.InstructorID,
		&video.Category,
		pq.Array(&video.Tags),
		&video.Duration,
		&video.Thumbnail,
		&video.Status,
		&video.IsPublic,
		&video.ViewCount,
		&video.Rating,
		&video.CreatedAt,
		&video.UpdatedAt,
		&instructorID,
		&instructorEmail,
		&instructorFirstName,
		&instructorLastName,
		&instructorRole,
		&instructorAvatar,
		&instructorCreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.VideoNotFound
		}
		return nil, fmt.Errorf("failed to get video by ID: %w", err)
	}

	// Set instructor if exists
	if instructorID.Valid {
		instructor.ID = uuid.MustParse(instructorID.String)
		instructor.Email = instructorEmail.String
		instructor.FirstName = instructorFirstName.String
		instructor.LastName = instructorLastName.String
		instructor.Role = instructorRole.String
		instructor.Avatar = instructorAvatar.String
		if instructorCreatedAt.Valid {
			instructor.CreatedAt = instructorCreatedAt.Time
		}
		video.Instructor = &instructor
	}

	// Get video files
	videoFiles, err := r.GetVideoFiles(ctx, video.ID)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to get video files for video %s: %v\n", video.ID, err)
	} else {
		video.VideoFiles = videoFiles
	}

	return video, nil
}

func (r *VideoRepository) Update(ctx context.Context, video *entities.Video) error {
	query := `
		UPDATE videos
		SET title = $2, description = $3, category = $4, tags = $5, duration = $6,
		    thumbnail = $7, status = $8, is_public = $9, view_count = $10, rating = $11, updated_at = $12
		WHERE id = $1
	`

	video.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		video.ID,
		video.Title,
		video.Description,
		video.Category,
		pq.Array(video.Tags),
		video.Duration,
		video.Thumbnail,
		video.Status,
		video.IsPublic,
		video.ViewCount,
		video.Rating,
		video.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.VideoNotFound
	}

	return nil
}

func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete video files first
	_, err = tx.ExecContext(ctx, "DELETE FROM video_files WHERE video_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete video files: %w", err)
	}

	// Delete video
	result, err := tx.ExecContext(ctx, "DELETE FROM videos WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.VideoNotFound
	}

	return tx.Commit()
}

func (r *VideoRepository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*entities.Video, int64, error) {
	offset := (page - 1) * limit

	// Build WHERE clause
	whereClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "instructor_id":
			whereClauses = append(whereClauses, fmt.Sprintf("v.instructor_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "category":
			whereClauses = append(whereClauses, fmt.Sprintf("v.category = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "status":
			whereClauses = append(whereClauses, fmt.Sprintf("v.status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_public":
			whereClauses = append(whereClauses, fmt.Sprintf("v.is_public = $%d", argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM videos v %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get video count: %w", err)
	}

	// Get videos
	query := fmt.Sprintf(`
		SELECT v.id, v.title, v.description, v.instructor_id, v.category, v.tags, 
		       v.duration, v.thumbnail, v.status, v.is_public, v.view_count, v.rating,
		       v.created_at, v.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.avatar, u.created_at
		FROM videos v
		LEFT JOIN users u ON v.instructor_id = u.id
		%s
		ORDER BY v.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list videos: %w", err)
	}
	defer rows.Close()

	videos := []*entities.Video{}
	for rows.Next() {
		video, err := r.scanVideoWithInstructor(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate videos: %w", err)
	}

	return videos, total, nil
}

func (r *VideoRepository) Search(ctx context.Context, searchReq entities.VideoSearchRequest) (*entities.VideoSearchResponse, error) {
	offset := (searchReq.Page - 1) * searchReq.Limit

	// Build search query
	whereClauses := []string{"v.is_public = true"} // Only search public videos
	args := []interface{}{}
	argIndex := 1

	if searchReq.Query != "" {
		searchQuery := "%" + strings.ToLower(searchReq.Query) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(v.title) LIKE $%d OR LOWER(v.description) LIKE $%d)", argIndex, argIndex))
		args = append(args, searchQuery)
		argIndex++
	}

	if searchReq.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("v.category = $%d", argIndex))
		args = append(args, searchReq.Category)
		argIndex++
	}

	if len(searchReq.Tags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("v.tags && $%d", argIndex))
		args = append(args, pq.Array(searchReq.Tags))
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM videos v %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get search count: %w", err)
	}

	// Order by relevance and other factors
	orderBy := "ORDER BY v.view_count DESC, v.rating DESC, v.created_at DESC"

	// Get videos
	query := fmt.Sprintf(`
		SELECT v.id, v.title, v.description, v.instructor_id, v.category, v.tags, 
		       v.duration, v.thumbnail, v.status, v.is_public, v.view_count, v.rating,
		       v.created_at, v.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.avatar, u.created_at
		FROM videos v
		LEFT JOIN users u ON v.instructor_id = u.id
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, searchReq.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search videos: %w", err)
	}
	defer rows.Close()

	videos := []entities.Video{}
	for rows.Next() {
		video, err := r.scanVideoWithInstructor(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, *video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate videos: %w", err)
	}

	return &entities.VideoSearchResponse{
		Videos: videos,
		Total:  total,
		Page:   searchReq.Page,
		Limit:  searchReq.Limit,
	}, nil
}

func (r *VideoRepository) GetByCategory(ctx context.Context, category string, page, limit int) ([]*entities.Video, int64, error) {
	filters := map[string]interface{}{
		"category":  category,
		"is_public": true,
	}
	return r.List(ctx, filters, page, limit)
}

func (r *VideoRepository) GetByInstructor(ctx context.Context, instructorID uuid.UUID, page, limit int) ([]*entities.Video, int64, error) {
	filters := map[string]interface{}{
		"instructor_id": instructorID,
	}
	return r.List(ctx, filters, page, limit)
}

func (r *VideoRepository) GetPublicVideos(ctx context.Context, page, limit int) ([]*entities.Video, int64, error) {
	filters := map[string]interface{}{
		"is_public": true,
		"status":    entities.VideoStatusReady,
	}
	return r.List(ctx, filters, page, limit)
}

func (r *VideoRepository) GetFeaturedVideos(ctx context.Context, limit int) ([]*entities.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.instructor_id, v.category, v.tags, 
		       v.duration, v.thumbnail, v.status, v.is_public, v.view_count, v.rating,
		       v.created_at, v.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.avatar, u.created_at
		FROM videos v
		LEFT JOIN users u ON v.instructor_id = u.id
		WHERE v.is_public = true AND v.status = $1
		ORDER BY v.view_count DESC, v.rating DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, entities.VideoStatusReady, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get featured videos: %w", err)
	}
	defer rows.Close()

	videos := []*entities.Video{}
	for rows.Next() {
		video, err := r.scanVideoWithInstructor(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate videos: %w", err)
	}

	return videos, nil
}

func (r *VideoRepository) GetPopularVideos(ctx context.Context, limit int) ([]*entities.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.instructor_id, v.category, v.tags, 
		       v.duration, v.thumbnail, v.status, v.is_public, v.view_count, v.rating,
		       v.created_at, v.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.avatar, u.created_at
		FROM videos v
		LEFT JOIN users u ON v.instructor_id = u.id
		WHERE v.is_public = true AND v.status = $1 AND v.created_at > NOW() - INTERVAL '30 days'
		ORDER BY v.view_count DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, entities.VideoStatusReady, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular videos: %w", err)
	}
	defer rows.Close()

	videos := []*entities.Video{}
	for rows.Next() {
		video, err := r.scanVideoWithInstructor(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate videos: %w", err)
	}

	return videos, nil
}

// Video Files methods
func (r *VideoRepository) CreateVideoFile(ctx context.Context, videoFile *entities.VideoFile) error {
	query := `
		INSERT INTO video_files (id, video_id, quality, file_path, file_size, bitrate, format)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	videoFile.ID = uuid.New()

	_, err := r.db.ExecContext(ctx, query,
		videoFile.ID,
		videoFile.VideoID,
		videoFile.Quality,
		videoFile.FilePath,
		videoFile.FileSize,
		videoFile.Bitrate,
		videoFile.Format,
	)

	if err != nil {
		return fmt.Errorf("failed to create video file: %w", err)
	}

	return nil
}

func (r *VideoRepository) GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entities.VideoFile, error) {
	query := `
		SELECT id, video_id, quality, file_path, file_size, bitrate, format
		FROM video_files
		WHERE video_id = $1
		ORDER BY 
			CASE quality
				WHEN '1080p' THEN 1
				WHEN '720p' THEN 2
				WHEN '480p' THEN 3
				WHEN '360p' THEN 4
				ELSE 5
			END
	`

	rows, err := r.db.QueryContext(ctx, query, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video files: %w", err)
	}
	defer rows.Close()

	videoFiles := []*entities.VideoFile{}
	for rows.Next() {
		videoFile := &entities.VideoFile{}
		err := rows.Scan(
			&videoFile.ID,
			&videoFile.VideoID,
			&videoFile.Quality,
			&videoFile.FilePath,
			&videoFile.FileSize,
			&videoFile.Bitrate,
			&videoFile.Format,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video file: %w", err)
		}
		videoFiles = append(videoFiles, videoFile)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate video files: %w", err)
	}

	return videoFiles, nil
}

func (r *VideoRepository) UpdateVideoFile(ctx context.Context, videoFile *entities.VideoFile) error {
	query := `
		UPDATE video_files
		SET quality = $2, file_path = $3, file_size = $4, bitrate = $5, format = $6
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		videoFile.ID,
		videoFile.Quality,
		videoFile.FilePath,
		videoFile.FileSize,
		videoFile.Bitrate,
		videoFile.Format,
	)

	if err != nil {
		return fmt.Errorf("failed to update video file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrRecordNotFound
	}

	return nil
}

func (r *VideoRepository) DeleteVideoFile(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM video_files WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete video file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrRecordNotFound
	}

	return nil
}

// Statistics methods
func (r *VideoRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE videos SET view_count = view_count + 1, updated_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

func (r *VideoRepository) UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error {
	query := `UPDATE videos SET rating = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, rating, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	return nil
}

func (r *VideoRepository) GetVideoStats(ctx context.Context, id uuid.UUID) (*entities.VideoStats, error) {
	// This would be implemented with a proper stats table
	stats := &entities.VideoStats{
		VideoID:   id,
		Views:     0,
		Likes:     0,
		Comments:  0,
		Shares:    0,
		WatchTime: 0,
	}

	return stats, nil
}

func (r *VideoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus) error {
	query := `UPDATE videos SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.VideoNotFound
	}

	return nil
}

func (r *VideoRepository) GetVideosByStatus(ctx context.Context, status entities.VideoStatus) ([]*entities.Video, error) {
	query := `
		SELECT v.id, v.title, v.description, v.instructor_id, v.category, v.tags, 
		       v.duration, v.thumbnail, v.status, v.is_public, v.view_count, v.rating,
		       v.created_at, v.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.avatar, u.created_at
		FROM videos v
		LEFT JOIN users u ON v.instructor_id = u.id
		WHERE v.status = $1
		ORDER BY v.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos by status: %w", err)
	}
	defer rows.Close()

	videos := []*entities.Video{}
	for rows.Next() {
		video, err := r.scanVideoWithInstructor(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan video: %w", err)
		}
		videos = append(videos, video)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate videos: %w", err)
	}

	return videos, nil
}

// Helper method to scan video with instructor
func (r *VideoRepository) scanVideoWithInstructor(rows *sql.Rows) (*entities.Video, error) {
	video := &entities.Video{}
	var instructor entities.UserProfile
	var instructorID sql.NullString
	var instructorEmail sql.NullString
	var instructorFirstName sql.NullString
	var instructorLastName sql.NullString
	var instructorRole sql.NullString
	var instructorAvatar sql.NullString
	var instructorCreatedAt sql.NullTime

	err := rows.Scan(
		&video.ID,
		&video.Title,
		&video.Description,
		&video.InstructorID,
		&video.Category,
		pq.Array(&video.Tags),
		&video.Duration,
		&video.Thumbnail,
		&video.Status,
		&video.IsPublic,
		&video.ViewCount,
		&video.Rating,
		&video.CreatedAt,
		&video.UpdatedAt,
		&instructorID,
		&instructorEmail,
		&instructorFirstName,
		&instructorLastName,
		&instructorRole,
		&instructorAvatar,
		&instructorCreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Set instructor if exists
	if instructorID.Valid {
		instructor.ID = uuid.MustParse(instructorID.String)
		instructor.Email = instructorEmail.String
		instructor.FirstName = instructorFirstName.String
		instructor.LastName = instructorLastName.String
		instructor.Role = instructorRole.String
		instructor.Avatar = instructorAvatar.String
		if instructorCreatedAt.Valid {
			instructor.CreatedAt = instructorCreatedAt.Time
		}
		video.Instructor = &instructor
	}

	return video, nil
}
