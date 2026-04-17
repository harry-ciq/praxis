package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConnectedProvider struct {
	ID               string     `json:"id"`
	UserID           string     `json:"userId"`
	Provider         string     `json:"provider"`
	ProviderUsername string     `json:"providerUsername"`
	LastSyncedAt     *time.Time `json:"lastSyncedAt,omitempty"`
	SyncStatus       string     `json:"syncStatus"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type ProviderRepo struct {
	pool *pgxpool.Pool
}

func NewProviderRepo(pool *pgxpool.Pool) *ProviderRepo {
	return &ProviderRepo{pool: pool}
}

func (r *ProviderRepo) Create(ctx context.Context, userID, provider, providerUsername string) (*ConnectedProvider, error) {
	var cp ConnectedProvider
	err := r.pool.QueryRow(ctx,
		`INSERT INTO connected_providers (user_id, provider, provider_username, sync_status)
		 VALUES ($1, $2, $3, 'IDLE')
		 RETURNING id, user_id, provider, provider_username, last_synced_at, sync_status, created_at`,
		userID, provider, providerUsername,
	).Scan(&cp.ID, &cp.UserID, &cp.Provider, &cp.ProviderUsername, &cp.LastSyncedAt, &cp.SyncStatus, &cp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &cp, nil
}

func (r *ProviderRepo) Get(ctx context.Context, userID, provider string) (*ConnectedProvider, error) {
	var cp ConnectedProvider
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, provider_username, last_synced_at, sync_status, created_at
		 FROM connected_providers WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	).Scan(&cp.ID, &cp.UserID, &cp.Provider, &cp.ProviderUsername, &cp.LastSyncedAt, &cp.SyncStatus, &cp.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cp, nil
}

func (r *ProviderRepo) UpdateSyncStatus(ctx context.Context, userID, provider, status string) error {
	var query string
	if status == "IDLE" {
		query = `UPDATE connected_providers SET sync_status = $3, last_synced_at = NOW() WHERE user_id = $1 AND provider = $2`
	} else {
		query = `UPDATE connected_providers SET sync_status = $3 WHERE user_id = $1 AND provider = $2`
	}
	_, err := r.pool.Exec(ctx, query, userID, provider, status)
	return err
}

func (r *ProviderRepo) List(ctx context.Context, userID string) ([]ConnectedProvider, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, provider_username, last_synced_at, sync_status, created_at
		 FROM connected_providers WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ConnectedProvider
	for rows.Next() {
		var cp ConnectedProvider
		if err := rows.Scan(&cp.ID, &cp.UserID, &cp.Provider, &cp.ProviderUsername, &cp.LastSyncedAt, &cp.SyncStatus, &cp.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []ConnectedProvider{}
	}
	return results, nil
}

// ListAll returns every connected provider across all users — used by the
// periodic sync scheduler.
func (r *ProviderRepo) ListAll(ctx context.Context) ([]ConnectedProvider, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, provider_username, last_synced_at, sync_status, created_at
		 FROM connected_providers
		 ORDER BY last_synced_at NULLS FIRST`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ConnectedProvider
	for rows.Next() {
		var cp ConnectedProvider
		if err := rows.Scan(&cp.ID, &cp.UserID, &cp.Provider, &cp.ProviderUsername, &cp.LastSyncedAt, &cp.SyncStatus, &cp.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *ProviderRepo) Delete(ctx context.Context, userID, provider string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM connected_providers WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)
	return err
}
