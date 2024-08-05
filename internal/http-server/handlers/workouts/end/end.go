package end

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewEndHandler returns a handler function to end a workout session.
func NewEndHandler(log *slog.Logger, woRepo storage.WorkoutRepository, sm *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.end.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Retrieve user ID from JWT token in context.
		userID := jwt.GetUserIDFromContext(r.Context())

		// Get the current workout session for the user.
		workoutSession := sm.GetSession(userID)
		if err := woRepo.EndWorkout(workoutSession.SessionID); err != nil {
			log.Error("Failed to end workout", slog.String("error", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// End the session for the user.
		sm.EndSession(userID)
		log.Debug("Workout ended", slog.String("user", userID))

		// Respond with a success message.
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
