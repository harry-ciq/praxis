package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
)

// withChiURLParam sets a chi URL parameter on the request context.
func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// withAuth injects user claims into the request context using the auth middleware context key.
func withAuth(r *http.Request, userID, username string) *http.Request {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	token := signTestToken(userID, username, 15*time.Minute)

	// Use the real middleware to inject claims properly
	var authedReq *http.Request
	mw := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedReq = r
	}))
	rr := httptest.NewRecorder()
	r.Header.Set("Authorization", "Bearer "+token)
	mw.ServeHTTP(rr, r)

	return authedReq
}

// --- GetProfile handler tests ---

func TestGetProfile_MissingUsername(t *testing.T) {
	h := NewUserHandler(nil)

	req := httptest.NewRequest("GET", "/api/v1/users/", nil)
	req = withChiURLParam(req, "username", "")
	rec := httptest.NewRecorder()

	h.GetProfile(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "username is required", resp["error"])
}

// --- UpdateProfile handler tests ---

func TestUpdateProfile_Unauthenticated(t *testing.T) {
	h := NewUserHandler(nil)

	body := bytes.NewBufferString(`{"name":"New Name"}`)
	req := httptest.NewRequest("PATCH", "/api/v1/users/me", body)
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUpdateProfile_InvalidBody(t *testing.T) {
	h := NewUserHandler(nil)

	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest("PATCH", "/api/v1/users/me", body)
	req = withAuth(req, "user-123", "testuser")
	require.NotNil(t, req)
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateProfile_EmptyUpdate(t *testing.T) {
	h := NewUserHandler(nil)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest("PATCH", "/api/v1/users/me", body)
	req = withAuth(req, "user-123", "testuser")
	require.NotNil(t, req)
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "at least one field must be provided", resp["error"])
}

// --- Follow handler tests ---

func TestFollow_Unauthenticated(t *testing.T) {
	h := NewUserHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/users/someone/follow", nil)
	req = withChiURLParam(req, "username", "someone")
	rec := httptest.NewRecorder()

	h.Follow(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestFollow_MissingUsername(t *testing.T) {
	h := NewUserHandler(nil)

	req := httptest.NewRequest("POST", "/api/v1/users//follow", nil)
	req = withAuth(req, "user-123", "testuser")
	require.NotNil(t, req)
	// Add empty username chi param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("username", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.Follow(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Unfollow handler tests ---

func TestUnfollow_Unauthenticated(t *testing.T) {
	h := NewUserHandler(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/users/someone/follow", nil)
	req = withChiURLParam(req, "username", "someone")
	rec := httptest.NewRecorder()

	h.Unfollow(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUnfollow_MissingUsername(t *testing.T) {
	h := NewUserHandler(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/users//follow", nil)
	req = withAuth(req, "user-123", "testuser")
	require.NotNil(t, req)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("username", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.Unfollow(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- TryAuthMiddleware tests ---

func TestTryAuthMiddleware_ValidToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	token := signTestToken("user-123", "testuser", 15*time.Minute)

	var capturedClaims *service.Claims
	handler := middleware.TryAuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "user-123", capturedClaims.UserID)
}

func TestTryAuthMiddleware_NoToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	var capturedClaims *service.Claims
	handler := middleware.TryAuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should not reject, just proceed without claims
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, capturedClaims)
}

func TestTryAuthMiddleware_InvalidToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	var capturedClaims *service.Claims
	handler := middleware.TryAuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = middleware.GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should not reject, just proceed without claims
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, capturedClaims)
}
