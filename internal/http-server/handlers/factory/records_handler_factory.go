package factory

import (
	"GYMBRO/internal/http-server/handlers/records/add"
	"GYMBRO/internal/http-server/handlers/records/delete"
	"GYMBRO/internal/http-server/handlers/records/update"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

// RecordsHandlerFactory defines the interface for creating record-related handlers.
type RecordsHandlerFactory interface {
	CreateAddHandler() http.HandlerFunc
	CreateDeleteHandler() http.HandlerFunc
	CreateUpdateHandler() http.HandlerFunc
}

// RecordHandlerFactory implements the RecordsHandlerFactory interface.
type RecordHandlerFactory struct {
	log   *slog.Logger
	srepo storage.SessionRepository
}

// NewRecordHandlerFactory creates a new instance of RecordHandlerFactory.
func NewRecordHandlerFactory(log *slog.Logger, srepo storage.SessionRepository) *RecordHandlerFactory {
	return &RecordHandlerFactory{
		log:   log,
		srepo: srepo,
	}
}

func (f *RecordHandlerFactory) CreateAddHandler() http.HandlerFunc {
	return add.NewAddHandler(f.log, f.srepo)
}

func (f *RecordHandlerFactory) CreateDeleteHandler() http.HandlerFunc {
	return delete.NewDeleteHandler(f.log, f.srepo)
}

func (f *RecordHandlerFactory) CreateUpdateHandler() http.HandlerFunc {
	return update.NewUpdateHandler(f.log, f.srepo)
}
