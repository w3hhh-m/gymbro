package logout

import (
	"GYMBRO/internal/lib/jwt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

func New(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.logout.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		uid := jwt.GetUserIDFromContext(r.Context())
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			MaxAge:   -1, // Delete the cookie.
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: "",
		})
		log.Info("User logged out", slog.Int("uid", uid))
		render.JSON(w, r, "Successfully logged out")
	}
}
