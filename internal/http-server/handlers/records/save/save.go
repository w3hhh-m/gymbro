package save

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
	"time"
)

// Response defines the structure of the response for the POST request
type Response struct {
	resp.Response
	Id int `json:"id"`
}

// responseOK sends a successful response with the record ID
func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Id:       id,
	})
}

// NewSaveHandler returns a handler function to save a record
func NewSaveHandler(log *slog.Logger, recordRepo storage.RecordRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.save.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Decode the JSON request body into a Record struct
		var rec storage.Record
		err := render.DecodeJSON(r.Body, &rec)
		if err != nil {
			log.Error("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request"))
			return
		}
		log.Info("Request body decoded", slog.Any("request", rec))

		// Validate the decoded Record struct
		if err := validator.New().Struct(rec); err != nil {
			log.Info("Failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Set the creation time if not provided
		if rec.CreatedAt.IsZero() {
			rec.CreatedAt = time.Now()
		}

		// Set the user ID from the JWT context
		rec.FkUserId = jwt.GetUserIDFromContext(r.Context())

		// Save the record to the database
		id, err := recordRepo.SaveRecord(rec)
		if err != nil {
			log.Error("Failed to save record", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}
		log.Info("Record saved", slog.Int("id", id))

		// Send a successful response
		responseOK(w, r, id)
	}
}
