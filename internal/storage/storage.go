package storage

import (
	"errors"
	"time"
)

// Common errors for storages
// TODO: add more errors :)

var (
	ErrExerciseNotFound = errors.New("exercise not found")
)

// Common structures

type Exercise struct {
	Id        int64
	Username  string
	Name      string
	Sets      int
	Rps       int
	Weight    int
	Timestamp time.Time
}
