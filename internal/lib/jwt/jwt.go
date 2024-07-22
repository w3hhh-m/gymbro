package jwt

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/storage"
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
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

func WithJWTAuth(handlerFunc http.HandlerFunc, userProvider storage.UserProvider, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromRequest(r)
		token, err := validateJWT(tokenString, secret)
		if err != nil {
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("failed to validate token"))
			return
		}
		if !token.Valid {
			render.Status(r, 401)
			render.JSON(w, r, resp.Error("invalid token"))
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		userID := int(claims["uid"].(float64))
		if err != nil {
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("failed to retrieve uid"))
			return
		}
		u, err := userProvider.GetUserByID(userID)
		if err != nil {
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("failed to get user"))
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserKey, u.UserId)
		r = r.WithContext(ctx)
		handlerFunc(w, r)
	}
}

func getTokenFromRequest(r *http.Request) string {
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
