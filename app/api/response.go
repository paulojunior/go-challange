// Package api provides HTTP response utilities for JSON responses.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/logger"
	"github.com/mytheresa/go-hiring-challenge/app/middleware"
)

// OKResponse sends a JSON response with status 200 OK.
func OKResponse(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode JSON response",
			slog.String("request_id", middleware.GetRequestID(r.Context())),
			slog.String("error", err.Error()),
		)
	}
}

// CreatedResponse sends a JSON response with status 201 Created.
func CreatedResponse(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode JSON response",
			slog.String("request_id", middleware.GetRequestID(r.Context())),
			slog.String("error", err.Error()),
		)
	}
}
