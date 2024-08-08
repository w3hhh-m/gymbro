package start

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
)

// NewStartHandler returns a handler function to initiate a new workout session for a user.
func NewStartHandler(log *slog.Logger, sessionRepo storage.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.start.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		userID := jwt.GetUserIDFromContext(r.Context())
		activeSession, err := sessionRepo.GetSession(userID)

		// Check if the user already has an active workout session
		if activeSession != nil {
			log.Debug("User already has active workout", slog.String("user_id", userID))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, resp.Error("Already has active workout", resp.CodeActiveWorkout, "End current workout to start new one"))
			return
		}

		if !errors.Is(err, storage.ErrNoSession) {
			log.Error("Cant get session", slog.String("user_id", userID), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		session := storage.WorkoutSession{
			SessionID:   storage.GenerateUID(),
			UserID:      userID,
			StartTime:   time.Now(),
			LastUpdated: time.Now(),
			IsActive:    true,
			Records:     []storage.Record{},
			Points:      0,
		}

		// Save the new workout to the sessionrepo
		if err := sessionRepo.CreateSession(session); err != nil {
			log.Error("Failed to create session", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		log.Debug("Workout started", slog.String("session_id", session.SessionID))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
