package coherence

import (
	"testing"

	"dungeons/internal/adventure"
)

// narrativeAdv builds a temp adventure with journal files, sessions (with
// summaries) and an optional campaign plan.
func narrativeAdv(t *testing.T, files map[int][]adventure.JournalEntry, sessions []adventure.Session, plan *adventure.CampaignPlan) *adventure.Adventure {
	t.Helper()
	adv := adventure.New("Narrative Test", "Test")
	adv.SetBasePath(t.TempDir())
	if err := adv.SaveJournalMetadata(&adventure.JournalMetadata{NextID: 1, Categories: []string{"story"}}); err != nil {
		t.Fatal(err)
	}
	for sid, entries := range files {
		if err := adv.SaveSessionJournal(&adventure.SessionJournal{SessionID: sid, Entries: entries}); err != nil {
			t.Fatal(err)
		}
	}
	if err := adv.SaveSessions(&adventure.SessionHistory{Sessions: sessions}); err != nil {
		t.Fatal(err)
	}
	if plan != nil {
		if err := adv.SaveCampaignPlan(plan); err != nil {
			t.Fatal(err)
		}
	}
	return adv
}

func TestNarrativeThinSession(t *testing.T) {
	adv := narrativeAdv(t,
		map[int][]adventure.JournalEntry{
			1: {{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)}},
			2: {
				{ID: 2, SessionID: 2, Type: "story", Content: "b", Timestamp: ts(5)},
				{ID: 3, SessionID: 2, Type: "story", Content: "c", Timestamp: ts(6)},
				{ID: 4, SessionID: 2, Type: "story", Content: "d", Timestamp: ts(7)},
			},
		},
		[]adventure.Session{{ID: 1, Status: "completed"}, {ID: 2, Status: "completed"}},
		nil,
	)
	lr, err := AnalyzeNarrative(adv)
	if err != nil {
		t.Fatalf("AnalyzeNarrative() error = %v", err)
	}
	f := findRule(lr, "narrative.thin_session")
	if f == nil || f.Severity != SeverityInfo {
		t.Fatalf("expected info narrative.thin_session for session 1, got %+v", lr.Findings)
	}
	if f.Context["session"] != 1 {
		t.Errorf("expected session 1 flagged thin, got %v", f.Context["session"])
	}
}

func TestNarrativeOpenLoops(t *testing.T) {
	plan := &adventure.CampaignPlan{
		Progression: adventure.Progression{ActiveThreads: []string{"thread_a", "thread_b"}},
		Foreshadows: adventure.ForeshadowsContainer{
			Active: []adventure.ForeshadowLinked{{ID: "fsh_1", Importance: adventure.ImportanceMajor}},
		},
	}
	adv := narrativeAdv(t,
		map[int][]adventure.JournalEntry{
			1: {{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)}, {ID: 2, SessionID: 1, Type: "story", Content: "b", Timestamp: ts(1)}, {ID: 3, SessionID: 1, Type: "story", Content: "c", Timestamp: ts(2)}},
		},
		[]adventure.Session{{ID: 1, Status: "completed"}},
		plan,
	)
	lr, err := AnalyzeNarrative(adv)
	if err != nil {
		t.Fatalf("AnalyzeNarrative() error = %v", err)
	}
	f := findRule(lr, "narrative.open_loops")
	if f == nil {
		t.Fatalf("expected narrative.open_loops, got %+v", lr.Findings)
	}
	if f.Context["open_threads"] != 2 || f.Context["pending_foreshadows"] != 1 {
		t.Errorf("open loops counts wrong: %+v", f.Context)
	}
}

