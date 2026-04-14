-- name: CreateJob :one
INSERT INTO jobs (company_id, title, description, location, job_type, salary_range, required_achievements)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetJobByID :one
SELECT j.*, c.name as company_name, c.logo_url as company_logo_url, c.website as company_website
FROM jobs j
JOIN companies c ON j.company_id = c.id
WHERE j.id = $1;

-- name: ListActiveJobs :many
SELECT j.*, c.name as company_name, c.logo_url as company_logo_url, c.website as company_website
FROM jobs j
JOIN companies c ON j.company_id = c.id
WHERE j.status = 'ACTIVE'
ORDER BY j.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateJobApplication :one
INSERT INTO job_applications (job_id, user_id, cover_note)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserApplication :one
SELECT * FROM job_applications WHERE job_id = $1 AND user_id = $2;

-- name: CreateCompany :one
INSERT INTO companies (name, logo_url, description, website)
VALUES ($1, $2, $3, $4)
RETURNING *;
