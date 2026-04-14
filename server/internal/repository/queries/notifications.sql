-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body, data)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkNotificationRead :exec
UPDATE notifications SET read = true WHERE id = $1 AND user_id = $2;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false;