func TestNarrativeNeverEmitsErrors(t *testing.T) {
	plan := &adventure.CampaignPlan{
		Progression: adventure.Progression{ActiveThreads: []string{"t"}},
	}
	adv := narrativeAdv(t,
		map[int][]adventure.JournalEntry{1: {{ID: 1, SessionID: 1, Type: "note", Content: "", Timestamp: ts(0)}}},
		[]adventure.Session{{ID: 1, Status: "completed"}},
		plan,
	)
	lr, err := AnalyzeNarrative(adv)
	if err != nil {
		t.Fatalf("AnalyzeNarrative() error = %v", err)
	}
	for _, f := range lr.Findings {
		if f.Severity == SeverityError {
			t.Errorf("narrative layer must never emit errors, got %+v", f)
		}
	}
	if lr.HasErrors() {
		t.Error("narrative layer should never report HasErrors")
	}
}

func TestBuildNarrativeBriefDigest(t *testing.T) {
	adv := narrativeAdv(t,
		map[int][]adventure.JournalEntry{
			1: {
				{ID: 1, SessionID: 1, Type: "combat", Content: "Round 1: coup", Timestamp: ts(0)},
				{ID: 2, SessionID: 1, Type: "quest", Content: "Quête acceptée", Timestamp: ts(1)},
				{ID: 3, SessionID: 1, Type: "story", Content: "Plot point completed: x", Timestamp: ts(2)},
			},
		},
		[]adventure.Session{{ID: 1, Status: "completed", Summary: "recap session 1"}},
		nil,
	)
	brief, err := BuildNarrativeBrief(adv)
	if err != nil {
		t.Fatalf("BuildNarrativeBrief() error = %v", err)
	}

	if brief.PlayedSessions != 1 {
		t.Errorf("PlayedSessions = %d, want 1", brief.PlayedSessions)
	}
	if len(brief.Sessions) != 1 {
		t.Fatalf("expected 1 session digest, got %d", len(brief.Sessions))
	}
	sd := brief.Sessions[0]
	if sd.EntryCount != 3 {
		t.Errorf("EntryCount = %d, want 3", sd.EntryCount)
	}
	if sd.CombatLogLines != 1 {
		t.Errorf("CombatLogLines = %d, want 1", sd.CombatLogLines)
	}
	// quest type + "Plot point completed" content -> 2 progression markers.
	if sd.ProgressionMarkers != 2 {
		t.Errorf("ProgressionMarkers = %d, want 2", sd.ProgressionMarkers)
	}
	if sd.Summary != "recap session 1" {
		t.Errorf("Summary = %q, want 'recap session 1'", sd.Summary)
	}
	if brief.TypeTotals["combat"] != 1 || brief.TypeTotals["quest"] != 1 || brief.TypeTotals["story"] != 1 {
		t.Errorf("TypeTotals wrong: %+v", brief.TypeTotals)
	}
	if len(brief.Notes) != len(narrativeDataNotes) {
		t.Errorf("Notes count = %d, want %d", len(brief.Notes), len(narrativeDataNotes))
	}
}

func TestEngineVersionOrV0(t *testing.T) {
	if engineVersionOrV0("") != "v0" {
		t.Error("empty engine version should map to v0 (legacy)")
	}
	if engineVersionOrV0("abc123") != "abc123" {
		t.Error("non-empty engine version should pass through")
	}
}

func TestIsProgressionMarker(t *testing.T) {
	cases := []struct {
		entry adventure.JournalEntry
		want  bool
	}{
		{adventure.JournalEntry{Type: "quest", Content: "x"}, true},
		{adventure.JournalEntry{Type: "levelup", Content: "x"}, true},
		{adventure.JournalEntry{Type: "story", Content: "Plot point completed: y"}, true},
		{adventure.JournalEntry{Type: "story", Content: "New narrative thread: z"}, true},
		{adventure.JournalEntry{Type: "combat", Content: "Round 3 coup"}, false},
		{adventure.JournalEntry{Type: "story", Content: "le groupe avance"}, false},
	}
	for _, c := range cases {
		if got := isProgressionMarker(c.entry); got != c.want {
			t.Errorf("isProgressionMarker(%q/%q) = %v, want %v", c.entry.Type, c.entry.Content, got, c.want)
		}
	}
}
