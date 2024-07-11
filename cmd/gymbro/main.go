package main

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/storage"
	"GYMBRO/internal/storage/sqlite"
	"fmt"
	"log/slog"
	"os"
	"time"
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
	_ = db
	log.Info("Storage loaded.")
	id, err := db.SaveExercise(storage.Exercise{
		Username:  "qwerty",
		Name:      "push ups",
		Sets:      5,
		Rps:       20,
		Weight:    0,
		Timestamp: time.Now(),
	})
	if err != nil {
		log.Error("Error saving exercise.", slog.Any("error", err))
	}
	log.Info("Exercise saved.", slog.Any("exercise", id))
	exercise, err := db.GetExercise(id)
	if err != nil {
		log.Error("Error retrieving exercise.", slog.Any("error", err))
	}
	fmt.Print(exercise)
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
