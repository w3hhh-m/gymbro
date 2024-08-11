package login_test

import (
	"GYMBRO/internal/config"
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/http-server/handlers/users/login"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginHandler(t *testing.T) {
	cfg := &config.Config{JWTCfg: config.JWTCfg{SecretKey: "test", JWTLifetime: time.Hour}}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	emailValue := "test@example.com"
	email := &emailValue

	tests := []struct {
		name               string
		reqBody            interface{}
		setupMock          func(userRepo *mocks.UserRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name: "Success",
			reqBody: login.Request{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(&storage.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:               "InvalidRequest",
			reqBody:            "invalid-json",
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
		{
			name: "ValidationError",
			reqBody: login.Request{
				Email:    "invalid-email",
				Password: "",
			},
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeValidationError},
		},
		{
			name: "UserNotFound",
			reqBody: login.Request{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(nil, storage.ErrUserNotFound)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
		{
			name: "InvalidCredentials",
			reqBody: login.Request{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(&storage.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
		{
			name: "InternalServerError",
			reqBody: login.Request{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(nil, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tt.setupMock(userRepo)
			handler := login.NewLoginHandler(logger, userRepo, cfg)
			reqBody, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/users/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

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
