package factory

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/storage"
	"log/slog"
)

// AbstractHandlerFactory defines the interface for the handler factory.
type AbstractHandlerFactory interface {
	GetMiddlewaresHandlerFactory() MiddlewaresHandlerFactory
	GetUsersHandlerFactory() UsersHandlerFactory
	GetWorkoutsHandlerFactory() WorkoutsHandlerFactory
}

// ConcreteHandlerFactory implements the AbstractHandlerFactory interface.
type ConcreteHandlerFactory struct {
	log   *slog.Logger
	urepo storage.UserRepository
	wrepo storage.WorkoutRepository
	cfg   *config.Config
	sm    *session.Manager
}

// NewConcreteHandlerFactory creates a new instance of ConcreteHandlerFactory.
func NewConcreteHandlerFactory(log *slog.Logger, urepo storage.UserRepository, wrepo storage.WorkoutRepository, cfg *config.Config, sm *session.Manager) *ConcreteHandlerFactory {
	return &ConcreteHandlerFactory{
		log:   log,
		urepo: urepo,
		wrepo: wrepo,
		cfg:   cfg,
		sm:    sm,
	}
}

func (f *ConcreteHandlerFactory) GetMiddlewaresHandlerFactory() MiddlewaresHandlerFactory {
	return NewMiddlewareHandlerFactory(f.log, f.urepo, f.cfg, f.sm)
}

func (f *ConcreteHandlerFactory) GetUsersHandlerFactory() UsersHandlerFactory {
	return NewUserHandlerFactory(f.log, f.urepo, f.cfg)
}

func (f *ConcreteHandlerFactory) GetWorkoutsHandlerFactory() WorkoutsHandlerFactory {
	return NewWorkoutHandlerFactory(f.log, f.wrepo, f.sm)
}
