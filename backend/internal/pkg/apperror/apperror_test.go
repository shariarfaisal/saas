package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeNotFound, "user not found")
	if err.Code != CodeNotFound {
		t.Errorf("expected code %s, got %s", CodeNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("unexpected message: %s", err.Message)
	}
	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, err.HTTPStatus())
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("db connection failed")
	err := Wrap(CodeInternal, "failed to fetch user", original)

	if !errors.Is(err, original) {
		t.Error("expected wrapped error to unwrap to original")
	}

	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.HTTPStatus())
	}
}

func TestWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field": "email",
		"rule":  "required",
	}
	err := ValidationError("validation failed", details)

	if err.Details["field"] != "email" {
		t.Error("expected details to contain field=email")
	}
}

func TestConvenienceConstructors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected int
	}{
		{"NotFound", NotFound("user"), http.StatusNotFound},
		{"Forbidden", Forbidden("access denied"), http.StatusForbidden},
		{"Unauthorized", Unauthorized("invalid token"), http.StatusUnauthorized},
		{"Conflict", Conflict("already exists"), http.StatusConflict},
		{"BadRequest", BadRequest("invalid input"), http.StatusBadRequest},
		{"RateLimited", RateLimited(), http.StatusTooManyRequests},
		{"Internal", Internal("db error", errors.New("conn refused")), http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.HTTPStatus() != tc.expected {
				t.Errorf("expected status %d, got %d", tc.expected, tc.err.HTTPStatus())
			}
		})
	}
}

func TestError_String(t *testing.T) {
	err := NotFound("order")
	expected := "[NOT_FOUND] order not found"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	wrapped := Internal("db fail", errors.New("timeout"))
	if wrapped.Error() == "" {
		t.Error("expected non-empty error string for wrapped error")
	}
}
