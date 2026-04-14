package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/ws"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepo
	hub       *ws.Hub
	logger    *zap.Logger
}

func NewNotificationService(
	notifRepo *repository.NotificationRepo,
	hub *ws.Hub,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		notifRepo: notifRepo,
		hub:       hub,
		logger:    logger,
	}
}

// Create creates a notification and sends it via WebSocket in real-time.
func (s *NotificationService) Create(ctx context.Context, userID, notifType, title, body string, data map[string]interface{}) error {
	notif, err := s.notifRepo.Create(ctx, userID, notifType, title, body, data)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	// Send real-time notification via WebSocket
	s.hub.SendToUser(userID, &ws.WSMessage{
		Type: "notification",
		Data: notif,
	})

	s.logger.Debug("notification created",
		zap.String("userId", userID),
		zap.String("type", notifType),
		zap.String("title", title),
	)

	return nil
}

// List returns paginated notifications for a user.
func (s *NotificationService) List(ctx context.Context, userID string, limit, offset int) ([]repository.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.notifRepo.List(ctx, userID, limit, offset)
}

// MarkRead marks a notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	return s.notifRepo.MarkRead(ctx, id, userID)
}

// CountUnread returns the unread notification count for a user.
func (s *NotificationService) CountUnread(ctx context.Context, userID string) (int, error) {
	return s.notifRepo.CountUnread(ctx, userID)
}
