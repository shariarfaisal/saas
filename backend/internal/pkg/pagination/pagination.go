package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Cursor holds the opaque pagination state.
type Cursor struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
}

// Params holds parsed pagination parameters from a request.
type Params struct {
	Limit  int
	Cursor *Cursor
}

// Meta is the pagination metadata included in list responses.
type Meta struct {
	Total      int64  `json:"total"`
	Page       int    `json:"page,omitempty"`
	PerPage    int    `json:"per_page"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// PagedResponse wraps list data with pagination metadata.
type PagedResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// ParseFromRequest extracts pagination params from query string.
func ParseFromRequest(r *http.Request) Params {
	q := r.URL.Query()

	limit := DefaultPageSize
	if l := q.Get("per_page"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > MaxPageSize {
				limit = MaxPageSize
			}
		}
	}

	var cursor *Cursor
	if c := q.Get("cursor"); c != "" {
		cursor = DecodeCursor(c)
	}

	return Params{
		Limit:  limit,
		Cursor: cursor,
	}
}

// EncodeCursor encodes a cursor to a base64 string.
func EncodeCursor(c Cursor) string {
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeCursor decodes a base64 cursor string.
func DecodeCursor(encoded string) *Cursor {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil
	}
	return &c
}

// NewMeta creates pagination metadata for a response.
func NewMeta(total int64, perPage int, lastID string) Meta {
	m := Meta{
		Total:   total,
		PerPage: perPage,
	}
	if lastID != "" {
		m.NextCursor = EncodeCursor(Cursor{ID: lastID})
	}
	return m
}

// NewResponse creates a paged response.
func NewResponse(data interface{}, total int64, perPage int, lastID string) PagedResponse {
	return PagedResponse{
		Data: data,
		Meta: NewMeta(total, perPage, lastID),
	}
}

// FormatLimitOffset returns SQL LIMIT and OFFSET values for traditional pagination.
func FormatLimitOffset(page, perPage int) (limit, offset int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = DefaultPageSize
	}
	if perPage > MaxPageSize {
		perPage = MaxPageSize
	}
	return perPage, (page - 1) * perPage
}

// FormatCursorCondition returns a SQL WHERE condition for cursor-based pagination.
func FormatCursorCondition(cursor *Cursor, column string) string {
	if cursor == nil || cursor.ID == "" {
		return ""
	}
	return fmt.Sprintf("%s < '%s'", column, cursor.ID)
}
