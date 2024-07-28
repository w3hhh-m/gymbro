package jwt

import (
	"GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/storage"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const UserKey contextKey = "uid"

// NewToken generates a new JWT token for a user with a specified duration and secret key
func NewToken(usr storage.User, duration time.Duration, secret string) (string, error) {
	claims := jwt.MapClaims{
		"uid":      usr.UserId,
		"username": usr.Username,
		"exp":      time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// WithJWTAuth middleware checks for a valid JWT token and adds the user ID to the context
func WithJWTAuth(log *slog.Logger, userRepo storage.UserRepository, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "lib.jwt.WithJWTAuth"
			reqID := middleware.GetReqID(r.Context())
			log = log.With(slog.String("op", op), slog.Any("request_id", reqID))

			tokenString := GetTokenFromRequest(r)
			if tokenString == "" {
				log.Info("User is not authenticated")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("You are not authenticated"))
				return
			}
			token, err := validateJWT(tokenString, secret)
			if err != nil {
				log.Error("Failed to validate JWT", "error", err)
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error"))
				return
			}
			if !token.Valid {
				log.Info("Got invalid token")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("Invalid token"))
				return
			}

			claims := token.Claims.(jwt.MapClaims)
			userID := int(claims["uid"].(float64))
			if err != nil {
				log.Error("Failed to extract user ID", "error", err)
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error"))
				return
			}

			u, err := userRepo.GetUserByID(userID)
			if err != nil {
				log.Error("Failed to retrieve user", slog.Any("error", err), slog.Int("uid", userID))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error"))
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, u.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetTokenFromRequest extracts the JWT token from the Authorization header, query parameters, or cookies
func GetTokenFromRequest(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		return authHeader
	}
	if tokenQuery := r.URL.Query().Get("token"); tokenQuery != "" {
		return tokenQuery
	}
	if tokenCookie, err := r.Cookie("jwt"); err == nil {
		return tokenCookie.Value
	}
	return ""
}

// validateJWT parses and validates the JWT token using the provided secret
func validateJWT(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) int {
	if userID, ok := ctx.Value(UserKey).(int); ok {
		return userID
	}
	return -1
}
