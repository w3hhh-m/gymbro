package delete

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewDeleteHandler returns a handler function to delete a workout record.
func NewDeleteHandler(log *slog.Logger, sessionRepo storage.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.records.delete.New"
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

		points := 0

		// Get recordID from URL parameters.
		recordID := chi.URLParam(r, "recordID")

		// Delete the record
		found := false
		for i, rec := range activeSession.Records {
			if rec.RecordId == recordID {
				points = rec.Weight * rec.Reps
				activeSession.Records = append(activeSession.Records[:i], activeSession.Records[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			log.Debug("Record not found", slog.Any("id", recordID))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("Record not found", resp.CodeNotFound, "Maybe this record doesnt exist"))
			return
		}

		activeSession.Points -= points

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
