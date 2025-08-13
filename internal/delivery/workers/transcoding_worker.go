// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package workers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"streaming-platform/internal/domain/entities"
	"streaming-platform/internal/usecase/video"
	"streaming-platform/pkg/logger"
)

type TranscodingWorker struct {
	videoUsecase video.VideoUsecase
	ffmpegPath   string
	storagePath  string
	logger       logger.Logger
}

func NewTranscodingWorker(videoUsecase video.VideoUsecase, ffmpegPath, storagePath string) *TranscodingWorker {
	return &TranscodingWorker{
		videoUsecase: videoUsecase,
		ffmpegPath:   ffmpegPath,
		storagePath:  storagePath,
		logger:       logger.NewLogger("info"),
	}
}

func (w *TranscodingWorker) ProcessJob(ctx context.Context, job map[string]interface{}) error {
	videoIDStr, ok := job["video_id"].(string)
	if !ok {
		return fmt.Errorf("invalid video_id in job")
	}

	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		return fmt.Errorf("invalid video_id format: %v", err)
	}

	filePath, ok := job["file_path"].(string)
	if !ok {
		return fmt.Errorf("invalid file_path in job")
	}

	w.logger.Info("Starting transcoding for video %s", videoID)

	// Crear directorio de salida
	outputDir := filepath.Join(w.storagePath, "videos", videoID.String())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Obtener información del video
	videoInfo, err := w.getVideoInfo(filePath)
	if err != nil {
		return fmt.Errorf("error getting video info: %v", err)
	}

	// Definir perfiles de calidad
	profiles := []QualityProfile{
		{Quality: "1080p", Width: 1920, Height: 1080, Bitrate: "5000k", AudioBitrate: "128k"},
		{Quality: "720p", Width: 1280, Height: 720, Bitrate: "2500k", AudioBitrate: "128k"},
		{Quality: "480p", Width: 854, Height: 480, Bitrate: "1000k", AudioBitrate: "96k"},
		{Quality: "360p", Width: 640, Height: 360, Bitrate: "500k", AudioBitrate: "64k"},
	}

	var videoFiles []entities.VideoFile
	var masterPlaylistContent strings.Builder
	masterPlaylistContent.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n\n")

	// Transcodificar cada perfil
	for _, profile := range profiles {
		// Solo transcodificar si la resolución original es mayor o igual
		if videoInfo.Width < profile.Width {
			continue
		}

		hlsDir := filepath.Join(outputDir, "hls", profile.Quality)
		if err := os.MkdirAll(hlsDir, 0755); err != nil {
			w.logger.Error("Error creating HLS directory: %v", err)
			continue
		}

		playlistPath := filepath.Join(hlsDir, "playlist.m3u8")
		
		if err := w.transcodeToHLS(filePath, playlistPath, profile); err != nil {
			w.logger.Error("Error transcoding %s: %v", profile.Quality, err)
			continue
		}

		// Calcular tamaño total de archivos
		fileSize, err := w.calculateHLSSize(hlsDir)
		if err != nil {
			w.logger.Error("Error calculating size for %s: %v", profile.Quality, err)
			fileSize = 0
		}

		// Crear registro de archivo de video
		videoFile := entities.VideoFile{
			ID:       uuid.New(),
			VideoID:  videoID,
			Quality:  profile.Quality,
			FilePath: filepath.Join("videos", videoID.String(), "hls", profile.Quality, "playlist.m3u8"),
			FileSize: fileSize,
			Bitrate:  parseBitrate(profile.Bitrate),
			Format:   "hls",
		}

		videoFiles = append(videoFiles, videoFile)

		// Agregar al master playlist
		bandwidth := parseBitrate(profile.Bitrate) * 1000 // convertir a bps
		masterPlaylistContent.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", 
			bandwidth, profile.Width, profile.Height))
		masterPlaylistContent.WriteString(fmt.Sprintf("hls/%s/playlist.m3u8\n", profile.Quality))
	}

	// Crear master playlist
	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
	if err := os.WriteFile(masterPlaylistPath, []byte(masterPlaylistContent.String()), 0644); err != nil {
		return fmt.Errorf("error creating master playlist: %v", err)
	}

	// Generar thumbnail
	thumbnailPath := filepath.Join(outputDir, "thumbnail.jpg")
	if err := w.generateThumbnail(filePath, thumbnailPath); err != nil {
		w.logger.Error("Error generating thumbnail: %v", err)
	}

	// Actualizar video en base de datos
	video, err := w.videoUsecase.GetVideoByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("error getting video: %v", err)
	}

	video.Duration = videoInfo.Duration
	video.Status = entities.VideoStatusReady
	video.Thumbnail = filepath.Join("videos", videoID.String(), "thumbnail.jpg")
	video.VideoFiles = videoFiles

	if err := w.videoUsecase.UpdateVideoWithFiles(ctx, video); err != nil {
		return fmt.Errorf("error updating video: %v", err)
	}

	// Limpiar archivo temporal
	if err := os.Remove(filePath); err != nil {
		w.logger.Error("Error removing temp file: %v", err)
	}

	w.logger.Info("Transcoding completed for video %s", videoID)
	return nil
}

