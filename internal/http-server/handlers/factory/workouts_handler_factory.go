package factory

import (
	"GYMBRO/internal/http-server/handlers/workouts/end"
	"GYMBRO/internal/http-server/handlers/workouts/records/add"
	"GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/http-server/handlers/workouts/start"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

// WorkoutsHandlerFactory defines the interface for creating workout-related handlers.
type WorkoutsHandlerFactory interface {
	CreateStartHandler() http.HandlerFunc
	CreateEndHandler() http.HandlerFunc
	CreateAddHandler() http.HandlerFunc
}

// WorkoutHandlerFactory implements the WorkoutsHandlerFactory interface.
type WorkoutHandlerFactory struct {
	log  *slog.Logger
	repo storage.WorkoutRepository
	sm   *session.Manager
}

// NewWorkoutHandlerFactory creates a new instance of WorkoutHandlerFactory.
func NewWorkoutHandlerFactory(log *slog.Logger, repo storage.WorkoutRepository, sm *session.Manager) *WorkoutHandlerFactory {
	return &WorkoutHandlerFactory{
		log:  log,
		repo: repo,
		sm:   sm,
	}
}

func (f *WorkoutHandlerFactory) CreateStartHandler() http.HandlerFunc {
	return start.NewStartHandler(f.log, f.repo, f.sm)
}

func (f *WorkoutHandlerFactory) CreateEndHandler() http.HandlerFunc {
	return end.NewEndHandler(f.log, f.repo, f.sm)
}

func (f *WorkoutHandlerFactory) CreateAddHandler() http.HandlerFunc {
	return add.NewAddHandler(f.log, f.repo, f.sm)
}
