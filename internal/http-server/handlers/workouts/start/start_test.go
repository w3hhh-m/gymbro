package start_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/http-server/handlers/workouts/start"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStartHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name               string
		userID             string
		setupMock          func(sessionRepo *mocks.SessionRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:   "Success",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(nil, storage.ErrNoSession)
				sessionRepo.On("CreateSession", mock.AnythingOfType("storage.WorkoutSession")).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:   "CreateSessionError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(nil, storage.ErrNoSession)
				sessionRepo.On("CreateSession", mock.AnythingOfType("storage.WorkoutSession")).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "UserHasActiveSession",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				activeSession := &storage.WorkoutSession{
					SessionID:   "session123",
					UserID:      "user123",
					StartTime:   time.Now(),
					LastUpdated: time.Now(),
					IsActive:    true,
				}
				sessionRepo.On("GetSession", "user123").Return(activeSession, nil)
			},
			expectedStatusCode: http.StatusConflict,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeActiveWorkout},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewSessionRepository(t)
			tt.setupMock(sessionRepo)

			handler := start.NewStartHandler(logger, sessionRepo)

			req := httptest.NewRequest("POST", "/workouts/start", nil)
			ctx := context.WithValue(req.Context(), jwt.UserKey, tt.userID)
			req = req.WithContext(ctx)

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

			sessionRepo.AssertExpectations(t)
		})
	}
}
