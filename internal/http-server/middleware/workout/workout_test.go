package mwworkout_test

import (
	resp "GYMBRO/internal/http-server/handlers/response"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	mwworkout "GYMBRO/internal/http-server/middleware/workout"
	"GYMBRO/internal/lib/jwt"
	"context"
	"encoding/json"
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
	sessionManager := session.NewSessionManager()

	tests := []struct {
		name               string
		setupSession       func()
		userID             string
		expectedStatusCode int
		expectedResponse   resp.DetailedResponse
	}{
		{
			name: "ActiveSessionExists",
			setupSession: func() {
				sessionManager.StartSession("user123", "session123")
			},
			userID:             "user123",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusOK},
		},
		{
			name:               "NoActiveSession",
			setupSession:       func() {},
			userID:             "user456",
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   resp.DetailedResponse{Status: resp.StatusError, Code: resp.CodeNoActiveWorkout},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupSession()

			handler := mwworkout.WithActiveSessionCheck(logger, sessionManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		})
	}
}
