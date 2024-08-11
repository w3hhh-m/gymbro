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

// NewDeleteHandler creates an HTTP handler to delete a workout record.
// It retrieves the user's active session, removes the specified record,
// updates the session, and adjusts the user's points. (2 sessionRepo calls)
func NewDeleteHandler(log *slog.Logger, sessionRepo storage.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.workouts.records.delete.New"
		userID := jwt.GetUserIDFromContext(r.Context())
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("user_id", userID))

		activeSession, err := sessionRepo.GetSession(&userID)
		if err != nil {
			log.Error("Cant GET session", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		points := 0

		recordID := chi.URLParam(r, "recordID")

		found := false
		for i, record := range activeSession.Records {
			if record.RecordId == recordID {
				points = record.Points
				activeSession.Records = append(activeSession.Records[:i], activeSession.Records[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			log.Debug("Record not found", slog.Any("record_id", recordID))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("Record not found", resp.CodeNotFound, "Maybe this record doesnt exist"))
			return
		}

		activeSession.Points -= points

		if err := sessionRepo.UpdateSession(&userID, activeSession); err != nil {
			log.Error("Failed to UPDATE workout", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
	}
}
