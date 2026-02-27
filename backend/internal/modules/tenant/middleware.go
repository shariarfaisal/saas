package tenant

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
)

const tenantCacheTTL = 60 * time.Second

// Resolver resolves the tenant from the request.
type Resolver struct {
	repo  *Repository
	redis *redisclient.Client
}

// NewResolver creates a new tenant resolver.
func NewResolver(repo *Repository, redis *redisclient.Client) *Resolver {
	return &Resolver{repo: repo, redis: redis}
}

// Middleware resolves the tenant and injects it into context.
// If no tenant can be resolved, the request proceeds without a tenant.
func (res *Resolver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, err := res.resolve(r)
		if err != nil {
			var appErr *apperror.AppError
			if errors.As(err, &appErr) {
				respond.Error(w, appErr)
			} else {
				respond.Error(w, apperror.Internal("failed to resolve tenant", err))
			}
			return
		}
		if t != nil {
			if t.Status == sqlc.TenantStatusSuspended || t.Status == sqlc.TenantStatusCancelled {
				respond.Error(w, apperror.Forbidden("tenant is suspended or cancelled"))
				return
			}
			r = r.WithContext(WithContext(r.Context(), t))
		}
		next.ServeHTTP(w, r)
	})
}

// RequireTenant is middleware that enforces a tenant must be present in context.
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if FromContext(r.Context()) == nil {
			respond.Error(w, apperror.NotFound("tenant"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (res *Resolver) resolve(r *http.Request) (*sqlc.Tenant, error) {
	ctx := r.Context()

	// Strategy 1: subdomain from Host header
	if slug := extractSubdomain(r.Host); slug != "" {
		return res.getTenantBySlug(ctx, slug)
	}

	// Strategy 2: tenant_id claim in JWT (parsed without full verification)
	if tenantID := extractTenantIDFromJWT(r); tenantID != uuid.Nil {
		return res.getTenantByID(ctx, tenantID)
	}

	// Strategy 3: X-Tenant-ID header
	if idStr := r.Header.Get("X-Tenant-ID"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, apperror.BadRequest("invalid X-Tenant-ID header")
		}
		return res.getTenantByID(ctx, id)
	}

	return nil, nil
}

func (res *Resolver) getTenantBySlug(ctx context.Context, slug string) (*sqlc.Tenant, error) {
	cacheKey := fmt.Sprintf("tenant:slug:%s", slug)
	return res.getWithCache(ctx, cacheKey, func() (*sqlc.Tenant, error) {
		t, err := res.repo.GetBySlug(ctx, slug)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("tenant")
		}
		return t, err
	})
}

func (res *Resolver) getTenantByID(ctx context.Context, id uuid.UUID) (*sqlc.Tenant, error) {
	cacheKey := fmt.Sprintf("tenant:id:%s", id.String())
	return res.getWithCache(ctx, cacheKey, func() (*sqlc.Tenant, error) {
		t, err := res.repo.GetByID(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("tenant")
		}
		return t, err
	})
}

func (res *Resolver) getWithCache(ctx context.Context, key string, fetch func() (*sqlc.Tenant, error)) (*sqlc.Tenant, error) {
	// TODO: implement Redis caching using key and tenantCacheTTL once Redis is required.
	// For now, always fetch from DB to keep the dependency optional.
	_ = key
	return fetch()
}

func extractSubdomain(host string) string {
	// Strip port
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	parts := strings.Split(host, ".")
	// sub.domain.tld → 3+ parts
	if len(parts) >= 3 {
		sub := parts[0]
		if sub != "www" && sub != "api" && sub != "app" && sub != "admin" {
			return sub
		}
	}
	// sub.localhost for local dev
	if len(parts) == 2 && parts[1] == "localhost" {
		return parts[0]
	}
	return ""
}

func extractTenantIDFromJWT(r *http.Request) uuid.UUID {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return uuid.Nil
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return uuid.Nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil
	}
	tidStr, _ := claims["tenant_id"].(string)
	id, err := uuid.Parse(tidStr)
	if err != nil {
		return uuid.Nil
	}
	return id
}
