-- name: UpsertRiderLocation :one
INSERT INTO rider_locations (rider_id, tenant_id, geo_lat, geo_lng, heading, speed_kmh, accuracy_meters)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (rider_id) DO UPDATE SET
  geo_lat = EXCLUDED.geo_lat,
  geo_lng = EXCLUDED.geo_lng,
  heading = EXCLUDED.heading,
  speed_kmh = EXCLUDED.speed_kmh,
  accuracy_meters = EXCLUDED.accuracy_meters,
  updated_at = NOW()
RETURNING *;

-- name: GetRiderLocation :one
SELECT * FROM rider_locations WHERE rider_id = $1 LIMIT 1;

-- name: ListRiderLocationsByTenant :many
SELECT * FROM rider_locations WHERE tenant_id = $1 ORDER BY updated_at DESC;

-- name: AppendLocationHistory :one
INSERT INTO rider_location_history (rider_id, tenant_id, order_id, geo_lat, geo_lng, event_type, distance_from_prev_km)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListLocationHistoryByRider :many
SELECT * FROM rider_location_history
WHERE rider_id = $1 AND created_at >= sqlc.arg(since)::timestamptz
ORDER BY created_at DESC
LIMIT $2;
