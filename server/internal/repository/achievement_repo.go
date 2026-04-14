package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Achievement struct {
	ID               string          `json:"id"`
	UserID           string          `json:"userId"`
	Type             string          `json:"type"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Metadata         json.RawMessage `json:"metadata"`
	ProofURL         string          `json:"proofUrl"`
	ProofData        json.RawMessage `json:"proofData"`
	Source           string          `json:"source"`
	SourceID         string          `json:"sourceId"`
	VerificationHash *string         `json:"verificationHash,omitempty"`
	Status           string          `json:"status"` // "active" or "archived"
	CreatedAt        time.Time       `json:"createdAt"`
}

type AchievementWithUser struct {
	Achievement
	UserUsername  string `json:"userUsername"`
	UserName     string `json:"userName"`
	UserAvatarURL string `json:"userAvatarUrl"`
}

type ReactionCounts struct {
	Clap   int `json:"clap"`
	Fire   int `json:"fire"`
	Rocket int `json:"rocket"`
	Total  int `json:"total"`
}

type AchievementRepo struct {
	pool *pgxpool.Pool
}

func NewAchievementRepo(pool *pgxpool.Pool) *AchievementRepo {
	return &AchievementRepo{pool: pool}
}

func (r *AchievementRepo) Create(ctx context.Context, achievement *Achievement) (*Achievement, error) {
	metaBytes := achievement.Metadata
	if metaBytes == nil {
		metaBytes = json.RawMessage(`{}`)
	}
	proofBytes := achievement.ProofData
	if proofBytes == nil {
		proofBytes = json.RawMessage(`{}`)
	}

	// Use the provided CreatedAt if set, otherwise let the DB default to now()
	var a Achievement
	if !achievement.CreatedAt.IsZero() {
		err := r.pool.QueryRow(ctx,
			`INSERT INTO achievements (user_id, type, title, description, metadata, proof_url, proof_data, source, source_id, verification_hash, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			 RETURNING id, user_id, type, title, description, metadata, proof_url, proof_data, source, source_id, verification_hash, status, created_at`,
			achievement.UserID, achievement.Type, achievement.Title, achievement.Description,
			metaBytes, achievement.ProofURL, proofBytes,
			achievement.Source, achievement.SourceID, achievement.VerificationHash,
			achievement.CreatedAt,
		).Scan(&a.ID, &a.UserID, &a.Type, &a.Title, &a.Description, &a.Metadata,
			&a.ProofURL, &a.ProofData, &a.Source, &a.SourceID, &a.VerificationHash, &a.Status, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.pool.QueryRow(ctx,
			`INSERT INTO achievements (user_id, type, title, description, metadata, proof_url, proof_data, source, source_id, verification_hash)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			 RETURNING id, user_id, type, title, description, metadata, proof_url, proof_data, source, source_id, verification_hash, status, created_at`,
			achievement.UserID, achievement.Type, achievement.Title, achievement.Description,
			metaBytes, achievement.ProofURL, proofBytes,
			achievement.Source, achievement.SourceID, achievement.VerificationHash,
		).Scan(&a.ID, &a.UserID, &a.Type, &a.Title, &a.Description, &a.Metadata,
			&a.ProofURL, &a.ProofData, &a.Source, &a.SourceID, &a.VerificationHash, &a.Status, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
	}
	return &a, nil
}

func (r *AchievementRepo) GetByID(ctx context.Context, id string) (*AchievementWithUser, error) {
	var a AchievementWithUser
	err := r.pool.QueryRow(ctx,
		`SELECT a.id, a.user_id, a.type, a.title, a.description, a.metadata, a.proof_url, a.proof_data,
		        a.source, a.source_id, a.verification_hash, a.status, a.created_at,
		        u.username, u.name, u.avatar_url
		 FROM achievements a
		 JOIN users u ON u.id = a.user_id
		 WHERE a.id = $1`,
		id,
	).Scan(&a.ID, &a.UserID, &a.Type, &a.Title, &a.Description, &a.Metadata,
		&a.ProofURL, &a.ProofData, &a.Source, &a.SourceID, &a.VerificationHash, &a.Status, &a.CreatedAt,
		&a.UserUsername, &a.UserName, &a.UserAvatarURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AchievementRepo) GetBySourceID(ctx context.Context, sourceID string) (*Achievement, error) {
	var a Achievement
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, type, title, description, metadata, proof_url, proof_data, source, source_id, verification_hash, status, created_at
		 FROM achievements WHERE source_id = $1`,
		sourceID,
	).Scan(&a.ID, &a.UserID, &a.Type, &a.Title, &a.Description, &a.Metadata,
		&a.ProofURL, &a.ProofData, &a.Source, &a.SourceID, &a.VerificationHash, &a.Status, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AchievementRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]AchievementWithUser, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.user_id, a.type, a.title, a.description, a.metadata, a.proof_url, a.proof_data,
		        a.source, a.source_id, a.verification_hash, a.status, a.created_at,
		        u.username, u.name, u.avatar_url
		 FROM achievements a
		 JOIN users u ON u.id = a.user_id
		 WHERE a.user_id = $1
		 ORDER BY a.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAchievementsWithUser(rows)
}

