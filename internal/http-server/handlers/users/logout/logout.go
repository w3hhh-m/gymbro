package logout

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewLogoutHandler creates an HTTP handler for user logout.
// It clears the JWT cookie, logs the event, and redirects the user to the home page.
func NewLogoutHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.logout.New"
		userID := jwt.GetUserIDFromContext(r.Context())
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())), slog.String("user_id", userID))

		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			MaxAge:   -1,
			Name:     "jwt",
			Value:    "",
		})

		log.Debug("User logged out")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp.OK())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
