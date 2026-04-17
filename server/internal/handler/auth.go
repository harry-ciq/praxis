package handler

import (
	"net/http"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	userRepo    *repository.UserRepo
}

func NewAuthHandler(authService *service.AuthService, userRepo *repository.UserRepo) *AuthHandler {
	return &AuthHandler{authService: authService, userRepo: userRepo}
}

// GitHubLogin returns the GitHub OAuth authorization URL.
// GET /api/v1/auth/github
func (h *AuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := h.authService.GitHubAuthURL()
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// GitHubCallback exchanges the GitHub code for tokens and returns a JWT pair.
// POST /api/v1/auth/github/callback
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	tokenPair, user, err := h.authService.HandleGitHubCallback(r.Context(), body.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "authentication failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tokens": tokenPair,
		"user":   user,
	})
}

// YouTubeLogin returns the Google OAuth authorization URL for YouTube.
// GET /api/v1/auth/youtube
func (h *AuthHandler) YouTubeLogin(w http.ResponseWriter, r *http.Request) {
	url := h.authService.GoogleAuthURL()
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// YouTubeCallback exchanges the Google OAuth code and links the YouTube account.
// POST /api/v1/auth/youtube/callback
func (h *AuthHandler) YouTubeCallback(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	err := h.authService.HandleGoogleYouTubeCallback(r.Context(), body.Code, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "youtube connection failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "connected"})
}

// RefreshToken issues a new token pair from a valid refresh token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refreshToken is required")
		return
	}

	tokenPair, err := h.authService.RefreshTokens(r.Context(), body.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	writeJSON(w, http.StatusOK, tokenPair)
}

// Logout invalidates a refresh token.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "refreshToken is required")
		return
	}

	if err := h.authService.Logout(r.Context(), body.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me returns the currently authenticated user.
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if h.userRepo == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"name":      user.Name,
		"email":     user.Email,
		"avatarUrl": user.AvatarURL,
		"headline":  user.Headline,
	})
}
