-- name: CreateReview :one
INSERT INTO reviews (
    tenant_id, order_id, user_id, restaurant_id,
    restaurant_rating, rider_rating, comment, images
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetReviewByID :one
SELECT * FROM reviews WHERE id = $1 AND tenant_id = $2 LIMIT 1;

-- name: GetReviewByOrderAndUser :one
SELECT * FROM reviews WHERE order_id = $1 AND user_id = $2 LIMIT 1;

-- name: ListReviewsByRestaurant :many
SELECT * FROM reviews
WHERE restaurant_id = $1 AND tenant_id = $2 AND is_published = true
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountReviewsByRestaurant :one
SELECT COUNT(*) FROM reviews
WHERE restaurant_id = $1 AND tenant_id = $2 AND is_published = true;

-- name: UpdateRestaurantReply :one
UPDATE reviews SET restaurant_reply = $3, restaurant_reply_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: PublishReview :one
UPDATE reviews SET is_published = true
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: GetRestaurantAvgRating :one
SELECT
    COALESCE(AVG(restaurant_rating), 0)::NUMERIC(3,2) AS avg_rating,
    COUNT(*)::INT AS review_count
FROM reviews
WHERE restaurant_id = $1 AND is_published = true;

-- name: UpdateRestaurantRating :exec
UPDATE restaurants SET rating_avg = $2, rating_count = $3
WHERE id = $1;
