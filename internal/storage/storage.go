package storage

import (
	"errors"
	"time"
)

// Common errors for storages

var (
	ErrRecordNotFound = errors.New("records not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrUserExists     = errors.New("users already exists")
)

// Common structures

type Record struct {
	RecordId     int       `json:"record_id"`
	FkUserId     int       `json:"fk_user_id" validate:"required"`
	FkExerciseId int       `json:"fk_exercise_id" validate:"required"`
	Reps         int       `json:"reps" validate:"required"`
	Weight       int       `json:"weight" validate:"required"`
	CreatedAt    time.Time `json:"created_at"`
}

type Exercise struct {
	ExerciseId int    `json:"exercise_id"`
	Name       string `json:"name"`
}

type User struct {
	UserId      int       `json:"user_id"`
	Username    string    `json:"username" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}
