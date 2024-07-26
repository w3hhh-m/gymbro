package login

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
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

// responseOK sends a successful response with the JWT token
func responseOK(w http.ResponseWriter, r *http.Request, token string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Token:    token,
	})
}

// NewLoginHandler returns a handler function to authenticate user login
func NewLoginHandler(log *slog.Logger, userRepo storage.UserRepository, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.login.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		var req Request
		// Decode the request body into a Request struct
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request"))
			return
		}

		log.Info("Request body decoded", slog.Any("request", req))

		// Validate the Request struct
		if err := validator.New().Struct(req); err != nil {
			log.Info("Failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Retrieve the user by email
		usr, err := userRepo.GetUserByEmail(req.Email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("User not found", slog.Any("email", req.Email))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("User not found"))
				return
			}
			log.Error("Failed to get user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Failed to login"))
			return
		}

		// Compare the hashed password with the provided password
		if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(req.Password)); err != nil {
			log.Info("Invalid password", slog.Any("error", err))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, resp.Error("Invalid credentials"))
			return
		}

		// Generate a new JWT token
		token, err := jwt.NewToken(usr, 24*time.Hour, secret)
		if err != nil {
			log.Error("Failed to generate token", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}

		// Set the JWT token as a cookie
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: token,
		})

		responseOK(w, r, token)
	}
}
