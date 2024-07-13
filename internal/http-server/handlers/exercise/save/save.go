package save

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
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
		// TODO: add validation
		// TODO: make user choose time
		req.Timestamp = time.Now()
		id, err := exSaver.SaveExercise(req)
		if err != nil {
			log.Error("failed to save exercise", slog.Any("error", err))
			render.JSON(w, r, resp.Error("failed to save exercise"))
			return
		}
		log.Info("saved exercise", slog.Int64("id", id))
		responseOK(w, r, id)
	}
}

// TODO: make tests
// TODO: mocks
