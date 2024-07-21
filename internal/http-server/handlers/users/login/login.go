package login

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/jwt"
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
	Token string `json:"token"`
}

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func responseOK(w http.ResponseWriter, r *http.Request, token string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Token:    token,
	})
}

type UserProvider interface {
	GetUser(email string) (storage.User, error)
}

func New(log *slog.Logger, usrProvider UserProvider, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.login.New"
		log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", slog.Any("error", err))
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			log.Info("failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, 400)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}
		usr, err := usrProvider.GetUser(req.Email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("user not found", slog.Any("email", req.Email))
				render.Status(r, 404)
				render.JSON(w, r, resp.Error("user not found"))
				return
			}
			log.Error("failed to get user", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("failed to login"))
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(req.Password)); err != nil {
			log.Info("invalid password", slog.Any("error", err))
			render.Status(r, 401)
			render.JSON(w, r, resp.Error("invalid credentials"))
			return
		}
		token, err := jwt.NewToken(usr, time.Duration(24*time.Hour), secret)
		if err != nil {
			log.Error("failed to generate token", slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		responseOK(w, r, token)
	}
}
