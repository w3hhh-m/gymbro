package register

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/validation"
	"GYMBRO/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
)

// NewRegisterHandler creates an HTTP handler for user registration.
// It decodes the request body, validates the user data, checks for existing users,
// hashes the password, and registers the new user, redirecting to the login page upon success. (2 userRepo calls)
func NewRegisterHandler(log *slog.Logger, userRepo storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.register.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		var user storage.User
		err := render.DecodeJSON(r.Body, &user)
		if err != nil {
			log.Warn("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request", resp.CodeBadRequest, "Check the request fields for typos or naming errors"))
			return
		}
		log.Debug("Request body decoded", slog.Any("request", user))

		if err := validation.ValidateStruct(log, &user); err != nil {
			validation.HandleValidationError(w, r, err)
			return
		}
		if user.UserId != "" || user.Points != 0 || user.GoogleId != "" || user.FkGymId != 0 || user.FkClanId != "" {
			log.Warn("User wanted to set restricted field", slog.Any("user", user))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("You do not have permission to set some fields", resp.CodeBadRequest, "Check the request fields for extras"))
			return
		}

		existingUser, err := userRepo.GetUserByEmail(&user.Email)
		if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("Failed to GET user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}
		if existingUser != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("User already exists", resp.CodeUserExists, "User with this email already exists. Check email for typos or try to login"))
			return
		}

		passHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("Failed to GENERATE password", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		user.Password = string(passHash)
		user.UserId = storage.GenerateUID()

		_, err = userRepo.RegisterNewUser(&user)
		if err != nil {
			log.Error("Failed to SAVE user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		log.Debug("Registered user")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
		http.Redirect(w, r, "/users/login", http.StatusTemporaryRedirect)
	}
}
