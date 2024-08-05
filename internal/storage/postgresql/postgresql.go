package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"GYMBRO/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage struct holds the PostgreSQL database connection pool
type Storage struct {
	db *pgxpool.Pool
}

// New initializes a new PostgreSQL storage connection using the provided connection string
func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"
	dbpool, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: dbpool}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

// RegisterNewUser registers a new user in the database and returns the user ID or an error
func (s *Storage) RegisterNewUser(usr storage.User) (*string, error) {
	const op = "storage.postgresql.RegisterNewUser"
	_, err := s.db.Exec(context.Background(),
		`INSERT INTO users (user_id, username, email, password_hash, date_of_birth, google_id, fk_clan_id, fk_gym_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		usr.UserId, usr.Username, usr.Email, usr.Password, usr.DateOfBirth, usr.GoogleId, "0", 0)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation error code
			return nil, storage.ErrUserExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &usr.UserId, nil
}

// GetUserByID retrieves a user's data by their ID
func (s *Storage) GetUserByID(id string) (*storage.User, error) {
	const op = "storage.postgresql.GetUserByID"
	var user storage.User
	row := s.db.QueryRow(context.Background(), `SELECT user_id, username, email, password_hash, date_of_birth, google_id, fk_clan_id, fk_gym_id, created_at FROM users WHERE user_id = $1`, id)
	err := row.Scan(&user.UserId, &user.Username, &user.Email, &user.Password, &user.DateOfBirth, &user.GoogleId, &user.FkClanId, &user.FkGymId, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user's data by their email
func (s *Storage) GetUserByEmail(email string) (*storage.User, error) {
	const op = "storage.postgresql.GetUserByEmail"
	var user storage.User
	row := s.db.QueryRow(context.Background(), `SELECT user_id, username, email, password_hash, date_of_birth, google_id, fk_clan_id, fk_gym_id, created_at FROM users WHERE email = $1`, email)
	err := row.Scan(&user.UserId, &user.Username, &user.Email, &user.Password, &user.DateOfBirth, &user.GoogleId, &user.FkClanId, &user.FkGymId, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

// CreateWorkout inserts a new workout record into the database
func (s *Storage) CreateWorkout(workout storage.Workout) error {
	const op = "storage.postgresql.CreateWorkout"
	_, err := s.db.Exec(context.Background(), `INSERT INTO workouts (workout_id, fk_user_id, start_time, is_active) 
		VALUES ($1, $2, $3, $4)`, workout.WorkoutId, workout.FkUserId, workout.StartTime, workout.IsActive)
	return err
}

// EndWorkout updates an existing workout record to mark it as completed
func (s *Storage) EndWorkout(workoutID string) error {
	const op = "storage.postgresql.EndWorkout"
	_, err := s.db.Exec(context.Background(), `UPDATE workouts SET end_time = $1, is_active = $2 WHERE workout_id = $3`,
		time.Now(), false, workoutID)
	return err
}

// AddRecord inserts a new record into the database and updates the workout points.
func (s *Storage) AddRecord(record storage.Record) error {
	const op = "storage.postgresql.AddRecord"
	ctx := context.Background()

	// Start a transaction.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	// Defer transaction rollback in case of error.
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// Insert the record into the database.
	_, err = tx.Exec(ctx, `INSERT INTO records (record_id, fk_workout_id, fk_exercise_id, reps, weight) 
		VALUES ($1, $2, $3, $4, $5)`,
		record.RecordId, record.FkWorkoutId, record.FkExerciseId, record.Reps, record.Weight)
	if err != nil {
		return err
	}

	// Calculate points.
	points := record.Reps * record.Weight

	// Update the workout points.
	_, err = tx.Exec(ctx, `UPDATE workouts 
		SET points = COALESCE(points, 0) + $1
		WHERE workout_id = $2`,
		points, record.FkWorkoutId)
	if err != nil {
		return err
	}

	// Commit the transaction.
	return tx.Commit(ctx)
}
