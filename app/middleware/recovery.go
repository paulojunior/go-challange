package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/mytheresa/go-hiring-challenge/app/logger"
)

// Recovery is a middleware that recovers from panics and logs the error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				logger.Error("Panic recovered",
					slog.String("request_id", GetRequestID(r.Context())),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Any("panic", err),
					slog.String("stack", string(debug.Stack())),
				)

				// Return 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"code":"internal_error","message":"An internal error occurred"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
