package middleware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/auth"
	tenantmod "github.com/munchies/platform/backend/internal/modules/tenant"
)

const idempotencyTTL = 24 * time.Hour

// Idempotency returns middleware that deduplicates POST/PATCH requests using the
// Idempotency-Key header. Requires sqlc.Queries to store/retrieve keys.
func Idempotency(q *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to mutating methods
			if r.Method != http.MethodPost && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Read and buffer the request body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			requestHash := hashBytes(bodyBytes)
			endpoint := r.URL.Path

			// Determine tenant and user UUIDs
			tenantPgUUID := pgtype.UUID{}
			if t := tenantmod.FromContext(r.Context()); t != nil {
				tenantPgUUID = pgtype.UUID{Bytes: t.ID, Valid: true}
			}
			userPgUUID := pgtype.UUID{}
			if u := auth.UserFromContext(r.Context()); u != nil {
				userPgUUID = pgtype.UUID{Bytes: u.ID, Valid: true}
			}

			existing, err := q.GetIdempotencyKey(r.Context(), sqlc.GetIdempotencyKeyParams{
				TenantID: tenantPgUUID,
				UserID:   userPgUUID,
				Key:      key,
				Endpoint: endpoint,
			})
			if err == nil {
				// Key exists — return cached response
				if existing.ResponseStatus != nil && existing.ResponseBody != nil {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Idempotency-Replayed", "true")
					w.WriteHeader(int(*existing.ResponseStatus))
					_, _ = w.Write(existing.ResponseBody)
					return
				}
				// Key exists but response not yet stored (in-flight); proceed
				next.ServeHTTP(w, r)
				return
			}

			// Create new key record
			rec, err := q.CreateIdempotencyKey(r.Context(), sqlc.CreateIdempotencyKeyParams{
				TenantID:    tenantPgUUID,
				UserID:      userPgUUID,
				Key:         key,
				Endpoint:    endpoint,
				RequestHash: requestHash,
				ExpiresAt:   time.Now().Add(idempotencyTTL),
			})
			if err != nil {
				// Conflict or DB error — proceed without idempotency
				next.ServeHTTP(w, r)
				return
			}

			// Intercept the response writer to capture status + body
			rw := &responseRecorder{ResponseWriter: w, body: &bytes.Buffer{}}
			next.ServeHTTP(rw, r)

			// Store the captured response
			status := int32(rw.status)
			_ = q.UpdateIdempotencyKeyResponse(r.Context(), sqlc.UpdateIdempotencyKeyResponseParams{
				ID:             rec.ID,
				ResponseStatus: &status,
				ResponseBody:   rw.body.Bytes(),
			})
		})
	}
}

// responseRecorder captures the HTTP response for idempotency storage.
type responseRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func hashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}
