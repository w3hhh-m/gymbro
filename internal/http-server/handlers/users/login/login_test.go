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
	cfg := &config.Config{
		JWTLifetime: 1 * time.Hour,
		SecretKey:   "test_secret_key",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

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
				userRepo.On("GetUserByEmail", "test@example.com").Return(&storage.User{
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
			name: "UserNotFound",
			reqBody: login.Request{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", "test@example.com").Return(nil, storage.ErrUserNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeNotFound},
		},
		{
			name: "InvalidCredentials",
			reqBody: login.Request{
				Email:    "test@example.com",
				Password: "wrong",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", "test@example.com").Return(&storage.User{
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
				userRepo.On("GetUserByEmail", "test@example.com").Return(nil, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name: "InvalidValidation",
			reqBody: map[string]string{
				"email": "invalid-email",
			},
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeValidationError},
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
			if tt.expectedResponse.Error != "" {
				require.Contains(t, response.Error, tt.expectedResponse.Error)
			}

			userRepo.AssertExpectations(t)
		})
	}
}
