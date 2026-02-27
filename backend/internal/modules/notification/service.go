package notification

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/rs/zerolog/log"
)

// Service implements notification business logic.
type Service struct {
	q    *sqlc.Queries
	fcm  PushSender
	sms  SMSSender
	mail EmailSender
}

// PushSender sends push notifications.
type PushSender interface {
	Send(ctx context.Context, token, title, body string, data map[string]string) error
}

// SMSSender sends SMS messages.
type SMSSender interface {
	Send(ctx context.Context, phone, message string) error
}

// EmailSender sends emails.
type EmailSender interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

// NewService creates a new notification service.
func NewService(q *sqlc.Queries, fcm PushSender, sms SMSSender, mail EmailSender) *Service {
	return &Service{q: q, fcm: fcm, sms: sms, mail: mail}
}

// SendPush sends a push notification to a user and persists it.
func (s *Service) SendPush(ctx context.Context, tenantID *uuid.UUID, userID uuid.UUID, title, body string, data map[string]string) error {
	// Persist notification
	actionPayload, _ := json.Marshal(data)
	_, err := s.q.CreateNotification(ctx, sqlc.CreateNotificationParams{
		TenantID:      toPgUUIDPtr(tenantID),
		UserID:        userID,
		Channel:       sqlc.NotificationChannelPush,
		Title:         title,
		Body:          body,
		ActionType:    toNullStringVal(data["action_type"]),
		ActionPayload: actionPayload,
		Status:        sqlc.NotificationStatusPending,
	})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to persist notification")
	}

	if s.fcm == nil {
		return nil
	}

	// Get user's push token
	tokenNs, err := s.q.GetUserDevicePushToken(ctx, userID)
	if err != nil {
		return nil
	}
	if !tokenNs.Valid || tokenNs.String == "" {
		return nil
	}

	// Send push notification
	if err := s.fcm.Send(ctx, tokenNs.String, title, body, data); err != nil {
		// If token is invalid, clear it
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("push notification failed")
		return err
	}

	return nil
}

// SendSMS sends an SMS message.
func (s *Service) SendSMS(ctx context.Context, phone, message string) error {
	if s.sms == nil {
		return nil
	}
	return s.sms.Send(ctx, phone, message)
}

// SendEmail sends an email.
func (s *Service) SendEmail(ctx context.Context, to, subject, htmlBody string) error {
	if s.mail == nil {
		return nil
	}
	return s.mail.Send(ctx, to, subject, htmlBody)
}

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toNullStringVal(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
