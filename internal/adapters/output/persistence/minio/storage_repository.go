// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"streaming-platform/internal/core/ports/output"
	"streaming-platform/pkg/database"
	"streaming-platform/pkg/logger"
)

type storageRepository struct {
	client *database.MinIOClient
	logger logger.Logger
}

// NewStorageRepository crea una nueva instancia del repositorio de almacenamiento
func NewStorageRepository(client *database.MinIOClient, log logger.Logger) output.StorageRepository {
	return &storageRepository{
		client: client,
		logger: log,
	}
}

func (r *storageRepository) UploadFile(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error {
	r.logger.Info("Uploading file to MinIO: %s (size: %d bytes)", path, size)

	err := r.client.UploadFile(ctx, path, reader, size, contentType)
	if err != nil {
		return fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	r.logger.Info("File uploaded successfully: %s", path)
	return nil
}

func (r *storageRepository) DownloadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	r.logger.Info("Downloading file from MinIO: %s", path)

	reader, err := r.client.DownloadFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from MinIO: %w", err)
	}

	return reader, nil
}

func (r *storageRepository) DeleteFile(ctx context.Context, path string) error {
	r.logger.Info("Deleting file from MinIO: %s", path)

	err := r.client.DeleteFile(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	r.logger.Info("File deleted successfully: %s", path)
	return nil
}

func (r *storageRepository) FileExists(ctx context.Context, path string) (bool, error) {
	exists, err := r.client.FileExists(ctx, path)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return exists, nil
}

func (r *storageRepository) GetFileInfo(ctx context.Context, path string) (*output.FileInfo, error) {
	info, err := r.client.GetFileInfo(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &output.FileInfo{
		Name:         info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		ETag:         info.ETag,
	}, nil
}

func (r *storageRepository) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	files, err := r.client.ListFiles(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

func (r *storageRepository) GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	url, err := r.client.GetPresignedURL(ctx, path, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	r.logger.Info("Generated presigned URL for: %s (expires in %v)", path, expiry)
	return url, nil
}

func (r *storageRepository) CopyFile(ctx context.Context, sourcePath, destPath string) error {
	r.logger.Info("Copying file in MinIO: %s -> %s", sourcePath, destPath)

	err := r.client.CopyFile(ctx, sourcePath, destPath)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	r.logger.Info("File copied successfully")
	return nil
}

func (r *storageRepository) DeleteMultipleFiles(ctx context.Context, paths []string) error {
	r.logger.Info("Deleting %d files from MinIO", len(paths))

	err := r.client.DeleteMultipleFiles(ctx, paths)
	if err != nil {
		return fmt.Errorf("failed to delete multiple files: %w", err)
	}

	r.logger.Info("Files deleted successfully: %d", len(paths))
	return nil
}

func (r *storageRepository) HealthCheck(ctx context.Context) error {
	return r.client.HealthCheck(ctx)
}
