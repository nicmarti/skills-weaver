package coherence

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"dungeons/internal/adventure"
)

func ts(min int) time.Time {
	return time.Date(2026, 2, 14, 18, 0, 0, 0, time.UTC).Add(time.Duration(min) * time.Minute)
}

// buildAdv creates a temp adventure with the given session journals and metadata.
func buildAdv(t *testing.T, meta *adventure.JournalMetadata, files map[int][]adventure.JournalEntry) *adventure.Adventure {
	t.Helper()
	adv := adventure.New("Coherence Test", "Test")
	adv.SetBasePath(t.TempDir())
	for sid, entries := range files {
		if err := adv.SaveSessionJournal(&adventure.SessionJournal{SessionID: sid, Entries: entries}); err != nil {
			t.Fatal(err)
		}
	}
	if meta != nil {
		if err := adv.SaveJournalMetadata(meta); err != nil {
			t.Fatal(err)
		}
	}
	return adv
}

// findRule reports whether the report contains a finding with the given rule.
func findRule(lr LayerReport, rule string) *Finding {
	for i := range lr.Findings {
		if lr.Findings[i].Rule == rule {
			return &lr.Findings[i]
		}
	}
	return nil
}

func analyze(t *testing.T, adv *adventure.Adventure) LayerReport {
	t.Helper()
	lr, err := AnalyzeIntegrity(adv)
	if err != nil {
		t.Fatalf("AnalyzeIntegrity() error = %v", err)
	}
	return lr
}

func TestIntegrityCleanJournalScores100(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 3, Categories: []string{"combat", "loot"}},
		map[int][]adventure.JournalEntry{
			1: {
				{ID: 1, SessionID: 1, Type: "combat", Content: "a", Timestamp: ts(0)},
				{ID: 2, SessionID: 1, Type: "loot", Content: "b", Timestamp: ts(5)},
			},
		},
	)
	// Declare session 1 in sessions.json so it is not flagged as orphaned,
	// and keep session_count consistent.
	if err := adv.SaveSessions(&adventure.SessionHistory{
		Sessions: []adventure.Session{{ID: 1, Status: "completed"}},
	}); err != nil {
		t.Fatal(err)
	}
	adv.SessionCount = 1

	lr := analyze(t, adv)
	if len(lr.Findings) != 0 {
		t.Errorf("clean journal should have no findings, got %+v", lr.Findings)
	}
	if lr.Score != 100 {
		t.Errorf("clean journal score = %d, want 100", lr.Score)
	}
}

func TestIntegrityDuplicateIDIsError(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 2, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			1: {
				{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
				{ID: 1, SessionID: 1, Type: "story", Content: "b", Timestamp: ts(5)},
			},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "journal.duplicate_id")
	if f == nil {
		t.Fatalf("expected journal.duplicate_id, got %+v", lr.Findings)
	}
	if f.Severity != SeverityError {
		t.Errorf("duplicate_id severity = %q, want error", f.Severity)
	}
	if !lr.HasErrors() {
		t.Error("layer with duplicate id should report HasErrors")
	}
}

func TestIntegrityIDGapIsInfo(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 4, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			1: {
				{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
				{ID: 3, SessionID: 1, Type: "story", Content: "c", Timestamp: ts(5)}, // 2 missing
			},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "journal.id_gap")
	if f == nil || f.Severity != SeverityInfo {
		t.Fatalf("expected info journal.id_gap, got %+v", lr.Findings)
	}
}

func TestIntegrityNextIDDesyncIsWarning(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 99, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			1: {{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)}},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "journal.next_id_desync")
	if f == nil || f.Severity != SeverityWarning {
		t.Fatalf("expected warning journal.next_id_desync, got %+v", lr.Findings)
	}
}

