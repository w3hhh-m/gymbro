package jwt

import (
	resp "GYMBRO/internal/http-server/handlers/response"
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

func NewToken(usr storage.User, duration time.Duration, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = usr.UserId
	claims["username"] = usr.Username
	claims["exp"] = time.Now().Add(duration).Unix()
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func WithJWTAuth(handlerFunc http.HandlerFunc, log *slog.Logger, userProvider storage.UserProvider, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "lib.jwt.WithJWTAuth"
		log = log.With(slog.String("op", op), slog.Any("request_id", middleware.GetReqID(r.Context())))
		tokenString := GetTokenFromRequest(r)
		token, err := validateJWT(tokenString, secret)
		if err != nil {
			log.Error("Failed to validate JWT", "error", err)
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}
		if !token.Valid {
			log.Info("Got invalid token")
			render.Status(r, 401)
			render.JSON(w, r, resp.Error("invalid token"))
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		userID := int(claims["uid"].(float64))
		if err != nil {
			log.Error("Failed to get user ID", "error", err)
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}
		u, err := userProvider.GetUserByID(userID)
		if err != nil {
			log.Error("Failed to get user", "error", err)
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("Internal error"))
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.UserId)
		r = r.WithContext(ctx)
		handlerFunc(w, r)
	}
}

func GetTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")
	tokenQuery := r.URL.Query().Get("token")
	tokenCookie, err := r.Cookie("jwt")
	if err != nil {
		tokenCookie = &http.Cookie{}
	}
	if tokenAuth != "" {
		return tokenAuth
	}
	if tokenQuery != "" {
		return tokenQuery
	}
	if tokenCookie.Value != "" {
		return tokenCookie.Value
	}
	return ""
}

func validateJWT(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return -1
	}
	return userID
}
