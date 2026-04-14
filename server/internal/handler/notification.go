package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// ListNotifications returns paginated notifications for the authenticated user.
// GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	notifications, err := h.notificationService.List(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}

	writeJSON(w, http.StatusOK, notifications)
}

// MarkRead marks a notification as read.
// PATCH /api/v1/notifications/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "notification id is required")
		return
	}

	if err := h.notificationService.MarkRead(r.Context(), id, claims.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UnreadCount returns the count of unread notifications.
// GET /api/v1/notifications/unread-count
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	count, err := h.notificationService.CountUnread(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}
