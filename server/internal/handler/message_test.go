package handler

import (
	"bytes"
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

func newTestMessageHandler() *MessageHandler {
	hub := ws.NewHub(zap.NewNop())
	// MessageService requires repos that need a DB, so we test at the handler
	// level using nil repos and verify error paths / validation.
	msgService := service.NewMessageService(nil, nil, hub, zap.NewNop())
	return NewMessageHandler(msgService)
}

// --- ListConversations ---

func TestListConversations_Unauthenticated(t *testing.T) {
	h := newTestMessageHandler()

	req := httptest.NewRequest("GET", "/api/v1/conversations", nil)
	rec := httptest.NewRecorder()

	h.ListConversations(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "not authenticated", resp["error"])
}

func TestListConversations_Authenticated_ReachesHandler(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	// Use a simple handler that proves auth middleware passes through
	var capturedClaims *service.Claims
	testHandler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserFromContext(r.Context())
		writeJSON(w, http.StatusOK, []interface{}{})
	}))

	req := httptest.NewRequest("GET", "/api/v1/conversations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	testHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "user-123", capturedClaims.UserID)
}

// --- CreateConversation ---

func TestCreateConversation_MissingUsername(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	hub := ws.NewHub(zap.NewNop())
	msgService := service.NewMessageService(nil, nil, hub, zap.NewNop())
	msgHandler := NewMessageHandler(msgService)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(msgHandler.CreateConversation))

	body := bytes.NewBufferString(`{"username":""}`)
	req := httptest.NewRequest("POST", "/api/v1/conversations", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "username is required", resp["error"])
}

func TestCreateConversation_InvalidBody(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	hub := ws.NewHub(zap.NewNop())
	msgService := service.NewMessageService(nil, nil, hub, zap.NewNop())
	msgHandler := NewMessageHandler(msgService)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(msgHandler.CreateConversation))

	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest("POST", "/api/v1/conversations", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- SendMessage ---

func TestSendMessage_EmptyContent(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	hub := ws.NewHub(zap.NewNop())
	msgService := service.NewMessageService(nil, nil, hub, zap.NewNop())
	msgHandler := NewMessageHandler(msgService)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	// Set up chi context with URL param
	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(authSvc))
	r.Post("/api/v1/conversations/{id}/messages", msgHandler.SendMessage)

	body := bytes.NewBufferString(`{"content":""}`)
	req := httptest.NewRequest("POST", "/api/v1/conversations/conv-123/messages", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "content is required", resp["error"])
}

func TestSendMessage_InvalidBody(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	hub := ws.NewHub(zap.NewNop())
	msgService := service.NewMessageService(nil, nil, hub, zap.NewNop())
	msgHandler := NewMessageHandler(msgService)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware(authSvc))
	r.Post("/api/v1/conversations/{id}/messages", msgHandler.SendMessage)

	body := bytes.NewBufferString(`not valid json`)
	req := httptest.NewRequest("POST", "/api/v1/conversations/conv-123/messages", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendMessage_Unauthenticated(t *testing.T) {
	h := newTestMessageHandler()

	body := bytes.NewBufferString(`{"content":"hello"}`)
	req := httptest.NewRequest("POST", "/api/v1/conversations/conv-123/messages", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.SendMessage(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --- MarkRead ---

func TestMarkRead_Unauthenticated(t *testing.T) {
	h := newTestMessageHandler()

	req := httptest.NewRequest("PATCH", "/api/v1/conversations/conv-123/read", nil)
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
