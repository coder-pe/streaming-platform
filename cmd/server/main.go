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
	elasticsearchAdapter "streaming-platform/internal/adapters/output/persistence/elasticsearch"
	minioAdapter "streaming-platform/internal/adapters/output/persistence/minio"
	postgresAdapter "streaming-platform/internal/adapters/output/persistence/postgres"
	rabbitmqAdapter "streaming-platform/internal/adapters/output/persistence/rabbitmq"
	redisAdapter "streaming-platform/internal/adapters/output/persistence/redis"
	"streaming-platform/internal/core/ports/output"
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

	// RabbitMQ connection
	rabbitMQClient, err := database.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Error("Error connecting to RabbitMQ: %v", err)
		os.Exit(1)
	}
	defer rabbitMQClient.Close()

	// MinIO connection
	minioClient, err := database.NewMinIO(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Error("Error connecting to MinIO: %v", err)
		os.Exit(1)
	}

	// Elasticsearch connection
	esClient, err := database.NewElasticsearch([]string{cfg.ElasticsearchURL})
	if err != nil {
		log.Error("Error connecting to Elasticsearch: %v", err)
		os.Exit(1)
	}

	// Output Adapters (Repositorios) - Implementan los puertos de salida
	userRepo := postgresAdapter.NewUserRepository(db)
	videoRepo := postgresAdapter.NewVideoRepository(db)
	cacheRepo := redisAdapter.NewCacheRepository(redisClient)
	queueRepo := rabbitmqAdapter.NewQueueRepository(rabbitMQClient, log)
	_ = minioAdapter.NewStorageRepository(minioClient, log) // Preparado para uso futuro (Fase 3)
	searchRepo := elasticsearchAdapter.NewSearchRepository(esClient, log)

	// Initialize Elasticsearch index
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := searchRepo.InitializeIndex(ctx); err != nil {
		log.Error("Warning: Failed to initialize Elasticsearch index: %v", err)
		// No salir, solo registrar el error
	}
	cancel()

	// Core Services - Implementan los puertos de entrada, usan los puertos de salida
	authService := services.NewAuthService(userRepo, cacheRepo, cfg.JWTSecret)
	userService := services.NewUserService(userRepo, cacheRepo)
	videoService := services.NewVideoService(videoRepo, cacheRepo)
	streamingService := services.NewStreamingService(videoRepo, cacheRepo, cfg.CDNBaseURL, cfg.JWTSecret)

	// Worker Pool - Debe crearse antes de los handlers
	workerPool := workers.NewWorkerPool(cfg.WorkerPoolSize, log)
	transcodingWorker := workers.NewTranscodingWorker(videoService, cfg.FFmpegPath, cfg.StoragePath)
	workerPool.RegisterWorker("transcoding", transcodingWorker)
	workerPool.Start()
	defer workerPool.Stop()

	// Iniciar consumo de trabajos desde RabbitMQ
	go func() {
		err := queueRepo.ConsumeJobs(context.Background(), "transcoding_queue", func(job output.JobMessage) error {
			return workerPool.ProcessJobFromQueue(job)
		})
		if err != nil {
			log.Error("Error starting job consumer: %v", err)
		}
	}()

	// Input Adapters (Handlers HTTP) - Usan los servicios del core
	authHandler := httpHandlers.NewAuthHandler(authService, log)
	userHandler := httpHandlers.NewUserHandler(userService, log)
	videoHandler := httpHandlers.NewVideoHandler(videoService, queueRepo, cfg.StoragePath, log)
	streamingHandler := httpHandlers.NewStreamingHandler(streamingService, log)

	// Router
	router := mux.NewRouter()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(log))
	router.Use(middleware.RateLimit(redisClient))
	router.Use(middleware.CacheControl()) // Cache control para archivos estáticos

	// Rutas públicas
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods("POST")
	router.HandleFunc("/api/videos", videoHandler.GetPublicVideos).Methods("GET")
	router.HandleFunc("/api/videos/{id}", videoHandler.GetVideo).Methods("GET")
	router.HandleFunc("/api/users/{id}/stats", userHandler.GetUserStats).Methods("GET")
	router.HandleFunc("/api/users/{id}", userHandler.GetPublicProfile).Methods("GET")

	// Rutas protegidas
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))

	// User routes
	protected.HandleFunc("/user/profile", userHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/user/profile", userHandler.UpdateProfile).Methods("PUT")
	protected.HandleFunc("/user/stats", userHandler.GetUserStats).Methods("GET")
	protected.HandleFunc("/users/{id}/avatar", userHandler.UploadAvatar).Methods("POST")

	// Video routes
	protected.HandleFunc("/videos", videoHandler.UploadVideo).Methods("POST")
	protected.HandleFunc("/videos/{id}", videoHandler.UpdateVideo).Methods("PUT")
	protected.HandleFunc("/videos/{id}", videoHandler.DeleteVideo).Methods("DELETE")

	// Streaming routes
	protected.HandleFunc("/stream/{id}/master.m3u8", streamingHandler.GetMasterPlaylist).Methods("GET")
	protected.HandleFunc("/stream/{id}/{quality}/playlist.m3u8", streamingHandler.GetPlaylist).Methods("GET")
	protected.HandleFunc("/stream/{id}/{quality}/{segment}", streamingHandler.GetSegment).Methods("GET")

	// Static files y SPA fallback
	// Servir archivos estáticos primero
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// SPA fallback - sirve index.html para todas las rutas que no son API o archivos estáticos
	// Esto permite que el router del frontend (History API) maneje las rutas
	spaHandler := middleware.SPAHandler("./web", "./web/index.html")
	router.PathPrefix("/").Handler(spaHandler(http.FileServer(http.Dir("./web/"))))

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
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
