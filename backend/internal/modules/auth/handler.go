package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/munchies/platform/backend/internal/db/sqlc"
	"github.com/munchies/platform/backend/internal/modules/tenant"
	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles auth HTTP requests.
type Handler struct {
	svc    *Service
	tokens TokenConfig
}

// NewHandler creates a new auth handler.
func NewHandler(svc *Service, tokens TokenConfig) *Handler {
	return &Handler{svc: svc, tokens: tokens}
}

// SendOTP handles POST /auth/otp/send
func (h *Handler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone   string `json:"phone"`
		Purpose string `json:"purpose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Phone == "" {
		respond.Error(w, apperror.BadRequest("phone is required"))
		return
	}
	if req.Purpose == "" {
		req.Purpose = "login"
	}

	t := tenant.FromContext(r.Context())
	if err := h.svc.SendOTP(r.Context(), tenantUUIDPtr(t), req.Phone, req.Purpose); err != nil {
		var appErr *apperror.AppError
		if e, ok := err.(*apperror.AppError); ok {
			appErr = e
		} else {
			appErr = apperror.Internal("send otp", err)
		}
		respond.Error(w, appErr)
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{"message": "OTP sent successfully"})
}

// VerifyOTP handles POST /auth/otp/verify
func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone   string `json:"phone"`
		Purpose string `json:"purpose"`
		Code    string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Phone == "" || req.Code == "" {
		respond.Error(w, apperror.BadRequest("phone and code are required"))
		return
	}
	if req.Purpose == "" {
		req.Purpose = "login"
	}

	t := tenant.FromContext(r.Context())
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.RemoteAddr
	}

	pair, user, err := h.svc.VerifyOTP(r.Context(), tenantUUIDPtr(t), req.Phone, req.Purpose, req.Code, ip)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.setRefreshTokenCookie(w, pair.RefreshToken)
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
		"user":         userResponse(user),
	})
}

// Login handles POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Email == "" || req.Password == "" {
		respond.Error(w, apperror.BadRequest("email and password are required"))
		return
	}

	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.RemoteAddr
	}

	pair, user, err := h.svc.Login(r.Context(), req.Email, req.Password, ip)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.setRefreshTokenCookie(w, pair.RefreshToken)
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
		"user":         userResponse(user),
	})
}

// Refresh handles POST /auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	rawToken := h.getRefreshToken(r)
	if rawToken == "" {
		respond.Error(w, apperror.Unauthorized("refresh token required"))
		return
	}

	pair, user, err := h.svc.Refresh(r.Context(), rawToken)
	if err != nil {
		respond.Error(w, toAppError(err))
		return
	}

	h.setRefreshTokenCookie(w, pair.RefreshToken)
	respond.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token": pair.AccessToken,
		"expires_in":   pair.ExpiresIn,
		"user":         userResponse(user),
	})
}

// Logout handles POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	rawToken := h.getRefreshToken(r)
	if rawToken != "" {
		_ = h.svc.Logout(r.Context(), rawToken)
	}
	h.clearRefreshTokenCookie(w)
	respond.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// RequestPasswordReset handles POST /auth/password/reset-request
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	// Always return success to prevent enumeration
	_ = h.svc.ResetPasswordRequest(r.Context(), req.Email)
	respond.JSON(w, http.StatusOK, map[string]string{"message": "if the email exists, a reset link has been sent"})
}

// ResetPassword handles POST /auth/password/reset
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperror.BadRequest("invalid request body"))
		return
	}
	if req.Token == "" || req.NewPassword == "" {
		respond.Error(w, apperror.BadRequest("token and new_password are required"))
		return
	}
	if err := h.svc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		respond.Error(w, toAppError(err))
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func (h *Handler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.tokens.RefreshExpiry.Seconds()),
		Expires:  time.Now().Add(h.tokens.RefreshExpiry),
	})
}

func (h *Handler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func (h *Handler) getRefreshToken(r *http.Request) string {
	// Try cookie first
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		return cookie.Value
	}
	// Fallback: JSON body
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	return body.RefreshToken
}

func tenantUUIDPtr(t *sqlc.Tenant) *uuid.UUID {
	if t == nil {
		return nil
	}
	id := t.ID
	return &id
}

func userResponse(u *sqlc.User) map[string]interface{} {
	if u == nil {
		return nil
	}
	return map[string]interface{}{
		"id":         u.ID,
		"name":       u.Name,
		"phone":      u.Phone,
		"email":      u.Email,
		"role":       u.Role,
		"status":     u.Status,
		"avatar_url": u.AvatarUrl,
		"created_at": u.CreatedAt,
	}
}

func toAppError(err error) *apperror.AppError {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperror.Internal("unexpected error", err)
}
