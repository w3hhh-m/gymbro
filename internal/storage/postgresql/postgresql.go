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
	"time"
)

type Storage struct {
	db *pgxpool.Pool
}

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
func (s *Storage) RegisterNewUser(user *storage.User) (*string, error) {
	const op = "storage.postgresql.RegisterNewUser"
	_, err := s.db.Exec(context.Background(),
		`INSERT INTO users (user_id, username, email, password_hash, date_of_birth, google_id, fk_clan_id, fk_gym_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.UserId, user.Username, user.Email, user.Password, user.DateOfBirth, user.GoogleId, "0", 0)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation error code
			return nil, storage.ErrUserExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user.UserId, nil
}

// GetUserByID retrieves a user's data by their ID
func (s *Storage) GetUserByID(id *string) (*storage.User, error) {
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
func (s *Storage) GetUserByEmail(email *string) (*storage.User, error) {
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

// ChangeStatus updates the active status and last active timestamp for a user.
func (s *Storage) ChangeStatus(userID *string, status bool) error {
	const op = "storage.postgresql.ChangeStatus"
	_, err := s.db.Exec(context.Background(), `UPDATE users SET is_active = $1, last_active = $2 WHERE user_id = $3`, status, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetUserMax retrieves the maximum weight and reps for a specific exercise.
func (s *Storage) GetUserMax(userID *string, exercise *int) (*storage.Max, error) {
	const op = "storage.postgresql.GetUserMax"
	var userMax storage.Max
	row := s.db.QueryRow(context.Background(), `SELECT user_id, exercise_id, max_weight, reps FROM userexercisemaxweights WHERE user_id = $1 AND exercise_id = $2`, userID, exercise)
	err := row.Scan(&userMax.UserID, &userMax.ExerciseId, &userMax.MaxWeight, &userMax.Reps)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNoMaxes
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &userMax, nil
}

// GetUserMaxes retrieves all maximum weight records for a user.
func (s *Storage) GetUserMaxes(userID *string) ([]*storage.Max, error) {
	const op = "storage.postgresql.GetUserMaxes"
	var userMaxes []*storage.Max
	rows, err := s.db.Query(context.Background(), `SELECT user_id, exercise_id, max_weight, reps FROM userexercisemaxweights WHERE user_id = $1`, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userMaxes, storage.ErrNoMaxes
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		userMax := &storage.Max{}
		err := rows.Scan(&userMax.UserID, &userMax.ExerciseId, &userMax.MaxWeight, &userMax.Reps)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		if userMax.UserID != "" {
			userMaxes = append(userMaxes, userMax)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userMaxes, nil
}

// SetUserMax inserts or updates the maximum weight and reps for a user's exercise.
func (s *Storage) SetUserMax(userID *string, max *storage.Max) error {
	const op = "storage.postgresql.SetUserMax"
	_, err := s.db.Exec(context.Background(), `INSERT INTO userexercisemaxweights (user_id, exercise_id, max_weight, reps) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, exercise_id) DO UPDATE SET max_weight = EXCLUDED.max_weight, reps = EXCLUDED.reps`, userID, max.ExerciseId, max.MaxWeight, max.Reps)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetWorkout retrieves a workout record by its ID.
func (s *Storage) GetWorkout(workoutID *string) (*storage.WorkoutWithRecords, error) {
	const op = "storage.postgresql.GetWorkout"

	query := `SELECT w.workout_id, w.fk_user_id, w.start_time, w.end_time, w.points, r.record_id, r.fk_workout_id, r.fk_exercise_id, r.reps, r.weight, r.points
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
		WorkoutID: *workoutID,
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
			&record.Points,
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
			err := tx.Rollback(ctx)
			if err != nil {
				return
			}
		}
	}()

	if len(workout.Records) < 1 {
		err := tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	userQuery := `UPDATE users SET points = points + $1 WHERE user_id = $2`

	_, err = tx.Exec(ctx, userQuery, workout.Points, workout.UserID)
	if err != nil {
		return fmt.Errorf("%s, userQuery: %w", op, err)
	}

	workoutQuery := `INSERT INTO workouts (workout_id, fk_user_id, start_time, end_time, points) VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(ctx, workoutQuery,
		workout.SessionID,
		workout.UserID,
		workout.StartTime,
		workout.LastUpdated,
		workout.Points,
	)
	if err != nil {
		return fmt.Errorf("%s, workoutQuery: %w", op, err)
	}

	inParams := make([]string, 0, len(workout.Records))
	args := make([]interface{}, 0, len(workout.Records)*6)

	for i, record := range workout.Records {
		inParams = append(inParams, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6))
		args = append(args, record.RecordId, record.FkWorkoutId, record.FkExerciseId, record.Reps, record.Weight, record.Points)
	}

	recordQuery := fmt.Sprintf(`INSERT INTO records (record_id, fk_workout_id, fk_exercise_id, reps, weight, points) VALUES %s`, strings.Join(inParams, ", "))

	_, err = tx.Exec(ctx, recordQuery, args...)
	if err != nil {
		return fmt.Errorf("%s, recordQuery: %w", op, err)
	}

	return tx.Commit(ctx)
}
