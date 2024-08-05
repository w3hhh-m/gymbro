package start_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/http-server/handlers/workouts/start"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage/mocks"
	"bytes"
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
)

func TestStartHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	sessionManager := session.NewSessionManager()

	tests := []struct {
		name               string
		userID             string
		setupMock          func(workoutRepo *mocks.WorkoutRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:   "Success",
			userID: "user123",
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				workoutRepo.On("CreateWorkout", mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:   "CreateWorkoutError",
			userID: "user1234",
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				workoutRepo.On("CreateWorkout", mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "UserHasActiveSession",
			userID: "user123",
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				sessionManager.StartSession("user123", "session123")
			},
			expectedStatusCode: http.StatusConflict,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeActiveWorkout},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutRepo := mocks.NewWorkoutRepository(t)
			tt.setupMock(workoutRepo)
			handler := start.NewStartHandler(logger, workoutRepo, sessionManager)

			reqBody, _ := json.Marshal("")
			req := httptest.NewRequest("POST", "/workouts/start", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
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

			workoutRepo.AssertExpectations(t)
		})
	}
}
