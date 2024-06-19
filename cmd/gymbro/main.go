package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/storage/sqlite"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env) // TODO: make logger look pretty
	log.Info("Configuration loaded.")
	log.Info("Logger loaded.")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("Error initializing storage.", slog.Any("error", err))
		os.Exit(1)
	}
	_ = storage
	log.Info("Storage loaded.")

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
