package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/repository"
	"github.com/praxis-social/praxis/server/internal/service"
)

type UserHandler struct {
	userService        *service.UserService
	achievementService *service.AchievementService
	experienceRepo     *repository.ExperienceRepo
	skillRepo          *repository.SkillRepo
}

func NewUserHandler(
	userService *service.UserService,
	achievementService *service.AchievementService,
	experienceRepo *repository.ExperienceRepo,
	skillRepo *repository.SkillRepo,
) *UserHandler {
	return &UserHandler{
		userService:        userService,
		achievementService: achievementService,
		experienceRepo:     experienceRepo,
		skillRepo:          skillRepo,
	}
}

// SearchUsers searches for users by name or username.
// GET /api/v1/users/search?q=harry&limit=10
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" || len(query) < 2 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"users": []interface{}{},
		})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	var viewerID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		viewerID = claims.UserID
	}

	users, err := h.userService.SearchUsers(r.Context(), query, viewerID, limit, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search users")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
	})
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

	if req.Name == nil && req.Bio == nil && req.Headline == nil && req.Location == nil && req.WebsiteURL == nil && req.SocialLinks == nil {
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

// ListFollowers returns a paginated list of followers for a user.
// GET /api/v1/users/{username}/followers?limit=20&offset=0
func (h *UserHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var viewerID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		viewerID = claims.UserID
	}

	followers, err := h.userService.GetFollowers(r.Context(), username, viewerID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get followers")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": followers,
	})
}

// ListFollowing returns a paginated list of users that a user follows.
// GET /api/v1/users/{username}/following?limit=20&offset=0
func (h *UserHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var viewerID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		viewerID = claims.UserID
	}

	following, err := h.userService.GetFollowing(r.Context(), username, viewerID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get following")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users": following,
	})
}

// GetUserAchievements returns paginated achievements for a user.
// GET /api/v1/users/{username}/achievements?limit=20&offset=0
func (h *UserHandler) GetUserAchievements(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Resolve username to userID
	var viewerID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		viewerID = claims.UserID
	}

	// We need to look up the user by username to get the userID
	profile, err := h.userService.GetProfile(r.Context(), username, viewerID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	achievements, err := h.achievementService.GetUserAchievements(r.Context(), profile.ID, limit+1, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get achievements")
		return
	}

	var nextCursor *string
	if len(achievements) > limit {
		achievements = achievements[:limit]
		next := strconv.Itoa(offset + limit)
		nextCursor = &next
	}

	// Enrich with reaction data (same format as feed.go)
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
		var userReaction string
		if viewerID != "" {
			userReaction, _ = h.achievementService.GetUserReaction(r.Context(), viewerID, a.ID)
		}

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

// CreateExperience creates a new experience for the authenticated user.
// POST /api/v1/users/me/experiences
func (h *UserHandler) CreateExperience(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		CompanyName string  `json:"companyName"`
		Role        string  `json:"role"`
		StartDate   string  `json:"startDate"`
		EndDate     *string `json:"endDate"`
		Description string  `json:"description"`
		IsCurrent   bool    `json:"isCurrent"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CompanyName == "" || req.Role == "" || req.StartDate == "" {
		writeError(w, http.StatusBadRequest, "companyName, role, and startDate are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "startDate must be in YYYY-MM-DD format")
		return
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "endDate must be in YYYY-MM-DD format")
			return
		}
		endDate = &parsed
	}

	exp := &repository.Experience{
		UserID:      claims.UserID,
		CompanyName: req.CompanyName,
		Role:        req.Role,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
		IsCurrent:   req.IsCurrent,
	}

	created, err := h.experienceRepo.Create(r.Context(), exp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create experience")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// UpdateExperience updates an existing experience for the authenticated user.
// PATCH /api/v1/users/me/experiences/{id}
func (h *UserHandler) UpdateExperience(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "experience id is required")
		return
	}

	// Verify ownership
	existing, err := h.experienceRepo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get experience")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "experience not found")
		return
	}
	if existing.UserID != claims.UserID {
		writeError(w, http.StatusForbidden, "not authorized")
		return
	}

	var req struct {
		CompanyName string  `json:"companyName"`
		Role        string  `json:"role"`
		StartDate   string  `json:"startDate"`
		EndDate     *string `json:"endDate"`
		Description string  `json:"description"`
		IsCurrent   bool    `json:"isCurrent"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CompanyName == "" || req.Role == "" || req.StartDate == "" {
		writeError(w, http.StatusBadRequest, "companyName, role, and startDate are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "startDate must be in YYYY-MM-DD format")
		return
	}

	var endDate *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "endDate must be in YYYY-MM-DD format")
			return
		}
		endDate = &parsed
	}

	exp := &repository.Experience{
		CompanyName: req.CompanyName,
		Role:        req.Role,
		StartDate:   startDate,
		EndDate:     endDate,
		Description: req.Description,
		IsCurrent:   req.IsCurrent,
	}

	updated, err := h.experienceRepo.Update(r.Context(), id, exp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update experience")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeleteExperience deletes an experience for the authenticated user.
// DELETE /api/v1/users/me/experiences/{id}
func (h *UserHandler) DeleteExperience(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "experience id is required")
		return
	}

	err := h.experienceRepo.Delete(r.Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "experience not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete experience")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateSkill creates a new skill for the authenticated user.
// POST /api/v1/users/me/skills
func (h *UserHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	skill, err := h.skillRepo.Create(r.Context(), claims.UserID, req.Name, false, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create skill")
		return
	}

	writeJSON(w, http.StatusCreated, skill)
}

// DeleteSkill deletes a skill for the authenticated user.
// DELETE /api/v1/users/me/skills/{id}
func (h *UserHandler) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "skill id is required")
		return
	}

	err := h.skillRepo.Delete(r.Context(), id, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete skill")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
