package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conversation struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConversationWithDetails struct {
	ID                  string                    `json:"id"`
	LastMessageContent  *string                   `json:"lastMessageContent"`
	LastMessageSenderID *string                   `json:"lastMessageSenderId"`
	LastMessageAt       *time.Time                `json:"lastMessageAt"`
	UnreadCount         int                       `json:"unreadCount"`
	UpdatedAt           time.Time                 `json:"updatedAt"`
	Participants        []ConversationParticipant `json:"participants"`
}

type ConversationParticipant struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	SenderID       string    `json:"senderId"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MessageRepo struct {
	pool *pgxpool.Pool
}

func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

func (r *MessageRepo) CreateConversation(ctx context.Context) (*Conversation, error) {
	var c Conversation
	err := r.pool.QueryRow(ctx,
		`INSERT INTO conversations DEFAULT VALUES
		 RETURNING id, created_at, updated_at`,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *MessageRepo) AddParticipant(ctx context.Context, conversationID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO conversation_participants (conversation_id, user_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		conversationID, userID,
	)
	return err
}

func (r *MessageRepo) ListUserConversations(ctx context.Context, userID string, limit, offset int) ([]ConversationWithDetails, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT
			c.id,
			lm.content AS last_message_content,
			lm.sender_id AS last_message_sender_id,
			lm.created_at AS last_message_at,
			COALESCE(
				(SELECT COUNT(*) FROM messages m2
				 WHERE m2.conversation_id = c.id
				   AND m2.created_at > COALESCE(
				     (SELECT last_read_at FROM conversation_participants
				      WHERE conversation_id = c.id AND user_id = $1), '1970-01-01'
				   )
				   AND m2.sender_id != $1
				), 0
			) AS unread_count,
			c.updated_at
		 FROM conversations c
		 JOIN conversation_participants cp ON cp.conversation_id = c.id
		 LEFT JOIN LATERAL (
			SELECT content, sender_id, created_at
			FROM messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		 ) lm ON true
		 WHERE cp.user_id = $1
		 ORDER BY c.updated_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationWithDetails
	for rows.Next() {
		var c ConversationWithDetails
		if err := rows.Scan(
			&c.ID,
			&c.LastMessageContent,
			&c.LastMessageSenderID,
			&c.LastMessageAt,
			&c.UnreadCount,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch participants for each conversation
	for i := range conversations {
		participants, err := r.GetParticipants(ctx, conversations[i].ID)
		if err != nil {
			return nil, err
		}
		conversations[i].Participants = participants
	}

	return conversations, nil
}

func (r *MessageRepo) CreateMessage(ctx context.Context, conversationID, senderID, content string) (*Message, error) {
	var m Message
	err := r.pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, sender_id, content)
		 VALUES ($1, $2, $3)
		 RETURNING id, conversation_id, sender_id, content, created_at`,
		conversationID, senderID, content,
	).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MessageRepo) ListMessages(ctx context.Context, conversationID string, limit, offset int) ([]Message, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, conversation_id, sender_id, content, created_at
		 FROM messages
		 WHERE conversation_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		conversationID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *MessageRepo) UpdateConversationTimestamp(ctx context.Context, conversationID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE conversations SET updated_at = NOW() WHERE id = $1`,
		conversationID,
	)
	return err
}

func (r *MessageRepo) UpdateLastRead(ctx context.Context, conversationID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE conversation_participants SET last_read_at = NOW()
		 WHERE conversation_id = $1 AND user_id = $2`,
		conversationID, userID,
	)
	return err
}

func (r *MessageRepo) GetParticipants(ctx context.Context, conversationID string) ([]ConversationParticipant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, u.name, u.avatar_url
		 FROM conversation_participants cp
		 JOIN users u ON u.id = cp.user_id
		 WHERE cp.conversation_id = $1`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []ConversationParticipant
	for rows.Next() {
		var p ConversationParticipant
		if err := rows.Scan(&p.ID, &p.Username, &p.Name, &p.AvatarURL); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, rows.Err()
}

func (r *MessageRepo) FindDirectConversation(ctx context.Context, userID1, userID2 string) (string, error) {
	var conversationID string
	err := r.pool.QueryRow(ctx,
		`SELECT cp1.conversation_id
		 FROM conversation_participants cp1
		 JOIN conversation_participants cp2 ON cp1.conversation_id = cp2.conversation_id
		 JOIN conversations c ON c.id = cp1.conversation_id
		 WHERE cp1.user_id = $1 AND cp2.user_id = $2
		   AND (SELECT COUNT(*) FROM conversation_participants
		        WHERE conversation_id = cp1.conversation_id) = 2
		 LIMIT 1`,
		userID1, userID2,
	).Scan(&conversationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return conversationID, nil
}

func (r *MessageRepo) DeleteMessage(ctx context.Context, messageID, senderID string) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM messages WHERE id = $1 AND sender_id = $2`,
		messageID, senderID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MessageRepo) DeleteConversation(ctx context.Context, conversationID string) error {
	// Messages are cascade-deleted via FK
	_, err := r.pool.Exec(ctx,
		`DELETE FROM conversations WHERE id = $1`,
		conversationID,
	)
	return err
}

func (r *MessageRepo) IsParticipant(ctx context.Context, conversationID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM conversation_participants
			WHERE conversation_id = $1 AND user_id = $2
		)`,
		conversationID, userID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
