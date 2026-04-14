-- name: CreateFollow :exec
INSERT INTO follows (follower_id, following_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteFollow :exec
DELETE FROM follows WHERE follower_id = $1 AND following_id = $2;

-- name: IsFollowing :one
SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2) as is_following;

-- name: ListFollowers :many
SELECT u.* FROM users u
JOIN follows f ON u.id = f.follower_id
WHERE f.following_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListFollowing :many
SELECT u.* FROM users u
JOIN follows f ON u.id = f.following_id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;
