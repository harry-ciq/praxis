-- name: CreateConversation :one
INSERT INTO conversations DEFAULT VALUES
RETURNING *;

-- name: AddConversationParticipant :exec
INSERT INTO conversation_participants (conversation_id, user_id)
VALUES ($1, $2);

-- name: ListUserConversations :many
SELECT c.*,
    (SELECT content FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1) as last_message_content,
    (SELECT sender_id FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1) as last_message_sender_id,
    (SELECT created_at FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1) as last_message_at,
    (SELECT COUNT(*) FROM messages m WHERE m.conversation_id = c.id AND m.created_at > COALESCE(cp.last_read_at, '1970-01-01')) as unread_count
FROM conversations c
JOIN conversation_participants cp ON c.id = cp.conversation_id AND cp.user_id = $1
ORDER BY c.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListMessages :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateConversationTimestamp :exec
UPDATE conversations SET updated_at = NOW() WHERE id = $1;

-- name: UpdateLastRead :exec
UPDATE conversation_participants
SET last_read_at = NOW()
WHERE conversation_id = $1 AND user_id = $2;

-- name: GetConversationParticipants :many
SELECT u.id, u.username, u.name, u.avatar_url
FROM users u
JOIN conversation_participants cp ON u.id = cp.user_id
WHERE cp.conversation_id = $1;

-- name: FindDirectConversation :one
SELECT cp1.conversation_id FROM conversation_participants cp1
JOIN conversation_participants cp2 ON cp1.conversation_id = cp2.conversation_id
WHERE cp1.user_id = $1 AND cp2.user_id = $2
LIMIT 1;
