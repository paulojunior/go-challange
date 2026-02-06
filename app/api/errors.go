package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/logger"
	"github.com/mytheresa/go-hiring-challenge/app/middleware"
	"github.com/mytheresa/go-hiring-challenge/app/services"
	"gorm.io/gorm"
)

// ErrorCode represents a standardized error code.
type ErrorCode string

const (
	ErrCodeInvalidInput ErrorCode = "invalid_input"
	ErrCodeNotFound     ErrorCode = "not_found"
	ErrCodeInternal     ErrorCode = "internal_error"
)

// ErrorResponse represents a standardized error response.
type ErrorResponseBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// HandleError maps application errors to HTTP responses.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	var status int
	var code ErrorCode
	var message string

	switch {
	case errors.Is(err, services.ErrInvalidOffset):
		status = http.StatusBadRequest
		code = ErrCodeInvalidInput
		message = err.Error()
	case errors.Is(err, services.ErrInvalidLimit):
		status = http.StatusBadRequest
		code = ErrCodeInvalidInput
		message = err.Error()
	case errors.Is(err, services.ErrInvalidPrice):
		status = http.StatusBadRequest
		code = ErrCodeInvalidInput
		message = err.Error()
	case errors.Is(err, services.ErrNegativePrice):
		status = http.StatusBadRequest
		code = ErrCodeInvalidInput
		message = err.Error()
	case errors.Is(err, services.ErrInvalidCategoryInput):
		status = http.StatusBadRequest
		code = ErrCodeInvalidInput
		message = err.Error()
	case errors.Is(err, services.ErrInvalidInput):
		status = http.StatusBadRequest
		code = ErrCodeInvalidInput
		message = "Invalid input provided"
	case errors.Is(err, services.ErrNotFound):
		status = http.StatusNotFound
		code = ErrCodeNotFound
		message = "Resource not found"
	case errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound
		code = ErrCodeNotFound
		message = "Resource not found"
	default:
		status = http.StatusInternalServerError
		code = ErrCodeInternal
		message = "An internal error occurred"

		// Log internal errors with full details
		logger.Error("Internal server error",
			slog.String("request_id", middleware.GetRequestID(r.Context())),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponseBody{
		Code:    code,
		Message: message,
	}

	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		logger.Error("Failed to encode error response",
			slog.String("request_id", middleware.GetRequestID(r.Context())),
			slog.String("error", encErr.Error()),
		)
	}
}
