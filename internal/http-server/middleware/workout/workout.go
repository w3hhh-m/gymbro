package mwworkout

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/lib/jwt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// WithActiveSessionCheck middleware ensures that a user has an active workout session
func WithActiveSessionCheck(log *slog.Logger, sm *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.mwworkout.WithActiveSessionCheck"
			reqID := middleware.GetReqID(r.Context())
			log = log.With(slog.String("op", op), slog.Any("request_id", reqID))

			// Extract user ID from the context
			userID := jwt.GetUserIDFromContext(r.Context())

			// Retrieve the active session for the user
			session := sm.GetSession(userID)
			if session == nil {
				// Log and respond if there is no active session
				log.Debug("No active session", slog.String("user_id", userID))
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, resp.Error("No active workout at this time", resp.CodeNoActiveWorkout, "You need to start workout first"))
				return
			}

			// Proceed to the next handler if an active session is found
			next.ServeHTTP(w, r)
		})
	}
}
