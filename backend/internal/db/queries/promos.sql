-- ============================================================
-- Promos SQLC Queries
-- ============================================================

-- name: CreatePromo :one
INSERT INTO promos (
    tenant_id, code, title, description, promo_type,
    discount_amount, max_discount_cap, cashback_amount, funded_by,
    applies_to, min_order_amount, max_total_uses, max_uses_per_user,
    include_stores, is_active, starts_at, ends_at, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
RETURNING *;

-- name: GetPromoByID :one
SELECT * FROM promos
WHERE id = $1 AND tenant_id = $2;

-- name: GetPromoByCode :one
SELECT * FROM promos
WHERE code = $1 AND tenant_id = $2;

-- name: ListPromos :many
SELECT * FROM promos
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPromos :one
SELECT COUNT(*) FROM promos
WHERE tenant_id = $1;

-- name: UpdatePromo :one
UPDATE promos SET
    title = COALESCE(sqlc.narg(title), title),
    description = COALESCE(sqlc.narg(description), description),
    discount_amount = COALESCE(sqlc.narg(discount_amount), discount_amount),
    max_discount_cap = COALESCE(sqlc.narg(max_discount_cap), max_discount_cap),
    cashback_amount = COALESCE(sqlc.narg(cashback_amount), cashback_amount),
    min_order_amount = COALESCE(sqlc.narg(min_order_amount), min_order_amount),
    max_total_uses = COALESCE(sqlc.narg(max_total_uses), max_total_uses),
    max_uses_per_user = COALESCE(sqlc.narg(max_uses_per_user), max_uses_per_user),
    starts_at = COALESCE(sqlc.narg(starts_at), starts_at),
    ends_at = COALESCE(sqlc.narg(ends_at), ends_at)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: DeactivatePromo :one
UPDATE promos
SET is_active = false
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: IncrementPromoUsage :exec
UPDATE promos
SET total_uses = total_uses + 1,
    total_discount_given = total_discount_given + sqlc.arg(discount_amount)
WHERE id = sqlc.arg(id) AND tenant_id = sqlc.arg(tenant_id);

-- name: CreatePromoUsage :one
INSERT INTO promo_usages (
    promo_id, user_id, order_id, tenant_id,
    discount_amount, cashback_amount
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUsageCountByUserAndPromo :one
SELECT COUNT(*) FROM promo_usages
WHERE user_id = $1 AND promo_id = $2;

-- name: ListPromoRestaurantRestrictions :many
SELECT restaurant_id FROM promo_restaurant_restrictions
WHERE promo_id = $1;

-- name: ListPromoCategoryRestrictions :many
SELECT category_id FROM promo_category_restrictions
WHERE promo_id = $1;

-- name: ListPromoUserEligibility :many
SELECT user_id FROM promo_user_eligibility
WHERE promo_id = $1;

-- name: CheckPromoUserEligibility :one
SELECT COUNT(*) FROM promo_user_eligibility
WHERE promo_id = $1 AND user_id = $2;

-- name: AddPromoRestaurantRestriction :exec
INSERT INTO promo_restaurant_restrictions (promo_id, restaurant_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: AddPromoCategoryRestriction :exec
INSERT INTO promo_category_restrictions (promo_id, category_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: AddPromoUserEligibility :exec
INSERT INTO promo_user_eligibility (promo_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemovePromoRestaurantRestrictions :exec
DELETE FROM promo_restaurant_restrictions
WHERE promo_id = $1;

-- name: RemovePromoCategoryRestrictions :exec
DELETE FROM promo_category_restrictions
WHERE promo_id = $1;

-- name: RemovePromoUserEligibility :exec
DELETE FROM promo_user_eligibility
WHERE promo_id = $1;

-- name: GetActivePromoByCode :one
SELECT * FROM promos
WHERE code = $1 AND tenant_id = $2
  AND is_active = true
  AND starts_at <= NOW()
  AND (ends_at IS NULL OR ends_at > NOW());
