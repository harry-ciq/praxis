package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/repository"
)

// JobCompany is the company info nested inside the job response.
type JobCompany struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logoUrl"`
	Website string `json:"website"`
}

// JobResponse is the API shape for a job — match frontend Job type.
type JobResponse struct {
	ID                   string          `json:"id"`
	Title                string          `json:"title"`
	Description          string          `json:"description"`
	Location             string          `json:"location"`
	JobType              string          `json:"jobType"`
	SalaryRange          string          `json:"salaryRange"`
	RequiredAchievements []string        `json:"requiredAchievements"`
	Status               string          `json:"status"`
	CreatedAt            time.Time       `json:"createdAt"`
	Company              JobCompany      `json:"company"`
	MatchedAchievements  int             `json:"matchedAchievements"`
	TotalRequired        int             `json:"totalRequired"`
	MatchedRequirements  []string        `json:"matchedRequirements"`
}

func toJobResponse(job repository.JobWithCompany, matched int, matchedList []string) JobResponse {
	var reqs []string
	if err := json.Unmarshal(job.RequiredAchievements, &reqs); err != nil || reqs == nil {
		reqs = []string{}
	}
	if matchedList == nil {
		matchedList = []string{}
	}
	return JobResponse{
		ID:                   job.ID,
		Title:                job.Title,
		Description:          job.Description,
		Location:             job.Location,
		JobType:              job.JobType,
		SalaryRange:          job.SalaryRange,
		RequiredAchievements: reqs,
		Status:               job.Status,
		CreatedAt:            job.CreatedAt,
		Company: JobCompany{
			ID:      job.CompanyID,
			Name:    job.CompanyName,
			LogoURL: job.CompanyLogoURL,
			Website: job.CompanyWebsite,
		},
		MatchedAchievements: matched,
		TotalRequired:       len(reqs),
		MatchedRequirements: matchedList,
	}
}

type JobService struct {
	jobRepo         *repository.JobRepo
	achievementRepo *repository.AchievementRepo
	logger          *zap.Logger
}

func NewJobService(
	jobRepo *repository.JobRepo,
	achievementRepo *repository.AchievementRepo,
	logger *zap.Logger,
) *JobService {
	return &JobService{
		jobRepo:         jobRepo,
		achievementRepo: achievementRepo,
		logger:          logger,
	}
}

// ListJobs returns active jobs with pagination, scored against the optional userID.
// If qualifiedOnly is true, jobs are filtered to those where the user meets all requirements.
func (s *JobService) ListJobs(ctx context.Context, userID string, qualifiedOnly bool, limit, offset int) ([]JobResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	jobs, err := s.jobRepo.ListActive(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Pre-fetch user achievements once if a user is provided.
	var userAchievements []repository.AchievementWithUser
	if userID != "" {
		userAchievements, err = s.achievementRepo.ListByUserID(ctx, userID, 1000, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list user achievements: %w", err)
		}
	}

	results := make([]JobResponse, 0, len(jobs))
	for _, job := range jobs {
		var reqs []string
		_ = json.Unmarshal(job.RequiredAchievements, &reqs)

		matched, matchedList := scoreRequirements(reqs, userAchievements)

		// If qualified filter requested, only include jobs where user meets all requirements
		if qualifiedOnly && (len(reqs) == 0 || matched < len(reqs)) {
			continue
		}
		results = append(results, toJobResponse(job, matched, matchedList))
	}

	return results, nil
}

// GetJob returns a single job with match info for the given user.
func (s *JobService) GetJob(ctx context.Context, jobID, userID string) (*JobResponse, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}

	var reqs []string
	_ = json.Unmarshal(job.RequiredAchievements, &reqs)

	var userAchievements []repository.AchievementWithUser
	if userID != "" {
		userAchievements, err = s.achievementRepo.ListByUserID(ctx, userID, 1000, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to list user achievements: %w", err)
		}
	}

	matched, matchedList := scoreRequirements(reqs, userAchievements)
	resp := toJobResponse(*job, matched, matchedList)
	return &resp, nil
}

// Apply applies to a job for the given user.
func (s *JobService) Apply(ctx context.Context, jobID, userID, coverNote string) (*repository.JobApplication, error) {
	// Check if job exists
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}
	if job.Status != "ACTIVE" {
		return nil, fmt.Errorf("job is not accepting applications")
	}

	// Check for existing application
	existing, err := s.jobRepo.GetUserApplication(ctx, jobID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing application: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("already applied to this job")
	}

	app, err := s.jobRepo.CreateApplication(ctx, jobID, userID, coverNote)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return app, nil
}

