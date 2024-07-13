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
	Id        int64     `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Sets      int       `json:"sets"`
	Rps       int       `json:"rps"`
	Weight    int       `json:"weight"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}
