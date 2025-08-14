// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"streaming-platform/internal/delivery/http/handlers"
	"streaming-platform/internal/delivery/http/middleware"
	"streaming-platform/internal/delivery/workers"
	"streaming-platform/internal/repository/postgres"
	"streaming-platform/internal/repository/redis"
	"streaming-platform/internal/usecase/auth"
	"streaming-platform/internal/usecase/streaming"
	"streaming-platform/internal/usecase/user"
	"streaming-platform/internal/usecase/video"
	"streaming-platform/pkg/config"
	"streaming-platform/pkg/database"
	"streaming-platform/pkg/logger"

	"github.com/gorilla/mux"
)

func main() {
	// Configuración
	cfg := config.Load()

	// Logger
	log := logger.NewLogger()

	// Conexiones a base de datos
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Error("Error connecting to PostgreSQL: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Error("Error connecting to Redis: %v", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Repositorios
	userRepo := postgres.NewUserRepository(db)
	videoRepo := postgres.NewVideoRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Use cases - Ahora devuelven interfaces, no punteros
	authUsecase := auth.NewAuthUsecase(userRepo, cacheRepo, cfg.JWTSecret)
	userUsecase := user.NewUserUsecase(userRepo, cacheRepo)
	videoUsecase := video.NewVideoUsecase(videoRepo, cacheRepo)
	streamingUsecase := streaming.NewStreamingUsecase(videoRepo, cacheRepo, cfg.CDNBaseURL, cfg.JWTSecret)

	// Handlers - Reciben interfaces directamente
	authHandler := handlers.NewAuthHandler(authUsecase, log)
	userHandler := handlers.NewUserHandler(userUsecase, log)
	videoHandler := handlers.NewVideoHandler(videoUsecase, log)
	streamingHandler := handlers.NewStreamingHandler(streamingUsecase, log)

	// Worker Pool
	workerPool := workers.NewWorkerPool(cfg.WorkerPoolSize, log)
	transcodingWorker := workers.NewTranscodingWorker(videoUsecase, cfg.FFmpegPath, cfg.StoragePath)
	workerPool.RegisterWorker("transcoding", transcodingWorker)
	workerPool.Start()
	defer workerPool.Stop()

	// Router
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(log))
	router.Use(middleware.RateLimit(redisClient))

	// Rutas públicas
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/videos", videoHandler.GetPublicVideos).Methods("GET")
	router.HandleFunc("/api/videos/{id}", videoHandler.GetVideo).Methods("GET")

	// Rutas protegidas
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))

	// User routes
	protected.HandleFunc("/user/profile", userHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/user/profile", userHandler.UpdateProfile).Methods("PUT")

	// Video routes
	protected.HandleFunc("/videos", videoHandler.UploadVideo).Methods("POST")
	protected.HandleFunc("/videos/{id}", videoHandler.UpdateVideo).Methods("PUT")
	protected.HandleFunc("/videos/{id}", videoHandler.DeleteVideo).Methods("DELETE")

	// Streaming routes
	protected.HandleFunc("/stream/{id}/master.m3u8", streamingHandler.GetMasterPlaylist).Methods("GET")
	protected.HandleFunc("/stream/{id}/{quality}/playlist.m3u8", streamingHandler.GetPlaylist).Methods("GET")
	protected.HandleFunc("/stream/{id}/{quality}/{segment}", streamingHandler.GetSegment).Methods("GET")

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Info("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Server is shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
