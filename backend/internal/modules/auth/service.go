package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	redisclient "github.com/munchies/platform/backend/internal/platform/redis"
	"github.com/munchies/platform/backend/internal/platform/sms"
	"golang.org/x/crypto/bcrypt"
)

const (
	otpRateLimit     = 3
	otpWindow        = 10 * time.Minute
	otpTTL           = 10 * time.Minute
	denyListTTL      = 7 * 24 * time.Hour
	passwordResetTTL = 1 * time.Hour
)

// Service implements authentication business logic.
type Service struct {
	q      *sqlc.Queries
	redis  *redisclient.Client
	sms    sms.Sender
	tokens TokenConfig
}

// NewService creates a new auth service.
func NewService(q *sqlc.Queries, redis *redisclient.Client, smsSender sms.Sender, tokens TokenConfig) *Service {
	return &Service{q: q, redis: redis, sms: smsSender, tokens: tokens}
}

// SendOTP generates and sends an OTP for phone-based auth.
func (s *Service) SendOTP(ctx context.Context, tenantID *uuid.UUID, phone, purpose string) error {
	since := time.Now().Add(-otpWindow)
	count, err := s.q.CountRecentOTPs(ctx, sqlc.CountRecentOTPsParams{
		Phone:     phone,
		Purpose:   purpose,
		CreatedAt: since,
	})
	if err != nil {
		return apperror.Internal("count otps", err)
	}
	if count >= otpRateLimit {
		return apperror.RateLimited()
	}

	otpCode, err := generateOTP()
	if err != nil {
		return apperror.Internal("generate otp", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal("hash otp", err)
	}

	pgTenantID := uuidToPgtype(tenantID)
	_, err = s.q.CreateOTPVerification(ctx, sqlc.CreateOTPVerificationParams{
		TenantID:  pgTenantID,
		Phone:     phone,
		Purpose:   purpose,
		OtpHash:   string(hash),
		ExpiresAt: time.Now().Add(otpTTL),
	})
	if err != nil {
		return apperror.Internal("store otp", err)
	}

	msg := fmt.Sprintf("Your Munchies verification code is: %s. Valid for 10 minutes.", otpCode)
	if err := s.sms.Send(ctx, phone, msg); err != nil {
		return apperror.Internal("send sms", err)
	}

	return nil
}

// VerifyOTP verifies an OTP and returns JWT tokens plus the user.
func (s *Service) VerifyOTP(ctx context.Context, tenantID *uuid.UUID, phone, purpose, code, ipStr string) (*TokenPair, *sqlc.User, error) {
	otp, err := s.q.GetLatestOTP(ctx, sqlc.GetLatestOTPParams{
		Phone:   phone,
		Purpose: purpose,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apperror.BadRequest("OTP not found or expired")
		}
		return nil, nil, apperror.Internal("get otp", err)
	}

	otp, err = s.q.IncrementOTPAttempts(ctx, otp.ID)
	if err != nil {
		return nil, nil, apperror.Internal("increment attempts", err)
	}

	if otp.Attempts > otp.MaxAttempts {
		return nil, nil, apperror.BadRequest("too many OTP attempts")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otp.OtpHash), []byte(code)); err != nil {
		return nil, nil, apperror.BadRequest("invalid OTP")
	}

	if err := s.q.MarkOTPVerified(ctx, otp.ID); err != nil {
		return nil, nil, apperror.Internal("mark verified", err)
	}

	// Find or create user for the given tenant
	var user sqlc.User
	if tenantID != nil {
		pgTenantID := uuidToPgtype(tenantID)
		existing, err := s.q.GetUserByPhone(ctx, sqlc.GetUserByPhoneParams{
			TenantID: pgTenantID,
			Phone:    sql.NullString{String: phone, Valid: true},
		})
		if errors.Is(err, pgx.ErrNoRows) {
			newUser, err := s.q.CreateUser(ctx, sqlc.CreateUserParams{
				TenantID: pgTenantID,
				Phone:    sql.NullString{String: phone, Valid: true},
				Name:     "",
				Role:     sqlc.UserRoleCustomer,
				Status:   sqlc.UserStatusActive,
				Metadata: json.RawMessage("{}"),
			})
			if err != nil {
				return nil, nil, apperror.Internal("create user", err)
			}
			user = newUser
		} else if err != nil {
			return nil, nil, apperror.Internal("get user", err)
		} else {
			user = existing
		}
	}

	// Update last login
	now := time.Now()
	ip := parseIP(ipStr)
	updated, err := s.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:          user.ID,
		LastLoginAt: pgtype.Timestamptz{Time: now, Valid: true},
		LastLoginIp: ip,
	})
	if err == nil {
		user = updated
	}

	return s.issueTokens(ctx, &user)
}

