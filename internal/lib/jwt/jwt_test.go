package jwt

import (
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewToken(t *testing.T) {
	//There won`t be any validation error bc data is validating before calling NewToken
	user := storage.User{
		UserId:   1,
		Username: "test",
	}

	secret := "secret"
	duration := time.Hour

	tokenString, err := NewToken(user, duration, secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)
}

func TestGetTokenFromRequest(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		expectedToken string
	}{
		{
			name: "Authorization Header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "Bearer test")
				return req
			},
			expectedToken: "Bearer test",
		},
		{
			name: "Query Parameter",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/?token=test", nil)
				return req
			},
			expectedToken: "test",
		},
		{
			name: "Cookie",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(&http.Cookie{Name: "jwt", Value: "test"})
				return req
			},
			expectedToken: "test",
		},
		{
			name: "No Token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				return req
			},
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			token := GetTokenFromRequest(req)
			require.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "secret"

	test := []struct {
		name     string
		claims   jwt.MapClaims
		method   jwt.SigningMethod
		expected bool
	}{
		{
			name: "Valid JWT",
			claims: jwt.MapClaims{
				"uid":      1,
				"username": "test",
				"exp":      time.Now().Add(time.Hour).Unix(),
			},
			method:   jwt.SigningMethodHS256,
			expected: true,
		},
		{
			name: "Invalid Method",
			claims: jwt.MapClaims{
				"uid":      1,
				"username": "test",
				"exp":      time.Now().Add(time.Hour).Unix(),
			},
			method:   jwt.SigningMethodNone,
			expected: true,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Invalid Method" {
				_, err := jwt.NewWithClaims(tt.method, tt.claims).SignedString([]byte(secret))
				require.Error(t, err)
			} else {
				tokenString, err := jwt.NewWithClaims(tt.method, tt.claims).SignedString([]byte(secret))
				require.NoError(t, err)
				require.NotEmpty(t, tokenString)

				if tt.name == "Invalid Method" {

				}
				token, err := validateJWT(tokenString, secret)
				require.NoError(t, err)
				require.Equal(t, token.Valid, tt.expected)
			}
		})
	}
}

func TestWithJWTAuth(t *testing.T) {
	secret := "secret"
	user := storage.User{
		UserId:   1,
		Username: "test",
	}

	userRepo := mocks.NewUserRepository(t)
	userRepo.On("GetUserByID", user.UserId).Once().Return(user, nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tokenString, err := NewToken(user, time.Hour, secret)
	require.NoError(t, err)
	require.NotEmpty(t, tokenString)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
	}{
		{
			name: "Valid Token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", tokenString)
				return req
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid Token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Authorization", "invalid")
				return req
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "No Token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				return req
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := WithJWTAuth(logger, userRepo, secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := tt.setupRequest()
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
