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

type RecordGetter interface {
	GetRecord(id int) (storage.Record, error)
}

type Response struct {
	resp.Response
	Record storage.Record `json:"record"`
}

func responseOK(w http.ResponseWriter, r *http.Request, record storage.Record) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Record:   record,
	})
}
func New(log *slog.Logger, recGetter RecordGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.record.get.New"
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
			log.Info("nan id in request url", slog.String("id", id))
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("nan id in request url"))
			return
		}
		resRec, err := recGetter.GetRecord(idnum)
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Info("no such record", slog.Int("id", idnum))
			render.Status(r, 404)
			render.JSON(w, r, resp.Error("no such record"))
			return
		}
		if err != nil {
			log.Info("record not found", slog.Int("id", idnum))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		responseOK(w, r, resRec)
	}
}
