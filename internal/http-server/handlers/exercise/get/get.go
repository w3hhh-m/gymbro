package get

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type ExerciseGetter interface {
	GetExercise(id int64) (storage.Exercise, error)
}

type Response struct {
	resp.Response
	Exercise storage.Exercise `json:"exercise"`
}

func responseOK(w http.ResponseWriter, r *http.Request, exercise storage.Exercise) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Exercise: exercise,
	})
}
func New(log *slog.Logger, exGetter ExerciseGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.exercise.get.New"
		log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("empty id in request url")
			render.JSON(w, r, resp.Error("empty id in request url"))
			return
		}
		idnum, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.Info("nan id in request url")
			render.JSON(w, r, resp.Error("nan id in request url"))
			return
		}
		resEx, err := exGetter.GetExercise(idnum)
		if errors.Is(err, storage.ErrExerciseNotFound) {
			log.Info("no such exercise")
			render.JSON(w, r, resp.Error("no such exercise"))
			return
		}
		if err != nil {
			log.Info("exercise not found")
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		responseOK(w, r, resEx)
	}
}
