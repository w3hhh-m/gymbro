package postgresql

import (
	"GYMBRO/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
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

// GetWorkout retrieves a workout record by its ID.
func (s *Storage) GetWorkout(workoutID string) (*storage.WorkoutWithRecords, error) {
	const op = "storage.postgresql.GetWorkout"

	query := `SELECT w.workout_id, w.fk_user_id, w.start_time, w.end_time, w.points, r.record_id, r.fk_workout_id, r.fk_exercise_id, r.reps, r.weight
	FROM workouts w
	LEFT JOIN records r ON w.workout_id = r.fk_workout_id
	WHERE w.workout_id = $1`

	rows, err := s.db.Query(context.Background(), query, workoutID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrWorkoutNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	workoutWithRecords := &storage.WorkoutWithRecords{
		WorkoutID: workoutID,
	}

	for rows.Next() {
		var record storage.Record
		err := rows.Scan(
			&workoutWithRecords.WorkoutID,
			&workoutWithRecords.UserID,
			&workoutWithRecords.StartTime,
			&workoutWithRecords.EndTime,
			&workoutWithRecords.Points,
			&record.RecordId,
			&record.FkWorkoutId,
			&record.FkExerciseId,
			&record.Reps,
			&record.Weight,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if record.RecordId != "" {
			workoutWithRecords.Records = append(workoutWithRecords.Records, record)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return workoutWithRecords, nil
}

func (s *Storage) SaveWorkout(workout *storage.WorkoutSession) error {
	const op = "storage.postgresql.SaveWorkout"
	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	if len(workout.Records) < 1 {
		tx.Commit(ctx)
		return nil
	}

	userQuery := `UPDATE users SET points = points + $1 WHERE user_id = $2`

	_, err = tx.Exec(ctx, userQuery, workout.Points, workout.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	workoutQuery := `INSERT INTO workouts (workout_id, fk_user_id, start_time, end_time, points, is_active) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err = tx.Exec(ctx, workoutQuery,
		workout.SessionID,
		workout.UserID,
		workout.StartTime,
		workout.LastUpdated,
		workout.Points,
		workout.IsActive,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	inParams := make([]string, 0, len(workout.Records)*5)
	args := make([]interface{}, 0, len(workout.Records)*5)

	for i, record := range workout.Records {
		inParams = append(inParams, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		args = append(args, record.RecordId, record.FkWorkoutId, record.FkExerciseId, record.Reps, record.Weight)
	}

	recordQuery := fmt.Sprintf(`INSERT INTO records (record_id, fk_workout_id, fk_exercise_id, reps, weight) VALUES %s`, strings.Join(inParams, ", "))

	_, err = tx.Exec(ctx, recordQuery, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit(ctx)
}
