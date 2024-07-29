package save

import (
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/mocks"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSaveNewHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	tests := []struct {
		name           string
		record         storage.Record
		mockSave       func(recordRepo *mocks.RecordRepository)
		userID         int
		expectedStatus int
	}{
		{
			name: "Success",
			record: storage.Record{
				FkExerciseId: 1,
				Reps:         100,
				Weight:       100,
				CreatedAt:    time.Now(),
			},
			mockSave: func(recordRepo *mocks.RecordRepository) {
				recordRepo.On("SaveRecord", mock.Anything).Once().Return(1, nil)
			},
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name: "SuccessWithZeroCreatedAt",
			record: storage.Record{
				FkExerciseId: 1,
				Reps:         100,
				Weight:       100,
				CreatedAt:    time.Time{},
			},
			mockSave: func(recordRepo *mocks.RecordRepository) {
				recordRepo.On("SaveRecord", mock.Anything).Once().Return(1, nil)
			},
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name: "DecodeError",
			record: storage.Record{
				FkExerciseId: 1,
				Reps:         100,
				Weight:       100,
				CreatedAt:    time.Now(),
			},
			mockSave:       func(recordRepo *mocks.RecordRepository) {},
			userID:         1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "ValidationError",
			record: storage.Record{
				FkExerciseId: 0,
				Reps:         -2,
				Weight:       100,
				CreatedAt:    time.Now(),
			},
			mockSave:       func(recordRepo *mocks.RecordRepository) {},
			userID:         1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "SaveError",
			record: storage.Record{
				FkExerciseId: 1,
				Reps:         100,
				Weight:       100,
				CreatedAt:    time.Now(),
			},
			mockSave: func(recordRepo *mocks.RecordRepository) {
				recordRepo.On("SaveRecord", mock.Anything).Once().Return(0, errors.New("internal server error"))
			},
			userID:         1,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recordRepo := mocks.NewRecordRepository(t)
			tt.mockSave(recordRepo)

			handler := NewSaveHandler(logger, recordRepo)
			r := chi.NewRouter()
			r.Post("/records", handler)

			var body []byte
			var err error
			if tt.name == "DecodeError" {
				body = []byte("invalid json blablabla")
			} else {
				body, err = json.Marshal(tt.record)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
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
