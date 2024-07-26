package postgresql

import (
	"context"
	"errors"
	"fmt"

	"GYMBRO/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage struct contains only one field which is connection to database, made in process of New function
type Storage struct {
	db *pgxpool.Pool
}

// New initializes PostgreSQL storage connection
func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"
	dbpool, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: dbpool}, nil
}

// SaveRecord saves the given record data to the database
func (s *Storage) SaveRecord(rec storage.Record) (int, error) {
	const op = "storage.postgresql.SaveRecord"
	var id int
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO records (fk_user_id, fk_exercise_id, reps, weight, created_at) 
		VALUES ($1, $2, $3, $4, $5) RETURNING record_id`,
		rec.FkUserId, rec.FkExerciseId, rec.Reps, rec.Weight, rec.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// GetRecord retrieves the data of a specific record by its ID
func (s *Storage) GetRecord(id int) (storage.Record, error) {
	const op = "storage.postgresql.GetRecord"
	var rec storage.Record
	row := s.db.QueryRow(context.Background(), `SELECT * FROM records WHERE record_id = $1`, id)
	err := row.Scan(&rec.RecordId, &rec.FkUserId, &rec.FkExerciseId, &rec.Reps, &rec.Weight, &rec.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, storage.ErrRecordNotFound
	}
	if err != nil {
		return rec, fmt.Errorf("%s: %w", op, err)
	}
	return rec, nil
}

// DeleteRecord deletes a specific record from the database by its ID
func (s *Storage) DeleteRecord(id int) error {
	const op = "storage.postgresql.DeleteRecord"
	_, err := s.db.Exec(context.Background(), `DELETE FROM records WHERE record_id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// RegisterNewUser registers a new user in the database
func (s *Storage) RegisterNewUser(usr storage.User) (int, error) {
	const op = "storage.postgresql.RegisterNewUser"
	var id int
	err := s.db.QueryRow(context.Background(),
		`INSERT INTO users (username, email, phone, password, date_of_birth, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING user_id`,
		usr.Username, usr.Email, usr.Phone, usr.Password, usr.DateOfBirth, usr.CreatedAt).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation error code
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// GetUserByID retrieves a user's data by their ID
func (s *Storage) GetUserByID(id int) (storage.User, error) {
	const op = "storage.postgresql.GetUserByID"
	var user storage.User
	row := s.db.QueryRow(context.Background(), `SELECT * FROM users WHERE user_id = $1`, id)
	err := row.Scan(&user.UserId, &user.Username, &user.Email, &user.Phone, &user.Password, &user.DateOfBirth, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, storage.ErrUserNotFound
	}
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user's data by their email
func (s *Storage) GetUserByEmail(email string) (storage.User, error) {
	const op = "storage.postgresql.GetUserByEmail"
	var user storage.User
	row := s.db.QueryRow(context.Background(), `SELECT * FROM users WHERE email = $1`, email)
	err := row.Scan(&user.UserId, &user.Username, &user.Email, &user.Phone, &user.Password, &user.DateOfBirth, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, storage.ErrUserNotFound
	}
	if err != nil {
		return user, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}
