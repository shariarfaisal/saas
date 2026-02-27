-- name: CreateBanner :one
INSERT INTO banners (
    tenant_id, title, subtitle, image_url, mobile_image_url,
    link_type, link_value, platform, sort_order, is_active, hub_ids, starts_at, ends_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetBannerByID :one
SELECT * FROM banners WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: UpdateBanner :one
UPDATE banners SET
    title = $3, subtitle = $4, image_url = $5, mobile_image_url = $6,
    link_type = $7, link_value = $8, platform = $9, sort_order = $10,
    is_active = $11, hub_ids = $12, starts_at = $13, ends_at = $14
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteBanner :exec
DELETE FROM banners WHERE id = $1 AND tenant_id = $2;

-- name: ListBannersByTenant :many
SELECT * FROM banners WHERE tenant_id = $1
ORDER BY sort_order ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountBannersByTenant :one
SELECT COUNT(*) FROM banners WHERE tenant_id = $1;

-- name: ListActiveBanners :many
SELECT * FROM banners
WHERE tenant_id = $1 AND is_active = true
  AND (starts_at IS NULL OR starts_at <= NOW())
  AND (ends_at IS NULL OR ends_at >= NOW())
ORDER BY sort_order ASC;

-- name: CreateStory :one
INSERT INTO stories (
    tenant_id, restaurant_id, title, media_url, media_type,
    thumbnail_url, link_type, link_value, expires_at, sort_order, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetStoryByID :one
SELECT * FROM stories WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: DeleteStory :exec
DELETE FROM stories WHERE id = $1 AND tenant_id = $2;

-- name: ListStoriesByTenant :many
SELECT * FROM stories WHERE tenant_id = $1
ORDER BY sort_order ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountStoriesByTenant :one
SELECT COUNT(*) FROM stories WHERE tenant_id = $1;

-- name: ListActiveStories :many
SELECT * FROM stories
WHERE tenant_id = $1 AND is_active = true AND expires_at > NOW()
ORDER BY sort_order ASC;

-- name: GetSectionByID :one
SELECT * FROM homepage_sections WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: UpdateSection :one
UPDATE homepage_sections SET
    title = $3, subtitle = $4, content_type = $5, item_ids = $6,
    filter_rule = $7, sort_order = $8, is_active = $9, hub_ids = $10
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: ListSectionsByTenant :many
SELECT * FROM homepage_sections WHERE tenant_id = $1
ORDER BY sort_order ASC;

-- name: ListActiveSections :many
SELECT * FROM homepage_sections
WHERE tenant_id = $1 AND is_active = true
ORDER BY sort_order ASC;
