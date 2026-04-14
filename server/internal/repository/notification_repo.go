package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        string          `json:"id"`
	UserID    string          `json:"userId"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Data      json.RawMessage `json:"data"`
	Read      bool            `json:"read"`
	CreatedAt time.Time       `json:"createdAt"`
}

type NotificationRepo struct {
	pool *pgxpool.Pool
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, userID, notifType, title, body string, data map[string]interface{}) (*Notification, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		dataBytes = []byte("{}")
	}

	var n Notification
	err = r.pool.QueryRow(ctx,
		`INSERT INTO notifications (user_id, type, title, body, data)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, type, title, body, data, read, created_at`,
		userID, notifType, title, body, dataBytes,
	).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.Read, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NotificationRepo) List(ctx context.Context, userID string, limit, offset int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, type, title, body, data, read, created_at
		 FROM notifications
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []Notification{}
	}
	return results, nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET read = true WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	return err
}

func (r *NotificationRepo) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`,
		userID,
	).Scan(&count)
	return count, err
}
