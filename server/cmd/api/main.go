package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/config"
	"github.com/praxis-social/praxis/server/internal/handler"
	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/provider"
	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/service"
	"github.com/praxis-social/praxis/server/internal/ws"
)

func main() {
	cfg := config.Load()

	var logger *zap.Logger
	var err error
	if cfg.IsDevelopment() {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to PostgreSQL
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}
	logger.Info("connected to database")

	// Connect to Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal("failed to parse redis URL", zap.Error(err))
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to ping redis", zap.Error(err))
	}
	logger.Info("connected to redis")

	// Set up router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(logger))
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(middleware.CORSConfig(cfg.FrontendURL)))

	// Health check
	r.Get("/health", handler.HealthCheck())

	// Repositories
	userRepo := repository.NewUserRepo(pool)
	achievementRepo := repository.NewAchievementRepo(pool)
	providerRepo := repository.NewProviderRepo(pool)
	followRepo := repository.NewFollowRepo(pool)
	messageRepo := repository.NewMessageRepo(pool)
	jobRepo := repository.NewJobRepo(pool)
	notificationRepo := repository.NewNotificationRepo(pool)

	// Provider registry
	providerRegistry := provider.NewRegistry()
	providerRegistry.Register(provider.NewGitHubProvider())

	// WebSocket hub
	hub := ws.NewHub(logger)
	go hub.Run()

	// Services
	authService := service.NewAuthService(userRepo, rdb, cfg)
	achievementService := service.NewAchievementService(achievementRepo, providerRepo, userRepo, providerRegistry, logger)
	userService := service.NewUserService(userRepo, followRepo, achievementRepo, providerRepo)
	messageService := service.NewMessageService(messageRepo, userRepo, hub, logger)
	jobService := service.NewJobService(jobRepo, achievementRepo, logger)
	notificationService := service.NewNotificationService(notificationRepo, hub, logger)

	// Handlers
	authHandler := handler.NewAuthHandler(authService, userRepo)
	achievementHandler := handler.NewAchievementHandler(achievementService)
	feedHandler := handler.NewFeedHandler(achievementService)
	userHandler := handler.NewUserHandler(userService)
	messageHandler := handler.NewMessageHandler(messageService)
	jobHandler := handler.NewJobHandler(jobService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	wsHandler := handler.NewWebSocketHandler(hub, authService, logger)

	// WebSocket route (uses query param token, not auth middleware)
	r.Get("/ws", wsHandler.HandleWS)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Get("/github", authHandler.GitHubLogin)
			r.Post("/github/callback", authHandler.GitHubCallback)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(authService))
				r.Get("/me", authHandler.Me)
			})
		})

		// User profile (public, with optional auth for isFollowing)
		r.Group(func(r chi.Router) {
			r.Use(middleware.TryAuthMiddleware(authService))
			r.Get("/users/{username}", userHandler.GetProfile)
		})

		// Job routes (public listing and detail)
		r.Group(func(r chi.Router) {
			r.Use(middleware.TryAuthMiddleware(authService))
			r.Get("/jobs", jobHandler.ListJobs)
			r.Get("/jobs/{id}", jobHandler.GetJob)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(authService))

			// User routes
			r.Patch("/users/me", userHandler.UpdateProfile)
			r.Post("/users/{username}/follow", userHandler.Follow)
			r.Delete("/users/{username}/follow", userHandler.Unfollow)

			// Achievement routes
			r.Route("/achievements", func(r chi.Router) {
				r.Post("/sync", achievementHandler.Sync)
				r.Get("/{id}", achievementHandler.GetByID)
				r.Post("/{id}/reactions", achievementHandler.AddReaction)
				r.Delete("/{id}/reactions", achievementHandler.RemoveReaction)
			})

			// Feed route
			r.Get("/feed", feedHandler.GetFeed)

			// Job routes (protected)
			r.Post("/jobs", jobHandler.CreateJob)
			r.Post("/jobs/{id}/apply", jobHandler.Apply)

			// Notification routes
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notificationHandler.ListNotifications)
				r.Get("/unread-count", notificationHandler.UnreadCount)
				r.Patch("/{id}/read", notificationHandler.MarkRead)
			})

			// Conversation / messaging routes
			r.Route("/conversations", func(r chi.Router) {
				r.Get("/", messageHandler.ListConversations)
				r.Post("/", messageHandler.CreateConversation)
				r.Get("/{id}/messages", messageHandler.ListMessages)
				r.Post("/{id}/messages", messageHandler.SendMessage)
				r.Patch("/{id}/read", messageHandler.MarkRead)
			})
		})
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("server starting", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-done
	logger.Info("server shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
