package end

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewEndHandler returns a handler function to end a workout session.
func NewEndHandler(log *slog.Logger, sessionRepo storage.SessionRepository, woRepo storage.WorkoutRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.end.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Retrieve user ID from JWT token in context.
		userID := jwt.GetUserIDFromContext(r.Context())

		// Get the current workout session for the user.
		activeSession, err := sessionRepo.GetSession(userID)
		if err != nil {
			log.Error("Cant get session", slog.String("user_id", userID), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		err = woRepo.SaveWorkout(activeSession)
		if err != nil {
			log.Error("Cant save workout", slog.String("user_id", userID), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		err = sessionRepo.DeleteSession(userID)
		if err != nil {
			log.Error("Cant end session", slog.String("user_id", userID), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		log.Debug("Workout ended", slog.String("session_id", activeSession.SessionID))

		// Respond with a success message.
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