// CreateJob creates a new job posting for a company.
func (s *JobService) CreateJob(ctx context.Context, companyID, title, description, location, jobType, salary string, requiredAchievements []string) (*repository.Job, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	reqBytes, err := json.Marshal(requiredAchievements)
	if err != nil {
		reqBytes = []byte("[]")
	}

	job := &repository.Job{
		CompanyID:            companyID,
		Title:                title,
		Description:          description,
		Location:             location,
		JobType:              jobType,
		SalaryRange:          salary,
		RequiredAchievements: reqBytes,
		Status:               "ACTIVE",
	}

	return s.jobRepo.CreateJob(ctx, job)
}

// scoreRequirements checks how many of the requirements the user satisfies.
//
// Requirement formats:
//   - "TYPE"           — user needs at least one achievement of this type
//   - "TYPE:N"         — meaning depends on the type:
//       * threshold types (STARS_MILESTONE, SUBSCRIBERS_MILESTONE, VIEWS_MILESTONE):
//           user needs an achievement where metadata.current >= N (or metadata.threshold >= N)
//       * streak type (COMMIT_STREAK): user needs metadata.maxStreak >= N (fallback metadata.days)
//       * count types (anything else, e.g. PR_MERGED, REPO_CREATED, VIDEO_PUBLISHED):
//           user needs at least N achievements of this type
//
// Returns the count matched and the list of requirements that matched.
func scoreRequirements(requirements []string, userAchievements []repository.AchievementWithUser) (int, []string) {
	if len(requirements) == 0 {
		return 0, nil
	}

	// Group user achievements by type for fast lookup
	byType := make(map[string][]repository.AchievementWithUser)
	for _, a := range userAchievements {
		byType[a.Type] = append(byType[a.Type], a)
	}

	matched := 0
	matchedList := make([]string, 0, len(requirements))

	for _, req := range requirements {
		reqType := req
		var threshold int
		var hasThreshold bool

		if idx := strings.Index(req, ":"); idx > 0 {
			reqType = req[:idx]
			if n, err := strconv.Atoi(strings.TrimSpace(req[idx+1:])); err == nil {
				threshold = n
				hasThreshold = true
			}
		}

		userAch, ok := byType[reqType]
		if !ok {
			continue
		}

		if !hasThreshold {
			matched++
			matchedList = append(matchedList, req)
			continue
		}

		if requirementSatisfied(reqType, threshold, userAch) {
			matched++
			matchedList = append(matchedList, req)
		}
	}

	return matched, matchedList
}

// requirementSatisfied determines whether the user achievements satisfy a TYPE:N requirement.
func requirementSatisfied(reqType string, threshold int, achievements []repository.AchievementWithUser) bool {
	// Threshold-based types — check metadata.current or metadata.threshold
	thresholdTypes := map[string]bool{
		"STARS_MILESTONE":       true,
		"SUBSCRIBERS_MILESTONE": true,
		"VIEWS_MILESTONE":       true,
	}
	if thresholdTypes[reqType] {
		for _, a := range achievements {
			var meta map[string]interface{}
			if err := json.Unmarshal(a.Metadata, &meta); err != nil {
				continue
			}
			if v, ok := meta["current"]; ok {
				if n, ok := toInt(v); ok && n >= threshold {
					return true
				}
			}
			if v, ok := meta["threshold"]; ok {
				if n, ok := toInt(v); ok && n >= threshold {
					return true
				}
			}
		}
		return false
	}

	if reqType == "COMMIT_STREAK" {
		for _, a := range achievements {
			var meta map[string]interface{}
			if err := json.Unmarshal(a.Metadata, &meta); err != nil {
				continue
			}
			if v, ok := meta["maxStreak"]; ok {
				if n, ok := toInt(v); ok && n >= threshold {
					return true
				}
			}
			if v, ok := meta["days"]; ok {
				if n, ok := toInt(v); ok && n >= threshold {
					return true
				}
			}
		}
		return false
	}

	// Count-based types (e.g. PR_MERGED, REPO_CREATED, VIDEO_PUBLISHED)
	return len(achievements) >= threshold
}

// toInt converts a JSON-decoded number to int.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}