// Login authenticates with email + password (partner/admin users).
func (s *Service) Login(ctx context.Context, email, password, ipStr string) (*TokenPair, *sqlc.User, error) {
	user, err := s.q.GetUserByEmail(ctx, sqlc.GetUserByEmailParams{
		TenantID: pgtype.UUID{},
		Email:    sql.NullString{String: email, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apperror.Unauthorized("invalid credentials")
		}
		return nil, nil, apperror.Internal("get user", err)
	}

	if !user.PasswordHash.Valid {
		return nil, nil, apperror.Unauthorized("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(password)); err != nil {
		return nil, nil, apperror.Unauthorized("invalid credentials")
	}

	if user.Status != sqlc.UserStatusActive {
		return nil, nil, apperror.Forbidden("account is not active")
	}

	now := time.Now()
	ip := parseIP(ipStr)
	updated, err := s.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:          user.ID,
		LastLoginAt: pgtype.Timestamptz{Time: now, Valid: true},
		LastLoginIp: ip,
	})
	if err == nil {
		user = updated
	}

	return s.issueTokens(ctx, &user)
}

// Refresh rotates a refresh token.
func (s *Service) Refresh(ctx context.Context, rawToken string) (*TokenPair, *sqlc.User, error) {
	denyKey := fmt.Sprintf("token:deny:%s", hashToken(rawToken))
	n, _ := s.redis.Exists(ctx, denyKey)
	if n > 0 {
		return nil, nil, apperror.Unauthorized("token revoked")
	}

	hash := hashToken(rawToken)
	rt, err := s.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, apperror.Unauthorized("invalid refresh token")
		}
		return nil, nil, apperror.Internal("get refresh token", err)
	}

	if err := s.q.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return nil, nil, apperror.Internal("revoke token", err)
	}

	user, err := s.q.GetUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, nil, apperror.Internal("get user", err)
	}

	return s.issueTokens(ctx, &user)
}

// Logout invalidates a refresh token.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	rt, err := s.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return apperror.Internal("get refresh token", err)
	}
	if err == nil {
		if err := s.q.RevokeRefreshToken(ctx, rt.ID); err != nil {
			return apperror.Internal("revoke token", err)
		}
	}
	denyKey := fmt.Sprintf("token:deny:%s", hash)
	_ = s.redis.Set(ctx, denyKey, "1", denyListTTL)
	return nil
}

// ResetPasswordRequest stores a reset token in Redis (email sending is TODO).
func (s *Service) ResetPasswordRequest(ctx context.Context, email string) error {
	user, err := s.q.GetUserByEmail(ctx, sqlc.GetUserByEmailParams{
		TenantID: pgtype.UUID{},
		Email:    sql.NullString{String: email, Valid: true},
	})
	if err != nil {
		// Silently succeed to prevent user enumeration
		return nil
	}

	token, err := NewRefreshToken()
	if err != nil {
		return apperror.Internal("generate reset token", err)
	}

	key := fmt.Sprintf("password:reset:%s", hashToken(token))
	if err := s.redis.Set(ctx, key, user.ID.String(), passwordResetTTL); err != nil {
		return apperror.Internal("store reset token", err)
	}

	// NOTE: email delivery is not yet implemented; a follow-up task will add the email service.
	// The reset token is available in Redis for passwordResetTTL (1 hour).
	return nil
}

// ResetPassword applies a password reset using the given token.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	key := fmt.Sprintf("password:reset:%s", hashToken(token))
	userIDStr, err := s.redis.Get(ctx, key)
	if err != nil {
		return apperror.BadRequest("invalid or expired reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.Internal("parse user id", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal("hash password", err)
	}

	if err := s.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
	}); err != nil {
		return apperror.Internal("update password", err)
	}

	_ = s.redis.Del(ctx, key)

	if err := s.q.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		return apperror.Internal("revoke tokens", err)
	}

	return nil
}

func (s *Service) issueTokens(ctx context.Context, user *sqlc.User) (*TokenPair, *sqlc.User, error) {
	tenantID := pgtypeToUUID(user.TenantID)
	accessToken, err := NewAccessToken(s.tokens, user.ID, tenantID, string(user.Role))
	if err != nil {
		return nil, nil, apperror.Internal("sign access token", err)
	}

	rawRefresh, err := NewRefreshToken()
	if err != nil {
		return nil, nil, apperror.Internal("generate refresh token", err)
	}

	hash := hashToken(rawRefresh)
	expiresAt := time.Now().Add(s.tokens.RefreshExpiry)

	_, err = s.q.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, nil, apperror.Internal("store refresh token", err)
	}

	pair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(s.tokens.AccessExpiry.Seconds()),
	}
	return pair, user, nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

func uuidToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgtypeToUUID(p pgtype.UUID) *uuid.UUID {
	if !p.Valid {
		return nil
	}
	id := uuid.UUID(p.Bytes)
	return &id
}

func parseIP(ipStr string) *netip.Addr {
	if ipStr == "" {
		return nil
	}
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil
	}
	return &addr
}