type QualityProfile struct {
	Quality      string
	Width        int
	Height       int
	Bitrate      string
	AudioBitrate string
}

type VideoInfo struct {
	Duration int // en segundos
	Width    int
	Height   int
	Bitrate  int
}

func (w *TranscodingWorker) getVideoInfo(filePath string) (*VideoInfo, error) {
	cmd := exec.Command(w.ffmpegPath, "-i", filePath, "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil, fmt.Errorf("ffmpeg should return error for null output")
	}

	// Parsear output de ffmpeg
	outputStr := string(output)
	info := &VideoInfo{}

	// Extraer duración
	if strings.Contains(outputStr, "Duration:") {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Duration:") {
				parts := strings.Split(line, "Duration: ")
				if len(parts) > 1 {
					timeStr := strings.Split(parts[1], ",")[0]
					duration, _ := w.parseDuration(timeStr)
					info.Duration = duration
				}
				break
			}
		}
	}

	// Extraer resolución
	if strings.Contains(outputStr, "Video:") {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Video:") && strings.Contains(line, "x") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.Contains(part, "x") && strings.Contains(part, "0") {
						resParts := strings.Split(part, "x")
						if len(resParts) == 2 {
							width, _ := strconv.Atoi(resParts[0])
							height, _ := strconv.Atoi(strings.Split(resParts[1], ",")[0])
							info.Width = width
							info.Height = height
						}
						break
					}
				}
				break
			}
		}
	}

	return info, nil
}

func (w *TranscodingWorker) transcodeToHLS(inputPath, playlistPath string, profile QualityProfile) error {
	segmentDir := filepath.Dir(playlistPath)
	segmentPattern := filepath.Join(segmentDir, "segment_%03d.ts")

	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-c:a", "aac",
		"-b:v", profile.Bitrate,
		"-b:a", profile.AudioBitrate,
		"-vf", fmt.Sprintf("scale=%d:%d", profile.Width, profile.Height),
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	}

	cmd := exec.Command(w.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}

	return nil
}

func (w *TranscodingWorker) generateThumbnail(inputPath, outputPath string) error {
	args := []string{
		"-i", inputPath,
		"-ss", "00:00:05", // 5 segundos desde el inicio
		"-vframes", "1",
		"-vf", "scale=640:360",
		"-q:v", "2",
		outputPath,
	}

	cmd := exec.Command(w.ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("thumbnail generation error: %v, output: %s", err, string(output))
	}

	return nil
}

func (w *TranscodingWorker) calculateHLSSize(hlsDir string) (int64, error) {
	var totalSize int64
	
	err := filepath.Walk(hlsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	
	return totalSize, err
}

func (w *TranscodingWorker) parseDuration(timeStr string) (int, error) {
	// Formato: HH:MM:SS.mmm
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	totalSeconds := hours*3600 + minutes*60 + int(seconds)
	return totalSeconds, nil
}

func parseBitrate(bitrateStr string) int {
	// Remover 'k' y convertir a int
	bitrateStr = strings.TrimSuffix(bitrateStr, "k")
	bitrate, _ := strconv.Atoi(bitrateStr)
	return bitrate
}
