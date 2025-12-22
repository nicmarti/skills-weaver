package adventure

import (
	"testing"
	"time"
)

func TestLoadSessionsNonExistent(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	history, err := adv.LoadSessions()
	if err != nil {
		t.Fatalf("LoadSessions() error = %v, want nil", err)
	}
	if len(history.Sessions) != 0 {
		t.Errorf("LoadSessions() sessions = %v, want []", history.Sessions)
	}
	if history.CurrentSession != nil {
		t.Errorf("LoadSessions() current session = %v, want nil", history.CurrentSession)
	}
}

func TestStartSession(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	session, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}

	if session == nil {
		t.Errorf("StartSession() returned nil")
	}
	if session.ID != 1 {
		t.Errorf("First session ID = %d, want 1", session.ID)
	}
	if session.Status != "active" {
		t.Errorf("Session status = %q, want 'active'", session.Status)
	}
	if session.StartedAt.IsZero() {
		t.Errorf("Session StartedAt is zero")
	}

	// Verify we can get current session
	current, err := adv.GetCurrentSession()
	if err != nil {
		t.Fatalf("GetCurrentSession() error = %v, want nil", err)
	}
	if current == nil {
		t.Fatal("GetCurrentSession() returned nil, want active session")
	}
	if current.ID != 1 {
		t.Errorf("Current session ID = %d, want 1", current.ID)
	}
}

func TestStartSessionMultiple(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Start and end first session
	s1, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() 1st error = %v, want nil", err)
	}
	if s1.ID != 1 {
		t.Errorf("First session ID = %d, want 1", s1.ID)
	}

	_, err = adv.EndSession("First session completed")
	if err != nil {
		t.Fatalf("EndSession() 1st error = %v, want nil", err)
	}

	// Start and end second session
	s2, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() 2nd error = %v, want nil", err)
	}
	if s2.ID != 2 {
		t.Errorf("Second session ID = %d, want 2", s2.ID)
	}

	_, err = adv.EndSession("Second session completed")
	if err != nil {
		t.Fatalf("EndSession() 2nd error = %v, want nil", err)
	}

	// Verify session count
	sessions, err := adv.GetAllSessions()
	if err != nil {
		t.Fatalf("GetAllSessions() error = %v, want nil", err)
	}
	if len(sessions) != 2 {
		t.Errorf("GetAllSessions() count = %d, want 2", len(sessions))
	}
}

func TestStartSessionAlreadyActive(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Start first session
	_, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() 1st error = %v, want nil", err)
	}

	// Try to start another session without ending the first
	_, err = adv.StartSession()
	if err == nil {
		t.Errorf("StartSession() should fail when session already active, got nil error")
	}
}

func TestEndSession(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Start a session
	_, err := adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}

	// End the session
	summary := "Explored the dungeon"
	session, err := adv.EndSession(summary)
	if err != nil {
		t.Fatalf("EndSession() error = %v, want nil", err)
	}

	if session == nil {
		t.Fatal("EndSession() returned nil, want completed session")
	}
	if session.Status != "completed" {
		t.Errorf("Session status = %q, want %q", session.Status, "completed")
	}
	if session.Summary != summary {
		t.Errorf("Session summary = %q, want %q", session.Summary, summary)
	}
	if session.EndedAt.IsZero() {
		t.Errorf("Session EndedAt is zero, want non-zero time")
	}

	// Verify no current session
	current, err := adv.GetCurrentSession()
	if err != nil {
		t.Fatalf("GetCurrentSession() error = %v, want nil", err)
	}
	if current != nil {
		t.Errorf("GetCurrentSession() = %v, want nil after ending session", current)
	}
}

func TestEndSessionNoActive(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	_, err := adv.EndSession("Summary")
	if err == nil {
		t.Errorf("EndSession() should fail when no active session")
	}
}

func TestGetCurrentSession(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Initially no current session
	current, err := adv.GetCurrentSession()
	if err != nil {
		t.Fatalf("GetCurrentSession() error = %v, want nil", err)
	}
	if current != nil {
		t.Errorf("GetCurrentSession() = %v, want nil initially", current)
	}

	// Start a session
	_, err = adv.StartSession()
	if err != nil {
		t.Fatalf("StartSession() error = %v, want nil", err)
	}

	// Now should have current session
	current, err = adv.GetCurrentSession()
	if err != nil {
		t.Fatalf("GetCurrentSession() error = %v, want nil", err)
	}
	if current == nil {
		t.Fatal("GetCurrentSession() = nil, want active session")
	}
	if current.ID != 1 {
		t.Errorf("GetCurrentSession().ID = %d, want 1", current.ID)
	}

	// End session
	adv.EndSession("Done")

	// No current session after ending
	current, err = adv.GetCurrentSession()
	if err != nil {
		t.Fatalf("GetCurrentSession() error = %v, want nil", err)
	}
	if current != nil {
		t.Errorf("GetCurrentSession() = %v, want nil after ending", current)
	}
}

func TestGetSession(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.EndSession("Session 1")
	adv.StartSession()
	adv.EndSession("Session 2")

	// Get specific session
	s1, err := adv.GetSession(1)
	if err != nil {
		t.Fatalf("GetSession(1) error = %v", err)
	}
	if s1.Summary != "Session 1" {
		t.Errorf("Session 1 summary = %q, want 'Session 1'", s1.Summary)
	}

	// Get non-existent session
	_, err = adv.GetSession(999)
	if err == nil {
		t.Errorf("GetSession(999) should fail for non-existent session")
	}
}

