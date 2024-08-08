package services

import (
	"GYMBRO/internal/config"
	"GYMBRO/internal/storage"
	"log/slog"
	"time"
)

type SessionScheduler struct {
	sessionRepo       storage.SessionRepository
	woRepo            storage.WorkoutRepository
	checkInterval     time.Duration
	inactivityTimeout time.Duration
	log               *slog.Logger
}

func NewSessionScheduler(sessionRepo storage.SessionRepository, woRepo storage.WorkoutRepository, cfg *config.Config, log *slog.Logger) *SessionScheduler {
	return &SessionScheduler{
		sessionRepo:       sessionRepo,
		woRepo:            woRepo,
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
		s.log.Error("Scheduler cant get sessions", slog.Any("error", err))
	}

	endedSessions := 0

	for _, session := range sessions {
		if session != nil && time.Since(session.LastUpdated) > inactivityDuration {
			session.IsActive = false

			err := s.woRepo.SaveWorkout(session)
			if err != nil {
				s.log.Error("Scheduler cant save workout", slog.Any("error", err))
				continue
			}
			err = s.sessionRepo.DeleteSession(session.UserID)
			if err != nil {
				s.log.Error("Scheduler cant delete session", slog.Any("error", err))
			}
			endedSessions++
		}
	}

	return endedSessions
}
