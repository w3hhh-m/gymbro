package register

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
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

type UserRegisterer interface {
	RegisterNewUser(usr storage.User) (int, error)
}

func New(log *slog.Logger, usrRegisterer UserRegisterer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.register.New"
		log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		var usr storage.User
		err := render.DecodeJSON(r.Body, &usr)
		if err != nil {
			log.Error("failed to decode request", slog.Any("error", err))
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", usr))
		if err := validator.New().Struct(usr); err != nil {
			log.Info("failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, 400)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}
		if usr.CreatedAt.IsZero() {
			usr.CreatedAt = time.Now()
		}

		passHash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		usr.Password = string(passHash)

		id, err := usrRegisterer.RegisterNewUser(usr)
		if err != nil {
			if errors.Is(err, storage.ErrUserExists) {
				render.Status(r, 400)
				render.JSON(w, r, resp.Error("user already exists"))
				return
			}
			log.Error("failed to register user", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("registered user", slog.Int("id", id))
		responseOK(w, r, id)
	}
}
