package postgresql

import (
	"GYMBRO/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Initializing postgresql storage

// Storage struct contains only one field which is connection to database, made in process of New function
type Storage struct {
	db *pgxpool.Pool
}

// New initializes postgresql storage connection, creating tables if not exist
func New(storagePath string) (*Storage, error) {
	const op = "storage.postgresql.New"
	dbpool, err := pgxpool.New(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = dbpool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS users (user_id serial PRIMARY KEY, username varchar(255) NOT NULL UNIQUE, email varchar(255) NOT NULL UNIQUE, password varchar(255) NOT NULL, date_of_birth date NOT NULL, created_at timestamp NOT NULL)`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = dbpool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS exercises (exercise_id serial PRIMARY KEY, name varchar(255) NOT NULL UNIQUE)`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = dbpool.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS records (record_id serial PRIMARY KEY, fk_user_id integer REFERENCES users(user_id) NOT NULL, fk_exercise_id integer REFERENCES exercises(exercise_id) NOT NULL, reps integer NOT NULL, weight integer NOT NULL, created_at timestamp NOT NULL)`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: dbpool}, nil
}

// SaveRecord is a method of Storage struct, that is saving given data of records to a database
func (s *Storage) SaveRecord(rec storage.Record) (int, error) {
	const op = "storage.postgresql.SaveRecord"
	var id int
	err := s.db.QueryRow(context.Background(), `INSERT INTO records (fk_user_id, fk_exercise_id, reps, weight, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING record_id`, rec.FkUserId, rec.FkExerciseId, rec.Reps, rec.Weight, rec.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// GetRecord is a method of Storage struct that is getting data of one exact records by given id
func (s *Storage) GetRecord(id int) (storage.Record, error) {
	const op = "storage.postgresql.GetRecord"
	var rec storage.Record
	row := s.db.QueryRow(context.Background(), `SELECT * FROM records WHERE record_id = $1`, id)
	err := row.Scan(&rec.RecordId, &rec.FkUserId, &rec.FkExerciseId, &rec.Reps, &rec.Weight, &rec.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return rec, storage.ErrRecordNotFound
	}
	if err != nil {
		return rec, fmt.Errorf("%s: %w", op, err)
	}
	return rec, nil
}

// DeleteRecord is a method of Storage struct that is deleting exact one records from database by its id
func (s *Storage) DeleteRecord(id int) error {
	const op = "storage.postgresql.DeleteRecord"
	_, err := s.db.Exec(context.Background(), `DELETE FROM records WHERE record_id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) RegisterNewUser(usr storage.User) (int, error) {
	const op = "storage.postgresql.RegisterNewUser"
	var id int
	err := s.db.QueryRow(context.Background(), `INSERT INTO users (username, email, password, date_of_birth, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING user_id`, usr.Username, usr.Email, usr.Password, usr.DateOfBirth, usr.CreatedAt).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, storage.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}
