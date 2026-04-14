-- name: CreateAuthAccount :one
INSERT INTO auth_accounts (user_id, provider, provider_account_id, access_token, refresh_token, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAuthAccount :one
SELECT * FROM auth_accounts
WHERE provider = $1 AND provider_account_id = $2;

-- name: UpdateAuthTokens :exec
UPDATE auth_accounts
SET access_token = $1, refresh_token = $2, expires_at = $3
WHERE id = $4;
