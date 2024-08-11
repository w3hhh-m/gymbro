package register_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/http-server/handlers/users/register"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

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
			reqBody: storage.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(nil, storage.ErrUserNotFound)
				userRepo.On("RegisterNewUser", mock.Anything).Return(func() *string {
					id := "new_user_id"
					return &id
				}(), nil)
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
			name: "ValidationFailed",
			reqBody: storage.User{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeValidationError},
		},
		{
			name: "UserAlreadyExists",
			reqBody: storage.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(&storage.User{
					Email: "test@example.com",
				}, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeUserExists},
		},
		{
			name: "RegistrationError",
			reqBody: storage.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(nil, storage.ErrUserNotFound)
				userRepo.On("RegisterNewUser", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name: "InternalServerError",
			reqBody: storage.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", email).Return(nil, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name: "RestrictedFieldSet",
			reqBody: storage.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				UserId:   "some_id",
				Points:   100,
			},
			setupMock:          func(userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tt.setupMock(userRepo)
			handler := register.NewRegisterHandler(logger, userRepo)
			reqBody, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/users/register", bytes.NewBuffer(reqBody))
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
