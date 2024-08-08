package update_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"GYMBRO/internal/http-server/handlers/records/update"
	resp "GYMBRO/internal/http-server/handlers/response"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
)

func TestUpdateHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	originalRecord := storage.Record{
		RecordId:     "record1",
		FkWorkoutId:  "session123",
		FkExerciseId: 1,
		Reps:         10,
		Weight:       100,
	}
	updatedRecord := storage.Record{
		RecordId:     "record1",
		FkWorkoutId:  "session123",
		FkExerciseId: 1,
		Reps:         15,
		Weight:       120,
	}

	tests := []struct {
		name               string
		userID             string
		recordID           string
		reqBody            interface{}
		setupMock          func(sessionRepo *mocks.SessionRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:     "Success",
			userID:   "user123",
			recordID: "record1",
			reqBody:  updatedRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(&storage.WorkoutSession{
					SessionID: "session123",
					IsActive:  true,
					Records:   []storage.Record{originalRecord},
					Points:    1000,
				}, nil)
				sessionRepo.On("UpdateSession", "user123", mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:     "RecordNotFound",
			userID:   "user123",
			recordID: "record2",
			reqBody:  updatedRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(&storage.WorkoutSession{
					SessionID: "session123",
					IsActive:  true,
					Records:   []storage.Record{originalRecord},
					Points:    1000,
				}, nil)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeNotFound},
		},
		{
			name:     "GetSessionError",
			userID:   "user123",
			recordID: "record1",
			reqBody:  updatedRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:     "UpdateSessionError",
			userID:   "user123",
			recordID: "record1",
			reqBody:  updatedRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", "user123").Return(&storage.WorkoutSession{
					SessionID: "session123",
					IsActive:  true,
					Records:   []storage.Record{originalRecord},
					Points:    1000,
				}, nil)
				sessionRepo.On("UpdateSession", "user123", mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:               "InvalidRequest",
			userID:             "user123",
			recordID:           "record1",
			reqBody:            "xxx",
			setupMock:          func(sessionRepo *mocks.SessionRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewSessionRepository(t)
			tt.setupMock(sessionRepo)
			handler := update.NewUpdateHandler(logger, sessionRepo)

			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Put("/workouts/update/{recordID}", handler)

			reqBody, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("PUT", "/workouts/update/"+tt.recordID, bytes.NewBuffer(reqBody))
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

			sessionRepo.AssertExpectations(t)
		})
	}
}
