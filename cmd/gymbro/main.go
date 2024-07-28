package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/records/delete"
	"GYMBRO/internal/http-server/handlers/records/get"
	"GYMBRO/internal/http-server/handlers/records/save"
	"GYMBRO/internal/http-server/handlers/users/login"
	"GYMBRO/internal/http-server/handlers/users/logout"
	"GYMBRO/internal/http-server/handlers/users/oauth"
	"GYMBRO/internal/http-server/handlers/users/register"
	mwlogger "GYMBRO/internal/http-server/middleware/logger"
	"GYMBRO/internal/lib/jwt"
	"GYMBRO/internal/lib/prettylogger"
	"GYMBRO/internal/storage/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

// TODO: make tests with mocks

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("Configuration loaded")
	log.Info("Logger loaded")

	db, err := postgresql.New(cfg.StoragePath)
	if err != nil {
		log.Error("Error initializing storage", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Storage loaded")

	router := setupRouter(cfg, log, db)
	startServer(cfg, router, log)
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case "production":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "local":
		//got it from https://github.com/dusted-go/logging
		prettyHandler := prettylogger.NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   false,
			ReplaceAttr: nil,
		})
		return slog.New(prettyHandler)
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

func setupRouter(cfg *config.Config, log *slog.Logger, db *postgresql.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat) // to extract {var} from url

	oauth.NewOAuth(cfg)

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(jwt.WithJWTAuth(log, db, cfg.SecretKey))
		r.Post("/records", save.NewSaveHandler(log, db))
		r.Get("/records/{id}", get.NewGetHandler(log, db))
		r.Delete("/records/{id}", delete.NewDeleteHandler(log, db))
		r.Get("/users/logout", logout.NewLogoutHandler(log))
	})

	// Public routes
	router.Post("/users/register", register.NewRegisterHandler(log, db))
	router.Get("/users/login", login.NewLoginHandler(log, db, cfg.SecretKey))

	// OAuth routes
	router.Get("/users/oauth/{provider}/callback", oauth.NewCallbackHandler(log, db, cfg.SecretKey))
	router.Get("/users/oauth/{provider}/logout", oauth.NewLogoutHandler(log))
	router.Get("/users/oauth/{provider}", oauth.NewLoginHandler(log))

	return router
}

func startServer(cfg *config.Config, router *chi.Mux, log *slog.Logger) {
	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		WriteTimeout: cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("Starting server", slog.String("address", cfg.Address))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Error starting server", slog.Any("error", err))
	}

	log.Error("Server shutdown", slog.String("address", cfg.Address))
}
