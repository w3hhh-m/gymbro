package register

import (
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name           string
		user           storage.User
		mockRegister   func(userRepo *mocks.UserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			user: storage.User{
				Username:    "user",
				Email:       "test@example.com",
				Password:    "password",
				Phone:       "70000000000",
				DateOfBirth: time.Now(),
			},
			mockRegister: func(userRepo *mocks.UserRepository) {
				userRepo.On("RegisterNewUser", mock.Anything).Once().Return(1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"OK","id":1}`,
		},
		{
			name:           "DecodeError",
			user:           storage.User{},
			mockRegister:   func(userRepo *mocks.UserRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"Failed to decode request"}`,
		},
		{
			name: "InvalidData",
			user: storage.User{
				Username: "user",
			},
			mockRegister:   func(userRepo *mocks.UserRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"field email is a required field, field phone is a required field, field password is a required field, field dateofbirth is a required field"}`,
		},
		{
			name: "UserExists",
			user: storage.User{
				Username:    "user",
				Email:       "test@example.com",
				Password:    "password",
				Phone:       "70000000000",
				DateOfBirth: time.Now(),
			},
			mockRegister: func(userRepo *mocks.UserRepository) {
				userRepo.On("RegisterNewUser", mock.Anything).Once().Return(0, storage.ErrUserExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"User already exists"}`,
		},
		{
			name: "UserRegisterError",
			user: storage.User{
				Username:    "user",
				Email:       "test@example.com",
				Password:    "password",
				Phone:       "70000000000",
				DateOfBirth: time.Now(),
			},
			mockRegister: func(userRepo *mocks.UserRepository) {
				userRepo.On("RegisterNewUser", mock.Anything).Once().Return(0, errors.New("test error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":"ERROR","error":"Internal error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tt.mockRegister(userRepo)

			handler := NewRegisterHandler(logger, userRepo)
			r := chi.NewRouter()
			r.Post("/users", handler)

			var body []byte
			var err error
			if tt.name == "DecodeError" {
				body = []byte("blablabla")
			} else {
				body, err = json.Marshal(tt.user)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			require.JSONEq(t, tt.expectedBody, rr.Body.String())

			userRepo.AssertExpectations(t)
		})
	}
}
