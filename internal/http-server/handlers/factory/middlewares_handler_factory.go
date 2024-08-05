package factory

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/workouts/sessions"
	mwjwt "GYMBRO/internal/http-server/middleware/jwt"
	mwworkout "GYMBRO/internal/http-server/middleware/workout"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

// MiddlewaresHandlerFactory defines the interface for creating middleware handlers.
type MiddlewaresHandlerFactory interface {
	CreateJWTAuthHandler() func(http.Handler) http.Handler
	CreateActiveSessionHandler() func(http.Handler) http.Handler
}

// MiddlewareHandlerFactory implements the MiddlewaresHandlerFactory interface.
type MiddlewareHandlerFactory struct {
	log  *slog.Logger
	repo storage.UserRepository
	cfg  *config.Config
	sm   *session.Manager
}

// NewMiddlewareHandlerFactory creates a new instance of MiddlewareHandlerFactory.
func NewMiddlewareHandlerFactory(log *slog.Logger, repo storage.UserRepository, cfg *config.Config, sm *session.Manager) *MiddlewareHandlerFactory {
	return &MiddlewareHandlerFactory{
		log:  log,
		repo: repo,
		cfg:  cfg,
		sm:   sm,
	}
}

func (f *MiddlewareHandlerFactory) CreateJWTAuthHandler() func(http.Handler) http.Handler {
	return mwjwt.WithJWTAuth(f.log, f.repo, f.cfg)
}

func (f *MiddlewareHandlerFactory) CreateActiveSessionHandler() func(http.Handler) http.Handler {
	return mwworkout.WithActiveSessionCheck(f.log, f.sm)
}
