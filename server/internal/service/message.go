package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/ws"
)

var (
	ErrRecipientNotFound    = errors.New("recipient not found")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrNotParticipant       = errors.New("not a participant in this conversation")
	ErrEmptyMessage         = errors.New("message content is empty")
	ErrCannotMessageSelf    = errors.New("cannot message yourself")
)

type MessageService struct {
	messageRepo *repository.MessageRepo
	userRepo    *repository.UserRepo
	hub         *ws.Hub
	logger      *zap.Logger
}

func NewMessageService(
	messageRepo *repository.MessageRepo,
	userRepo *repository.UserRepo,
	hub *ws.Hub,
	logger *zap.Logger,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		userRepo:    userRepo,
		hub:         hub,
		logger:      logger,
	}
}

// CreateOrGetConversation creates a new DM conversation with a recipient or
// returns the existing conversation ID if one already exists.
func (s *MessageService) CreateOrGetConversation(ctx context.Context, senderID, recipientUsername string) (string, error) {
	recipient, err := s.userRepo.GetUserByUsername(ctx, recipientUsername)
	if err != nil {
		return "", err
	}
	if recipient == nil {
		return "", ErrRecipientNotFound
	}

	if senderID == recipient.ID {
		return "", ErrCannotMessageSelf
	}

	// Check if a direct conversation already exists between these two users
	existingID, err := s.messageRepo.FindDirectConversation(ctx, senderID, recipient.ID)
	if err != nil {
		return "", err
	}
	if existingID != "" {
		return existingID, nil
	}

	// Create a new conversation
	conv, err := s.messageRepo.CreateConversation(ctx)
	if err != nil {
		return "", err
	}

	// Add both participants
	if err := s.messageRepo.AddParticipant(ctx, conv.ID, senderID); err != nil {
		return "", err
	}
	if err := s.messageRepo.AddParticipant(ctx, conv.ID, recipient.ID); err != nil {
		return "", err
	}

	return conv.ID, nil
}

// SendMessage sends a message in a conversation and broadcasts it to
// participants via WebSocket.
func (s *MessageService) SendMessage(ctx context.Context, conversationID, senderID, content string) (*repository.Message, error) {
	if content == "" {
		return nil, ErrEmptyMessage
	}

	// Verify sender is a participant
	isParticipant, err := s.messageRepo.IsParticipant(ctx, conversationID, senderID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	// Create the message
	msg, err := s.messageRepo.CreateMessage(ctx, conversationID, senderID, content)
	if err != nil {
		return nil, err
	}

	// Update conversation timestamp
	if err := s.messageRepo.UpdateConversationTimestamp(ctx, conversationID); err != nil {
		s.logger.Error("failed to update conversation timestamp", zap.Error(err))
	}

	// Update sender's last read
	if err := s.messageRepo.UpdateLastRead(ctx, conversationID, senderID); err != nil {
		s.logger.Error("failed to update last read", zap.Error(err))
	}

	// Broadcast to all participants via WebSocket
	participants, err := s.messageRepo.GetParticipants(ctx, conversationID)
	if err != nil {
		s.logger.Error("failed to get participants for broadcast", zap.Error(err))
	} else {
		userIDs := make([]string, len(participants))
		for i, p := range participants {
			userIDs[i] = p.ID
		}
		s.logger.Info("broadcasting message via WebSocket",
			zap.String("conversationID", conversationID),
			zap.String("messageID", msg.ID),
			zap.Strings("targetUserIDs", userIDs),
		)
		s.hub.SendToUsers(userIDs, &ws.WSMessage{
			Type: "message",
			Data: msg,
		})
	}

	return msg, nil
}

// ListConversations returns the authenticated user's conversations with details.
func (s *MessageService) ListConversations(ctx context.Context, userID string, limit, offset int) ([]repository.ConversationWithDetails, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.messageRepo.ListUserConversations(ctx, userID, limit, offset)
}

// ListMessages returns messages in a conversation, verifying the user is a participant.
func (s *MessageService) ListMessages(ctx context.Context, conversationID, userID string, limit, offset int) ([]repository.Message, error) {
	isParticipant, err := s.messageRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.messageRepo.ListMessages(ctx, conversationID, limit, offset)
}

// DeleteMessage deletes a message if the user is the sender.
func (s *MessageService) DeleteMessage(ctx context.Context, conversationID, messageID, userID string) error {
	isParticipant, err := s.messageRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return ErrNotParticipant
	}

	if err := s.messageRepo.DeleteMessage(ctx, messageID, userID); err != nil {
		return err
	}

	// Notify participants about the deletion via WebSocket
	participants, err := s.messageRepo.GetParticipants(ctx, conversationID)
	if err == nil {
		userIDs := make([]string, len(participants))
		for i, p := range participants {
			userIDs[i] = p.ID
		}
		s.hub.SendToUsers(userIDs, &ws.WSMessage{
			Type: "message_deleted",
			Data: map[string]string{
				"messageId":      messageID,
				"conversationId": conversationID,
			},
		})
	}

	return nil
}

// DeleteConversation deletes a conversation if the user is a participant.
func (s *MessageService) DeleteConversation(ctx context.Context, conversationID, userID string) error {
	isParticipant, err := s.messageRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return ErrNotParticipant
	}

	return s.messageRepo.DeleteConversation(ctx, conversationID)
}

// MarkRead marks a conversation as read for the user.
func (s *MessageService) MarkRead(ctx context.Context, conversationID, userID string) error {
	isParticipant, err := s.messageRepo.IsParticipant(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return ErrNotParticipant
	}

	return s.messageRepo.UpdateLastRead(ctx, conversationID, userID)
}
