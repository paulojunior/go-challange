// Package api provides HTTP response utilities for JSON responses.
package api

import "net/http"

// HandlerFunc defines an HTTP handler that can return an error.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ErrorHandler wraps a HandlerFunc and centralizes error handling.
func ErrorHandler(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			HandleError(w, r, err)
		}
	})
}
