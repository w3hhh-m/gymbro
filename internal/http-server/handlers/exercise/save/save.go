package save

import (
	resp "GYMBRO/internal/http-server/handlers/response"
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
	Id int64 `json:"id"`
}

func responseOK(w http.ResponseWriter, r *http.Request, id int64) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Id:       id,
	})
}

type ExerciseSaver interface {
	SaveExercise(ex storage.Exercise) (int64, error)
}

func New(log *slog.Logger, exSaver ExerciseSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.exercise.save.New"
		log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		var req storage.Exercise
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", slog.Any("error", err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			log.Info("failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}
		if req.Timestamp == "" {
			req.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		} else {
			_, err := time.Parse("2006-01-02 15:04:05", req.Timestamp)
			if err != nil {
				log.Info("failed to parse timestamp", slog.Any("error", err))
				render.JSON(w, r, resp.Error("timestamp must be in the format 2006-01-02 15:04:05"))
				return
			}
		}
		id, err := exSaver.SaveExercise(req)
		if err != nil {
			log.Error("failed to save exercise", slog.Any("error", err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("saved exercise", slog.Int64("id", id))
		responseOK(w, r, id)
	}
}

// TODO: make tests
// TODO: mocks
