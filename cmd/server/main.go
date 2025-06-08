package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
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
)

func main() {
	// Configuración
	cfg := config.Load()
	
	// Logger
	log := logger.NewLogger(cfg.LogLevel)
	
	// Conexiones a base de datos
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Error connecting to PostgreSQL: %v", err)
	}
	defer db.Close()

	redisClient, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("Error connecting to Redis: %v", err)
	}
	defer redisClient.Close()

	// Repositorios
	userRepo := postgres.NewUserRepository(db)
	videoRepo := postgres.NewVideoRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)

	// Use cases
	authUsecase := auth.NewAuthUsecase(userRepo, cacheRepo, cfg.JWTSecret)
	userUsecase := user.NewUserUsecase(userRepo, cacheRepo)
	videoUsecase := video.NewVideoUsecase(videoRepo, cacheRepo)
	streamingUsecase := streaming.NewStreamingUsecase(videoRepo, cacheRepo, cfg.CDNBaseURL)

	// Handlers
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
	protected.HandleFunc("/videos/upload", videoHandler.UploadVideo).Methods("POST")
	protected.HandleFunc("/videos/{id}/edit", videoHandler.UpdateVideo).Methods("PUT")
	protected.HandleFunc("/videos/{id}/delete", videoHandler.DeleteVideo).Methods("DELETE")
	
	// Streaming routes
	protected.HandleFunc("/stream/{id}/playlist", streamingHandler.GetHLSPlaylist).Methods("GET")
	protected.HandleFunc("/stream/{id}/segment/{segment}", streamingHandler.GetHLSSegment).Methods("GET")

	// Servir archivos estáticos
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	// Servidor HTTP
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + cfg.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: %v", err)
		}
	}()

	// Esperar señal de interrupción
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: %v", err)
	}
	
	log.Info("Server exited")
}

