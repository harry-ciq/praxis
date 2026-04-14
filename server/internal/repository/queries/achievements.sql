-- name: CreateAchievement :one
INSERT INTO achievements (user_id, type, title, description, metadata, proof_url, proof_data, source, source_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetAchievementByID :one
SELECT a.*, u.username as user_username, u.name as user_name, u.avatar_url as user_avatar_url
FROM achievements a
JOIN users u ON a.user_id = u.id
WHERE a.id = $1;

-- name: GetAchievementBySourceID :one
SELECT * FROM achievements WHERE source_id = $1;

-- name: ListUserAchievements :many
SELECT a.*, u.username as user_username, u.name as user_name, u.avatar_url as user_avatar_url
FROM achievements a
JOIN users u ON a.user_id = u.id
WHERE a.user_id = $1
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListFeedAchievements :many
SELECT a.*, u.username as user_username, u.name as user_name, u.avatar_url as user_avatar_url
FROM achievements a
JOIN users u ON a.user_id = u.id
WHERE a.user_id = ANY(
    SELECT following_id FROM follows WHERE follower_id = $1
    UNION SELECT $1::uuid
)
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetReactionCounts :one
SELECT
    COUNT(*) FILTER (WHERE type = 'CLAP') as clap_count,
    COUNT(*) FILTER (WHERE type = 'FIRE') as fire_count,
    COUNT(*) FILTER (WHERE type = 'ROCKET') as rocket_count,
    COUNT(*) as total_count
FROM reactions WHERE achievement_id = $1;

-- name: GetUserReaction :one
SELECT * FROM reactions WHERE user_id = $1 AND achievement_id = $2;

-- name: CreateReaction :one
INSERT INTO reactions (user_id, achievement_id, type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteReaction :exec
DELETE FROM reactions WHERE user_id = $1 AND achievement_id = $2;
