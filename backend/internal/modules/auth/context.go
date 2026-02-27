package auth

import (
	"context"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/contextkey"
)

// UserFromContext returns the authenticated user from context.
func UserFromContext(ctx context.Context) *sqlc.User {
	u, _ := ctx.Value(contextkey.UserKey).(*sqlc.User)
	return u
}

// WithUser returns a new context with the user attached.
func WithUser(ctx context.Context, u *sqlc.User) context.Context {
	return context.WithValue(ctx, contextkey.UserKey, u)
}
