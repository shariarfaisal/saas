-- name: CreateOTPVerification :one
INSERT INTO otp_verifications (tenant_id, phone, purpose, otp_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLatestOTP :one
SELECT * FROM otp_verifications
WHERE phone = $1 AND purpose = $2 AND verified_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementOTPAttempts :one
UPDATE otp_verifications SET attempts = attempts + 1 WHERE id = $1 RETURNING *;

-- name: MarkOTPVerified :exec
UPDATE otp_verifications SET verified_at = NOW() WHERE id = $1;

-- name: CountRecentOTPs :one
SELECT COUNT(*) FROM otp_verifications
WHERE phone = $1 AND purpose = $2 AND created_at > $3;
