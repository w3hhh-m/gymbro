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
	"github.com/stretchr/testify/mock"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	userIDValue := "user123"
	userID := &userIDValue

	tests := []struct {
		name               string
		userID             string
		setupMock          func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository)
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name:   "Success",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					Records: []storage.Record{
						{FkExerciseId: 1, Weight: 100, Reps: 10},
					},
				}
				sessionRepo.On("GetSession", userID).Return(session, nil)
				userRepo.On("GetUserMaxes", userID).Return([]*storage.Max{}, nil)
				userRepo.On("SetUserMax", userID, mock.Anything).Return(nil)
				workoutRepo.On("SaveWorkout", session).Return(nil)
				sessionRepo.On("DeleteSession", userID).Return(nil)
				userRepo.On("ChangeStatus", userID, false).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:   "SessionNotFound",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				sessionRepo.On("GetSession", userID).Return(nil, storage.ErrNoSession)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "SaveWorkoutError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
				}
				userRepo.On("GetUserMaxes", userID).Return([]*storage.Max{}, nil)
				sessionRepo.On("GetSession", userID).Return(session, nil)
				workoutRepo.On("SaveWorkout", session).Return(errors.New("save workout error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "DeleteSessionError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
				}
				userRepo.On("GetUserMaxes", userID).Return([]*storage.Max{}, nil)
				sessionRepo.On("GetSession", userID).Return(session, nil)
				workoutRepo.On("SaveWorkout", session).Return(nil)
				sessionRepo.On("DeleteSession", userID).Return(errors.New("delete session error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "UserStatusError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
				}
				userRepo.On("GetUserMaxes", userID).Return([]*storage.Max{}, nil)
				sessionRepo.On("GetSession", userID).Return(session, nil)
				workoutRepo.On("SaveWorkout", session).Return(nil)
				sessionRepo.On("DeleteSession", userID).Return(nil)
				userRepo.On("ChangeStatus", userID, false).Return(errors.New("user status error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "GetUserMaxesError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					Records: []storage.Record{
						{FkExerciseId: 1, Weight: 100, Reps: 10},
					},
				}
				sessionRepo.On("GetSession", userID).Return(session, nil)
				userRepo.On("GetUserMaxes", userID).Return(nil, errors.New("get user maxes error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
		{
			name:   "NewRecordSetSuccess",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					Records: []storage.Record{
						{FkExerciseId: 1, Weight: 100, Reps: 10},
					},
				}
				userMax := &storage.Max{ExerciseId: 1, MaxWeight: 90, Reps: 8}
				sessionRepo.On("GetSession", userID).Return(session, nil)
				userRepo.On("GetUserMaxes", userID).Return([]*storage.Max{userMax}, nil)
				userRepo.On("SetUserMax", userID, mock.AnythingOfType("*storage.Max")).Return(nil)
				workoutRepo.On("SaveWorkout", session).Return(nil)
				sessionRepo.On("DeleteSession", userID).Return(nil)
				userRepo.On("ChangeStatus", userID, false).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:   "SetUserMaxError",
			userID: "user123",
			setupMock: func(sessionRepo *mocks.SessionRepository, workoutRepo *mocks.WorkoutRepository, userRepo *mocks.UserRepository) {
				session := &storage.WorkoutSession{
					UserID:    "user123",
					SessionID: "session123",
					Records: []storage.Record{
						{FkExerciseId: 1, Weight: 100, Reps: 10},
					},
				}
				userMax := &storage.Max{ExerciseId: 1, MaxWeight: 90, Reps: 8}
				sessionRepo.On("GetSession", userID).Return(session, nil)
				userRepo.On("GetUserMaxes", userID).Return([]*storage.Max{userMax}, nil)
				userRepo.On("SetUserMax", userID, mock.AnythingOfType("*storage.Max")).Return(errors.New("set user max error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewSessionRepository(t)
			workoutRepo := mocks.NewWorkoutRepository(t)
			userRepo := mocks.NewUserRepository(t)
			tt.setupMock(sessionRepo, workoutRepo, userRepo)

			handler := end.NewEndHandler(logger, sessionRepo, workoutRepo, userRepo)

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
			userRepo.AssertExpectations(t)
		})
	}
}
