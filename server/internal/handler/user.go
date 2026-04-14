package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile returns a user's public profile.
// GET /api/v1/users/{username}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	var viewerID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		viewerID = claims.UserID
	}

	profile, err := h.userService.GetProfile(r.Context(), username, viewerID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get profile")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile updates the authenticated user's profile.
// PATCH /api/v1/users/me
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req service.UpdateProfileRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == nil && req.Bio == nil && req.Headline == nil {
		writeError(w, http.StatusBadRequest, "at least one field must be provided")
		return
	}

	user, err := h.userService.UpdateProfile(r.Context(), claims.UserID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// Follow follows a user by username.
// POST /api/v1/users/{username}/follow
func (h *UserHandler) Follow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	err := h.userService.Follow(r.Context(), claims.UserID, username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, service.ErrCannotFollowSelf) {
			writeError(w, http.StatusBadRequest, "cannot follow yourself")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to follow user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Unfollow unfollows a user by username.
// DELETE /api/v1/users/{username}/follow
func (h *UserHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	err := h.userService.Unfollow(r.Context(), claims.UserID, username)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to unfollow user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
