package logout

import (
	"GYMBRO/internal/lib/jwt"
	"bytes"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogoutHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name           string
		userID         int
		expectedStatus int
	}{
		{
			name:           "SuccessfulLogout",
			userID:         1,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewLogoutHandler(logger)
			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Post("/logout", handler)

			req, err := http.NewRequest(http.MethodPost, "/logout", bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			ctx := req.Context()
			ctx = context.WithValue(ctx, jwt.UserKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			cookie := rr.Result().Cookies()
			require.Len(t, cookie, 1)
			assert.Equal(t, "jwt", cookie[0].Name)
			assert.Equal(t, "", cookie[0].Value)
			assert.Equal(t, -1, cookie[0].MaxAge)
		})
	}
}
