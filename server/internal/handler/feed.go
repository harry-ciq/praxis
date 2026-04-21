package handler

import (
	"net/http"
	"strconv"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
)

type FeedHandler struct {
	achievementService *service.AchievementService
	smartFeedService   *service.SmartFeedService
}

func NewFeedHandler(achievementService *service.AchievementService, smartFeedService *service.SmartFeedService) *FeedHandler {
	return &FeedHandler{
		achievementService: achievementService,
		smartFeedService:   smartFeedService,
	}
}

// GetSmartFeed returns an LLM-assisted digest of the user's raw feed.
// GET /api/v1/feed/smart?refresh=true
func (h *FeedHandler) GetSmartFeed(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	force := r.URL.Query().Get("refresh") == "true"

	digest, err := h.smartFeedService.Generate(r.Context(), claims.UserID, force)
	if err != nil {
		if err == service.ErrSmartFeedUnavailable {
			writeError(w, http.StatusServiceUnavailable, "smart feed is not configured on this server")
			return
		}
		writeError(w, http.StatusInternalServerError, "smart feed generation failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, digest)
}

// GetFeed returns a paginated feed of achievements.
// GET /api/v1/feed?limit=20&offset=0
func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	achievements, err := h.achievementService.GetFeed(r.Context(), claims.UserID, limit+1, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get feed")
		return
	}

	// Check if there are more results
	var nextCursor *string
	if len(achievements) > limit {
		achievements = achievements[:limit]
		next := strconv.Itoa(offset + limit)
		nextCursor = &next
	}

	// Enrich achievements with reaction data
	type enriched struct {
		ID               string      `json:"id"`
		UserID           string      `json:"userId"`
		Type             string      `json:"type"`
		Title            string      `json:"title"`
		Description      string      `json:"description"`
		Metadata         interface{} `json:"metadata"`
		ProofURL         string      `json:"proofUrl"`
		Source           string      `json:"source"`
		SourceID         string      `json:"sourceId"`
		VerificationHash *string     `json:"verificationHash"`
		Status           string      `json:"status"`
		CreatedAt        string      `json:"createdAt"`
		User             interface{} `json:"user"`
		Reactions        interface{} `json:"reactions"`
		UserReaction     *string     `json:"userReaction"`
	}

	results := make([]enriched, 0, len(achievements))
	for _, a := range achievements {
		reactionCounts, _ := h.achievementService.GetReactionCounts(r.Context(), a.ID)
		userReaction, _ := h.achievementService.GetUserReaction(r.Context(), claims.UserID, a.ID)

		var ur *string
		if userReaction != "" {
			ur = &userReaction
		}

		var reactions interface{}
		if reactionCounts != nil {
			reactions = reactionCounts
		} else {
			reactions = map[string]int{"clap": 0, "fire": 0, "rocket": 0, "total": 0}
		}

		results = append(results, enriched{
			ID:               a.ID,
			UserID:           a.UserID,
			Type:             a.Type,
			Title:            a.Title,
			Description:      a.Description,
			Metadata:         a.Metadata,
			ProofURL:         a.ProofURL,
			Source:           a.Source,
			SourceID:         a.SourceID,
			VerificationHash: a.VerificationHash,
			Status:           a.Status,
			CreatedAt:        a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			User: map[string]string{
				"username":  a.UserUsername,
				"name":      a.UserName,
				"avatarUrl": a.UserAvatarURL,
			},
			Reactions:    reactions,
			UserReaction: ur,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"achievements": results,
		"nextCursor":   nextCursor,
	})
}
