package end_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/http-server/handlers/workouts/end"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/lib/jwt"
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
	sessionManager := session.NewSessionManager()

	tests := []struct {
		name               string
		userID             string
		sessionID          string
		setupMock          func(workoutRepo *mocks.WorkoutRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:      "Success",
			userID:    "user123",
			sessionID: "session123",
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				sessionManager.StartSession("user123", "session123")
				workoutRepo.On("EndWorkout", "session123").Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:      "EndWorkoutError",
			userID:    "user123",
			sessionID: "session123",
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				sessionManager.StartSession("user123", "session123")
				workoutRepo.On("EndWorkout", "session123").Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutRepo := mocks.NewWorkoutRepository(t)
			tt.setupMock(workoutRepo)
			handler := end.NewEndHandler(logger, workoutRepo, sessionManager)

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

			workoutRepo.AssertExpectations(t)
		})
	}
}
