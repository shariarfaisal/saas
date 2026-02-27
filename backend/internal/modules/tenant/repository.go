package tenant

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
)

// Repository provides tenant data access.
type Repository struct {
	q *sqlc.Queries
}

// NewRepository creates a new tenant repository.
func NewRepository(q *sqlc.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*sqlc.Tenant, error) {
	t, err := r.q.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.Tenant, error) {
	t, err := r.q.GetTenantByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) GetByDomain(ctx context.Context, domain string) (*sqlc.Tenant, error) {
	t, err := r.q.GetTenantByDomain(ctx, sql.NullString{String: domain, Valid: true})
	if err != nil {
		return nil, err
	}
	return &t, nil
}
