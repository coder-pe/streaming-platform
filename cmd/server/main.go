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

	httpHandlers "streaming-platform/internal/adapters/input/http"
	"streaming-platform/internal/adapters/input/http/middleware"
	postgresAdapter "streaming-platform/internal/adapters/output/persistence/postgres"
	redisAdapter "streaming-platform/internal/adapters/output/persistence/redis"
	"streaming-platform/internal/core/services"
	"streaming-platform/internal/infrastructure/workers"
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

	// Output Adapters (Repositorios) - Implementan los puertos de salida
	userRepo := postgresAdapter.NewUserRepository(db)
	videoRepo := postgresAdapter.NewVideoRepository(db)
	cacheRepo := redisAdapter.NewCacheRepository(redisClient)

	// Core Services - Implementan los puertos de entrada, usan los puertos de salida
	authService := services.NewAuthService(userRepo, cacheRepo, cfg.JWTSecret)
	userService := services.NewUserService(userRepo, cacheRepo)
	videoService := services.NewVideoService(videoRepo, cacheRepo)
	streamingService := services.NewStreamingService(videoRepo, cacheRepo, cfg.CDNBaseURL, cfg.JWTSecret)

	// Input Adapters (Handlers HTTP) - Usan los servicios del core
	authHandler := httpHandlers.NewAuthHandler(authService, log)
	userHandler := httpHandlers.NewUserHandler(userService, log)
	videoHandler := httpHandlers.NewVideoHandler(videoService, log)
	streamingHandler := httpHandlers.NewStreamingHandler(streamingService, log)

	// Worker Pool
	// TODO: Actualizar worker para usar los servicios del core
	workerPool := workers.NewWorkerPool(cfg.WorkerPoolSize, log)
	// transcodingWorker := workers.NewTranscodingWorker(videoService, cfg.FFmpegPath, cfg.StoragePath)
	// workerPool.RegisterWorker("transcoding", transcodingWorker)
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
