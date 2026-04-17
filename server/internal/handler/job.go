package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/praxis-social/praxis/server/internal/middleware"
	"github.com/praxis-social/praxis/server/internal/service"
)

type JobHandler struct {
	jobService *service.JobService
}

func NewJobHandler(jobService *service.JobService) *JobHandler {
	return &JobHandler{jobService: jobService}
}

// ListJobs returns active job listings.
// GET /api/v1/jobs?qualified=true
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	qualified := r.URL.Query().Get("qualified") == "true"

	var userID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		userID = claims.UserID
	}

	jobs, err := h.jobService.ListJobs(r.Context(), userID, qualified, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list jobs")
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

// GetJob returns a single job by ID with match info.
// GET /api/v1/jobs/{id}
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "job id is required")
		return
	}

	var userID string
	if claims := middleware.GetUserFromContext(r.Context()); claims != nil {
		userID = claims.UserID
	}

	job, err := h.jobService.GetJob(r.Context(), id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get job")
		return
	}
	if job == nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// Apply applies to a job for the authenticated user.
// POST /api/v1/jobs/{id}/apply
func (h *JobHandler) Apply(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		writeError(w, http.StatusBadRequest, "job id is required")
		return
	}

	var body struct {
		CoverNote string `json:"coverNote"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	app, err := h.jobService.Apply(r.Context(), jobID, claims.UserID, body.CoverNote)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, app)
}

// CreateJob creates a new job posting.
// POST /api/v1/jobs
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var body struct {
		CompanyID            string   `json:"companyId"`
		Title                string   `json:"title"`
		Description          string   `json:"description"`
		Location             string   `json:"location"`
		JobType              string   `json:"jobType"`
		SalaryRange          string   `json:"salaryRange"`
		RequiredAchievements []string `json:"requiredAchievements"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if body.CompanyID == "" {
		writeError(w, http.StatusBadRequest, "companyId is required")
		return
	}

	job, err := h.jobService.CreateJob(r.Context(), body.CompanyID, body.Title, body.Description,
		body.Location, body.JobType, body.SalaryRange, body.RequiredAchievements)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create job")
		return
	}

	writeJSON(w, http.StatusCreated, job)
}
