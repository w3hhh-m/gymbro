package get

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
	"strconv"
)

// Response defines the structure of the response for the GET request
type Response struct {
	resp.Response
	Record storage.Record `json:"record"`
}

// responseOK sends a successful response with the record
func responseOK(w http.ResponseWriter, r *http.Request, record storage.Record) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Record:   record,
	})
}

// NewGetHandler returns a handler function to get a record
func NewGetHandler(log *slog.Logger, recordRepo storage.RecordRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.get.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Extract the record ID from the URL parameters
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("Empty id in request URL")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Empty id in request URL"))
			return
		}

		// Convert the ID to an integer
		idnum, err := strconv.Atoi(id)
		if err != nil {
			log.Info("Non-numeric id in request URL", slog.String("id", id))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Non-numeric id in request URL"))
			return
		}

		// Retrieve the record from the database
		resRec, err := recordRepo.GetRecord(idnum)
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Info("No such record", slog.Int("id", idnum))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, resp.Error("No such record"))
			return
		}
		if err != nil {
			log.Error("Record was not retrieved", slog.Int("id", idnum), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}

		// Check if the user is the owner of the record
		uid := jwt.GetUserIDFromContext(r.Context())
		if resRec.FkUserId != uid {
			log.Info("Attempting to get other user's record", slog.String("id", id))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, resp.Error("You are not the owner of this record"))
			return
		}

		// Send a successful response
		responseOK(w, r, resRec)
	}
}
