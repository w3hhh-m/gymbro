package mwjwt_test

import (
	"GYMBRO/internal/config"
	resp "GYMBRO/internal/http-server/handlers/response"
	mwjwt "GYMBRO/internal/http-server/middleware/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"encoding/json"
	"errors"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithJWTAuth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	cfg := &config.Config{JWTCfg: config.JWTCfg{SecretKey: "test"}}

	userIDValue := "user123"
	userID := &userIDValue

	tests := []struct {
		name               string
		token              string
		setupMock          func(userRepo *mocks.UserRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name: "Success",
			token: func() string {
				claims := jwt.MapClaims{
					"uid": "user123",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				x, _ := token.SignedString([]byte(cfg.SecretKey))
				return x
			}(),
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByID", userID).Return(&storage.User{UserId: "user123"}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:               "MissingToken",
			token:              "",
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeUnauthorized},
		},
		{
			name: "InvalidToken",
			token: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"uid": "user123",
				})
				tokenString, _ := token.SignedString([]byte("wrong_secret_key"))
				return tokenString
			}(),
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name: "UserNotFound",
			token: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"uid": "user123",
				})
				tokenString, _ := token.SignedString([]byte(cfg.SecretKey))
				return tokenString
			}(),
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByID", userID).Return(nil, errors.New("user not found"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tt.setupMock(userRepo)

			handler := mwjwt.WithJWTAuth(logger, userRepo, cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				render.Status(r, http.StatusOK)
				render.JSON(w, r, resp.OK())
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatusCode, rr.Code)

			var response resp.DetailedResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			require.Equal(t, tt.expectedResponse.Status, response.Status)
			if tt.expectedResponse.Code != "" {
				require.Equal(t, tt.expectedResponse.Code, response.Code)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
