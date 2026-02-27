package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
)

// Service implements analytics business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new analytics service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// DashboardResponse holds partner dashboard data.
type DashboardResponse struct {
	Today   *sqlc.GetDashboardTodayRow    `json:"today"`
	Trend   []sqlc.GetDashboardTrendRow   `json:"trend"`
	Top     []sqlc.GetTopProductsRow      `json:"top_products"`
	Pending int                           `json:"pending_orders"`
}

// GetDashboard returns partner dashboard KPIs.
func (s *Service) GetDashboard(ctx context.Context, tenantID uuid.UUID) (*DashboardResponse, error) {
	today, err := s.q.GetDashboardToday(ctx, tenantID)
	if err != nil {
		return nil, apperror.Internal("get dashboard today", err)
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)
	trend, err := s.q.GetDashboardTrend(ctx, sqlc.GetDashboardTrendParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, apperror.Internal("get dashboard trend", err)
	}

	topProducts, err := s.q.GetTopProducts(ctx, sqlc.GetTopProductsParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     3,
	})
	if err != nil {
		return nil, apperror.Internal("get top products", err)
	}

	pending, err := s.q.GetPendingOrderCount(ctx, tenantID)
	if err != nil {
		return nil, apperror.Internal("get pending count", err)
	}

	return &DashboardResponse{
		Today:   &today,
		Trend:   trend,
		Top:     topProducts,
		Pending: int(pending),
	}, nil
}

// GetSalesReport returns sales report data.
func (s *Service) GetSalesReport(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]sqlc.GetSalesReportRow, error) {
	return s.q.GetSalesReport(ctx, sqlc.GetSalesReportParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetPeakHours returns peak hour data.
func (s *Service) GetPeakHours(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]sqlc.GetPeakHoursRow, error) {
	return s.q.GetPeakHours(ctx, sqlc.GetPeakHoursParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetOrderBreakdown returns order status breakdown.
func (s *Service) GetOrderBreakdown(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]sqlc.GetOrderStatusBreakdownRow, error) {
	return s.q.GetOrderStatusBreakdown(ctx, sqlc.GetOrderStatusBreakdownParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetRiderAnalytics returns rider performance data.
func (s *Service) GetRiderAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]sqlc.GetRiderAnalyticsRow, error) {
	return s.q.GetRiderAnalytics(ctx, sqlc.GetRiderAnalyticsParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetTopProducts returns top selling products.
func (s *Service) GetTopProducts(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, limit int) ([]sqlc.GetTopProductsRow, error) {
	return s.q.GetTopProducts(ctx, sqlc.GetTopProductsParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     int32(limit),
	})
}

// GetAdminOverview returns platform-wide analytics.
func (s *Service) GetAdminOverview(ctx context.Context, startDate, endDate time.Time) (*sqlc.GetAdminOverviewRow, error) {
	overview, err := s.q.GetAdminOverview(ctx, sqlc.GetAdminOverviewParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, apperror.Internal("get admin overview", err)
	}
	return &overview, nil
}

// GetAdminRevenue returns platform revenue by period.
func (s *Service) GetAdminRevenue(ctx context.Context, startDate, endDate time.Time) ([]sqlc.GetAdminRevenueByPeriodRow, error) {
	return s.q.GetAdminRevenueByPeriod(ctx, sqlc.GetAdminRevenueByPeriodParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetAdminOrderVolume returns platform order volume.
func (s *Service) GetAdminOrderVolume(ctx context.Context, startDate, endDate time.Time) ([]sqlc.GetAdminOrderVolumeRow, error) {
	return s.q.GetAdminOrderVolume(ctx, sqlc.GetAdminOrderVolumeParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
}

// GetTenantAnalytics returns analytics for a specific tenant.
func (s *Service) GetTenantAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*sqlc.GetTenantAnalyticsRow, error) {
	row, err := s.q.GetTenantAnalytics(ctx, sqlc.GetTenantAnalyticsParams{
		TenantID:  tenantID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, apperror.Internal("get tenant analytics", err)
	}
	return &row, nil
}
