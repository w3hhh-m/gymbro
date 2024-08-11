package services

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/storage"
	"log/slog"
	"time"
)

type SessionScheduler struct {
	sessionRepo       storage.SessionRepository
	workoutRepo       storage.WorkoutRepository
	checkInterval     time.Duration
	inactivityTimeout time.Duration
	log               *slog.Logger
}

func NewSessionScheduler(sessionRepo storage.SessionRepository, workoutRepo storage.WorkoutRepository, cfg *config.Config, log *slog.Logger) *SessionScheduler {
	return &SessionScheduler{
		sessionRepo:       sessionRepo,
		workoutRepo:       workoutRepo,
		checkInterval:     cfg.SchedulerInterval,
		inactivityTimeout: cfg.SessionLifetime,
		log:               log,
	}
}

func (s *SessionScheduler) Start() {
	ticker := time.NewTicker(s.checkInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ended := s.processInactiveSessions(s.inactivityTimeout)
				s.log.Info("Scheduler processInactiveSessions finished", slog.Int("ended_sessions", ended))
			}
		}
	}()
}

func (s *SessionScheduler) processInactiveSessions(inactivityDuration time.Duration) int {
	sessions, err := s.sessionRepo.GetAllSessions()
	if err != nil {
		s.log.Error("Scheduler cant GET sessions", slog.Any("error", err))
	}

	endedSessions := 0

	for _, session := range sessions {
		if session != nil && time.Since(session.LastUpdated) > inactivityDuration {

			err := s.workoutRepo.SaveWorkout(session)
			if err != nil {
				s.log.Error("Scheduler cant SAVE workout", slog.Any("error", err))
				continue
			}
			err = s.sessionRepo.DeleteSession(&session.UserID)
			if err != nil {
				s.log.Error("Scheduler cant DELETE session", slog.Any("error", err))
			}
			endedSessions++
		}
	}

	return endedSessions
}
