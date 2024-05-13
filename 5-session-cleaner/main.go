//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

const (
	idleSessionTimeout    = 5 * time.Second
	sessionCleanerQuantom = time.Second
)

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions    map[string]Session
	StopCleaner chan struct{}
	mutex       sync.RWMutex
}

// Session stores the session's data
type Session struct {
	Data     map[string]interface{}
	LastUsed time.Time
}

func (s *Session) IsIdle(timeout time.Duration) bool {
	return time.Since(s.LastUsed) >= timeout
}

func (m *SessionManager) runSessionCleaner(quantom time.Duration, timeout time.Duration) {
	ticker := time.NewTicker(quantom)
	defer ticker.Stop()

	for {
		select {
		case <-m.StopCleaner:
			return
		case <-ticker.C:
			m.cleanIdleSessions(timeout)

		}
	}
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions:    make(map[string]Session),
		StopCleaner: make(chan struct{}),
	}
	// run session cleaner in the background
	go m.runSessionCleaner(sessionCleanerQuantom, idleSessionTimeout)

	return m
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[sessionID] = Session{
		Data:     make(map[string]interface{}),
		LastUsed: time.Now(),
	}

	return sessionID, nil
}

// Loops over all sessions and cleans inactive ones
func (m *SessionManager) cleanIdleSessions(timeout time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// run over all sessions and clean idle ones
	for sessionID, session := range m.sessions {
		if session.IsIdle(timeout) {
			delete(m.sessions, sessionID)
			log.Println("Cleaned session: ", sessionID)
		}
	}
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.mutex.RLock()
	_, ok := m.sessions[sessionID]
	m.mutex.RUnlock()

	if !ok {
		return ErrSessionNotFound
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[sessionID] = Session{
		Data:     data,
		LastUsed: time.Now(),
	}

	return nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
