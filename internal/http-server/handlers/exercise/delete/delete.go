package delete

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

type ExerciseDeleter interface {
	DeleteExercise(id int64) error
}

func New(log *slog.Logger, exDeleter ExerciseDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.exercise.delete.New"
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
		err = exDeleter.DeleteExercise(idnum)
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
		render.JSON(w, r, resp.OK())
	}
}
