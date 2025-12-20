package adventure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Session represents a game session.
type Session struct {
	ID        int       `json:"id"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Summary   string    `json:"summary,omitempty"`
	Location  string    `json:"location,omitempty"`  // Where the session took place in-game
	Notes     string    `json:"notes,omitempty"`     // DM notes
	XPAwarded int       `json:"xp_awarded,omitempty"`
	GoldFound int       `json:"gold_found,omitempty"`
	Status    string    `json:"status"` // active, completed, abandoned
}

// SessionHistory holds all sessions for an adventure.
type SessionHistory struct {
	Sessions      []Session `json:"sessions"`
	CurrentSession *int     `json:"current_session,omitempty"` // ID of active session
}

// LoadSessions loads the session history.
func (a *Adventure) LoadSessions() (*SessionHistory, error) {
	path := filepath.Join(a.basePath, "sessions.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &SessionHistory{
			Sessions:       []Session{},
			CurrentSession: nil,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading sessions.json: %w", err)
	}

	var history SessionHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("parsing sessions.json: %w", err)
	}

	return &history, nil
}

// SaveSessions saves the session history.
func (a *Adventure) SaveSessions(history *SessionHistory) error {
	path := filepath.Join(a.basePath, "sessions.json")
	return a.saveJSON(path, history)
}

// StartSession begins a new game session.
func (a *Adventure) StartSession() (*Session, error) {
	history, err := a.LoadSessions()
	if err != nil {
		return nil, err
	}

	// Check if there's already an active session
	if history.CurrentSession != nil {
		return nil, fmt.Errorf("session %d is already active, end it first", *history.CurrentSession)
	}

	// Create new session
	nextID := 1
	if len(history.Sessions) > 0 {
		nextID = history.Sessions[len(history.Sessions)-1].ID + 1
	}

	session := Session{
		ID:        nextID,
		StartedAt: time.Now(),
		Status:    "active",
	}

	history.Sessions = append(history.Sessions, session)
	history.CurrentSession = &session.ID

	// Update adventure
	a.SessionCount = len(history.Sessions)
	a.LastPlayed = time.Now()

	// Save both
	if err := a.SaveSessions(history); err != nil {
		return nil, err
	}
	if err := a.Save(filepath.Dir(a.basePath)); err != nil {
		return nil, err
	}

	// Log to journal
	a.LogEvent("session", fmt.Sprintf("Session %d démarrée", nextID))

	return &session, nil
}

// EndSession ends the current game session.
func (a *Adventure) EndSession(summary string) (*Session, error) {
	history, err := a.LoadSessions()
	if err != nil {
		return nil, err
	}

	if history.CurrentSession == nil {
		return nil, fmt.Errorf("no active session to end")
	}

	// Find and update the session
	var session *Session
	for i := range history.Sessions {
		if history.Sessions[i].ID == *history.CurrentSession {
			session = &history.Sessions[i]
			break
		}
	}

	if session == nil {
		return nil, fmt.Errorf("session %d not found", *history.CurrentSession)
	}

	session.EndedAt = time.Now()
	session.Status = "completed"
	session.Summary = summary
	session.Duration = formatDuration(session.EndedAt.Sub(session.StartedAt))

	history.CurrentSession = nil

	// Update adventure
	a.LastPlayed = time.Now()

	// Save both
	if err := a.SaveSessions(history); err != nil {
		return nil, err
	}
	if err := a.Save(filepath.Dir(a.basePath)); err != nil {
		return nil, err
	}

	// Log to journal
	a.LogEvent("session", fmt.Sprintf("Session %d terminée - %s", session.ID, summary))

	return session, nil
}

// GetCurrentSession returns the active session, if any.
func (a *Adventure) GetCurrentSession() (*Session, error) {
	history, err := a.LoadSessions()
	if err != nil {
		return nil, err
	}

	if history.CurrentSession == nil {
		return nil, nil
	}

	for _, s := range history.Sessions {
		if s.ID == *history.CurrentSession {
			return &s, nil
		}
	}

	return nil, nil
}

// GetSession returns a specific session by ID.
func (a *Adventure) GetSession(id int) (*Session, error) {
	history, err := a.LoadSessions()
	if err != nil {
		return nil, err
	}

	for _, s := range history.Sessions {
		if s.ID == id {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("session %d not found", id)
}

// GetAllSessions returns all sessions.
func (a *Adventure) GetAllSessions() ([]Session, error) {
	history, err := a.LoadSessions()
	if err != nil {
		return nil, err
	}

	// Sort by ID (chronological)
	sort.Slice(history.Sessions, func(i, j int) bool {
		return history.Sessions[i].ID < history.Sessions[j].ID
	})

	return history.Sessions, nil
}

// UpdateSessionNotes updates the notes of a session.
func (a *Adventure) UpdateSessionNotes(sessionID int, notes string) error {
	history, err := a.LoadSessions()
	if err != nil {
		return err
	}

	for i := range history.Sessions {
		if history.Sessions[i].ID == sessionID {
			history.Sessions[i].Notes = notes
			return a.SaveSessions(history)
		}
	}

	return fmt.Errorf("session %d not found", sessionID)
}

// UpdateSessionLocation updates the in-game location of a session.
func (a *Adventure) UpdateSessionLocation(sessionID int, location string) error {
	history, err := a.LoadSessions()
	if err != nil {
		return err
	}

	for i := range history.Sessions {
		if history.Sessions[i].ID == sessionID {
			history.Sessions[i].Location = location
			return a.SaveSessions(history)
		}
	}

	return fmt.Errorf("session %d not found", sessionID)
}

// AwardXP records XP awarded during a session.
func (a *Adventure) AwardXP(sessionID int, xp int) error {
	history, err := a.LoadSessions()
	if err != nil {
		return err
	}

	for i := range history.Sessions {
		if history.Sessions[i].ID == sessionID {
			history.Sessions[i].XPAwarded += xp
			a.LogEvent("xp", fmt.Sprintf("Session %d: %d XP attribués", sessionID, xp))
			return a.SaveSessions(history)
		}
	}

	return fmt.Errorf("session %d not found", sessionID)
}

// RecordGoldFound records gold found during a session.
func (a *Adventure) RecordGoldFound(sessionID int, gold int) error {
	history, err := a.LoadSessions()
	if err != nil {
		return err
	}

	for i := range history.Sessions {
		if history.Sessions[i].ID == sessionID {
			history.Sessions[i].GoldFound += gold
			return a.SaveSessions(history)
		}
	}

	return fmt.Errorf("session %d not found", sessionID)
}

// GetTotalXPAwarded returns total XP across all sessions.
func (a *Adventure) GetTotalXPAwarded() (int, error) {
	sessions, err := a.GetAllSessions()
	if err != nil {
		return 0, err
	}

	total := 0
	for _, s := range sessions {
		total += s.XPAwarded
	}

	return total, nil
}

// GetTotalGoldFound returns total gold across all sessions.
func (a *Adventure) GetTotalGoldFound() (int, error) {
	sessions, err := a.GetAllSessions()
	if err != nil {
		return 0, err
	}

	total := 0
	for _, s := range sessions {
		total += s.GoldFound
	}

	return total, nil
}

// GetTotalPlayTime returns total time spent playing.
func (a *Adventure) GetTotalPlayTime() (time.Duration, error) {
	sessions, err := a.GetAllSessions()
	if err != nil {
		return 0, err
	}

	var total time.Duration
	for _, s := range sessions {
		if s.Status == "completed" && !s.EndedAt.IsZero() {
			total += s.EndedAt.Sub(s.StartedAt)
		}
	}

	return total, nil
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
