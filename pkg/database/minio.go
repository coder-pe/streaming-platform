// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package database

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient wraps the MinIO client
type MinIOClient struct {
	client     *minio.Client
	bucketName string
}

// NewMinIO creates a new MinIO client connection
func NewMinIO(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinIOClient, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	client := &MinIOClient{
		client:     minioClient,
		bucketName: bucketName,
	}

	// Verify connection and create bucket if it doesn't exist
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket: %w", err)
	}

	return client, nil
}

// ensureBucket creates the bucket if it doesn't exist
func (m *MinIOClient) ensureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile uploads a file to MinIO
func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.client.PutObject(
		ctx,
		m.bucketName,
		objectName,
		reader,
		objectSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from MinIO
func (m *MinIOClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(
		ctx,
		m.bucketName,
		objectName,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

// DeleteFile deletes a file from MinIO
func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
	err := m.client.RemoveObject(
		ctx,
		m.bucketName,
		objectName,
		minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists in MinIO
func (m *MinIOClient) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.client.StatObject(
		ctx,
		m.bucketName,
		objectName,
		minio.StatObjectOptions{},
	)
	if err != nil {
		// Check if error is "object not found"
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// GetFileInfo returns information about a file
func (m *MinIOClient) GetFileInfo(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	info, err := m.client.StatObject(
		ctx,
		m.bucketName,
		objectName,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	return info, nil
}

// ListFiles lists all files with a given prefix
func (m *MinIOClient) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing files: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

// GetPresignedURL generates a presigned URL for temporary access to a file
func (m *MinIOClient) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(
		ctx,
		m.bucketName,
		objectName,
		expiry,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// CopyFile copies a file within MinIO
func (m *MinIOClient) CopyFile(ctx context.Context, sourceObject, destObject string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: m.bucketName,
		Object: sourceObject,
	}

	dstOpts := minio.CopyDestOptions{
		Bucket: m.bucketName,
		Object: destObject,
	}

	_, err := m.client.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// DeleteMultipleFiles deletes multiple files
func (m *MinIOClient) DeleteMultipleFiles(ctx context.Context, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{
				Key: name,
			}
		}
	}()

	errorCh := m.client.RemoveObjects(ctx, m.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete file %s: %w", err.ObjectName, err.Err)
		}
	}

	return nil
}

// HealthCheck performs a health check on the MinIO connection
func (m *MinIOClient) HealthCheck(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", m.bucketName)
	}

	return nil
}

// GetClient returns the underlying MinIO client
func (m *MinIOClient) GetClient() *minio.Client {
	return m.client
}

// GetBucketName returns the bucket name
func (m *MinIOClient) GetBucketName() string {
	return m.bucketName
}
