package factory

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/users/login"
	"GYMBRO/internal/http-server/handlers/users/logout"
	"GYMBRO/internal/http-server/handlers/users/oauth"
	"GYMBRO/internal/http-server/handlers/users/register"
	"GYMBRO/internal/storage"
	"log/slog"
	"net/http"
)

// UsersHandlerFactory defines the interface for creating user-related handlers.
type UsersHandlerFactory interface {
	CreateRegisterHandler() http.HandlerFunc
	CreateLoginHandler() http.HandlerFunc
	CreateLogoutHandler() http.HandlerFunc
	CreateOAuthCallbackHandler() http.HandlerFunc
	CreateOAuthLoginHandler() http.HandlerFunc
	CreateOAuthLogoutHandler() http.HandlerFunc
}

// UserHandlerFactory implements the UsersHandlerFactory interface.
type UserHandlerFactory struct {
	log  *slog.Logger
	repo storage.UserRepository
	cfg  *config.Config
}

// NewUserHandlerFactory creates a new instance of UserHandlerFactory.
func NewUserHandlerFactory(log *slog.Logger, repo storage.UserRepository, cfg *config.Config) *UserHandlerFactory {
	return &UserHandlerFactory{
		log:  log,
		repo: repo,
		cfg:  cfg,
	}
}

func (f *UserHandlerFactory) CreateRegisterHandler() http.HandlerFunc {
	return register.NewRegisterHandler(f.log, f.repo)
}

func (f *UserHandlerFactory) CreateLoginHandler() http.HandlerFunc {
	return login.NewLoginHandler(f.log, f.repo, f.cfg)
}

func (f *UserHandlerFactory) CreateLogoutHandler() http.HandlerFunc {
	return logout.NewLogoutHandler(f.log)
}

func (f *UserHandlerFactory) CreateOAuthCallbackHandler() http.HandlerFunc {
	return oauth.NewOAuthCallbackHandler(f.log, f.repo, f.cfg)
}

func (f *UserHandlerFactory) CreateOAuthLoginHandler() http.HandlerFunc {
	return oauth.NewOAuthLoginHandler(f.log)
}

func (f *UserHandlerFactory) CreateOAuthLogoutHandler() http.HandlerFunc {
	return oauth.NewOAuthLogoutHandler(f.log)
}
