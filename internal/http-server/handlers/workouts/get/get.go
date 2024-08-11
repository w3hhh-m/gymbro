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

// NewGetWorkoutHandler creates an HTTP handler to retrieve a workout by ID.
// It fetches the workout, checks user ownership, and responds with the workout data or handles errors. (1 workoutRepo call)
func NewGetWorkoutHandler(log *slog.Logger, workoutRepo storage.WorkoutRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.get.New"
		userID := jwt.GetUserIDFromContext(r.Context())
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("user_id", userID))

		workoutID := chi.URLParam(r, "workoutID")

		workout, err := workoutRepo.GetWorkout(&workoutID)
		if err != nil {
			if errors.Is(err, storage.ErrWorkoutNotFound) {
				log.Debug("Workout not found", slog.String("workout_id", workoutID))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("Workout not found", resp.CodeNotFound, "The requested workout does not exist"))
				return
			}
			log.Error("Failed to GET workout", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		if workout.UserID != userID {
			log.Debug("User does not own the workout", slog.String("workout_id", workoutID))
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, resp.Error("Forbidden", resp.CodeForbidden, "You do not have permission to access this workout"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.Data(workout))
	}
}
