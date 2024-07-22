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

type Response struct {
	resp.Response
	Id int `json:"id"`
}

func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Id:       id,
	})
}

func New(log *slog.Logger, recordProvider storage.RecordProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.save.New"
		log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		var rec storage.Record
		err := render.DecodeJSON(r.Body, &rec)
		if err != nil {
			log.Error("failed to decode request", slog.Any("error", err))
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", rec))
		if err := validator.New().Struct(rec); err != nil {
			log.Info("failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, 400)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}
		if rec.CreatedAt.IsZero() {
			rec.CreatedAt = time.Now()
		}
		rec.FkUserId = jwt.GetUserIDFromContext(r.Context())
		id, err := recordProvider.SaveRecord(rec)
		if err != nil {
			log.Error("failed to save records", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("saved records", slog.Int("id", id))
		responseOK(w, r, id)
	}
}

// TODO: make tests
// TODO: mocks
