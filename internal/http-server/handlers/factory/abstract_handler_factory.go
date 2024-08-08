package factory

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/storage"
	"log/slog"
)

// AbstractHandlerFactory defines the interface for the handler factory.
type AbstractHandlerFactory interface {
	GetMiddlewaresHandlerFactory() MiddlewaresHandlerFactory
	GetUsersHandlerFactory() UsersHandlerFactory
	GetWorkoutsHandlerFactory() WorkoutsHandlerFactory
	GetRecordsHandlerFactory() RecordsHandlerFactory
}

// ConcreteHandlerFactory implements the AbstractHandlerFactory interface.
type ConcreteHandlerFactory struct {
	log   *slog.Logger
	urepo storage.UserRepository
	wrepo storage.WorkoutRepository
	srepo storage.SessionRepository
	cfg   *config.Config
}

// NewConcreteHandlerFactory creates a new instance of ConcreteHandlerFactory.
func NewConcreteHandlerFactory(log *slog.Logger, urepo storage.UserRepository, wrepo storage.WorkoutRepository, srepo storage.SessionRepository, cfg *config.Config) *ConcreteHandlerFactory {
	return &ConcreteHandlerFactory{
		log:   log,
		urepo: urepo,
		wrepo: wrepo,
		cfg:   cfg,
		srepo: srepo,
	}
}

func (f *ConcreteHandlerFactory) GetMiddlewaresHandlerFactory() MiddlewaresHandlerFactory {
	return NewMiddlewareHandlerFactory(f.log, f.urepo, f.srepo, f.cfg)
}

func (f *ConcreteHandlerFactory) GetUsersHandlerFactory() UsersHandlerFactory {
	return NewUserHandlerFactory(f.log, f.urepo, f.cfg)
}

func (f *ConcreteHandlerFactory) GetWorkoutsHandlerFactory() WorkoutsHandlerFactory {
	return NewWorkoutHandlerFactory(f.log, f.wrepo, f.srepo)
}

func (f *ConcreteHandlerFactory) GetRecordsHandlerFactory() RecordsHandlerFactory {
	return NewRecordHandlerFactory(f.log, f.srepo)
}
