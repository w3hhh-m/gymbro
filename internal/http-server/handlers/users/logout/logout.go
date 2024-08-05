package logout

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// NewLogoutHandler returns a handler function to initiate user logout.
func NewLogoutHandler(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.logout.New"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))

		// Retrieve the user ID from the context.
		uid := jwt.GetUserIDFromContext(r.Context())

		// Delete the JWT cookie by setting its MaxAge to -1.
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			Path:     "/",
			MaxAge:   -1, // Delete the cookie.
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: "",
		})

		// Log the user logout event.
		log.Debug("User logged out", slog.String("uid", uid))

		// Send a successful logout response.
		render.JSON(w, r, resp.OK())

		// Redirect to the home page.
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
