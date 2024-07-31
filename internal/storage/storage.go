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
	FkUserId     int       `json:"fk_user_id"`
	FkExerciseId int       `json:"fk_exercise_id" validate:"required"`
	Reps         int       `json:"reps" validate:"required,gte=1"`
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
	Phone       string    `json:"phone" validate:"required,numeric,len=11"`
	Password    string    `json:"password" validate:"required"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
}

// Interfaces

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=RecordRepository --output=./mocks
type RecordRepository interface {
	GetRecord(id int) (Record, error)
	DeleteRecord(id int) error
	SaveRecord(ex Record) (int, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=UserRepository --output=./mocks
type UserRepository interface {
	GetUserByID(id int) (User, error)
	GetUserByEmail(email string) (User, error)
	RegisterNewUser(usr User) (int, error)
}
