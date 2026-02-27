-- name: CreateUser :one
INSERT INTO users (tenant_id, phone, email, name, password_hash, role, status, gender, date_of_birth, avatar_url, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE tenant_id = $1 AND phone = $2 AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE tenant_id IS NOT DISTINCT FROM $1 AND email = $2 AND deleted_at IS NULL LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateUser :one
UPDATE users SET
  name = COALESCE(sqlc.narg(name), name),
  email = COALESCE(sqlc.narg(email), email),
  avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
  date_of_birth = COALESCE(sqlc.narg(date_of_birth), date_of_birth),
  gender = COALESCE(sqlc.narg(gender), gender),
  device_push_token = COALESCE(sqlc.narg(device_push_token), device_push_token),
  device_platform = COALESCE(sqlc.narg(device_platform), device_platform),
  last_login_at = COALESCE(sqlc.narg(last_login_at), last_login_at),
  last_login_ip = COALESCE(sqlc.narg(last_login_ip), last_login_ip)
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users SET deleted_at = NOW() WHERE id = $1;
