package redis

import (
	"GYMBRO/internal/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisStorage manages workout sessions with Redis as a storage.
type RedisStorage struct {
	Client *redis.Client
	ctx    context.Context
}

// New creates a new instance of Manager with a Redis Client.
func New(redisAddr string, password string, db int) (*RedisStorage, error) {
	const op = "storage.redis.New"
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &RedisStorage{
		Client: rdb,
		ctx:    context.Background(),
	}, nil
}

// CreateSession initializes a new workout session for a user and stores it in Redis.
func (rs *RedisStorage) CreateSession(session storage.WorkoutSession) error {
	const op = "storage.redis.CreateSession"
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return rs.Client.Set(rs.ctx, session.UserID, data, 0).Err()
}

// GetSession retrieves the workout session for a specific sessionID from Redis.
func (rs *RedisStorage) GetSession(userID string) (*storage.WorkoutSession, error) {
	const op = "storage.redis.GetSession"
	data, err := rs.Client.Get(rs.ctx, userID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, storage.ErrNoSession
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var session storage.WorkoutSession
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &session, nil
}

// UpdateSession updates a session's details (e.g., points, records) and stores it in Redis.
func (rs *RedisStorage) UpdateSession(userID string, updatedSession *storage.WorkoutSession) error {
	const op = "storage.redis.UpdateSession"
	updatedSession.LastUpdated = time.Now()
	data, err := json.Marshal(updatedSession)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = rs.Client.Set(rs.ctx, userID, data, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// DeleteSession removes a session completely from Redis.
func (rs *RedisStorage) DeleteSession(userID string) error {
	const op = "storage.redis.DeleteSession"
	err := rs.Client.Del(rs.ctx, userID).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (rs *RedisStorage) GetAllSessions() ([]*storage.WorkoutSession, error) {
	const op = "storage.redis.GetAllSessions"
	var (
		cursor   uint64
		sessions []*storage.WorkoutSession
	)

	for {
		keys, newCursor, err := rs.Client.Scan(rs.ctx, cursor, "", 10).Result()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		for _, key := range keys {
			data, err := rs.Client.Get(rs.ctx, key).Result()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					continue
				}
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			var session storage.WorkoutSession
			if err := json.Unmarshal([]byte(data), &session); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}

			sessions = append(sessions, &session)
		}

		if newCursor == 0 {
			break
		}
		cursor = newCursor
	}

	return sessions, nil
}
