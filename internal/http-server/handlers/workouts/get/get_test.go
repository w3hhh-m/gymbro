package getwo_test

import (
	"GYMBRO/internal/http-server/handlers/workouts/get"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
)

func TestGetWorkoutHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name               string
		userID             string
		workoutID          string
		setupMock          func(woRepo *mocks.WorkoutRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:      "Success",
			userID:    "user123",
			workoutID: "workout123",
			setupMock: func(woRepo *mocks.WorkoutRepository) {
				woRepo.On("GetWorkout", "workout123").Return(&storage.WorkoutWithRecords{WorkoutID: "workout123", UserID: "user123"}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK, Code: resp.StatusOK},
		},
		{
			name:      "WorkoutNotFound",
			userID:    "user123",
			workoutID: "workout123",
			setupMock: func(woRepo *mocks.WorkoutRepository) {
				woRepo.On("GetWorkout", "workout123").Return(nil, storage.ErrWorkoutNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeNotFound},
		},
		{
			name:      "FailedToRetrieveWorkout",
			userID:    "user123",
			workoutID: "workout123",
			setupMock: func(woRepo *mocks.WorkoutRepository) {
				woRepo.On("GetWorkout", "workout123").Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:      "ForbiddenAccess",
			userID:    "user456",
			workoutID: "workout123",
			setupMock: func(woRepo *mocks.WorkoutRepository) {
				woRepo.On("GetWorkout", "workout123").Return(&storage.WorkoutWithRecords{WorkoutID: "workout123", UserID: "user123"}, nil)
			},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeForbidden},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			woRepo := mocks.NewWorkoutRepository(t)
			tt.setupMock(woRepo)

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Get("/workouts/{workoutID}", getwo.NewGetWorkoutHandler(logger, woRepo))

			req := httptest.NewRequest("GET", "/workouts/"+tt.workoutID, nil)
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), jwt.UserKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatusCode, rr.Code)

			var response resp.DetailedResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			require.Equal(t, tt.expectedResponse.Status, response.Status)
			if tt.expectedResponse.Code != "" {
				require.Equal(t, tt.expectedResponse.Code, response.Code)
			}
			if tt.expectedResponse.Data != nil {
				expectedData, _ := json.Marshal(tt.expectedResponse.Data)
				require.JSONEq(t, string(expectedData), string(rr.Body.Bytes()))
			}

			woRepo.AssertExpectations(t)
		})
	}
}
