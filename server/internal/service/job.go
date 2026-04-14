package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/praxis-social/praxis/server/internal/repository"
)

type JobWithMatchScore struct {
	repository.JobWithCompany
	MatchedAchievements int `json:"matchedAchievements"`
	TotalRequired       int `json:"totalRequired"`
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

// ListJobs returns active jobs with pagination.
func (s *JobService) ListJobs(ctx context.Context, limit, offset int) ([]JobWithMatchScore, error) {
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

	results := make([]JobWithMatchScore, len(jobs))
	for i, job := range jobs {
		results[i] = JobWithMatchScore{
			JobWithCompany: job,
		}
		// Parse requirements to set TotalRequired
		var reqs []string
		if err := json.Unmarshal(job.RequiredAchievements, &reqs); err == nil {
			results[i].TotalRequired = len(reqs)
		}
	}

	return results, nil
}

// GetJob returns a single job by ID.
func (s *JobService) GetJob(ctx context.Context, id string) (*repository.JobWithCompany, error) {
	return s.jobRepo.GetByID(ctx, id)
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

// CalculateMatchScore checks how many required achievements a user has.
// Requirements are formatted as "TYPE:VALUE" (e.g., "COMMIT_STREAK:30").
// It checks if the user has achievements matching those types.
func (s *JobService) CalculateMatchScore(ctx context.Context, userID string, requirements []string) (matched int, total int, err error) {
	total = len(requirements)
	if total == 0 {
		return 0, 0, nil
	}

	// Get all user achievements
	achievements, err := s.achievementRepo.ListByUserID(ctx, userID, 1000, 0)
	if err != nil {
		return 0, total, fmt.Errorf("failed to list user achievements: %w", err)
	}

	// Build a set of achievement types the user has
	userTypes := make(map[string]bool)
	for _, a := range achievements {
		userTypes[a.Type] = true
	}

	// Check each requirement
	for _, req := range requirements {
		// Parse "TYPE:VALUE" format - match on the TYPE part
		reqType := req
		if idx := strings.Index(req, ":"); idx > 0 {
			reqType = req[:idx]
		}
		if userTypes[reqType] {
			matched++
		}
	}

	return matched, total, nil
}
