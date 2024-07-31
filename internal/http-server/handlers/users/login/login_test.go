package login

import (
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoginHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	secret := "some secret"

	tests := []struct {
		name           string
		requestBody    Request
		mockGetUser    func(userRepo *mocks.UserRepository)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: Request{
				Email:    "test@example.com",
				Password: "password",
			},
			mockGetUser: func(userRepo *mocks.UserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				user := storage.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				userRepo.On("GetUserByEmail", "test@example.com").Once().Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "DecodeError",
			requestBody: Request{
				Email:    "test@example.com",
				Password: "password",
			},
			mockGetUser:    func(userRepo *mocks.UserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ValidationError",
			requestBody: Request{
				Email:    "invalid-email",
				Password: "",
			},
			mockGetUser:    func(userRepo *mocks.UserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "UserNotFound",
			requestBody: Request{
				Email:    "xxx@example.com",
				Password: "password",
			},
			mockGetUser: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", "xxx@example.com").Once().Return(storage.User{}, storage.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "UserGetError",
			requestBody: Request{
				Email:    "test@example.com",
				Password: "password",
			},
			mockGetUser: func(userRepo *mocks.UserRepository) {
				userRepo.On("GetUserByEmail", "test@example.com").Once().Return(storage.User{}, errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "InvalidPassword",
			requestBody: Request{
				Email:    "test@example.com",
				Password: "blablabla",
			},
			mockGetUser: func(userRepo *mocks.UserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				user := storage.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				userRepo.On("GetUserByEmail", "test@example.com").Once().Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		/*{ // Need to make jwt mock to test I think
			name: "GenerateTokenError",
			requestBody: Request{
				Email:    "test@example.com",
				Password: "password",
			},
			mockGetUser: func(userRepo *mocks.UserRepository) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				user := storage.User{
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				userRepo.On("GetUserByEmail", "test@example.com").Once().Return(user, nil)
			},
			expectedStatus: http.StatusInternalServerError,
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewUserRepository(t)
			tt.mockGetUser(userRepo)

			handler := NewLoginHandler(logger, userRepo, secret)
			r := chi.NewRouter()
			r.Post("/login", handler)

			var body []byte
			var err error
			if tt.name == "DecodeError" {
				body = []byte("blablabla")
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}
			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.name == "Success" { // check cookie set
				cookie := rr.Result().Cookies()
				require.Len(t, cookie, 1)
				assert.Equal(t, "jwt", cookie[0].Name)
				assert.NotEmpty(t, cookie[0].Value)
			}
			userRepo.AssertExpectations(t)
		})
	}
}

/*
	func FuzzLoginHandler(f *testing.F) {
		f.Add(`{"email": "test@example.com", "password": "password"}`)

		f.Fuzz(func(t *testing.T, input string) {
			fmt.Printf("Testing input: %s\n", input)
			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

			userRepo := mocks.NewUserRepository(t)
			userRepo.On("GetUserByEmail", mock.Anything).Return(func(email string) storage.User {
				if email == "test@example.com" {
					return storage.User{
						Email:    "test@example.com",
						Password: "$2a$10$7aIb9joXhtnCQZg4j.Qj6u.O/pDj4V5E3zSkHX9AF7rV0MPuquoaW",
					}
				}
				return storage.User{}
			}, func(email string) error {
				if email == "test@example.com" {
					return nil
				}
				return storage.ErrUserNotFound
			}).Maybe()
			secret := "secret"

			handler := NewLoginHandler(logger, userRepo, secret)
			r := chi.NewRouter()
			r.Post("/login", handler)

			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte(input)))
			require.NoError(t, err)

			// Add middleware for request ID
			//req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "request-id"))

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Check that the handler does not panic and responds with a status code
			require.NotEqual(t, http.StatusInternalServerError, rr.Code)
		})
	}
*/
func FuzzLoginHandler(f *testing.F) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	userRepo := mocks.NewUserRepository(f)
	handler := NewLoginHandler(logger, userRepo, "secret")
	r := chi.NewRouter()
	r.Post("/login", handler)

	file, err := os.OpenFile("fuzz_interesting_inputs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		f.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	f.Add("test@example.com", "password")

	f.Fuzz(func(t *testing.T, email, password string) {
		if email == "" || password == "" {
			t.Skip()
		}

		userRepo.On("GetUserByEmail", email).Return(storage.User{
			UserId:   1,
			Email:    email,
			Password: password,
		}, nil).Maybe()

		reqBody, err := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code == http.StatusInternalServerError {
			_, err := fmt.Fprintf(file, "Interesting input: email=%s, password=%s\n", email, password)
			require.NoError(t, err)
		}

		userRepo.AssertExpectations(t)
	})
}
