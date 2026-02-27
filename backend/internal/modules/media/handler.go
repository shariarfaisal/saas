package media

import (
	"net/http"

	"github.com/munchies/platform/backend/internal/pkg/apperror"
	"github.com/munchies/platform/backend/internal/pkg/respond"
)

// Handler handles media upload HTTP requests.
type Handler struct{}

// NewHandler creates a new media handler.
func NewHandler() *Handler {
	return &Handler{}
}

// Upload handles POST /api/v1/media/upload
// This is a stub — S3/R2 integration is future work.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respond.Error(w, apperror.BadRequest("invalid multipart form"))
		return
	}
	_, header, err := r.FormFile("file")
	if err != nil {
		respond.Error(w, apperror.BadRequest("file is required"))
		return
	}

	// Stub response — return a placeholder URL
	respond.JSON(w, http.StatusOK, map[string]string{
		"url":      "https://cdn.example.com/uploads/" + header.Filename,
		"filename": header.Filename,
	})
}
