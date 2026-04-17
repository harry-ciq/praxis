package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/service"
)

// Handlers groups the worker dependencies needed to process tasks.
type Handlers struct {
	achievementService *service.AchievementService
	logger             *zap.Logger
}

// NewHandlers constructs a Handlers instance.
func NewHandlers(achievementService *service.AchievementService, logger *zap.Logger) *Handlers {
	return &Handlers{
		achievementService: achievementService,
		logger:             logger,
	}
}

// HandleProviderSync re-syncs achievements for a user/provider combination.
func (h *Handlers) HandleProviderSync(ctx context.Context, t *asynq.Task) error {
	var p ProviderSyncPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w: %w", err, asynq.SkipRetry)
	}
	if p.UserID == "" || p.Provider == "" {
		return fmt.Errorf("missing userID or provider: %w", asynq.SkipRetry)
	}

	h.logger.Info("running provider sync task",
		zap.String("userId", p.UserID),
		zap.String("provider", p.Provider),
	)

	newCount, err := h.achievementService.SyncProvider(ctx, p.UserID, p.Provider)
	if err != nil {
		h.logger.Warn("provider sync failed",
			zap.String("userId", p.UserID),
			zap.String("provider", p.Provider),
			zap.Error(err),
		)
		return err
	}

	h.logger.Info("provider sync completed",
		zap.String("userId", p.UserID),
		zap.String("provider", p.Provider),
		zap.Int("newAchievements", newCount),
	)
	return nil
}

// HandleEmailNotify is a placeholder email handler that logs the would-be email.
// In production this would call SES, SendGrid, etc.
func (h *Handlers) HandleEmailNotify(ctx context.Context, t *asynq.Task) error {
	var p EmailNotifyPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal payload: %w: %w", err, asynq.SkipRetry)
	}

	h.logger.Info("would send email",
		zap.String("to", p.To),
		zap.String("subject", p.Subject),
		zap.String("template", p.Template),
		zap.Any("data", p.Data),
	)
	return nil
}
