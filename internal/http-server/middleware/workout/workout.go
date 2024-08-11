package mwworkout

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// WithActiveSessionCheck ensures that a user has an active workout session before proceeding.
func WithActiveSessionCheck(log *slog.Logger, sessionRepo storage.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.mwworkout.WithActiveSessionCheck"
			reqID := middleware.GetReqID(r.Context())
			userID := jwt.GetUserIDFromContext(r.Context())
			log = log.With(slog.String("op", op), slog.Any("request_id", reqID), slog.String("user_id", userID))

			session, err := sessionRepo.GetSession(&userID)

			if err != nil {
				if !errors.Is(err, storage.ErrNoSession) {
					log.Error("Cant GET session", slog.Any("error", err))
					render.Status(r, http.StatusInternalServerError)
					render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
					return
				}
			}

			if session == nil {
				log.Debug("No active session")
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, resp.Error("No active workout at this time", resp.CodeNoActiveWorkout, "You need to start workout first"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
