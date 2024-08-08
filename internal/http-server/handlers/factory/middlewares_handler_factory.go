package factory

import (
	"GYMBRO/internal/config"
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
	log   *slog.Logger
	urepo storage.UserRepository
	srepo storage.SessionRepository
	cfg   *config.Config
}

// NewMiddlewareHandlerFactory creates a new instance of MiddlewareHandlerFactory.
func NewMiddlewareHandlerFactory(log *slog.Logger, urepo storage.UserRepository, srepo storage.SessionRepository, cfg *config.Config) *MiddlewareHandlerFactory {
	return &MiddlewareHandlerFactory{
		log:   log,
		urepo: urepo,
		srepo: srepo,
		cfg:   cfg,
	}
}

func (f *MiddlewareHandlerFactory) CreateJWTAuthHandler() func(http.Handler) http.Handler {
	return mwjwt.WithJWTAuth(f.log, f.urepo, f.cfg)
}

func (f *MiddlewareHandlerFactory) CreateActiveSessionHandler() func(http.Handler) http.Handler {
	return mwworkout.WithActiveSessionCheck(f.log, f.srepo)
}
