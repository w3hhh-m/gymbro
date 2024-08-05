package storage

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

// Common errors for storages

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("users already exists")
)

// Common structures

type Exercise struct {
	ExerciseId int    `json:"exercise_id"`
	Name       string `json:"name"`
}

type User struct {
	UserId      string    `json:"user_id"`
	Username    string    `json:"username" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required"`
	DateOfBirth time.Time `json:"date_of_birth"`
	GoogleId    string    `json:"google_id"`
	FkClanId    string    `json:"fk_clan_id"`
	FkGymId     int       `json:"fk_gym_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Gym struct {
	GymId       int    `json:"gym_id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type Clan struct {
	ClanId      string    `json:"clan_id"`
	FkOwnerId   string    `json:"fk_owner_id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Points      int       `json:"points" validate:"gte=0"`
	CreatedAt   time.Time `json:"created_at"`
}

type Workout struct {
	WorkoutId string    `json:"workout_id"`
	FkUserId  string    `json:"fk_user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Points    int       `json:"points"`
	IsActive  bool      `json:"is_active"`
}

type Record struct {
	RecordId     string `json:"record_id"`
	FkWorkoutId  string `json:"fk_workout_id" validate:"required"`
	FkExerciseId int    `json:"fk_exercise_id" validate:"required"`
	Reps         int    `json:"reps" validate:"required,gte=1"`
	Weight       int    `json:"weight" validate:"required,gte=1"`
}

type Subscription struct {
	SubscriptionId string    `json:"subscription_id"`
	FkUserId       string    `json:"fk_user_id"`
	FkGymId        int       `json:"fk_gym_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	CreatedAt      time.Time `json:"created_at"`
}

// Interfaces

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=WorkoutRepository --output=./mocks
type WorkoutRepository interface {
	CreateWorkout(workout Workout) error
	EndWorkout(workoutID string) error
	AddRecord(record Record) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=UserRepository --output=./mocks
type UserRepository interface {
	GetUserByID(string) (*User, error)
	GetUserByEmail(string) (*User, error)
	RegisterNewUser(User) (*string, error)
}

func GenerateUID() string {
	return uuid.New().String()
}
