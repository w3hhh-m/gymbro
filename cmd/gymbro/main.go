package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/http-server/handlers/exercise/delete"
	"GYMBRO/internal/http-server/handlers/exercise/get"
	"GYMBRO/internal/http-server/handlers/exercise/save"
	mwlogger "GYMBRO/internal/http-server/middleware/logger"
	"GYMBRO/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env) // TODO: make logger look pretty
	log.Info("Configuration loaded.")
	log.Info("Logger loaded.")

	db, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("Error initializing storage.", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("Storage loaded.")

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/exercise", save.New(log, db))
	router.Get("/exercise/{id}", get.New(log, db))
	router.Delete("/exercise/{id}", delete.New(log, db))

	log.Info("starting server", slog.String("address", cfg.Address))

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
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}
