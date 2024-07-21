package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/records/delete"
	"GYMBRO/internal/http-server/handlers/records/get"
	"GYMBRO/internal/http-server/handlers/records/save"
	"GYMBRO/internal/http-server/handlers/users/login"
	"GYMBRO/internal/http-server/handlers/users/register"
	mwlogger "GYMBRO/internal/http-server/middleware/logger"
	"GYMBRO/internal/prettylogger"
	"GYMBRO/internal/storage/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

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

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat) //to extract {var} from url

	router.Post("/records", save.New(log, db))
	router.Get("/records/{id}", get.New(log, db))
	router.Delete("/records/{id}", delete.New(log, db))
	router.Post("/users", register.New(log, db))
	router.Get("/users", login.New(log, db, cfg.SecretKey))

	log.Info("Starting server", slog.String("address", cfg.Address))

	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		WriteTimeout: cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Error starting server", slog.Any("error", err))
	}

	log.Error("Server shutdown", slog.String("address", cfg.Address))
}

// setupLogger is a function that initialize logger depends on environment
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "production":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case "local":
		//got it from https://github.com/dusted-go/logging
		prettyHandler := prettylogger.NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   false,
			ReplaceAttr: nil,
		})
		log = slog.New(prettyHandler)
	}
	return log
}
