package sqlite

import (
	"GYMBRO/internal/storage"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3" //init sqlite3 driver
)

// Initializing sqlite storage

// Storage struct contains only one field which is connection to database, made in process of New function
type Storage struct {
	db *sql.DB
}

// New initializes sqlite storage, creating table exercises (id, username, exercise, sets, rps, weight, time)
func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// TODO: add migrations
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS exercises (id INTEGER PRIMARY KEY, username TEXT, exercise TEXT, sets INTEGER, rps INTEGER, weight INTEGER, time TEXT)`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}

// TODO: make indexes for usernames and exercises

// SaveExercise is a method of Storage struct, that is saving given data of exercise to a database
func (s *Storage) SaveExercise(ex storage.Exercise) (int64, error) {
	const op = "storage.sqlite.SaveExercise"
	stmt, err := s.db.Prepare(`INSERT INTO exercises (username, exercise, sets, rps, weight, time) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	res, err := stmt.Exec(ex.Username, ex.Name, ex.Sets, ex.Rps, ex.Weight, ex.Timestamp)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// GetExercise is a method of Storage struct that is getting data of one exact exercise by given id
func (s *Storage) GetExercise(id int64) (storage.Exercise, error) {
	const op = "storage.sqlite.GetExercise"
	var ex storage.Exercise
	stmt, err := s.db.Prepare(`SELECT * FROM exercises WHERE id = ?`)
	if err != nil {
		return ex, fmt.Errorf("%s: %w", op, err)
	}
	err = stmt.QueryRow(id).Scan(&ex.Id, &ex.Username, &ex.Name, &ex.Sets, &ex.Rps, &ex.Weight, &ex.Timestamp)
	if errors.Is(err, sql.ErrNoRows) {
		return ex, storage.ErrExerciseNotFound
	}
	if err != nil {
		return ex, fmt.Errorf("%s: %w", op, err)
	}
	return ex, nil
}

// DeleteExercise is a method of Storage struct that is deleting exact one exercise from database by its id
func (s *Storage) DeleteExercise(id int64) error {
	const op = "storage.sqlite.DeleteExercise"
	stmt, err := s.db.Prepare(`DELETE FROM exercises WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
