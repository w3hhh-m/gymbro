package jwt

import (
	"GYMBRO/internal/storage"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
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

// ValidateJWT parses and validates the JWT token using the provided secret
func ValidateJWT(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method matches the expected HMAC method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserKey).(string); ok {
		return userID
	}
	return ""
}
