package tenant

import (
	"context"

	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/contextkey"
)

// FromContext retrieves the tenant from the request context.
func FromContext(ctx context.Context) *sqlc.Tenant {
	t, _ := ctx.Value(contextkey.TenantKey).(*sqlc.Tenant)
	return t
}

// WithContext returns a new context with the tenant attached.
func WithContext(ctx context.Context, t *sqlc.Tenant) context.Context {
	return context.WithValue(ctx, contextkey.TenantKey, t)
}
