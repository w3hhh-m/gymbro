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

func NewOAuth(cfg *config.Config) {
	store := sessions.NewCookieStore([]byte(cfg.SecretKey))
	store.MaxAge(86400 * 7)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	if cfg.Env == "prod" {
		store.Options.Secure = true
	} else {
		store.Options.Secure = false
	}

	gothic.Store = store

	goth.UseProviders(google.New(cfg.GoogleKey, cfg.GoogleSecret, "http://"+cfg.Address+"/users/oauth/google/callback"))
}

func NewCB(log *slog.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.oauth.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Error("Failed to complete OAuth", slog.String("provider", provider), slog.Any("error", err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		_ = user
		// TODO: remember values
	})
}

func NewLogout() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
		err := gothic.Logout(w, r)
		if err != nil {
			return
		}
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}

func NewLogin() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// try to get the user without re-authenticating
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
		if gothUser, err := gothic.CompleteUserAuth(w, r); err == nil {
			_ = gothUser
			// TODO: remember values
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	})
}
