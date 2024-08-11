package storage

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("users already exists")
	ErrWorkoutNotFound = errors.New("workout not found")
	ErrNoSession       = errors.New("no session")
	ErrNoMaxes         = errors.New("no maxes")
)

type WorkoutWithRecords struct {
	UserID    string    `json:"user_id"`
	WorkoutID string    `json:"session_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Records   []Record  `json:"records"`
	Points    int       `json:"points"`
}

type Record struct {
	RecordId     string `json:"record_id"`
	FkWorkoutId  string `json:"fk_workout_id"`
	FkExerciseId int    `json:"fk_exercise_id" validate:"required"`
	Reps         int    `json:"reps" validate:"required,gte=1"`
	Weight       int    `json:"weight" validate:"required,gte=1"`
	Points       int    `json:"points"`
}

type Subscription struct {
	SubscriptionId string    `json:"subscription_id"`
	FkUserId       string    `json:"fk_user_id"`
	FkGymId        int       `json:"fk_gym_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	CreatedAt      time.Time `json:"created_at"`
}

type Workout struct {
	WorkoutId string    `json:"workout_id"`
	FkUserId  string    `json:"fk_user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Points    int       `json:"points"`
}

type Exercise struct {
	ExerciseId  int    `json:"exercise_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Picture     string `json:"picture"`
}

type User struct {
	UserId      string    `json:"user_id"`
	Username    string    `json:"username" validate:"required"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required"`
	Points      int       `json:"points"`
	DateOfBirth time.Time `json:"date_of_birth"`
	GoogleId    string    `json:"google_id"`
	FkClanId    string    `json:"fk_clan_id"`
	FkGymId     int       `json:"fk_gym_id"`
	IsActive    bool      `json:"is_active"`
	LastActive  time.Time `json:"last_active"`
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

type WorkoutSession struct {
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	StartTime   time.Time `json:"start_time"`
	LastUpdated time.Time `json:"last_updated"`
	Records     []Record  `json:"records"`
	Points      int       `json:"points"`
}

type Max struct {
	UserID     string `json:"user_id"`
	ExerciseId int    `json:"exercise_id"`
	MaxWeight  int    `json:"max_weight"`
	Reps       int    `json:"reps"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=WorkoutRepository --output=./mocks
type WorkoutRepository interface {
	GetWorkout(*string) (*WorkoutWithRecords, error)
	SaveWorkout(*WorkoutSession) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=SessionRepository --output=./mocks
type SessionRepository interface {
	CreateSession(*WorkoutSession) error
	UpdateSession(*string, *WorkoutSession) error
	DeleteSession(*string) error
	GetSession(*string) (*WorkoutSession, error)
	GetAllSessions() ([]*WorkoutSession, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.43.2 --name=UserRepository --output=./mocks
type UserRepository interface {
	GetUserByID(*string) (*User, error)
	GetUserByEmail(*string) (*User, error)
	RegisterNewUser(*User) (*string, error)
	ChangeStatus(*string, bool) error
	GetUserMax(*string, *int) (*Max, error)
	GetUserMaxes(*string) ([]*Max, error)
	SetUserMax(*string, *Max) error
}

func GenerateUID() string {
	return uuid.New().String()
}
