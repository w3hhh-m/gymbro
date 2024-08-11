package mwworkout_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	mwworkout "GYMBRO/internal/http-server/middleware/workout"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithActiveSessionCheck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	userIDValue := "user123"
	extraUserIDValue := "user456"
	userID := &userIDValue
	extraUserID := &extraUserIDValue

	tests := []struct {
		name               string
		setupMock          func(sessionRepo *mocks.SessionRepository)
		userID             string
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name: "ActiveSessionExists",
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				activeSession := &storage.WorkoutSession{
					SessionID: "session123",
					UserID:    "user123",
				}
				sessionRepo.On("GetSession", userID).Return(activeSession, nil)
			},
			userID:             "user123",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name: "NoActiveSession",
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", extraUserID).Return(nil, storage.ErrNoSession)
			},
			userID:             "user456",
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeNoActiveWorkout},
		},
		{
			name: "SessionRepoError",
			setupMock: func(sessionRepo *mocks.SessionRepository) {
				sessionRepo.On("GetSession", extraUserID).Return(nil, errors.New("db error"))
			},
			userID:             "user456",
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeInternalError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := mocks.NewSessionRepository(t)
			tt.setupMock(sessionRepo)

			handler := mwworkout.WithActiveSessionCheck(logger, sessionRepo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				render.Status(r, http.StatusOK)
				render.JSON(w, r, resp.OK())
			}))

			req := httptest.NewRequest("GET", "/", nil)
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