func TestIntegrityCategoriesEmptyIsWarning(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 2, Categories: []string{"story"}},
		map[int][]adventure.JournalEntry{
			1: {{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)}},
		},
	)
	// Simulate a legacy on-disk file with empty categories (SaveJournalMetadata
	// now normalizes, so the corrupted state must be written directly).
	raw := `{"next_id":2,"categories":[],"last_update":"2026-01-01T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(adv.BasePath(), "journal-meta.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	lr := analyze(t, adv)
	if findRule(lr, "journal.categories_empty") == nil {
		t.Fatalf("expected journal.categories_empty, got %+v", lr.Findings)
	}
}

func TestIntegrityMisfiledEntryIsError(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 3, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			// session_id=2 stored in the session-1 file.
			1: {
				{ID: 1, SessionID: 1, Type: "story", Content: "a", Timestamp: ts(0)},
				{ID: 2, SessionID: 2, Type: "story", Content: "misfiled", Timestamp: ts(5)},
			},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "journal.misfiled_entry")
	if f == nil || f.Severity != SeverityError {
		t.Fatalf("expected error journal.misfiled_entry, got %+v", lr.Findings)
	}
}

func TestIntegrityChronoOrderIsWarning(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 3, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			1: {
				{ID: 2, SessionID: 1, Type: "story", Content: "later", Timestamp: ts(10)},
				{ID: 1, SessionID: 1, Type: "story", Content: "earlier", Timestamp: ts(0)},
			},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "journal.chrono_order")
	if f == nil || f.Severity != SeverityWarning {
		t.Fatalf("expected warning journal.chrono_order, got %+v", lr.Findings)
	}
}

func TestIntegrityEmptyContentIsInfo(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 2, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			1: {{ID: 1, SessionID: 1, Type: "note", Content: "", Timestamp: ts(0)}},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "journal.empty_content")
	if f == nil || f.Severity != SeverityInfo {
		t.Fatalf("expected info journal.empty_content, got %+v", lr.Findings)
	}
}

func TestIntegrityOrphanedSession(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 2, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{
			// Session 5 has journal entries but sessions.json is empty.
			5: {{ID: 1, SessionID: 5, Type: "story", Content: "a", Timestamp: ts(0)}},
		},
	)
	lr := analyze(t, adv)
	f := findRule(lr, "session.orphaned")
	if f == nil || f.Severity != SeverityWarning {
		t.Fatalf("expected warning session.orphaned, got %+v", lr.Findings)
	}
}

// TestIntegrityCampaignPlanWarningsExcluded is the key contract of the scoping
// decision: a plan warning (cult keyword) must NOT pollute the integrity layer,
// while a structural error must.
func TestIntegrityCampaignPlanWarningsExcluded(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 1, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{},
	)
	// Structurally valid plan, but with a cult keyword -> exactly one warning, no error.
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{{
				Number:         1,
				Title:          "A",
				KeyEvents:      []string{"un culte ancien se réveille"},
				Goals:          []string{"survivre"},
				TargetSessions: []int{1},
			}},
		},
		PlotElements: adventure.PlotElements{
			Antagonist: adventure.Character{Name: "Vastren", Motivation: "le pouvoir"},
		},
	}
	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatal(err)
	}

	lr := analyze(t, adv)
	if findRule(lr, "campaign.plan_warning") != nil {
		t.Error("campaign plan warnings must be excluded from the integrity layer")
	}
	if findRule(lr, "campaign.plan_invalid") != nil {
		t.Errorf("structurally valid plan should produce no integrity error, got %+v", lr.Findings)
	}
}

func TestIntegrityCampaignPlanStructuralErrorIncluded(t *testing.T) {
	adv := buildAdv(t,
		&adventure.JournalMetadata{NextID: 1, Categories: []string{"x"}},
		map[int][]adventure.JournalEntry{},
	)
	// Antagonist missing name and motivation -> structural errors.
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{{Number: 1, Title: "A", KeyEvents: []string{"e"}, Goals: []string{"g"}, TargetSessions: []int{1}}},
		},
	}
	if err := adv.SaveCampaignPlan(plan); err != nil {
		t.Fatal(err)
	}

	lr := analyze(t, adv)
	f := findRule(lr, "campaign.plan_invalid")
	if f == nil || f.Severity != SeverityError {
		t.Fatalf("expected error campaign.plan_invalid, got %+v", lr.Findings)
	}
}