func TestGetAllSessions(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Create multiple sessions
	for i := 1; i <= 3; i++ {
		adv.StartSession()
		adv.EndSession("Session " + string(rune(48+i)))
	}

	sessions, err := adv.GetAllSessions()
	if err != nil {
		t.Fatalf("GetAllSessions() error = %v", err)
	}

	if len(sessions) != 3 {
		t.Errorf("GetAllSessions() returned %d sessions, want 3", len(sessions))
	}

	// Verify sessions are sorted by ID
	for i := 0; i < len(sessions)-1; i++ {
		if sessions[i].ID > sessions[i+1].ID {
			t.Errorf("Sessions not sorted by ID")
		}
	}
}

func TestUpdateSessionNotes(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.EndSession("Summary")

	notes := "The party discovered a hidden door"
	err := adv.UpdateSessionNotes(1, notes)
	if err != nil {
		t.Fatalf("UpdateSessionNotes() error = %v", err)
	}

	session, _ := adv.GetSession(1)
	if session.Notes != notes {
		t.Errorf("Session notes = %q, want %q", session.Notes, notes)
	}
}

func TestUpdateSessionLocation(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	location := "The Goblin Caves"
	err := adv.UpdateSessionLocation(1, location)
	if err != nil {
		t.Fatalf("UpdateSessionLocation() error = %v", err)
	}

	session, _ := adv.GetCurrentSession()
	if session.Location != location {
		t.Errorf("Session location = %q, want %q", session.Location, location)
	}
}

func TestAwardXP(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.EndSession("Summary")

	err := adv.AwardXP(1, 200)
	if err != nil {
		t.Fatalf("AwardXP() error = %v", err)
	}

	session, _ := adv.GetSession(1)
	if session.XPAwarded != 200 {
		t.Errorf("XPAwarded = %d, want 200", session.XPAwarded)
	}

	// Award more XP
	adv.AwardXP(1, 300)
	session, _ = adv.GetSession(1)
	if session.XPAwarded != 500 {
		t.Errorf("XPAwarded after second award = %d, want 500", session.XPAwarded)
	}
}

func TestRecordGoldFound(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.EndSession("Summary")

	err := adv.RecordGoldFound(1, 500)
	if err != nil {
		t.Fatalf("RecordGoldFound() error = %v", err)
	}

	session, _ := adv.GetSession(1)
	if session.GoldFound != 500 {
		t.Errorf("GoldFound = %d, want 500", session.GoldFound)
	}
}

func TestGetTotalXPAwarded(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Create sessions and award XP
	adv.StartSession()
	adv.EndSession("S1")
	adv.AwardXP(1, 100)

	adv.StartSession()
	adv.EndSession("S2")
	adv.AwardXP(2, 200)

	adv.StartSession()
	adv.EndSession("S3")
	adv.AwardXP(3, 300)

	total, err := adv.GetTotalXPAwarded()
	if err != nil {
		t.Fatalf("GetTotalXPAwarded() error = %v", err)
	}

	if total != 600 {
		t.Errorf("GetTotalXPAwarded() = %d, want 600", total)
	}
}

func TestGetTotalGoldFound(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.EndSession("S1")
	adv.RecordGoldFound(1, 250)

	adv.StartSession()
	adv.EndSession("S2")
	adv.RecordGoldFound(2, 500)

	total, err := adv.GetTotalGoldFound()
	if err != nil {
		t.Fatalf("GetTotalGoldFound() error = %v", err)
	}

	if total != 750 {
		t.Errorf("GetTotalGoldFound() = %d, want 750", total)
	}
}

func TestGetTotalPlayTime(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	_, _ = adv.StartSession()
	time.Sleep(10 * time.Millisecond)
	adv.EndSession("S1")

	_, _ = adv.StartSession()
	time.Sleep(10 * time.Millisecond)
	adv.EndSession("S2")

	total, err := adv.GetTotalPlayTime()
	if err != nil {
		t.Fatalf("GetTotalPlayTime() error = %v", err)
	}

	// Should be at least 20ms
	if total < 20*time.Millisecond {
		t.Errorf("GetTotalPlayTime() = %v, want >= 20ms", total)
	}
}

func TestGetTotalPlayTimeWithActiveSession(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.StartSession()
	adv.EndSession("Completed")
	adv.StartSession()
	// Don't end the second session

	total, err := adv.GetTotalPlayTime()
	if err != nil {
		t.Fatalf("GetTotalPlayTime() error = %v", err)
	}

	// Should only count completed session
	if total == 0 {
		t.Errorf("GetTotalPlayTime() should count completed sessions")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "minutes only",
			duration: 45 * time.Minute,
			expected: "45m",
		},
		{
			name:     "hours and minutes",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2h30m",
		},
		{
			name:     "zero",
			duration: 0,
			expected: "0m",
		},
		{
			name:     "seconds ignored",
			duration: 1*time.Hour + 5*time.Minute + 30*time.Second,
			expected: "1h05m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}

func TestSessionJSON(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Create and manipulate session
	adv.StartSession()
	adv.UpdateSessionLocation(1, "Dungeon")
	adv.UpdateSessionNotes(1, "Notes here")
	adv.AwardXP(1, 100)
	adv.RecordGoldFound(1, 50)
	adv.EndSession("Summary")

	// Load and verify all fields persisted
	session, _ := adv.GetSession(1)
	if session.Location != "Dungeon" {
		t.Errorf("Location not persisted: %q", session.Location)
	}
	if session.Notes != "Notes here" {
		t.Errorf("Notes not persisted: %q", session.Notes)
	}
	if session.XPAwarded != 100 {
		t.Errorf("XPAwarded not persisted: %d", session.XPAwarded)
	}
	if session.GoldFound != 50 {
		t.Errorf("GoldFound not persisted: %d", session.GoldFound)
	}
}
