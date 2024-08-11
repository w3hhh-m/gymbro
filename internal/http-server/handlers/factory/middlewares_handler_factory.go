package factory

import (
	"GYMBRO/internal/config"
	mwjwt "GYMBRO/internal/http-server/middleware/jwt"
	mwworkout "GYMBRO/internal/http-server/middleware/workout"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

type MiddlewaresHandlerFactory interface {
	CreateJWTAuthHandler() func(http.Handler) http.Handler
	CreateActiveSessionHandler() func(http.Handler) http.Handler
}

type MiddlewareHandlerFactory struct {
	log         *slog.Logger
	userRepo    storage.UserRepository
	sessionRepo storage.SessionRepository
	cfg         *config.Config
}

func NewMiddlewareHandlerFactory(log *slog.Logger, userRepo storage.UserRepository, sessionRepo storage.SessionRepository, cfg *config.Config) *MiddlewareHandlerFactory {
	return &MiddlewareHandlerFactory{
		log:         log,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (f *MiddlewareHandlerFactory) CreateJWTAuthHandler() func(http.Handler) http.Handler {
	return mwjwt.WithJWTAuth(f.log, f.userRepo, f.cfg)
}

func (f *MiddlewareHandlerFactory) CreateActiveSessionHandler() func(http.Handler) http.Handler {
	return mwworkout.WithActiveSessionCheck(f.log, f.sessionRepo)
}
