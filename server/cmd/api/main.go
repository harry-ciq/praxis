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
	"github.com/go-chi/httprate"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/config"
	"github.com/praxis-social/praxis/server/internal/handler"
	"github.com/praxis-social/praxis/server/internal/llm"
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
	// Global rate limit: 100 req/min per IP. Keeps generic abuse out without
	// being too aggressive on real usage.
	r.Use(httprate.LimitByIP(100, time.Minute))

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
	experienceRepo := repository.NewExperienceRepo(pool)
	skillRepo := repository.NewSkillRepo(pool)

	// Provider registry
	providerRegistry := provider.NewRegistry()
	providerRegistry.Register(provider.NewGitHubProvider())
	providerRegistry.Register(provider.NewYouTubeProvider())

	// WebSocket hub
	hub := ws.NewHub(logger)
	go hub.Run()

	// Asynq client for enqueueing background tasks
	asynqRedisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		logger.Fatal("failed to parse redis URI for asynq", zap.Error(err))
	}
	asynqClient := asynq.NewClient(asynqRedisOpt)
	defer asynqClient.Close()

	// Services
	authService := service.NewAuthService(userRepo, rdb, cfg)
	achievementService := service.NewAchievementService(achievementRepo, providerRepo, userRepo, skillRepo, providerRegistry, logger)
	llmClient := llm.NewClient(cfg.AnthropicAPIKey, cfg.SmartFeedModel, logger)
	smartFeedService := service.NewSmartFeedService(achievementRepo, userRepo, llmClient, rdb, cfg.SmartFeedCacheTTL, logger)
	userService := service.NewUserService(userRepo, followRepo, achievementRepo, providerRepo, experienceRepo, skillRepo)
	messageService := service.NewMessageService(messageRepo, userRepo, hub, logger)
	jobService := service.NewJobService(jobRepo, achievementRepo, logger)
	notificationService := service.NewNotificationService(notificationRepo, hub, logger)

	// Handlers
	authHandler := handler.NewAuthHandler(authService, userRepo)
	achievementHandler := handler.NewAchievementHandler(achievementService)
	feedHandler := handler.NewFeedHandler(achievementService, smartFeedService)
	userHandler := handler.NewUserHandler(userService, achievementService, smartFeedService, experienceRepo, skillRepo)
	messageHandler := handler.NewMessageHandler(messageService)
	jobHandler := handler.NewJobHandler(jobService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	wsHandler := handler.NewWebSocketHandler(hub, authService, logger)
	webhookHandler := handler.NewWebhookHandler(asynqClient, userRepo, cfg.GitHubWebhookSecret, logger)

	// WebSocket route (uses query param token, not auth middleware)
	r.Get("/ws", wsHandler.HandleWS)

	// Webhook routes (signature-validated, no auth middleware)
	r.Route("/api/v1/webhooks", func(r chi.Router) {
		r.Post("/github", webhookHandler.GitHub)
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public, tighter rate limit)
		r.Route("/auth", func(r chi.Router) {
			r.Use(httprate.LimitByIP(20, time.Minute))
			r.Get("/github", authHandler.GitHubLogin)
			r.Post("/github/callback", authHandler.GitHubCallback)
			r.Get("/youtube", authHandler.YouTubeLogin)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/logout", authHandler.Logout)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(authService))
				r.Get("/me", authHandler.Me)
				r.Post("/youtube/callback", authHandler.YouTubeCallback)
			})
		})

		// User routes (public, with optional auth for isFollowing)
		r.Group(func(r chi.Router) {
			r.Use(middleware.TryAuthMiddleware(authService))
			r.Get("/users/search", userHandler.SearchUsers)
			r.Get("/users/{username}", userHandler.GetProfile)
			r.Get("/users/{username}/achievements", userHandler.GetUserAchievements)
			r.Get("/users/{username}/followers", userHandler.ListFollowers)
			r.Get("/users/{username}/following", userHandler.ListFollowing)
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
			r.Post("/users/me/experiences", userHandler.CreateExperience)
			r.Patch("/users/me/experiences/{id}", userHandler.UpdateExperience)
			r.Delete("/users/me/experiences/{id}", userHandler.DeleteExperience)
			r.Post("/users/me/skills", userHandler.CreateSkill)
			r.Delete("/users/me/skills/{id}", userHandler.DeleteSkill)
			r.Post("/users/{username}/follow", userHandler.Follow)
			r.Delete("/users/{username}/follow", userHandler.Unfollow)

			// Achievement routes
			r.Route("/achievements", func(r chi.Router) {
				r.Post("/sync", achievementHandler.Sync)
				r.Get("/{id}", achievementHandler.GetByID)
				r.Post("/{id}/reactions", achievementHandler.AddReaction)
				r.Delete("/{id}/reactions", achievementHandler.RemoveReaction)
			})

			// Feed routes
			r.Get("/feed", feedHandler.GetFeed)
			// Smart feed is rate-limited more tightly since each call costs LLM tokens.
			r.Group(func(r chi.Router) {
				r.Use(httprate.LimitByIP(10, time.Minute))
				r.Get("/feed/smart", feedHandler.GetSmartFeed)
			})

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
				r.Get("/search", messageHandler.SearchMessages)
				r.Delete("/{id}", messageHandler.DeleteConversation)
				r.Get("/{id}/messages", messageHandler.ListMessages)
				r.Post("/{id}/messages", messageHandler.SendMessage)
				r.Delete("/{id}/messages/{messageId}", messageHandler.DeleteMessage)
				r.Get("/{id}/read-times", messageHandler.GetReadTimes)
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
