package mwlogger

import (
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

// New creates a middleware function for logging HTTP requests.
func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Enhance the logger with additional context for the middleware
		log = log.With(slog.String("component", "middleware/logger"))
		log.Info("Logger middleware initialized")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a logger entry with request details
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			// Wrap the response writer to capture the status and bytes written
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			// Log the request details after serving
			defer func() {
				entry.Info("Request served",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(start).String()),
				)
			}()

			// Pass the request to the next handler
			next.ServeHTTP(ww, r)
		})
	}
}
