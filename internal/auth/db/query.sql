-- name: GetApiKeyByHash :one
SELECT id, tenant_id, revoked, expires_at
FROM api_keys
WHERE key_hash = $1
LIMIT 1;

-- name: CreateApiKey :one
INSERT INTO api_keys (tenant_id, key_hash, key_prefix, name)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at;

-- name: RevokeApiKey :exec
UPDATE api_keys
SET revoked = true
WHERE id = $1;

-- name: ListApiKeys :many
SELECT id, name, key_prefix, created_at, last_used_at, revoked, expires_at
FROM api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: CreateTenant :one
INSERT INTO tenants (name, slug, contact_email)
VALUES ($1, $2, $3)
RETURNING id;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserByEmail :one
SELECT id, email, full_name
FROM users
WHERE email = $1;

-- name: GetUserByEmailWithPassword :one
SELECT id, email, password_hash
FROM users
WHERE email = $1;