package get

import (
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name           string
		recordID       string
		mockGet        func(recordRepo *mocks.RecordRepository)
		userID         int
		expectedStatus int
	}{
		{
			name:     "Success",
			recordID: "1",
			userID:   1,
			mockGet: func(recordRepo *mocks.RecordRepository) {
				record := storage.Record{
					RecordId:     1,
					FkExerciseId: 1,
					FkUserId:     1,
					Reps:         100,
					Weight:       100,
					CreatedAt:    time.Now(),
				}
				recordRepo.On("GetRecord", 1).Once().Return(record, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "RecordNotFound",
			recordID: "1",
			userID:   1,
			mockGet: func(recordRepo *mocks.RecordRepository) {
				recordRepo.On("GetRecord", 1).Once().Return(storage.Record{}, storage.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "NonNumericID",
			recordID:       "abc",
			userID:         1,
			mockGet:        func(recordRepo *mocks.RecordRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Unauthorized",
			recordID: "1",
			userID:   1,
			mockGet: func(recordRepo *mocks.RecordRepository) {
				record := storage.Record{
					RecordId:     1,
					FkExerciseId: 1,
					FkUserId:     3,
					Reps:         100,
					Weight:       100,
					CreatedAt:    time.Now(),
				}
				recordRepo.On("GetRecord", 1).Once().Return(record, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "GetError",
			recordID: "1",
			userID:   1,
			mockGet: func(recordRepo *mocks.RecordRepository) {
				recordRepo.On("GetRecord", 1).Once().Return(storage.Record{}, errors.New("could not get record"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordRepo := mocks.NewRecordRepository(t)
			tt.mockGet(recordRepo)

			handler := NewGetHandler(logger, recordRepo)
			r := chi.NewRouter()
			r.Use(middleware.URLFormat)
			r.Get("/records/{id}", handler)

			req, err := http.NewRequest(http.MethodGet, "/records/"+tt.recordID, nil)
			require.NoError(t, err)

			ctx := req.Context()
			ctx = context.WithValue(ctx, jwt.UserKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)
			recordRepo.AssertExpectations(t)
		})
	}
}
