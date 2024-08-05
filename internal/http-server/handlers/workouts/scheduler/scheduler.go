package scheduler

import (
	"GYMBRO/internal/config"
	session "GYMBRO/internal/http-server/handlers/workouts/sessions"
	"GYMBRO/internal/storage"
	"log/slog"
	"time"
)

// Scheduler manages the scheduling of workout sessions and their automatic ending.
type Scheduler struct {
	Log  *slog.Logger
	Repo storage.WorkoutRepository
	Sm   *session.Manager
	Cfg  *config.Config
}

// NewScheduler creates a new instance of Scheduler with the provided dependencies.
func NewScheduler(log *slog.Logger, repo storage.WorkoutRepository, sm *session.Manager, cfg *config.Config) *Scheduler {
	return &Scheduler{
		Log:  log,
		Repo: repo,
		Sm:   sm,
		Cfg:  cfg,
	}
}

// Start begins the scheduler's periodic tasks based on the configured interval.
func (s *Scheduler) Start() {
	ticker := time.NewTicker(s.Cfg.SchedulerInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ended := s.checkAndEndWorkouts()
				s.Log.Info("Scheduler checkAndEndWorkouts finished", slog.Any("ended_sessions", ended))

			}
		}
	}()
}

// checkAndEndWorkouts checks for active workout sessions and ends them if they have expired.
func (s *Scheduler) checkAndEndWorkouts() []string {
	const op = "handlers.workouts.scheduler.checkAndEndWorkouts"
	s.Log = s.Log.With(slog.String("op", op))
	var ended []string
	// Retrieve all active sessions that have exceeded the session lifetime.
	sessions := s.Sm.GetAllActiveSessions(s.Cfg.SessionLifetime)

	for _, session := range sessions {

		// Attempt to end the workout in the repository.
		err := s.Repo.EndWorkout(session.SessionID)
		if err != nil {
			s.Log.Error("Failed to end workout", slog.String("session_id", session.SessionID), slog.Any("error", err))
			continue
		}
		s.Log.Debug("Auto ended workout", slog.String("session_id", session.SessionID))
		ended = append(ended, session.SessionID)
		// End the session in the session manager.
		s.Sm.EndSession(session.SessionID)
	}
	return ended
}
