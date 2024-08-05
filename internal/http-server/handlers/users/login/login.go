package login

import (
	"GYMBRO/internal/config"
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

// Request represents the login request containing email and password.
type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// NewLoginHandler returns a handler function to authenticate user login.
func NewLoginHandler(log *slog.Logger, userRepo storage.UserRepository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.login.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		var req Request
		// Decode the request body into a Request struct.
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request", resp.CodeBadRequest, "Check the request fields for typos or naming errors"))
			return
		}

		log.Debug("Request body decoded", slog.Any("request", req))

		// Validate the Request struct.
		if err := validator.New().Struct(req); err != nil {
			log.Debug("Failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			if errors.As(err, &validateErr) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.ValidationError(validateErr))
			} else {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("Validation failed", resp.CodeValidationError, "Check the validation rules and request fields"))
			}
			return
		}

		// Retrieve the user by email.
		usr, err := userRepo.GetUserByEmail(req.Email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Debug("User not found", slog.Any("email", req.Email))
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("User not found", resp.CodeNotFound, "Verify the email address or register a new account"))
				return
			}
			log.Error("Failed to get user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Failed to login", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Compare the hashed password with the provided password.
		if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(req.Password)); err != nil {
			log.Debug("Invalid password", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Invalid credentials", resp.CodeBadRequest, "Check your email and password and try again"))
			return
		}

		// Generate a new JWT token.
		token, err := jwt.NewToken(*usr, cfg.JWTLifetime, cfg.SecretKey)
		if err != nil {
			log.Error("Failed to generate token", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Set the JWT token as a cookie.
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(cfg.JWTLifetime),
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: token,
		})

		render.JSON(w, r, resp.OK())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