func (r *AchievementRepo) ListFeed(ctx context.Context, userID string, limit, offset int) ([]AchievementWithUser, error) {
	// Feed shows all achievements, ordered by most recent.
	// The userID parameter is reserved for future use (e.g., showing only followed users).
	rows, err := r.pool.Query(ctx,
		`SELECT a.id, a.user_id, a.type, a.title, a.description, a.metadata, a.proof_url, a.proof_data,
		        a.source, a.source_id, a.verification_hash, a.status, a.created_at,
		        u.username, u.name, u.avatar_url
		 FROM achievements a
		 JOIN users u ON u.id = a.user_id
		 ORDER BY a.created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAchievementsWithUser(rows)
}

func (r *AchievementRepo) GetReactionCounts(ctx context.Context, achievementID string) (*ReactionCounts, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT type, COUNT(*) FROM reactions WHERE achievement_id = $1 GROUP BY type`,
		achievementID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := &ReactionCounts{}
	for rows.Next() {
		var rType string
		var count int
		if err := rows.Scan(&rType, &count); err != nil {
			return nil, err
		}
		switch rType {
		case "CLAP":
			counts.Clap = count
		case "FIRE":
			counts.Fire = count
		case "ROCKET":
			counts.Rocket = count
		}
		counts.Total += count
	}
	return counts, rows.Err()
}

func (r *AchievementRepo) GetUserReaction(ctx context.Context, userID, achievementID string) (string, error) {
	var rType string
	err := r.pool.QueryRow(ctx,
		`SELECT type FROM reactions WHERE user_id = $1 AND achievement_id = $2`,
		userID, achievementID,
	).Scan(&rType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return rType, nil
}

func (r *AchievementRepo) CreateReaction(ctx context.Context, userID, achievementID, reactionType string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reactions (user_id, achievement_id, type)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, achievement_id) DO UPDATE SET type = $3`,
		userID, achievementID, reactionType,
	)
	return err
}

func (r *AchievementRepo) DeleteReaction(ctx context.Context, userID, achievementID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM reactions WHERE user_id = $1 AND achievement_id = $2`,
		userID, achievementID,
	)
	return err
}

// ArchiveBySourceIDs marks achievements as archived when their source is no longer available.
func (r *AchievementRepo) ArchiveBySourceIDs(ctx context.Context, sourceIDs []string) error {
	if len(sourceIDs) == 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE achievements SET status = 'archived' WHERE source_id = ANY($1) AND status = 'active'`,
		sourceIDs,
	)
	return err
}

// ReactivateBySourceIDs marks achievements as active when their source is available again.
func (r *AchievementRepo) ReactivateBySourceIDs(ctx context.Context, sourceIDs []string) error {
	if len(sourceIDs) == 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE achievements SET status = 'active' WHERE source_id = ANY($1) AND status = 'archived'`,
		sourceIDs,
	)
	return err
}

// ListSourceIDsByUserAndPrefix returns source_ids for a user matching a prefix (e.g., "github:repo:").
func (r *AchievementRepo) ListSourceIDsByUserAndPrefix(ctx context.Context, userID, prefix string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT source_id FROM achievements WHERE user_id = $1 AND source_id LIKE $2`,
		userID, prefix+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *AchievementRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM achievements WHERE user_id = $1`,
		userID,
	).Scan(&count)
	return count, err
}

func scanAchievementsWithUser(rows pgx.Rows) ([]AchievementWithUser, error) {
	var results []AchievementWithUser
	for rows.Next() {
		var a AchievementWithUser
		if err := rows.Scan(&a.ID, &a.UserID, &a.Type, &a.Title, &a.Description, &a.Metadata,
			&a.ProofURL, &a.ProofData, &a.Source, &a.SourceID, &a.VerificationHash, &a.Status, &a.CreatedAt,
			&a.UserUsername, &a.UserName, &a.UserAvatarURL); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []AchievementWithUser{}
	}
	return results, nil
}
