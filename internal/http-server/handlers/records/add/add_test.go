package add_test

import (
	"GYMBRO/internal/http-server/handlers/records/add"
	resp "GYMBRO/internal/http-server/handlers/response"
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

	validRecord := storage.Record{
		FkWorkoutId:  "session123",
		FkExerciseId: 1,
		Reps:         10,
		Weight:       100,
	}

	tests := []struct {
		name               string
		userID             string
		reqBody            interface{}
		setupMock          func(sessionRepo *mocks.SessionRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:    "Success",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(&storage.WorkoutSession{
					SessionID: "session123",
					IsActive:  true,
				}, nil)
				sessionRepo.On("UpdateSession", "user123", mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:               "InvalidRequest",
			userID:             "user123",
			reqBody:            "xxx",
			setupMock:          func(sessionRepo *mocks.SessionRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
		{
			name:    "ValidationError",
			userID:  "user123",
			reqBody: storage.Record{FkWorkoutId: "", RecordId: ""}, // Invalid record
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(&storage.WorkoutSession{
					SessionID: "session123",
					IsActive:  true,
				}, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeValidationError},
		},
		{
			name:    "SessionNotFound",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(nil, storage.ErrNoSession)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:    "UpdateSessionError",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(&storage.WorkoutSession{
					SessionID: "session123",
					IsActive:  true,
				}, nil)
				sessionRepo.On("UpdateSession", "user123", mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewSessionRepository(t)
			tt.setupMock(sessionRepo)
			handler := add.NewAddHandler(logger, sessionRepo)

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

			sessionRepo.AssertExpectations(t)
		})
	}
}
