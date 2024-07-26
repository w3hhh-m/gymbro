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

// responseOK sends a successful response with the user ID
func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Id:       id,
	})
}

// NewRegisterHandler returns a handler function to initiate user registration
func NewRegisterHandler(log *slog.Logger, userRepo storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.register.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		var usr storage.User
		// Decode the request body into a User struct
		err := render.DecodeJSON(r.Body, &usr)
		if err != nil {
			log.Error("Failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Failed to decode request"))
			return
		}

		log.Info("Request body decoded", slog.Any("request", usr))

		// Validate the User struct
		if err := validator.New().Struct(usr); err != nil {
			log.Info("Failed to validate request", slog.Any("error", err))
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Set the creation time if it is not set
		if usr.CreatedAt.IsZero() {
			usr.CreatedAt = time.Now()
		}

		// Hash the user's password
		passHash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("Failed to generate password", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}

		usr.Password = string(passHash)

		// Register the new user
		id, err := userRepo.RegisterNewUser(usr)
		if err != nil {
			if errors.Is(err, storage.ErrUserExists) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("User already exists"))
				return
			}
			log.Error("Failed to register user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}

		log.Info("Registered user", slog.Int("id", id))
		responseOK(w, r, id)
	}
}
