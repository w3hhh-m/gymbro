package jwt

import (
	"GYMBRO/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

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
