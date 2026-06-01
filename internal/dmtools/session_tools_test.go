package dmtools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/coherence"
)

// fakeAgentManager implements the dmtools AgentManager interface for tests.
type fakeAgentManager struct{ calls int }

func (f *fakeAgentManager) InvokeAgent(agentName, question, contextInfo string, depth int) (string, error) {
	return "verdict", nil
}
func (f *fakeAgentManager) InvokeAgentSilent(agentName, question string, depth int) (string, error) {
	f.calls++
	return "verdict de " + agentName, nil
}

func TestRefreshNarrativeJudgment(t *testing.T) {
	adv := adventure.New("End Test", "")
	adv.SetBasePath(t.TempDir())
	if err := adv.SaveJournalMetadata(&adventure.JournalMetadata{NextID: 2, Categories: []string{"story"}}); err != nil {
		t.Fatal(err)
	}
	if err := adv.SaveSessionJournal(&adventure.SessionJournal{SessionID: 1, Entries: []adventure.JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "intro", Timestamp: time.Now()},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := adv.SaveSessions(&adventure.SessionHistory{Sessions: []adventure.Session{{ID: 1, Status: "completed"}}}); err != nil {
		t.Fatal(err)
	}

	// Nil manager => no-op, empty note.
	if note := refreshNarrativeJudgment(adv, nil); note != "" {
		t.Errorf("nil manager should produce no note, got %q", note)
	}

	// With a manager => judgment cached + success note.
	fam := &fakeAgentManager{}
	note := refreshNarrativeJudgment(adv, fam)
	if !strings.Contains(note, "rafraîchie") {
		t.Errorf("expected success note, got %q", note)
	}
	if fam.calls == 0 {
		t.Error("expected agent invocations")
	}
	cached, err := coherence.LoadNarrativeJudgment(adv)
	if err != nil || cached == nil {
		t.Fatalf("expected a cached judgment, got %v (err=%v)", cached, err)
	}
	if len(cached.Lenses) != 3 {
		t.Errorf("expected 3 lenses cached, got %d", len(cached.Lenses))
	}
}

func TestPreSessionCoherenceGate(t *testing.T) {
	adv := adventure.New("Gate Test", "")
	adv.SetBasePath(t.TempDir())

	// Minimal well-formed journal (non-empty categories => no integrity fixes).
	if err := adv.SaveJournalMetadata(&adventure.JournalMetadata{NextID: 2, Categories: []string{"story"}}); err != nil {
		t.Fatal(err)
	}
	if err := adv.SaveSessionJournal(&adventure.SessionJournal{SessionID: 1, Entries: []adventure.JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "intro", Timestamp: time.Now()},
	}}); err != nil {
		t.Fatal(err)
	}
	// 3 sessions played so a payoff due at session 2 is overdue.
	if err := adv.SaveSessions(&adventure.SessionHistory{Sessions: []adventure.Session{
		{ID: 1, Status: "completed"}, {ID: 2, Status: "completed"}, {ID: 3, Status: "completed"},
	}}); err != nil {
		t.Fatal(err)
	}
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{Acts: []adventure.Act{{Number: 1, Title: "A"}}},
		Progression:        adventure.Progression{CurrentAct: 1},
		Foreshadows: adventure.ForeshadowsContainer{
			Active: []adventure.ForeshadowLinked{
				{ID: "fsh_1", Description: "trahison", Importance: adventure.ImportanceCritical, TargetPayoffSession: 2},
			},
		},
	}
	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatal(err)
	}
	// A cached narrative judgment with a recognizable synthesis.
	if err := coherence.SaveNarrativeJudgment(adv, &coherence.NarrativeJudgment{
		GeneratedAt:    time.Now(),
		PlayedSessions: 2,
		Synthesis:      "SYNTH_MARKER les joueurs sont passifs",
	}); err != nil {
		t.Fatal(err)
	}

	human, brief := preSessionCoherenceGate(adv)

	if !strings.Contains(human, "Cohérence") {
		t.Errorf("human summary should mention coherence scores, got: %q", human)
	}
	if !strings.Contains(brief, "fsh_1") {
		t.Errorf("DM brief should surface the overdue foreshadow, got: %q", brief)
	}
	if !strings.Contains(brief, "SYNTH_MARKER") {
		t.Errorf("DM brief should inject the cached narrative synthesis, got: %q", brief)
	}
}

func TestPreSessionCoherenceGateAutoRepairs(t *testing.T) {
	adv := adventure.New("Gate Repair Test", "")
	adv.SetBasePath(t.TempDir())

	// Empty categories on disk (legacy) => repair should auto-fix. SaveJournalMetadata
	// now normalizes, so the corrupted state is written directly.
	raw := `{"next_id":2,"categories":[],"last_update":"2026-01-01T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(adv.BasePath(), "journal-meta.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	if err := adv.SaveSessionJournal(&adventure.SessionJournal{SessionID: 1, Entries: []adventure.JournalEntry{
		{ID: 1, SessionID: 1, Type: "story", Content: "intro", Timestamp: time.Now()},
	}}); err != nil {
		t.Fatal(err)
	}
	if err := adv.SaveSessions(&adventure.SessionHistory{Sessions: []adventure.Session{{ID: 1, Status: "completed"}}}); err != nil {
		t.Fatal(err)
	}

	human, _ := preSessionCoherenceGate(adv)
	if !strings.Contains(human, "auto-appliquée") {
		t.Errorf("expected auto-repair note for empty categories, got: %q", human)
	}
	// Categories must be repopulated on disk.
	meta, _ := adv.LoadJournal()
	if len(meta.Categories) == 0 {
		t.Error("categories should have been repaired")
	}
}
