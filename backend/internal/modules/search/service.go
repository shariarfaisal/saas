package search

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
)

// Service implements search business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new search service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// SearchResult holds combined search results.
type SearchResult struct {
	Restaurants []sqlc.Restaurant `json:"restaurants"`
	Products    []sqlc.Product    `json:"products"`
}

// Search performs full-text search across restaurants and products.
func (s *Service) Search(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, query, searchType string) (*SearchResult, error) {
	if query == "" {
		return nil, apperror.BadRequest("search query is required")
	}

	result := &SearchResult{}
	var resultCount int

	if searchType == "" || searchType == "all" || searchType == "restaurants" {
		restaurants, err := s.q.SearchRestaurants(ctx, sqlc.SearchRestaurantsParams{
			TenantID: tenantID,
			Query:    query,
			Limit:    20,
		})
		if err != nil {
			return nil, apperror.Internal("search restaurants", err)
		}
		result.Restaurants = restaurants
		resultCount += len(restaurants)
	}

	if searchType == "" || searchType == "all" || searchType == "products" {
		products, err := s.q.SearchProducts(ctx, sqlc.SearchProductsParams{
			TenantID: tenantID,
			Query:    query,
			Limit:    20,
		})
		if err != nil {
			return nil, apperror.Internal("search products", err)
		}
		result.Products = products
		resultCount += len(products)
	}

	// Log search
	filters, _ := json.Marshal(map[string]string{"type": searchType})
	go s.q.CreateSearchLog(context.Background(), sqlc.CreateSearchLogParams{
		TenantID:    tenantID,
		UserID:      userID,
		Query:       query,
		SearchType:  searchType,
		ResultCount: int32(resultCount),
		Filters:     filters,
	})

	return result, nil
}

// Autocomplete returns top 5 partial-match suggestions.
func (s *Service) Autocomplete(ctx context.Context, tenantID uuid.UUID, query string) (*SearchResult, error) {
	if query == "" {
		return &SearchResult{}, nil
	}

	restaurants, err := s.q.SearchRestaurants(ctx, sqlc.SearchRestaurantsParams{
		TenantID: tenantID,
		Query:    query,
		Limit:    5,
	})
	if err != nil {
		return nil, apperror.Internal("autocomplete restaurants", err)
	}

	products, err := s.q.SearchProducts(ctx, sqlc.SearchProductsParams{
		TenantID: tenantID,
		Query:    query,
		Limit:    5,
	})
	if err != nil {
		return nil, apperror.Internal("autocomplete products", err)
	}

	return &SearchResult{
		Restaurants: restaurants,
		Products:    products,
	}, nil
}

// GetTopSearchTerms returns top search terms for a tenant in the last 30 days.
func (s *Service) GetTopSearchTerms(ctx context.Context, tenantID uuid.UUID, limit int) ([]sqlc.GetTopSearchTermsRow, error) {
	since := time.Now().AddDate(0, 0, -30)
	return s.q.GetTopSearchTerms(ctx, sqlc.GetTopSearchTermsParams{
		TenantID: tenantID,
		Since:    since,
		Limit:    int32(limit),
	})
}
