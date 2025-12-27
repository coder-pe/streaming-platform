// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package output

import (
	"context"
	"io"
	"time"
)

// FileInfo contiene información sobre un archivo almacenado
type FileInfo struct {
	Name         string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}

// StorageRepository define el puerto de salida para almacenamiento de archivos
type StorageRepository interface {
	// UploadFile sube un archivo al almacenamiento
	UploadFile(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error

	// DownloadFile descarga un archivo del almacenamiento
	DownloadFile(ctx context.Context, path string) (io.ReadCloser, error)

	// DeleteFile elimina un archivo del almacenamiento
	DeleteFile(ctx context.Context, path string) error

	// FileExists verifica si un archivo existe
	FileExists(ctx context.Context, path string) (bool, error)

	// GetFileInfo obtiene información sobre un archivo
	GetFileInfo(ctx context.Context, path string) (*FileInfo, error)

	// ListFiles lista archivos con un prefijo dado
	ListFiles(ctx context.Context, prefix string) ([]string, error)

	// GetPresignedURL genera una URL temporal para acceder a un archivo
	GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)

	// CopyFile copia un archivo dentro del almacenamiento
	CopyFile(ctx context.Context, sourcePath, destPath string) error

	// DeleteMultipleFiles elimina múltiples archivos
	DeleteMultipleFiles(ctx context.Context, paths []string) error

	// HealthCheck verifica la salud de la conexión
	HealthCheck(ctx context.Context) error
}
