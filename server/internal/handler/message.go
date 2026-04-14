package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// ListConversations returns the authenticated user's conversations.
// GET /api/v1/conversations
func (h *MessageHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	conversations, err := h.messageService.ListConversations(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list conversations")
		return
	}

	if conversations == nil {
		conversations = []repository.ConversationWithDetails{}
	}

	writeJSON(w, http.StatusOK, conversations)
}

// CreateConversation creates a new DM conversation or returns an existing one.
// POST /api/v1/conversations
func (h *MessageHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	conversationID, err := h.messageService.CreateOrGetConversation(r.Context(), claims.UserID, req.Username)
	if err != nil {
		if errors.Is(err, service.ErrRecipientNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, service.ErrCannotMessageSelf) {
			writeError(w, http.StatusBadRequest, "cannot message yourself")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"id": conversationID})
}

// ListMessages returns messages in a conversation.
// GET /api/v1/conversations/{id}/messages
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		writeError(w, http.StatusBadRequest, "conversation id is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	messages, err := h.messageService.ListMessages(r.Context(), conversationID, claims.UserID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrNotParticipant) {
			writeError(w, http.StatusForbidden, "not a participant in this conversation")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	if messages == nil {
		messages = []repository.Message{}
	}

	writeJSON(w, http.StatusOK, messages)
}

// SendMessage sends a message in a conversation.
// POST /api/v1/conversations/{id}/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		writeError(w, http.StatusBadRequest, "conversation id is required")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	msg, err := h.messageService.SendMessage(r.Context(), conversationID, claims.UserID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrNotParticipant) {
			writeError(w, http.StatusForbidden, "not a participant in this conversation")
			return
		}
		if errors.Is(err, service.ErrEmptyMessage) {
			writeError(w, http.StatusBadRequest, "content is required")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to send message")
		return
	}

	writeJSON(w, http.StatusCreated, msg)
}

// MarkRead marks a conversation as read for the authenticated user.
// PATCH /api/v1/conversations/{id}/read
func (h *MessageHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if conversationID == "" {
		writeError(w, http.StatusBadRequest, "conversation id is required")
		return
	}

	err := h.messageService.MarkRead(r.Context(), conversationID, claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrNotParticipant) {
			writeError(w, http.StatusForbidden, "not a participant in this conversation")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to mark as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
