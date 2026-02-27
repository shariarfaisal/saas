package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset all env vars to test defaults
	envVars := []string{"PORT", "ENVIRONMENT", "DATABASE_URL", "REDIS_URL",
		"JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET", "ALLOWED_ORIGINS"}
	originals := make(map[string]string)
	for _, key := range envVars {
		originals[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	t.Cleanup(func() {
		for k, v := range originals {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Environment != EnvLocal {
		t.Errorf("expected default env local, got %s", cfg.Server.Environment)
	}

	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("expected default max open conns 25, got %d", cfg.Database.MaxOpenConns)
	}
}

func TestParseEnvironment(t *testing.T) {
	tests := []struct {
		input    string
		expected Environment
	}{
		{"local", EnvLocal},
		{"development", EnvDev},
		{"dev", EnvDev},
		{"staging", EnvStaging},
		{"production", EnvProduction},
		{"prod", EnvProduction},
		{"unknown", EnvLocal},
		{"", EnvLocal},
		{"  Production  ", EnvProduction},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseEnvironment(tc.input)
			if result != tc.expected {
				t.Errorf("parseEnvironment(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 1}, // default
		{"http://localhost:3000", 1},
		{"http://localhost:3000,http://localhost:3001", 2},
		{"http://a.com, http://b.com, http://c.com", 3},
	}

	for _, tc := range tests {
		result := parseOrigins(tc.input)
		if len(result) != tc.expected {
			t.Errorf("parseOrigins(%q) returned %d origins, want %d", tc.input, len(result), tc.expected)
		}
	}
}

func TestEnvironment_IsProduction(t *testing.T) {
	if !EnvProduction.IsProduction() {
		t.Error("expected EnvProduction.IsProduction() to be true")
	}
	if EnvLocal.IsProduction() {
		t.Error("expected EnvLocal.IsProduction() to be false")
	}
}

func TestEnvironment_IsDevelopment(t *testing.T) {
	if !EnvLocal.IsDevelopment() {
		t.Error("expected EnvLocal.IsDevelopment() to be true")
	}
	if !EnvDev.IsDevelopment() {
		t.Error("expected EnvDev.IsDevelopment() to be true")
	}
	if EnvProduction.IsDevelopment() {
		t.Error("expected EnvProduction.IsDevelopment() to be false")
	}
}
