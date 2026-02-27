package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// AuthMiddleware validates JWT access tokens and injects the user into context.
type AuthMiddleware struct {
	q      *sqlc.Queries
	tokens TokenConfig
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(q *sqlc.Queries, tokens TokenConfig) *AuthMiddleware {
	return &AuthMiddleware{q: q, tokens: tokens}
}

// Authenticate is HTTP middleware that validates the Bearer token and attaches the user to context.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.extractClaims(r)
		if err != nil {
			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				respond.Error(w, appErr)
			} else {
				respond.Error(w, apperror.Unauthorized("invalid token"))
			}
			return
		}

		user, err := m.q.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			respond.Error(w, apperror.Unauthorized("user not found"))
			return
		}

		if user.Status != sqlc.UserStatusActive {
			respond.Error(w, apperror.Forbidden("account is not active"))
			return
		}

		r = r.WithContext(WithUser(r.Context(), &user))
		next.ServeHTTP(w, r)
	})
}

// RequireRoles returns middleware that enforces the user has one of the given roles.
func RequireRoles(roles ...sqlc.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				respond.Error(w, apperror.Unauthorized("authentication required"))
				return
			}
			for _, role := range roles {
				if user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			respond.Error(w, apperror.Forbidden("insufficient permissions"))
		})
	}
}

func (m *AuthMiddleware) extractClaims(r *http.Request) (*Claims, error) {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		return ParseAccessToken(m.tokens, tokenStr)
	}

	// Fallback: access_token cookie
	cookie, err := r.Cookie("access_token")
	if err == nil {
		return ParseAccessToken(m.tokens, cookie.Value)
	}

	return nil, apperror.Unauthorized("authentication required")
}
