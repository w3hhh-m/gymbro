package delete

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

type RecordDeleter interface {
	DeleteRecord(id int) error
}

func New(log *slog.Logger, recDeleter RecordDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.record.delete.New"
		log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("empty id in request url") // TODO: not working
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("empty id in request url"))
			return
		}
		idnum, err := strconv.Atoi(id)
		if err != nil {
			log.Info("nan id in request url")
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("nan id in request url"))
			return
		}
		err = recDeleter.DeleteRecord(idnum)
		if err != nil {
			log.Info("record not found", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		render.JSON(w, r, resp.OK())
	}
}
