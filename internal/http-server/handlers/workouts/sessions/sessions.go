package session

import (
	"sync"
	"time"
)

// WorkoutSession represents a user's workout session.
type WorkoutSession struct {
	UserID      string
	SessionID   string
	StartTime   time.Time
	IsActive    bool
	LastUpdated time.Time
}

// Manager manages workout sessions with thread-safe operations.
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*WorkoutSession
}

// NewSessionManager creates a new instance of Manager.
func NewSessionManager() *Manager {
	return &Manager{
		sessions: make(map[string]*WorkoutSession),
	}
}

// StartSession initializes a new workout session for a user.
func (sm *Manager) StartSession(userID, sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[userID] = &WorkoutSession{
		UserID:      userID,
		SessionID:   sessionID,
		StartTime:   time.Now(),
		IsActive:    true,
		LastUpdated: time.Now(),
	}
}

// UpdateSession updates points data in user session.
func (sm *Manager) UpdateSession(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[userID].LastUpdated = time.Now()
}

// EndSession marks a user's session as inactive and removes it from the manager.
func (sm *Manager) EndSession(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[userID]; exists {
		session.IsActive = false
		session.LastUpdated = time.Now()
		delete(sm.sessions, userID)
	}
}

// GetSession retrieves the workout session for a specific user.
func (sm *Manager) GetSession(userID string) *WorkoutSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if session, exists := sm.sessions[userID]; exists {
		return session
	}
	return nil
}

// GetAllActiveSessions returns a list of active sessions that have exceeded the given lifetime.
func (sm *Manager) GetAllActiveSessions(lt time.Duration) []*WorkoutSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	var sessions []*WorkoutSession
	for _, session := range sm.sessions {
		if session.IsActive && time.Since(session.LastUpdated) > lt {
			sessions = append(sessions, session)
		}
	}
	return sessions
}
