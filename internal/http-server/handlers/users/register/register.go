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
)

// NewRegisterHandler returns a handler function to initiate user registration.
func NewRegisterHandler(log *slog.Logger, userRepo storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.register.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		var usr storage.User
		// Decode the request body into a User struct.
		err := render.DecodeJSON(r.Body, &usr)
		if err != nil {
			log.Warn("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request", resp.CodeBadRequest, "Check the request fields for typos or naming errors"))
			return
		}

		log.Debug("Request body decoded", slog.Any("request", usr))

		// Validate the User struct.
		if err := validator.New().Struct(usr); err != nil {
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

		// Check if user already exists
		existingUser, err := userRepo.GetUserByEmail(usr.Email)
		if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("Failed to get user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}
		if existingUser != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("User already exists", resp.CodeUserExists, "User with this email already exists. Check email for typos or try to login"))
			return
		}

		// Hash the user's password.
		passHash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Warn("Failed to generate password", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Update the user struct with the hashed password.
		usr.Password = string(passHash)
		usr.UserId = storage.GenerateUID()

		// Register the new user.
		id, err := userRepo.RegisterNewUser(usr)
		if err != nil {
			log.Error("Failed to register user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		log.Debug("Registered user", slog.String("id", *id))
		render.JSON(w, r, resp.OK())

		// Redirect to the login page after successful registration.
		http.Redirect(w, r, "/users/login", http.StatusTemporaryRedirect)
	}
}
