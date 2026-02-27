-- name: CreateTenant :one
INSERT INTO tenants (slug, name, status, plan, commission_rate, settings, contact_email, contact_phone, address, timezone, currency, locale, logo_url, favicon_url, primary_color, secondary_color, custom_domain)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1 LIMIT 1;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1 LIMIT 1;

-- name: GetTenantByDomain :one
SELECT * FROM tenants WHERE custom_domain = $1 LIMIT 1;

-- name: UpdateTenant :one
UPDATE tenants SET
  name = COALESCE(sqlc.narg(name), name),
  logo_url = COALESCE(sqlc.narg(logo_url), logo_url),
  favicon_url = COALESCE(sqlc.narg(favicon_url), favicon_url),
  primary_color = COALESCE(sqlc.narg(primary_color), primary_color),
  secondary_color = COALESCE(sqlc.narg(secondary_color), secondary_color),
  contact_email = COALESCE(sqlc.narg(contact_email), contact_email),
  contact_phone = COALESCE(sqlc.narg(contact_phone), contact_phone),
  settings = COALESCE(sqlc.narg(settings), settings),
  custom_domain = COALESCE(sqlc.narg(custom_domain), custom_domain),
  address = COALESCE(sqlc.narg(address), address),
  timezone = COALESCE(sqlc.narg(timezone), timezone),
  currency = COALESCE(sqlc.narg(currency), currency),
  locale = COALESCE(sqlc.narg(locale), locale)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateTenantStatus :one
UPDATE tenants SET status = $2 WHERE id = $1 RETURNING *;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2;
