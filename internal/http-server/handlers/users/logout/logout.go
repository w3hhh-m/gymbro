package logout

import (
	"net/http"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			HttpOnly: true,
			MaxAge:   -1, // Delete the cookie.
			SameSite: http.SameSiteLaxMode,
			// Uncomment below for HTTPS:
			// Secure: true,
			Name:  "jwt",
			Value: "",
		})
	}
}
