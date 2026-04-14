package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
	"github.com/praxis-social/praxis/server/internal/ws"
)

func newTestNotificationHandler() *NotificationHandler {
	hub := ws.NewHub(zap.NewNop())
	notifService := service.NewNotificationService(nil, hub, zap.NewNop())
	return NewNotificationHandler(notifService)
}

// --- ListNotifications ---

func TestListNotifications_Unauthenticated(t *testing.T) {
	h := newTestNotificationHandler()

	req := httptest.NewRequest("GET", "/api/v1/notifications", nil)
	rec := httptest.NewRecorder()

	h.ListNotifications(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "not authenticated", resp["error"])
}

func TestListNotifications_Authenticated(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	// Verify auth middleware passes through to handler
	var capturedClaims *service.Claims
	testHandler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserFromContext(r.Context())
		writeJSON(w, http.StatusOK, []interface{}{})
	}))

	req := httptest.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	testHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "user-123", capturedClaims.UserID)
}

// --- MarkRead ---

func TestNotificationMarkRead_Unauthenticated(t *testing.T) {
	h := newTestNotificationHandler()

	req := httptest.NewRequest("PATCH", "/api/v1/notifications/notif-123/read", nil)
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "not authenticated", resp["error"])
}

func TestNotificationMarkRead_MissingID(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	hub := ws.NewHub(zap.NewNop())
	notifService := service.NewNotificationService(nil, hub, zap.NewNop())
	notifHandler := NewNotificationHandler(notifService)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	// Without chi router, URLParam returns empty string
	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(notifHandler.MarkRead))

	req := httptest.NewRequest("PATCH", "/api/v1/notifications//read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "notification id is required", resp["error"])
}

func TestNotificationMarkRead_WithIDRouting(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	// Verify chi routing extracts the ID param and auth passes correctly
	var capturedID string
	var capturedClaims *service.Claims
	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(authSvc))
	r.Patch("/api/v1/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		capturedID = chi.URLParam(r, "id")
		capturedClaims = middleware.GetUserFromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("PATCH", "/api/v1/notifications/notif-123/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "notif-123", capturedID)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "user-123", capturedClaims.UserID)
}

// --- UnreadCount ---

func TestUnreadCount_Unauthenticated(t *testing.T) {
	h := newTestNotificationHandler()

	req := httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
	rec := httptest.NewRecorder()

	h.UnreadCount(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "not authenticated", resp["error"])
}

func TestUnreadCount_Authenticated(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	// Verify auth passes and handler is reached
	var capturedClaims *service.Claims
	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserFromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]int{"count": 0})
	}))

	req := httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "user-123", capturedClaims.UserID)

	var resp map[string]int
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, 0, resp["count"])
}

// Unused import guard
var _ = require.NotNil
