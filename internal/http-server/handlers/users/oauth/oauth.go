package oauth

import (
	"GYMBRO/internal/config"
	resp "GYMBRO/internal/http-server/handlers/response"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"log/slog"
	"net/http"
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
	goth.UseProviders(google.New(cfg.GoogleKey, cfg.GoogleSecret, "http://"+cfg.Address+"/users/oauth/google/callback", "profile", "email"))
}

// NewCallbackHandler returns a handler function to complete OAuth authentication
func NewCallbackHandler(log *slog.Logger) http.HandlerFunc {
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

		session, _ := gothic.Store.Get(r, "auth-session")
		session.Values["user_id"] = user.UserID
		session.Values["user_name"] = user.Name
		session.Values["user_email"] = user.Email
		err = session.Save(r, w)
		if err != nil {
			log.Error("Failed to save user data to session", slog.String("provider", provider), slog.Any("error", err))
		}
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
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// NewLoginHandler returns a handler function to initiate OAuth login
func NewLoginHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.NewLogoutHandler"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		provider := chi.URLParam(r, "provider")
		ctx := context.WithValue(r.Context(), "provider", provider)
		r = r.WithContext(ctx)

		log.Info("Starting OAuth login", slog.String("provider", provider))

		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
			session, _ := gothic.Store.Get(r, "auth-session")
			session.Values["user_id"] = gothUser.UserID
			session.Values["user_name"] = gothUser.Name
			session.Values["user_email"] = gothUser.Email
			err = session.Save(r, w)
			if err != nil {
				log.Error("Failed to save user data to session", slog.String("provider", provider), slog.Any("error", err))
			}
			log.Info("User already authenticated", slog.String("provider", provider), slog.Any("user", gothUser))
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	}
}
