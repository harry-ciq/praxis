package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Experience struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	CompanyName string     `json:"companyName"`
	Role        string     `json:"role"`
	StartDate   time.Time  `json:"startDate"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	Description string     `json:"description"`
	IsCurrent   bool       `json:"isCurrent"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type ExperienceRepo struct {
	pool *pgxpool.Pool
}

func NewExperienceRepo(pool *pgxpool.Pool) *ExperienceRepo {
	return &ExperienceRepo{pool: pool}
}

func (r *ExperienceRepo) Create(ctx context.Context, exp *Experience) (*Experience, error) {
	var e Experience
	err := r.pool.QueryRow(ctx,
		`INSERT INTO experiences (user_id, company_name, role, start_date, end_date, description, is_current)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, user_id, company_name, role, start_date, end_date, description, is_current, created_at, updated_at`,
		exp.UserID, exp.CompanyName, exp.Role, exp.StartDate, exp.EndDate, exp.Description, exp.IsCurrent,
	).Scan(&e.ID, &e.UserID, &e.CompanyName, &e.Role, &e.StartDate, &e.EndDate, &e.Description, &e.IsCurrent, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExperienceRepo) Update(ctx context.Context, id string, exp *Experience) (*Experience, error) {
	var e Experience
	err := r.pool.QueryRow(ctx,
		`UPDATE experiences
		 SET company_name = $2, role = $3, start_date = $4, end_date = $5, description = $6, is_current = $7, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, user_id, company_name, role, start_date, end_date, description, is_current, created_at, updated_at`,
		id, exp.CompanyName, exp.Role, exp.StartDate, exp.EndDate, exp.Description, exp.IsCurrent,
	).Scan(&e.ID, &e.UserID, &e.CompanyName, &e.Role, &e.StartDate, &e.EndDate, &e.Description, &e.IsCurrent, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExperienceRepo) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM experiences WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *ExperienceRepo) ListByUserID(ctx context.Context, userID string) ([]Experience, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, company_name, role, start_date, end_date, description, is_current, created_at, updated_at
		 FROM experiences
		 WHERE user_id = $1
		 ORDER BY is_current DESC, start_date DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var experiences []Experience
	for rows.Next() {
		var e Experience
		if err := rows.Scan(&e.ID, &e.UserID, &e.CompanyName, &e.Role, &e.StartDate, &e.EndDate, &e.Description, &e.IsCurrent, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		experiences = append(experiences, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if experiences == nil {
		experiences = []Experience{}
	}
	return experiences, nil
}

func (r *ExperienceRepo) GetByID(ctx context.Context, id string) (*Experience, error) {
	var e Experience
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, company_name, role, start_date, end_date, description, is_current, created_at, updated_at
		 FROM experiences WHERE id = $1`,
		id,
	).Scan(&e.ID, &e.UserID, &e.CompanyName, &e.Role, &e.StartDate, &e.EndDate, &e.Description, &e.IsCurrent, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}
