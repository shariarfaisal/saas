package respond

import (
	"encoding/json"
	"net/http"

	"github.com/munchies/platform/backend/internal/pkg/apperror"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

// Error writes an AppError as a JSON error response.
func Error(w http.ResponseWriter, err *apperror.AppError) {
	JSON(w, err.HTTPStatus(), map[string]interface{}{
		"error": map[string]interface{}{
			"code":    err.Code,
			"message": err.Message,
			"details": err.Details,
		},
	})
}
