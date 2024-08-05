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

// NewOAuth initializes OAuth settings and providers
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

// NewOAuthCallbackHandler returns a handler function to complete OAuth authentication
func NewOAuthCallbackHandler(log *slog.Logger, userRepo storage.UserRepository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewCallbackHandler"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Error("Failed to complete OAuth", slog.String("provider", provider), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeOAuthError, "Please try again later"))
			return
		}
		log.Debug("Completed OAuth", slog.String("provider", provider), slog.String("email", user.Email))

		dbUser, err := userRepo.GetUserByEmail(user.Email)
		if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("Failed to get user", slog.Any("error", err))
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
			id, err := userRepo.RegisterNewUser(newUser)
			if err != nil {
				log.Error("Failed to register user", slog.Any("error", err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
				return
			}
			log.Debug("Registered new OAuth user", slog.String("id", *id))
			dbUser = &newUser
			dbUser.UserId = *id
		} else {
			log.Debug("User already exists", slog.Any("user", dbUser))
		}

		token, err := jwt.NewToken(*dbUser, cfg.JWTLifetime, cfg.SecretKey)
		if err != nil {
			log.Error("Failed to generate token", slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
			return
		}

		// Set the JWT token as a cookie
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(cfg.JWTLifetime),
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: token,
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// NewOAuthLogoutHandler returns a handler function to log out a user from OAuth
func NewOAuthLogoutHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLogoutHandler"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		if err := gothic.Logout(w, r); err != nil {
			log.Error("Failed to logout", slog.String("provider", provider), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error", resp.CodeOAuthError, "Please try again later"))
		} else {
			session, _ := gothic.Store.Get(r, "auth-session")
			session.Options.MaxAge = -1
			err := session.Save(r, w)
			if err != nil {
				log.Error("Failed to delete user data from session", slog.String("provider", provider), slog.Any("error", err))
			}
		}
		http.Redirect(w, r, "/users/logout", http.StatusTemporaryRedirect)
	}
}

// NewOAuthLoginHandler returns a handler function to initiate OAuth login
func NewOAuthLoginHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLoginHandler"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		log.Debug("Starting OAuth login", slog.String("provider", provider))

		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
			log.Debug("User already authenticated", slog.String("provider", provider), slog.String("email", gothUser.Email))
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}
}
