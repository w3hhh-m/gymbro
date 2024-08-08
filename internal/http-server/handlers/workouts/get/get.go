package getwo

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewGetWorkoutHandler returns a handler function to get a workout by ID with its associated records.
func NewGetWorkoutHandler(log *slog.Logger, woRepo storage.WorkoutRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.get.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Retrieve user ID from JWT token in context.
		userID := jwt.GetUserIDFromContext(r.Context())

		// Get the workout ID from the URL parameters.
		workoutID := chi.URLParam(r, "workoutID")

		// Retrieve the workout from the repository.
		workout, err := woRepo.GetWorkout(workoutID)
		if err != nil {
			if errors.Is(err, storage.ErrWorkoutNotFound) {
				log.Debug("Workout not found", slog.String("workout_id", workoutID))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("Workout not found", resp.CodeNotFound, "The requested workout does not exist"))
				return
			}
			log.Error("Failed to retrieve workout", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Check if the workout belongs to the current user.
		if workout.UserID != userID {
			log.Debug("User does not own the workout", slog.String("user_id", userID))
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, resp.Error("Forbidden", resp.CodeForbidden, "You do not have permission to access this workout"))
			return
		}

		// Respond with the workout and its records.
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.Data(workout))
	}
}
