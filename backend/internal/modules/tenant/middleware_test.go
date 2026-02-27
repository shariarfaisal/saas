package tenant

import (
	"net/http"
	"testing"
)

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"acme.api.platform.com", "acme"},
		{"acme.localhost", "acme"},
		{"platform.com", ""},
		{"www.platform.com", ""},
		{"api.platform.com", ""},
		{"app.platform.com", ""},
		{"admin.platform.com", ""},
		{"localhost", ""},
		{"localhost:8080", ""},
		{"acme.localhost:8080", "acme"},
		{"kacchi.api.example.com", "kacchi"},
	}

	for _, tc := range tests {
		got := extractSubdomain(tc.host)
		if got != tc.want {
			t.Errorf("extractSubdomain(%q) = %q, want %q", tc.host, got, tc.want)
		}
	}
}

func TestExtractTenantIDFromJWT_NoHeader(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	id := extractTenantIDFromJWT(r)
	if id.String() != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("expected nil UUID, got %v", id)
	}
}

func TestExtractTenantIDFromJWT_MalformedToken(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	id := extractTenantIDFromJWT(r)
	if id.String() != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("expected nil UUID for malformed token, got %v", id)
	}
}
