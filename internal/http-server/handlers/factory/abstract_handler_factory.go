package factory

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/storage"
	"log/slog"
)

type AbstractHandlerFactory interface {
	GetMiddlewaresHandlerFactory() MiddlewaresHandlerFactory
	GetUsersHandlerFactory() UsersHandlerFactory
	GetWorkoutsHandlerFactory() WorkoutsHandlerFactory
	GetRecordsHandlerFactory() RecordsHandlerFactory
}

type ConcreteHandlerFactory struct {
	log         *slog.Logger
	userRepo    storage.UserRepository
	workoutRepo storage.WorkoutRepository
	sessionRepo storage.SessionRepository
	cfg         *config.Config
}

func NewConcreteHandlerFactory(log *slog.Logger, userRepo storage.UserRepository, workoutRepo storage.WorkoutRepository, sessionRepo storage.SessionRepository, cfg *config.Config) *ConcreteHandlerFactory {
	return &ConcreteHandlerFactory{
		log:         log,
		userRepo:    userRepo,
		workoutRepo: workoutRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (f *ConcreteHandlerFactory) GetMiddlewaresHandlerFactory() MiddlewaresHandlerFactory {
	return NewMiddlewareHandlerFactory(f.log, f.userRepo, f.sessionRepo, f.cfg)
}

func (f *ConcreteHandlerFactory) GetUsersHandlerFactory() UsersHandlerFactory {
	return NewUserHandlerFactory(f.log, f.userRepo, f.cfg)
}

func (f *ConcreteHandlerFactory) GetWorkoutsHandlerFactory() WorkoutsHandlerFactory {
	return NewWorkoutHandlerFactory(f.log, f.workoutRepo, f.sessionRepo, f.userRepo)
}

func (f *ConcreteHandlerFactory) GetRecordsHandlerFactory() RecordsHandlerFactory {
	return NewRecordHandlerFactory(f.log, f.sessionRepo, f.userRepo)
}
