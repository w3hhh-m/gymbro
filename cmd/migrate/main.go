package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "", "direction [up|down] default:up")
	flag.Parse()
	storagePath := os.Getenv("STORAGE_PATH") + "?sslmode=disable"
	migrationsPath := "file://cmd/migrate/migrations"
	m, err := migrate.New(migrationsPath, storagePath)
	if err != nil {
		panic(err)
	}
	if direction == "up" {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}
			panic(err)
		}
	}
	if direction == "down" {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to rollback")
				return
			}
			panic(err)
		}
	}
	fmt.Println("applied migrations")
}
