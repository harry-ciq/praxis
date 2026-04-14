-- name: CreateUser :one
INSERT INTO users (username, email, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET name = COALESCE(sqlc.narg('name'), name),
    bio = COALESCE(sqlc.narg('bio'), bio),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    headline = COALESCE(sqlc.narg('headline'), headline),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserProfile :one
SELECT u.*,
    (SELECT COUNT(*) FROM achievements WHERE user_id = u.id) as achievement_count,
    (SELECT COUNT(*) FROM follows WHERE following_id = u.id) as follower_count,
    (SELECT COUNT(*) FROM follows WHERE follower_id = u.id) as following_count
FROM users u WHERE u.username = $1;
