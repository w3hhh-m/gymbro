package end

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

// NewEndHandler creates an HTTP handler to end a workout session.
// It retrieves the active session, checks for new maxes, saves the workout data, deletes the session,
// and updates the user's status, responding with success or handling errors. (2 sessionRepo calls, 2+ userRepo calls, 1 workoutRepo calls
func NewEndHandler(log *slog.Logger, sessionRepo storage.SessionRepository, workoutRepo storage.WorkoutRepository, userRepo storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.end.New"
		userID := jwt.GetUserIDFromContext(r.Context())
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("user_id", userID))

		activeSession, err := sessionRepo.GetSession(&userID)
		if err != nil {
			log.Error("Cant GET session", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		maxSessionRecords := make(map[int]struct {
			RecordId int
			Reps     int
			Weight   int
		})

		for i, record := range activeSession.Records {
			current, exists := maxSessionRecords[record.FkExerciseId]
			if !exists || record.Weight > current.Weight || (record.Weight == current.Weight && record.Reps > current.Reps) {
				maxSessionRecords[record.FkExerciseId] = struct {
					RecordId int
					Reps     int
					Weight   int
				}{RecordId: i, Reps: record.Reps, Weight: record.Weight}
			}
		}

		maxDbRecords, err := userRepo.GetUserMaxes(&userID)
		if err != nil && !errors.Is(err, storage.ErrNoMaxes) {
			log.Error("Can't GET userMaxes", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		dbMaxMap := make(map[int]*storage.Max)
		for _, dbMax := range maxDbRecords {
			dbMaxMap[dbMax.ExerciseId] = dbMax
		}

		for exerciseId, sessionMax := range maxSessionRecords {
			dbMax, exists := dbMaxMap[exerciseId]
			if !exists || sessionMax.Weight > dbMax.MaxWeight || (sessionMax.Weight == dbMax.MaxWeight && sessionMax.Reps > dbMax.Reps) {
				err = userRepo.SetUserMax(&userID, &storage.Max{
					UserID:     userID,
					ExerciseId: exerciseId,
					MaxWeight:  sessionMax.Weight,
					Reps:       sessionMax.Reps,
				})
				if err != nil {
					log.Error("Can't SET userMax", slog.Any("error", err))
					render.Status(r, http.StatusInternalServerError)
					render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
					return
				}
				activeSession.Records[sessionMax.RecordId].Points += 50
				activeSession.Points += 50
			}
		}

		err = workoutRepo.SaveWorkout(activeSession)
		if err != nil {
			log.Error("Cant SAVE workout", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		err = sessionRepo.DeleteSession(&userID)
		if err != nil {
			log.Error("Cant DELETE session", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		if err := userRepo.ChangeStatus(&userID, false); err != nil {
			log.Error("Failed to CHANGE user status", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		log.Debug("Workout ended", slog.String("session_id", activeSession.SessionID))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
