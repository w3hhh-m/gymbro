package login

import (
	"GYMBRO/internal/config"
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/lib/validation"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"time"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// NewLoginHandler creates an HTTP handler for user authentication.
// It handles login requests by validating the input, checking credentials,
// and issuing a JWT token upon successful authentication. (1 userRepo call)
func NewLoginHandler(log *slog.Logger, userRepo storage.UserRepository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.login.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		var request Request
		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Warn("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request", resp.CodeBadRequest, "Check the request fields for typos or naming errors"))
			return
		}
		log.Debug("Request body decoded", slog.Any("request", request))

		if err := validation.ValidateStruct(log, &request); err != nil {
			validation.HandleValidationError(w, r, err)
			return
		}

		usr, err := userRepo.GetUserByEmail(&request.Email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Debug("Invalid credentials", slog.Any("request", request))
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("Invalid credentials", resp.CodeBadRequest, "Check your email and password and try again"))
				return
			}
			log.Error("Failed to GET user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Failed to login", resp.CodeInternalError, "Please try again later"))
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(request.Password)); err != nil {
			log.Debug("Invalid credentials", slog.Any("request", request))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Invalid credentials", resp.CodeBadRequest, "Check your email and password and try again"))
			return
		}

		token, err := jwt.NewToken(*usr, cfg.JWTLifetime, cfg.SecretKey)
		if err != nil {
			log.Error("Failed to GENERATE token", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(cfg.JWTLifetime),
			Name:     "jwt",
			Value:    token,
		})

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
