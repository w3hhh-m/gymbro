package mwjwt

import (
	"GYMBRO/internal/config"
	resp "GYMBRO/internal/http-server/handlers/response"
	jwtlib "GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"net/http"
)

// WithJWTAuth middleware checks for a valid JWT token and adds the user ID to the context.
func WithJWTAuth(log *slog.Logger, userRepo storage.UserRepository, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.mwjwt.WithJWTAuth"
			reqID := middleware.GetReqID(r.Context())
			log = log.With(slog.String("op", op), slog.Any("request_id", reqID))

			// Retrieve JWT token from the request
			tokenString := jwtlib.GetTokenFromRequest(r)
			if tokenString == "" {
				log.Debug("User is not authenticated")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("You are not authenticated", resp.CodeUnauthorized, "You need to login first"))
				return
			}

			// Validate the JWT token
			token, err := jwtlib.ValidateJWT(tokenString, cfg.SecretKey)
			if err != nil {
				log.Warn("Failed to validate JWT", slog.Any("error", err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please logout and login again, or try again later"))
				return
			}
			if !token.Valid {
				log.Warn("Got invalid token", slog.Any("token", token))
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("Invalid token", resp.CodeInternalError, "Please logout and login again, or try again later"))
				return
			}

			// Extract claims from the token
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Warn("Failed to parse JWT claims")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please logout and login again, or try again later"))
				return
			}

			// Retrieve user from the database using the user ID from claims
			userID := claims["uid"].(string)
			u, err := userRepo.GetUserByID(userID)
			if err != nil {
				log.Error("Failed to get user", slog.Any("error", err), slog.String("uid", userID))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal error", resp.CodeInternalError, "Please try again later"))
				return
			}

			// Add user ID to context and proceed to the next handler
			ctx := context.WithValue(r.Context(), jwtlib.UserKey, u.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
