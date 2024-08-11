package factory

import (
	"GYMBRO/internal/http-server/handlers/records/add"
	"GYMBRO/internal/http-server/handlers/records/delete"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

type RecordsHandlerFactory interface {
	CreateAddHandler() http.HandlerFunc
	CreateDeleteHandler() http.HandlerFunc
}

type RecordHandlerFactory struct {
	log         *slog.Logger
	sessionRepo storage.SessionRepository
	userRepo    storage.UserRepository
}

func NewRecordHandlerFactory(log *slog.Logger, sessionRepo storage.SessionRepository, userRepo storage.UserRepository) *RecordHandlerFactory {
	return &RecordHandlerFactory{
		log:         log,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (f *RecordHandlerFactory) CreateAddHandler() http.HandlerFunc {
	return add.NewAddHandler(f.log, f.sessionRepo, f.userRepo)
}

func (f *RecordHandlerFactory) CreateDeleteHandler() http.HandlerFunc {
	return delete.NewDeleteHandler(f.log, f.sessionRepo)
}
