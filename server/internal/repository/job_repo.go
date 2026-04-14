package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
	ID                   string          `json:"id"`
	CompanyID            string          `json:"companyId"`
	Title                string          `json:"title"`
	Description          string          `json:"description"`
	Location             string          `json:"location"`
	JobType              string          `json:"jobType"`
	SalaryRange          string          `json:"salaryRange"`
	RequiredAchievements json.RawMessage `json:"requiredAchievements"`
	Status               string          `json:"status"`
	CreatedAt            time.Time       `json:"createdAt"`
}

type JobWithCompany struct {
	Job
	CompanyName    string `json:"companyName"`
	CompanyLogoURL string `json:"companyLogoUrl"`
	CompanyWebsite string `json:"companyWebsite"`
}

type Company struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	LogoURL     string    `json:"logoUrl"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	CreatedAt   time.Time `json:"createdAt"`
}

type JobApplication struct {
	ID        string    `json:"id"`
	JobID     string    `json:"jobId"`
	UserID    string    `json:"userId"`
	CoverNote string    `json:"coverNote"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type JobRepo struct {
	pool *pgxpool.Pool
}

func NewJobRepo(pool *pgxpool.Pool) *JobRepo {
	return &JobRepo{pool: pool}
}

func (r *JobRepo) CreateJob(ctx context.Context, job *Job) (*Job, error) {
	reqAch := job.RequiredAchievements
	if reqAch == nil {
		reqAch = json.RawMessage(`[]`)
	}

	var j Job
	err := r.pool.QueryRow(ctx,
		`INSERT INTO jobs (company_id, title, description, location, job_type, salary_range, required_achievements, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, company_id, title, description, location, job_type, salary_range, required_achievements, status, created_at`,
		job.CompanyID, job.Title, job.Description, job.Location, job.JobType, job.SalaryRange, reqAch, job.Status,
	).Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description, &j.Location, &j.JobType,
		&j.SalaryRange, &j.RequiredAchievements, &j.Status, &j.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *JobRepo) GetByID(ctx context.Context, id string) (*JobWithCompany, error) {
	var j JobWithCompany
	err := r.pool.QueryRow(ctx,
		`SELECT j.id, j.company_id, j.title, j.description, j.location, j.job_type,
		        j.salary_range, j.required_achievements, j.status, j.created_at,
		        c.name, c.logo_url, c.website
		 FROM jobs j
		 JOIN companies c ON c.id = j.company_id
		 WHERE j.id = $1`,
		id,
	).Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description, &j.Location, &j.JobType,
		&j.SalaryRange, &j.RequiredAchievements, &j.Status, &j.CreatedAt,
		&j.CompanyName, &j.CompanyLogoURL, &j.CompanyWebsite)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}

func (r *JobRepo) ListActive(ctx context.Context, limit, offset int) ([]JobWithCompany, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT j.id, j.company_id, j.title, j.description, j.location, j.job_type,
		        j.salary_range, j.required_achievements, j.status, j.created_at,
		        c.name, c.logo_url, c.website
		 FROM jobs j
		 JOIN companies c ON c.id = j.company_id
		 WHERE j.status = 'ACTIVE'
		 ORDER BY j.created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJobsWithCompany(rows)
}

func (r *JobRepo) CreateApplication(ctx context.Context, jobID, userID, coverNote string) (*JobApplication, error) {
	var app JobApplication
	err := r.pool.QueryRow(ctx,
		`INSERT INTO job_applications (job_id, user_id, cover_note, status)
		 VALUES ($1, $2, $3, 'PENDING')
		 RETURNING id, job_id, user_id, cover_note, status, created_at`,
		jobID, userID, coverNote,
	).Scan(&app.ID, &app.JobID, &app.UserID, &app.CoverNote, &app.Status, &app.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *JobRepo) GetUserApplication(ctx context.Context, jobID, userID string) (*JobApplication, error) {
	var app JobApplication
	err := r.pool.QueryRow(ctx,
		`SELECT id, job_id, user_id, cover_note, status, created_at
		 FROM job_applications
		 WHERE job_id = $1 AND user_id = $2`,
		jobID, userID,
	).Scan(&app.ID, &app.JobID, &app.UserID, &app.CoverNote, &app.Status, &app.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *JobRepo) CreateCompany(ctx context.Context, name, logoURL, description, website string) (*Company, error) {
	var c Company
	err := r.pool.QueryRow(ctx,
		`INSERT INTO companies (name, logo_url, description, website)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, logo_url, description, website, created_at`,
		name, logoURL, description, website,
	).Scan(&c.ID, &c.Name, &c.LogoURL, &c.Description, &c.Website, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *JobRepo) GetCompany(ctx context.Context, id string) (*Company, error) {
	var c Company
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, logo_url, description, website, created_at
		 FROM companies WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Name, &c.LogoURL, &c.Description, &c.Website, &c.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func scanJobsWithCompany(rows pgx.Rows) ([]JobWithCompany, error) {
	var results []JobWithCompany
	for rows.Next() {
		var j JobWithCompany
		if err := rows.Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description, &j.Location, &j.JobType,
			&j.SalaryRange, &j.RequiredAchievements, &j.Status, &j.CreatedAt,
			&j.CompanyName, &j.CompanyLogoURL, &j.CompanyWebsite); err != nil {
			return nil, err
		}
		results = append(results, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []JobWithCompany{}
	}
	return results, nil
}
