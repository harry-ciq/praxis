package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/config"
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

	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		logger.Fatal("failed to parse redis URL", zap.Error(err))
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

	mux := asynq.NewServeMux()

	// Register task handlers here as they are implemented.
	// Example:
	//   mux.HandleFunc("email:welcome", handleWelcomeEmail)
	//   mux.HandleFunc("notification:push", handlePushNotification)

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
	srv.Shutdown()
	logger.Info("worker stopped")
}
