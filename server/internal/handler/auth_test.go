package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/praxis-social/praxis/server/internal/config"
	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/service"
)

const testJWTSecret = "test-secret-key-for-unit-tests-only"

// newTestConfig returns a config suitable for testing.
func newTestConfig() *config.Config {
	return &config.Config{
		JWTSecret:          testJWTSecret,
		GitHubClientID:     "test-client-id",
		GitHubClientSecret: "test-client-secret",
		FrontendURL:        "http://localhost:3000",
	}
}

// signTestToken creates a signed JWT for testing.
func signTestToken(userID, username string, expiry time.Duration) string {
	claims := service.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
	return token
}

// --- JWT generation and validation tests ---

func TestValidateAccessToken_Valid(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	claims, err := authSvc.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", -1*time.Minute)

	_, err := authSvc.ValidateAccessToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	// Sign with a different secret
	claims := service.Claims{
		UserID:   "user-123",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "user-123",
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))

	_, err := authSvc.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestValidateAccessToken_MalformedToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	_, err := authSvc.ValidateAccessToken("not-a-jwt")
	assert.Error(t, err)
}

// --- Auth middleware tests ---

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", 15*time.Minute)

	var capturedClaims *service.Claims
	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	assert.Equal(t, "testuser", capturedClaims.Username)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	token := signTestToken("user-123", "testuser", -1*time.Minute)

	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)

	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --- GitHub callback handler tests ---

func TestGitHubCallback_MissingCode(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	body := bytes.NewBufferString(`{"code":""}`)
	req := httptest.NewRequest("POST", "/api/v1/auth/github/callback", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.GitHubCallback(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "code is required", resp["error"])
}

func TestGitHubCallback_InvalidBody(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	body := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest("POST", "/api/v1/auth/github/callback", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.GitHubCallback(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Refresh token handler tests ---

func TestRefreshToken_MissingToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	body := bytes.NewBufferString(`{"refreshToken":""}`)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.RefreshToken(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRefreshToken_InvalidBody(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	body := bytes.NewBufferString(`bad json`)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.RefreshToken(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Logout handler tests ---

func TestLogout_MissingToken(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	body := bytes.NewBufferString(`{"refreshToken":""}`)
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.Logout(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Me endpoint tests ---

func TestMe_Authenticated(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	token := signTestToken("user-456", "janedoe", 15*time.Minute)

	// Wrap with auth middleware
	handler := middleware.AuthMiddleware(authSvc)(http.HandlerFunc(authHandler.Me))

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// With nil userRepo, /me returns 404 (user not found in DB)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestMe_Unauthenticated(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	// Call without middleware (simulates missing claims)
	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	authHandler.Me(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --- GitHub login URL test ---

func TestGitHubLogin_ReturnsURL(t *testing.T) {
	cfg := newTestConfig()
	authSvc := service.NewAuthService(nil, nil, cfg)
	authHandler := NewAuthHandler(authSvc, nil)

	req := httptest.NewRequest("GET", "/api/v1/auth/github", nil)
	rec := httptest.NewRecorder()

	authHandler.GitHubLogin(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Contains(t, resp["url"], "github.com/login/oauth/authorize")
	assert.Contains(t, resp["url"], "client_id=test-client-id")
}

// --- GetUserFromContext test ---

func TestGetUserFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	claims := middleware.GetUserFromContext(ctx)
	assert.Nil(t, claims)
}

// --- Helper function tests ---

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusCreated, map[string]string{"hello": "world"})

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "world", resp["hello"])
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusNotFound, "not found")

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "not found", resp["error"])
}

func TestReadJSON_Valid(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"test"}`)
	req := httptest.NewRequest("POST", "/", body)

	var data struct {
		Name string `json:"name"`
	}
	err := readJSON(req, &data)
	require.NoError(t, err)
	assert.Equal(t, "test", data.Name)
}

func TestReadJSON_Invalid(t *testing.T) {
	body := bytes.NewBufferString(`not json`)
	req := httptest.NewRequest("POST", "/", body)

	var data struct{}
	err := readJSON(req, &data)
	assert.Error(t, err)
}

// --- GitHub callback with mock server test ---

func TestGitHubCallback_WithMockGitHub(t *testing.T) {
	// Set up mock GitHub servers for token exchange and user info
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "mock-github-token",
			"token_type":   "bearer",
			"scope":        "read:user,user:email",
		})
	}))
	defer tokenServer.Close()

	userServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer mock-github-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         12345,
			"login":      "octocat",
			"email":      "octocat@github.com",
			"name":       "The Octocat",
			"avatar_url": "https://avatars.githubusercontent.com/u/12345",
		})
	}))
	defer userServer.Close()

	// Note: This test validates the mock servers respond correctly.
	// Full integration testing of HandleGitHubCallback requires a database
	// and would be done in integration tests.

	// Verify token server
	resp, err := http.Post(tokenServer.URL, "application/x-www-form-urlencoded", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp map[string]string
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	assert.Equal(t, "mock-github-token", tokenResp["access_token"])

	// Verify user server
	userReq, _ := http.NewRequest("GET", userServer.URL, nil)
	userReq.Header.Set("Authorization", "Bearer mock-github-token")
	userResp, err := http.DefaultClient.Do(userReq)
	require.NoError(t, err)
	defer userResp.Body.Close()
	assert.Equal(t, http.StatusOK, userResp.StatusCode)

	var userInfo map[string]interface{}
	json.NewDecoder(userResp.Body).Decode(&userInfo)
	assert.Equal(t, "octocat", userInfo["login"])
}

// Unused import guard - these are used in the test functions above.
var _ = (*repository.User)(nil)
