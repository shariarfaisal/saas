package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var testCfg = TokenConfig{
	AccessSecret:  "test-access-secret-32-bytes-long!",
	RefreshSecret: "test-refresh-secret-32-bytes-long",
	AccessExpiry:  15 * time.Minute,
	RefreshExpiry: 7 * 24 * time.Hour,
}

func TestNewAccessToken_RoundTrip(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	token, err := NewAccessToken(testCfg, userID, &tenantID, "customer")
	if err != nil {
		t.Fatalf("NewAccessToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ParseAccessToken(testCfg, token)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.TenantID == nil || *claims.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", claims.TenantID, tenantID)
	}
	if claims.Role != "customer" {
		t.Errorf("Role = %q, want customer", claims.Role)
	}
}

func TestNewAccessToken_NoTenant(t *testing.T) {
	userID := uuid.New()

	token, err := NewAccessToken(testCfg, userID, nil, "platform_admin")
	if err != nil {
		t.Fatalf("NewAccessToken error: %v", err)
	}

	claims, err := ParseAccessToken(testCfg, token)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}

	if claims.TenantID != nil {
		t.Errorf("expected nil TenantID, got %v", claims.TenantID)
	}
	if claims.Role != "platform_admin" {
		t.Errorf("Role = %q, want platform_admin", claims.Role)
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	userID := uuid.New()
	token, _ := NewAccessToken(testCfg, userID, nil, "customer")

	wrongCfg := TokenConfig{
		AccessSecret: "wrong-secret",
	}
	_, err := ParseAccessToken(wrongCfg, token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	expiredCfg := TokenConfig{
		AccessSecret: testCfg.AccessSecret,
		AccessExpiry: -1 * time.Second, // already expired
	}
	userID := uuid.New()
	token, err := NewAccessToken(expiredCfg, userID, nil, "customer")
	if err != nil {
		t.Fatalf("NewAccessToken error: %v", err)
	}

	_, err = ParseAccessToken(testCfg, token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestNewRefreshToken_Uniqueness(t *testing.T) {
	t1, err1 := NewRefreshToken()
	t2, err2 := NewRefreshToken()

	if err1 != nil || err2 != nil {
		t.Fatalf("NewRefreshToken errors: %v, %v", err1, err2)
	}
	if t1 == t2 {
		t.Error("expected unique refresh tokens")
	}
	if len(t1) != 64 { // 32 bytes * 2 hex chars
		t.Errorf("expected 64 char hex string, got len %d", len(t1))
	}
}
