package end_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/http-server/handlers/workouts/end"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"context"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEndHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name               string
		userID             string
		setupMock          func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:   "Success",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					IsActive:  true,
				}
				sessionRepo.On("GetSession", "user123").Return(session, nil)
				workoutRepo.On("SaveWorkout", session).Return(nil)
				sessionRepo.On("DeleteSession", "user123").Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:   "SessionNotFound",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository) {
				sessionRepo.On("GetSession", "user123").Return(nil, storage.ErrNoSession)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "SaveWorkoutError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					IsActive:  true,
				}
				sessionRepo.On("GetSession", "user123").Return(session, nil)
				workoutRepo.On("SaveWorkout", session).Return(errors.New("save workout error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "DeleteSessionError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					IsActive:  true,
				}
				sessionRepo.On("GetSession", "user123").Return(session, nil)
				workoutRepo.On("SaveWorkout", session).Return(nil)
				sessionRepo.On("DeleteSession", "user123").Return(errors.New("delete session error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewSessionRepository(t)
			workoutRepo := mocks.NewWorkoutRepository(t)
			tt.setupMock(sessionRepo, workoutRepo)

			handler := end.NewEndHandler(logger, sessionRepo, workoutRepo)

			req := httptest.NewRequest("POST", "/workouts/end", nil)
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
			workoutRepo.AssertExpectations(t)
		})
	}
}
