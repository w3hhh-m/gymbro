package add

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/lib/points"
	"GYMBRO/internal/lib/validation"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewAddHandler creates an HTTP handler for adding a new workout record.
// It decodes the request, validates it, updates the workout session with new points,
// and responds with the appropriate status. (2 sessionRepo calls, 1 userRepo call)
func NewAddHandler(log *slog.Logger, sessionRepo storage.SessionRepository, userRepo storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.records.add.New"
		userID := jwt.GetUserIDFromContext(r.Context())
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("user_id", userID))

		var record storage.Record
		if err := render.DecodeJSON(r.Body, &record); err != nil {
			log.Warn("Failed to decode request", slog.Any("error", err), slog.Any("request", r.Body))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request", resp.CodeBadRequest, "Check the request fields for typos or naming errors"))
			return
		}
		log.Debug("Request body decoded", slog.Any("record", record))

		if err := validation.ValidateStruct(log, &record); err != nil {
			validation.HandleValidationError(w, r, err)
			return
		}

		activeSession, err := sessionRepo.GetSession(&userID)
		if err != nil {
			log.Error("Can't GET session", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		record.FkWorkoutId = activeSession.SessionID
		record.RecordId = storage.GenerateUID()

		hasMax := true
		userMax, err := userRepo.GetUserMax(&userID, &record.FkExerciseId)
		if err != nil {
			if errors.Is(err, storage.ErrNoMaxes) {
				hasMax = false
			} else {
				log.Error("Failed to GET userMax", slog.Any("error", err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
				return
			}
		}

		if !hasMax || (userMax != nil && userMax.MaxWeight < record.Weight) {
			userMax = &storage.Max{
				UserID:     userID,
				ExerciseId: record.FkExerciseId,
				MaxWeight:  record.Weight,
				Reps:       record.Reps,
			}
		}

		record.Points = points.CalculatePoints(userMax.MaxWeight, userMax.Reps, record.Weight, record.Reps, 100)

		activeSession.Records = append(activeSession.Records, record)
		activeSession.Points += record.Points

		if err := sessionRepo.UpdateSession(&userID, activeSession); err != nil {
			log.Error("Failed to UPDATE session", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
