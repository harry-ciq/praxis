package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/config"
	"github.com/praxis-social/praxis/server/internal/provider"
	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/service"
	"github.com/praxis-social/praxis/server/internal/worker"
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

	// PostgreSQL
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Redis (for AuthService dependency, even though worker doesn't need it directly)
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal("failed to parse redis URL", zap.Error(err))
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	// Repositories
	userRepo := repository.NewUserRepo(pool)
	achievementRepo := repository.NewAchievementRepo(pool)
	providerRepo := repository.NewProviderRepo(pool)
	skillRepo := repository.NewSkillRepo(pool)

	// Provider registry
	registry := provider.NewRegistry()
	registry.Register(provider.NewGitHubProvider())
	registry.Register(provider.NewYouTubeProvider())

	// Services
	achievementService := service.NewAchievementService(achievementRepo, providerRepo, userRepo, skillRepo, registry, logger)

	// Asynq server
	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		logger.Fatal("failed to parse redis URI for asynq", zap.Error(err))
	}

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	handlers := worker.NewHandlers(achievementService, logger)
	mux := asynq.NewServeMux()
	mux.HandleFunc(worker.TaskProviderSync, handlers.HandleProviderSync)
	mux.HandleFunc(worker.TaskEmailNotify, handlers.HandleEmailNotify)

	// Asynq client (for the scheduler to enqueue tasks)
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	// Periodic scheduler
	intervalHours := 6
	if v := os.Getenv("SYNC_INTERVAL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalHours = n
		}
	}
	scheduler := worker.NewScheduler(client, providerRepo, time.Duration(intervalHours)*time.Hour, logger)
	go scheduler.Run(ctx)

	// Run worker
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("worker starting")
		if err := srv.Run(mux); err != nil {
			logger.Fatal("worker failed", zap.Error(err))
		}
	}()

	<-done
	logger.Info("worker shutting down")
	cancel()
	srv.Shutdown()
	logger.Info("worker stopped")
}
