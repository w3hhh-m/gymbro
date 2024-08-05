package start

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
)

// NewStartHandler returns a handler function to initiate a new workout session for a user.
func NewStartHandler(log *slog.Logger, woRepo storage.WorkoutRepository, sm *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.start.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		userID := jwt.GetUserIDFromContext(r.Context())
		activeSession := sm.GetSession(userID)

		// Check if the user already has an active workout session
		if activeSession != nil {
			log.Debug("User already has active workout", slog.String("user_id", userID))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, resp.Error("Already has active workout", resp.CodeActiveWorkout, "End current workout to start new one"))
			return
		}

		// Create a new workout record
		workout := storage.Workout{
			WorkoutId: storage.GenerateUID(),
			FkUserId:  userID,
			StartTime: time.Now(),
			IsActive:  true,
		}

		// Save the new workout to the repository
		if err := woRepo.CreateWorkout(workout); err != nil {
			log.Error("Failed to create workout", slog.String("error", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Start a new session for the user
		sm.StartSession(userID, workout.WorkoutId)
		log.Debug("Workout created", slog.String("workout_id", workout.WorkoutId))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
