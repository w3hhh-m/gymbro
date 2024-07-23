package delete

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

func New(log *slog.Logger, recordProvider storage.RecordProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.delete.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
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
		resRec, err := recordProvider.GetRecord(idnum)
		if errors.Is(err, storage.ErrRecordNotFound) {
			log.Info("no such records", slog.Int("id", idnum))
			render.Status(r, 404)
			render.JSON(w, r, resp.Error("no such records"))
			return
		}
		uid := jwt.GetUserIDFromContext(r.Context())
		if resRec.FkUserId != uid {
			log.Info("attempting to get other user record", slog.String("id", id))
			render.Status(r, 401)
			render.JSON(w, r, resp.Error("you are not the owner of this record"))
			return
		}
		err = recordProvider.DeleteRecord(idnum)
		if err != nil {
			log.Error("record was not deleted", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		render.JSON(w, r, resp.OK())
	}
}
