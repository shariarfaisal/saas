package issue

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/pagination"
)

// Service implements order issue business logic.
type Service struct {
	q *sqlc.Queries
}

// NewService creates a new issue service.
func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

// CreateIssueRequest holds fields for creating an issue.
type CreateIssueRequest struct {
	OrderID          uuid.UUID
	IssueType        sqlc.IssueType
	ReportedByID     uuid.UUID
	Details          string
	AccountableParty sqlc.Accountable
}

// CreateIssue creates a new order issue.
func (s *Service) CreateIssue(ctx context.Context, tenantID uuid.UUID, req CreateIssueRequest) (*sqlc.OrderIssue, error) {
	_, err := s.q.GetOrderByID(ctx, sqlc.GetOrderByIDParams{
		ID:       req.OrderID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("order")
	}
	if err != nil {
		return nil, err
	}

	issue, err := s.q.CreateOrderIssue(ctx, sqlc.CreateOrderIssueParams{
		OrderID:          req.OrderID,
		TenantID:         tenantID,
		IssueType:        req.IssueType,
		ReportedByID:     req.ReportedByID,
		Details:          req.Details,
		AccountableParty: req.AccountableParty,
	})
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// GetByID returns an order issue by ID.
func (s *Service) GetByID(ctx context.Context, tenantID, issueID uuid.UUID) (*sqlc.OrderIssue, error) {
	issue, err := s.q.GetOrderIssueByID(ctx, sqlc.GetOrderIssueByIDParams{
		ID:       issueID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("issue")
	}
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// ListByTenant returns paginated issues for a tenant.
func (s *Service) ListByTenant(ctx context.Context, tenantID uuid.UUID, page, perPage int) ([]sqlc.OrderIssue, pagination.Meta, error) {
	limit, offset := pagination.FormatLimitOffset(page, perPage)
	total, err := s.q.CountOrderIssuesByTenant(ctx, tenantID)
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("count issues", err)
	}
	items, err := s.q.ListOrderIssuesByTenant(ctx, sqlc.ListOrderIssuesByTenantParams{
		TenantID: tenantID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, pagination.Meta{}, apperror.Internal("list issues", err)
	}
	return items, pagination.NewMeta(total, limit, ""), nil
}

// Resolve resolves an order issue.
func (s *Service) Resolve(ctx context.Context, tenantID, issueID, resolvedByID uuid.UUID, note string) (*sqlc.OrderIssue, error) {
	issue, err := s.q.UpdateOrderIssueStatus(ctx, sqlc.UpdateOrderIssueStatusParams{
		ID:             issueID,
		TenantID:       tenantID,
		Status:         sqlc.IssueStatusResolved,
		ResolutionNote: toNullStringVal(note),
		ResolvedByID:   toPgUUID(resolvedByID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("issue")
	}
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// ApproveRefund approves a refund for an issue.
func (s *Service) ApproveRefund(ctx context.Context, tenantID, issueID uuid.UUID) (*sqlc.OrderIssue, error) {
	existing, err := s.q.GetOrderIssueByID(ctx, sqlc.GetOrderIssueByIDParams{
		ID:       issueID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("issue")
	}
	if err != nil {
		return nil, err
	}

	issue, err := s.q.UpdateOrderIssueRefund(ctx, sqlc.UpdateOrderIssueRefundParams{
		ID:           issueID,
		TenantID:     tenantID,
		RefundStatus: sqlc.RefundStatusApproved,
		RefundAmount: existing.RefundAmount,
	})
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// RejectRefund rejects a refund for an issue.
func (s *Service) RejectRefund(ctx context.Context, tenantID, issueID uuid.UUID) (*sqlc.OrderIssue, error) {
	existing, err := s.q.GetOrderIssueByID(ctx, sqlc.GetOrderIssueByIDParams{
		ID:       issueID,
		TenantID: tenantID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.NotFound("issue")
	}
	if err != nil {
		return nil, err
	}

	issue, err := s.q.UpdateOrderIssueRefund(ctx, sqlc.UpdateOrderIssueRefundParams{
		ID:           issueID,
		TenantID:     tenantID,
		RefundStatus: sqlc.RefundStatusRejected,
		RefundAmount: existing.RefundAmount,
	})
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// AddMessage adds a message to an issue thread.
func (s *Service) AddMessage(ctx context.Context, tenantID, issueID, senderID uuid.UUID, message string, attachments []string) (*sqlc.OrderIssueMessage, error) {
	msg, err := s.q.CreateOrderIssueMessage(ctx, sqlc.CreateOrderIssueMessageParams{
		IssueID:     issueID,
		TenantID:    tenantID,
		SenderID:    senderID,
		Message:     message,
		Attachments: attachments,
	})
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ListMessages lists all messages for an issue.
func (s *Service) ListMessages(ctx context.Context, tenantID, issueID uuid.UUID) ([]sqlc.OrderIssueMessage, error) {
	return s.q.ListOrderIssueMessages(ctx, sqlc.ListOrderIssueMessagesParams{
		IssueID:  issueID,
		TenantID: tenantID,
	})
}

// CreateAuditLog creates an audit log entry.
func (s *Service) CreateAuditLog(ctx context.Context, tenantID, actorID uuid.UUID, action, resourceType string, resourceID uuid.UUID, reason string) {
	changes, _ := json.Marshal(map[string]interface{}{})
	s.q.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		TenantID:     toPgUUID(tenantID),
		ActorID:      toPgUUID(actorID),
		ActorType:    sqlc.ActorTypePlatformAdmin,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   toPgUUID(resourceID),
		Changes:      changes,
		Reason:       toNullStringVal(reason),
	})
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func toNullStringVal(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
