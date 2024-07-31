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
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Not gonna test bc repos have same functionality as in default users handlers and sure that goth is carefully tested without me

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

// NewCallbackHandler returns a handler function to complete OAuth authentication
func NewCallbackHandler(log *slog.Logger, userRepo storage.UserRepository, secret string) http.HandlerFunc {
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
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}
		log.Info("Completed OAuth", slog.String("provider", provider), slog.Any("user", user))

		dbUser, err := userRepo.GetUserByEmail(user.Email)
		if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("Failed to retrieve user", slog.String("provider", provider), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}
		if dbUser.UserId == 0 {
			username := strings.Split(user.Email, "@")[0]
			newUser := storage.User{
				Email:     user.Email,
				Username:  username,
				CreatedAt: time.Now(),
			}
			passHash, err := bcrypt.GenerateFromPassword([]byte("RaNdOmPaSsWoRdFoRoAuThUsErS(rEpLaCe)"), bcrypt.DefaultCost)
			if err != nil {
				log.Error("Failed to generate password", slog.Any("error", err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error"))
				return
			}
			newUser.Password = string(passHash)
			id, err := userRepo.RegisterNewUser(newUser)
			if err != nil {
				log.Error("Failed to register user", slog.Any("error", err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error"))
				return
			}
			log.Info("Registered new OAuth user", slog.Int("id", id))
			dbUser = newUser
			dbUser.UserId = id
		} else {
			log.Info("User already exists", slog.Any("user", dbUser))
		}

		token, err := jwt.NewToken(dbUser, 24*time.Hour, secret)
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
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// NewLogoutHandler returns a handler function to log out a user from OAuth
func NewLogoutHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLogoutHandler"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		if err := gothic.Logout(w, r); err != nil {
			log.Error("Failed to logout", slog.String("provider", provider), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal error"))
		} else {
			session, _ := gothic.Store.Get(r, "auth-session")
			session.Options.MaxAge = -1
			err := session.Save(r, w)
			if err != nil {
				log.Error("Failed to delete user data from session", slog.String("provider", provider), slog.Any("error", err))
			}
		}
		http.Redirect(w, r, "/users/logout", http.StatusTemporaryRedirect)
		log.Info("User logged out", slog.String("provider", provider))
		render.JSON(w, r, "Successfully logged out")
	}
}

// NewLoginHandler returns a handler function to initiate OAuth login
func NewLoginHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLoginHandler"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		log.Info("Starting OAuth login", slog.String("provider", provider))

		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
			log.Info("User already authenticated", slog.String("provider", provider), slog.Any("user", gothUser))
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}
}
