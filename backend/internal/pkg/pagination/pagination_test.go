package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	original := Cursor{ID: "abc-123", CreatedAt: "2024-01-01T00:00:00Z"}
	encoded := EncodeCursor(original)
	if encoded == "" {
		t.Fatal("expected non-empty encoded cursor")
	}

	decoded := DecodeCursor(encoded)
	if decoded == nil {
		t.Fatal("expected non-nil decoded cursor")
	}

	if decoded.ID != original.ID {
		t.Errorf("decoded ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.CreatedAt != original.CreatedAt {
		t.Errorf("decoded CreatedAt = %q, want %q", decoded.CreatedAt, original.CreatedAt)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	result := DecodeCursor("not-valid-base64!!!")
	if result != nil {
		t.Error("expected nil for invalid cursor")
	}
}

func TestParseFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?per_page=50", nil)
	params := ParseFromRequest(req)

	if params.Limit != 50 {
		t.Errorf("limit = %d, want 50", params.Limit)
	}
	if params.Cursor != nil {
		t.Error("expected nil cursor")
	}
}

func TestParseFromRequest_MaxLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?per_page=500", nil)
	params := ParseFromRequest(req)

	if params.Limit != MaxPageSize {
		t.Errorf("limit = %d, want %d (max)", params.Limit, MaxPageSize)
	}
}

func TestParseFromRequest_WithCursor(t *testing.T) {
	cursor := EncodeCursor(Cursor{ID: "test-id"})
	req := httptest.NewRequest(http.MethodGet, "/?cursor="+cursor, nil)
	params := ParseFromRequest(req)

	if params.Cursor == nil {
		t.Fatal("expected non-nil cursor")
	}
	if params.Cursor.ID != "test-id" {
		t.Errorf("cursor ID = %q, want %q", params.Cursor.ID, "test-id")
	}
}

func TestFormatLimitOffset(t *testing.T) {
	limit, offset := FormatLimitOffset(3, 10)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 20 {
		t.Errorf("offset = %d, want 20", offset)
	}
}

func TestFormatLimitOffset_Defaults(t *testing.T) {
	limit, offset := FormatLimitOffset(0, 0)
	if limit != DefaultPageSize {
		t.Errorf("limit = %d, want %d", limit, DefaultPageSize)
	}
	if offset != 0 {
		t.Errorf("offset = %d, want 0", offset)
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse([]string{"a", "b"}, 100, 20, "last-id")
	if resp.Meta.Total != 100 {
		t.Errorf("total = %d, want 100", resp.Meta.Total)
	}
	if resp.Meta.NextCursor == "" {
		t.Error("expected non-empty next cursor")
	}
}
