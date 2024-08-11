package oauth

import (
	"GYMBRO/internal/config"
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func NewOAuth(cfg *config.Config) {
	store := sessions.NewCookieStore([]byte(cfg.SecretKey))
	store.MaxAge(86400 * 7)
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.Env == "prod",
	}

	gothic.Store = store
	goth.UseProviders(google.New(cfg.GoogleKey, cfg.GoogleSecret, "http://"+cfg.Address+"/users/oauth/google/callback", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"))
}

func NewOAuthCallbackHandler(log *slog.Logger, userRepo storage.UserRepository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewCallbackHandler"
		provider := chi.URLParam(r, "provider")
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("provider", provider))

		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Error("Failed to complete OAuth", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeOAuthError, "Please try again later"))
			return
		}
		log.Debug("Completed OAuth")

		dbUser, err := userRepo.GetUserByEmail(&user.Email)
		if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("Failed to GET user", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		if dbUser == nil {
			username := strings.Split(user.Email, "@")[0]
			newUser := storage.User{
				UserId:   storage.GenerateUID(),
				Email:    user.Email,
				Username: username,
				GoogleId: user.UserID,
			}
			id, err := userRepo.RegisterNewUser(&newUser)
			if err != nil {
				if errors.Is(err, storage.ErrUserExists) {
					log.Warn("User already exists")
					render.Status(r, http.StatusBadRequest)
					render.JSON(w, r, resp.Error("User exists", resp.CodeUserExists, "User with this username or email already exists. Try again with another username or email"))
					return
				}
				log.Error("Failed to SAVE user", slog.Any("error", err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
				return
			}
			log.Debug("Registered new OAuth user")
			dbUser = &newUser
			dbUser.UserId = *id
		} else {
			log.Debug("User already exists")
		}

		token, err := jwt.NewToken(*dbUser, cfg.JWTLifetime, cfg.SecretKey)
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

func NewOAuthLogoutHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLogoutHandler"
		provider := chi.URLParam(r, "provider")
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("provider", provider))

		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		if err := gothic.Logout(w, r); err != nil {
			log.Error("Failed to logout", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeOAuthError, "Please try again later"))
		} else {
			session, _ := gothic.Store.Get(r, "auth-session")
			session.Options.MaxAge = -1
			err := session.Save(r, w)
			if err != nil {
				log.Error("Failed to delete user data from session", slog.Any("error", err))
			}
		}
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
		http.Redirect(w, r, "/users/logout", http.StatusTemporaryRedirect)
	}
}

func NewOAuthLoginHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLoginHandler"
		provider := chi.URLParam(r, "provider")
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("provider", provider))

		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		log.Debug("Starting OAuth login")

		if _, err := gothic.CompleteUserAuth(w, r); err == nil {
			log.Debug("User already authenticated")
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}
}
