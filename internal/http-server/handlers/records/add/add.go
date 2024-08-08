package add

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

// NewAddHandler returns a handler function to add a new workout record.
func NewAddHandler(log *slog.Logger, sessionRepo storage.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.records.add.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Retrieve user ID from JWT token in context.
		userID := jwt.GetUserIDFromContext(r.Context())

		var rec storage.Record
		// Decode the request body into a Record struct.
		err := render.DecodeJSON(r.Body, &rec)
		if err != nil {
			log.Warn("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request", resp.CodeBadRequest, "Check the request fields for typos or naming errors"))
			return
		}

		// Get the current workout session for the user.
		activeSession, err := sessionRepo.GetSession(userID)
		if err != nil {
			log.Error("Cant get session", slog.String("user_id", userID), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Set the workout ID and generate a new record ID.
		rec.FkWorkoutId = activeSession.SessionID
		rec.RecordId = storage.GenerateUID()

		// Calculate points
		points := rec.Reps * rec.Weight

		log.Debug("Request body decoded", slog.Any("request", rec))

		// Validate the Record struct.
		if err := validator.New().Struct(rec); err != nil {
			log.Debug("Failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			if errors.As(err, &validateErr) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ValidationError(validateErr))
			} else {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("Validation failed", resp.CodeValidationError, "Check the validation rules and request fields"))
			}
			return
		}

		activeSession.Records = append(activeSession.Records, rec)
		activeSession.Points += points

		// Update session data
		if err := sessionRepo.UpdateSession(userID, activeSession); err != nil {
			log.Error("Failed to update workout", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Respond with a success message.
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
