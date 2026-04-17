package worker

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/repository"
)

// Scheduler periodically enqueues sync tasks for every connected provider.
type Scheduler struct {
	client       *asynq.Client
	providerRepo *repository.ProviderRepo
	interval     time.Duration
	logger       *zap.Logger
}

// NewScheduler creates a new periodic scheduler.
// interval controls how often the scheduler scans connected providers.
func NewScheduler(client *asynq.Client, providerRepo *repository.ProviderRepo, interval time.Duration, logger *zap.Logger) *Scheduler {
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	return &Scheduler{
		client:       client,
		providerRepo: providerRepo,
		interval:     interval,
		logger:       logger,
	}
}

// Run drives the scheduler loop until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	s.logger.Info("scheduler starting", zap.Duration("interval", s.interval))

	// Run immediately on startup, then on the interval tick.
	s.enqueueAll(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopping")
			return
		case <-ticker.C:
			s.enqueueAll(ctx)
		}
	}
}

func (s *Scheduler) enqueueAll(ctx context.Context) {
	providers, err := s.providerRepo.ListAll(ctx)
	if err != nil {
		s.logger.Warn("scheduler: failed to list providers", zap.Error(err))
		return
	}

	enqueued := 0
	for _, p := range providers {
		// Skip providers already syncing.
		if p.SyncStatus == "SYNCING" {
			continue
		}

		task, err := NewProviderSyncTask(p.UserID, p.Provider)
		if err != nil {
			s.logger.Warn("scheduler: failed to create task", zap.Error(err))
			continue
		}

		if _, err := s.client.EnqueueContext(ctx, task, asynq.Queue("default")); err != nil {
			s.logger.Warn("scheduler: failed to enqueue task",
				zap.String("userId", p.UserID),
				zap.String("provider", p.Provider),
				zap.Error(err),
			)
			continue
		}
		enqueued++
	}

	s.logger.Info("scheduler enqueued provider syncs",
		zap.Int("total", len(providers)),
		zap.Int("enqueued", enqueued),
	)
}
