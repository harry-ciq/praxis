package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
)

type AchievementHandler struct {
	achievementService *service.AchievementService
}

func NewAchievementHandler(achievementService *service.AchievementService) *AchievementHandler {
	return &AchievementHandler{achievementService: achievementService}
}

// Sync triggers a sync for the authenticated user's provider.
// POST /api/v1/achievements/sync
func (h *AchievementHandler) Sync(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var body struct {
		Provider string `json:"provider"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Provider == "" {
		writeError(w, http.StatusBadRequest, "provider is required")
		return
	}

	newCount, err := h.achievementService.SyncProvider(r.Context(), claims.UserID, body.Provider)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sync failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"newAchievements": newCount,
	})
}

// GetByID returns a single achievement with reaction counts.
// GET /api/v1/achievements/{id}
func (h *AchievementHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "achievement id is required")
		return
	}

	achievement, err := h.achievementService.GetAchievementByID(r.Context(), id, claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get achievement")
		return
	}
	if achievement == nil {
		writeError(w, http.StatusNotFound, "achievement not found")
		return
	}

	writeJSON(w, http.StatusOK, achievement)
}

// AddReaction adds a reaction to an achievement.
// POST /api/v1/achievements/{id}/reactions
func (h *AchievementHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	achievementID := chi.URLParam(r, "id")
	if achievementID == "" {
		writeError(w, http.StatusBadRequest, "achievement id is required")
		return
	}

	var body struct {
		Type string `json:"type"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Type == "" {
		writeError(w, http.StatusBadRequest, "reaction type is required")
		return
	}

	if err := h.achievementService.ToggleReaction(r.Context(), claims.UserID, achievementID, body.Type); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add reaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RemoveReaction removes a reaction from an achievement.
// DELETE /api/v1/achievements/{id}/reactions
func (h *AchievementHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	achievementID := chi.URLParam(r, "id")
	if achievementID == "" {
		writeError(w, http.StatusBadRequest, "achievement id is required")
		return
	}

	if err := h.achievementService.RemoveReaction(r.Context(), claims.UserID, achievementID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove reaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
