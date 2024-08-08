package factory

import (
	"GYMBRO/internal/http-server/handlers/workouts/end"
	getwo "GYMBRO/internal/http-server/handlers/workouts/get"
	"GYMBRO/internal/http-server/handlers/workouts/start"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

// WorkoutsHandlerFactory defines the interface for creating workout-related handlers.
type WorkoutsHandlerFactory interface {
	CreateStartHandler() http.HandlerFunc
	CreateEndHandler() http.HandlerFunc
	CreateGetWorkoutHandler() http.HandlerFunc
}

// WorkoutHandlerFactory implements the WorkoutsHandlerFactory interface.
type WorkoutHandlerFactory struct {
	log   *slog.Logger
	wrepo storage.WorkoutRepository
	srepo storage.SessionRepository
}

// NewWorkoutHandlerFactory creates a new instance of WorkoutHandlerFactory.
func NewWorkoutHandlerFactory(log *slog.Logger, wrepo storage.WorkoutRepository, srepo storage.SessionRepository) *WorkoutHandlerFactory {
	return &WorkoutHandlerFactory{
		log:   log,
		wrepo: wrepo,
		srepo: srepo,
	}
}

func (f *WorkoutHandlerFactory) CreateStartHandler() http.HandlerFunc {
	return start.NewStartHandler(f.log, f.srepo)
}

func (f *WorkoutHandlerFactory) CreateEndHandler() http.HandlerFunc {
	return end.NewEndHandler(f.log, f.srepo, f.wrepo)
}

func (f *WorkoutHandlerFactory) CreateGetWorkoutHandler() http.HandlerFunc {
	return getwo.NewGetWorkoutHandler(f.log, f.wrepo)
}
