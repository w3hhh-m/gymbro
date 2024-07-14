package storage

import (
	"errors"
)

// Common errors for storages

var (
	ErrExerciseNotFound = errors.New("exercise not found")
)

// Common structures

type Exercise struct {
	Id        int64  `json:"id"`
	Username  string `json:"username" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Sets      int    `json:"sets" validate:"required"`
	Rps       int    `json:"rps" validate:"required"`
	Weight    int    `json:"weight" validate:"required"`
	Timestamp string `json:"timestamp"`
}
