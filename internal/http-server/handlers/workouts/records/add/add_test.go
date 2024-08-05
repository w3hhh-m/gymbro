package add_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/http-server/handlers/workouts/records/add"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
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

func TestAddHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	sessionManager := session.NewSessionManager()
	sessionManager.StartSession("user123", "session123")
	validRecord := storage.Record{
		FkWorkoutId:  "session123",
		FkExerciseId: 1,
		//RecordId:    "record123",
		Reps:   10,
		Weight: 100,
	}

	tests := []struct {
		name               string
		userID             string
		sessionID          string
		reqBody            interface{}
		setupMock          func(workoutRepo *mocks.WorkoutRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:      "Success",
			userID:    "user123",
			sessionID: "session123",
			reqBody:   validRecord,
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				workoutRepo.On("AddRecord", mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:               "InvalidRequest",
			sessionID:          "session123",
			userID:             "user123",
			reqBody:            "xxx",
			setupMock:          func(workoutRepo *mocks.WorkoutRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
		{
			name:      "ValidationError",
			userID:    "user123",
			sessionID: "session123",
			reqBody:   storage.Record{FkWorkoutId: "", RecordId: ""}, // Invalid record
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeValidationError},
		},
		{
			name:      "AddRecordError",
			userID:    "user123",
			sessionID: "session123",
			reqBody:   validRecord,
			setupMock: func(workoutRepo *mocks.WorkoutRepository) {
				workoutRepo.On("AddRecord", mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workoutRepo := mocks.NewWorkoutRepository(t)
			tt.setupMock(workoutRepo)
			handler := add.NewAddHandler(logger, workoutRepo, sessionManager)

			reqBody, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/workouts/add", bytes.NewBuffer(reqBody))
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
