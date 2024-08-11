package factory

import (
	"GYMBRO/internal/http-server/handlers/workouts/end"
	getwo "GYMBRO/internal/http-server/handlers/workouts/get"
	"GYMBRO/internal/http-server/handlers/workouts/start"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

type WorkoutsHandlerFactory interface {
	CreateStartHandler() http.HandlerFunc
	CreateEndHandler() http.HandlerFunc
	CreateGetWorkoutHandler() http.HandlerFunc
}

type WorkoutHandlerFactory struct {
	log         *slog.Logger
	workoutRepo storage.WorkoutRepository
	sessionRepo storage.SessionRepository
	userRepo    storage.UserRepository
}

func NewWorkoutHandlerFactory(log *slog.Logger, workoutRepo storage.WorkoutRepository, sessionRepo storage.SessionRepository, userRepo storage.UserRepository) *WorkoutHandlerFactory {
	return &WorkoutHandlerFactory{
		log:         log,
		workoutRepo: workoutRepo,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (f *WorkoutHandlerFactory) CreateStartHandler() http.HandlerFunc {
	return start.NewStartHandler(f.log, f.sessionRepo, f.userRepo)
}

func (f *WorkoutHandlerFactory) CreateEndHandler() http.HandlerFunc {
	return end.NewEndHandler(f.log, f.sessionRepo, f.workoutRepo, f.userRepo)
}

func (f *WorkoutHandlerFactory) CreateGetWorkoutHandler() http.HandlerFunc {
	return getwo.NewGetWorkoutHandler(f.log, f.workoutRepo)
}
