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
		RecordId:     "record123",
		FkWorkoutId:  "session123",
		FkExerciseId: 1,
		Reps:         10,
		Weight:       100,
	}

	userIDValue := "user123"
	userID := &userIDValue

	tests := []struct {
		name               string
		userID             string
		reqBody            interface{}
		setupMock          func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:    "Success",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {
				sessionRepo.On("GetSession", userID).Return(&storage.WorkoutSession{
					SessionID: "session123",
				}, nil)
				userRepo.On("GetUserMax", userID, &validRecord.FkExerciseId).Return(&storage.Max{
					MaxWeight: 10,
					Reps:      10,
				}, nil)
				sessionRepo.On("UpdateSession", userID, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:               "InvalidRequest",
			userID:             "user123",
			reqBody:            "xxx",
			setupMock:          func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeBadRequest},
		},
		{
			name:    "ValidationError",
			userID:  "user123",
			reqBody: storage.Record{FkWorkoutId: "", RecordId: ""},
			setupMock: func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeValidationError},
		},
		{
			name:    "SessionNotFound",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {
				sessionRepo.On("GetSession", userID).Return(nil, storage.ErrNoSession)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:    "UpdateSessionError",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {
				sessionRepo.On("GetSession", userID).Return(&storage.WorkoutSession{
					SessionID: "session123",
				}, nil)
				userRepo.On("GetUserMax", userID, &validRecord.FkExerciseId).Return(&storage.Max{
					MaxWeight: 10,
					Reps:      10,
				}, nil)
				sessionRepo.On("UpdateSession", userID, mock.Anything).Return(errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:    "NoMax",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {
				sessionRepo.On("GetSession", userID).Return(&storage.WorkoutSession{
					SessionID: "session123",
				}, nil)
				userRepo.On("GetUserMax", userID, &validRecord.FkExerciseId).Return(nil, storage.ErrNoMaxes)
				sessionRepo.On("UpdateSession", userID, mock.Anything).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:    "CantGetMax",
			userID:  "user123",
			reqBody: validRecord,
			setupMock: func(sessionRepo *mocks.SessionRepository, userRepo *mocks.UserRepository) {
				sessionRepo.On("GetSession", userID).Return(&storage.WorkoutSession{
					SessionID: "session123",
				}, nil)
				userRepo.On("GetUserMax", userID, &validRecord.FkExerciseId).Return(nil, errors.New("some error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := new(mocks.SessionRepository)
			userRepo := new(mocks.UserRepository)

			tt.setupMock(sessionRepo, userRepo)

			reqBody, err := json.Marshal(tt.reqBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/records/add", bytes.NewBuffer(reqBody))
			ctx := context.WithValue(req.Context(), jwt.UserKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler := add.NewAddHandler(logger, sessionRepo, userRepo)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatusCode, rr.Code)

			var actualResp resp.DetailedResponse
			err = json.NewDecoder(rr.Body).Decode(&actualResp)
			require.NoError(t, err)

			require.Equal(t, tt.expectedResponse.Status, actualResp.Status)
		})
	}
}
