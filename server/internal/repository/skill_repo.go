package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Skill struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	Verified  bool      `json:"verified"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"createdAt"`
}

type SkillInput struct {
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
	Source   string `json:"source"`
}

type SkillRepo struct {
	pool *pgxpool.Pool
}

func NewSkillRepo(pool *pgxpool.Pool) *SkillRepo {
	return &SkillRepo{pool: pool}
}

func (r *SkillRepo) Create(ctx context.Context, userID, name string, verified bool, source string) (*Skill, error) {
	var s Skill
	err := r.pool.QueryRow(ctx,
		`INSERT INTO skills (user_id, name, verified, source)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, name) DO UPDATE SET verified = EXCLUDED.verified, source = EXCLUDED.source
		 RETURNING id, user_id, name, verified, source, created_at`,
		userID, name, verified, source,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.Verified, &s.Source, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SkillRepo) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM skills WHERE id = $1 AND user_id = $2`,
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

func (r *SkillRepo) ListByUserID(ctx context.Context, userID string) ([]Skill, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, verified, source, created_at
		 FROM skills
		 WHERE user_id = $1
		 ORDER BY verified DESC, name ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Verified, &s.Source, &s.CreatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if skills == nil {
		skills = []Skill{}
	}
	return skills, nil
}

func (r *SkillRepo) BulkUpsert(ctx context.Context, userID string, skills []SkillInput) error {
	for _, s := range skills {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO skills (user_id, name, verified, source)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id, name) DO UPDATE SET verified = EXCLUDED.verified, source = EXCLUDED.source`,
			userID, s.Name, s.Verified, s.Source,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
