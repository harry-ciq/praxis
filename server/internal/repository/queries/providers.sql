-- name: CreateConnectedProvider :one
INSERT INTO connected_providers (user_id, provider, provider_username)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConnectedProvider :one
SELECT * FROM connected_providers WHERE user_id = $1 AND provider = $2;

-- name: UpdateProviderSyncStatus :exec
UPDATE connected_providers
SET last_synced_at = NOW(), sync_status = $1
WHERE user_id = $2 AND provider = $3;

-- name: ListUserProviders :many
SELECT * FROM connected_providers WHERE user_id = $1;

-- name: DeleteConnectedProvider :exec
DELETE FROM connected_providers WHERE user_id = $1 AND provider = $2;
