package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FollowRepo struct {
	pool *pgxpool.Pool
}

func NewFollowRepo(pool *pgxpool.Pool) *FollowRepo {
	return &FollowRepo{pool: pool}
}

func (r *FollowRepo) Follow(ctx context.Context, followerID, followingID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO follows (follower_id, following_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		followerID, followingID,
	)
	return err
}

func (r *FollowRepo) Unfollow(ctx context.Context, followerID, followingID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	return err
}

func (r *FollowRepo) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	).Scan(&exists)
	return exists, err
}

func (r *FollowRepo) ListFollowers(ctx context.Context, userID string, limit, offset int) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, u.email, u.name, u.bio, u.avatar_url, u.headline, u.location, u.website_url, u.social_links, u.created_at, u.updated_at
		 FROM follows f
		 JOIN users u ON u.id = f.follower_id
		 WHERE f.following_id = $1
		 ORDER BY f.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUsers(rows)
}

func (r *FollowRepo) ListFollowing(ctx context.Context, userID string, limit, offset int) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, u.email, u.name, u.bio, u.avatar_url, u.headline, u.location, u.website_url, u.social_links, u.created_at, u.updated_at
		 FROM follows f
		 JOIN users u ON u.id = f.following_id
		 WHERE f.follower_id = $1
		 ORDER BY f.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUsers(rows)
}

func (r *FollowRepo) GetCounts(ctx context.Context, userID string) (followers int, following int, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT
			(SELECT COUNT(*) FROM follows WHERE following_id = $1),
			(SELECT COUNT(*) FROM follows WHERE follower_id = $1)`,
		userID,
	).Scan(&followers, &following)
	return
}

func scanUsers(rows pgx.Rows) ([]User, error) {
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Bio, &u.AvatarURL, &u.Headline, &u.Location, &u.WebsiteURL, &u.SocialLinks, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if users == nil {
		users = []User{}
	}
	return users, nil
}
