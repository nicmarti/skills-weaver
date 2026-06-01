package coherence

import (
	"testing"

	"dungeons/internal/adventure"
	"dungeons/internal/npc"
	"dungeons/internal/npcmanager"
)

// driftAdv builds a temp adventure with a campaign plan and a set of played
// sessions, plus a minimal valid journal/meta so AnalyzeDrift's loaders succeed.
func driftAdv(t *testing.T, plan *adventure.CampaignPlan, played int) *adventure.Adventure {
	t.Helper()
	adv := adventure.New("Drift Test", "Test")
	adv.SetBasePath(t.TempDir())
	if err := adv.SaveJournalMetadata(&adventure.JournalMetadata{NextID: 1, Categories: []string{"story"}}); err != nil {
		t.Fatal(err)
	}
	sessions := make([]adventure.Session, 0, played)
	for i := 1; i <= played; i++ {
		sessions = append(sessions, adventure.Session{ID: i, Status: "completed"})
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

func driftReport(t *testing.T, adv *adventure.Adventure) LayerReport {
	t.Helper()
	lr, err := AnalyzeDrift(adv)
	if err != nil {
		t.Fatalf("AnalyzeDrift() error = %v", err)
	}
	return lr
}

func TestDriftNoPlanIsClean(t *testing.T) {
	adv := driftAdv(t, nil, 3)
	lr := driftReport(t, adv)
	if len(lr.Findings) != 0 || lr.Score != 100 {
		t.Errorf("no plan should yield a clean drift layer, got score=%d findings=%+v", lr.Score, lr.Findings)
	}
}

func TestDriftOverduePayoffSeverityByImportance(t *testing.T) {
	cases := []struct {
		imp  adventure.Importance
		want Severity
	}{
		{adventure.ImportanceCritical, SeverityError},
		{adventure.ImportanceMajor, SeverityWarning},
		{adventure.ImportanceModerate, SeverityInfo},
	}
	for _, c := range cases {
		t.Run(string(c.imp), func(t *testing.T) {
			plan := &adventure.CampaignPlan{
				Foreshadows: adventure.ForeshadowsContainer{
					Active: []adventure.ForeshadowLinked{
						{ID: "fsh_x", Importance: c.imp, TargetPayoffSession: 3},
					},
				},
			}
			adv := driftAdv(t, plan, 5) // played 5 > payoff 3 -> overdue
			lr := driftReport(t, adv)
			f := findRule(lr, "foreshadow.overdue_payoff")
			if f == nil {
				t.Fatalf("expected foreshadow.overdue_payoff, got %+v", lr.Findings)
			}
			if f.Severity != c.want {
				t.Errorf("severity = %q, want %q", f.Severity, c.want)
			}
		})
	}
}

func TestDriftPayoffNotYetDueIsClean(t *testing.T) {
	plan := &adventure.CampaignPlan{
		Foreshadows: adventure.ForeshadowsContainer{
			Active: []adventure.ForeshadowLinked{
				{ID: "fsh_future", Importance: adventure.ImportanceCritical, TargetPayoffSession: 8},
			},
		},
	}
	adv := driftAdv(t, plan, 6) // payoff 8 > played 6 -> not overdue
	lr := driftReport(t, adv)
	if findRule(lr, "foreshadow.overdue_payoff") != nil {
		t.Errorf("payoff not yet due should not be flagged, got %+v", lr.Findings)
	}
}

func TestDriftActWindowElapsed(t *testing.T) {
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{
				{Number: 1, Title: "A", Status: "in_progress", TargetSessions: []int{1, 2, 3}},
			},
		},
		Progression: adventure.Progression{CurrentAct: 1},
	}
	adv := driftAdv(t, plan, 5) // max target 3 < played 5, not completed
	lr := driftReport(t, adv)
	f := findRule(lr, "act.window_elapsed")
	if f == nil || f.Severity != SeverityWarning {
		t.Fatalf("expected warning act.window_elapsed, got %+v", lr.Findings)
	}
}

func TestDriftCompletedActNotFlagged(t *testing.T) {
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{
				{Number: 1, Title: "A", Status: "completed", TargetSessions: []int{1, 2}},
			},
		},
		Progression: adventure.Progression{CurrentAct: 1},
	}
	adv := driftAdv(t, plan, 5)
	lr := driftReport(t, adv)
	if findRule(lr, "act.window_elapsed") != nil {
		t.Errorf("completed act should not be flagged, got %+v", lr.Findings)
	}
}

func TestDriftCurrentActOutOfRange(t *testing.T) {
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{
			Acts: []adventure.Act{{Number: 1, Title: "A"}, {Number: 2, Title: "B"}},
		},
		Progression: adventure.Progression{CurrentAct: 5},
	}
	adv := driftAdv(t, plan, 1)
	lr := driftReport(t, adv)
	if findRule(lr, "progression.current_act_out_of_range") == nil {
		t.Fatalf("expected progression.current_act_out_of_range, got %+v", lr.Findings)
	}
}

func TestDriftAntagonistAbsentAndPresent(t *testing.T) {
	mkPlan := func() *adventure.CampaignPlan {
		return &adventure.CampaignPlan{
			NarrativeStructure: adventure.NarrativeStructure{Acts: []adventure.Act{{Number: 1}}},
			Progression:        adventure.Progression{CurrentAct: 1},
			PlotElements: adventure.PlotElements{
				Antagonist: adventure.Character{Name: "Colonel Vastren", Motivation: "m", IntroductionSession: 2},
			},
		}
	}

	// Absent: nothing in journal mentions the antagonist.
	adv := driftAdv(t, mkPlan(), 4)
	lr := driftReport(t, adv)
	if findRule(lr, "antagonist.absent_from_journal") == nil {
		t.Fatalf("expected antagonist.absent_from_journal, got %+v", lr.Findings)
	}

	// Present: a journal entry mentions "Vastren" (the last name token).
	adv2 := driftAdv(t, mkPlan(), 4)
	if err := adv2.SaveSessionJournal(&adventure.SessionJournal{SessionID: 3, Entries: []adventure.JournalEntry{
		{ID: 1, SessionID: 3, Type: "story", Content: "Le groupe affronte Vastren", Timestamp: ts(0)},
	}}); err != nil {
		t.Fatal(err)
	}
	lr2 := driftReport(t, adv2)
	if findRule(lr2, "antagonist.absent_from_journal") != nil {
		t.Errorf("antagonist mentioned in journal should not be flagged, got %+v", lr2.Findings)
	}
}

func TestDriftPlannedNPCNeverUsed(t *testing.T) {
	plan := &adventure.CampaignPlan{
		NarrativeStructure: adventure.NarrativeStructure{Acts: []adventure.Act{{Number: 1}}},
		Progression:        adventure.Progression{CurrentAct: 1},
		PlotElements: adventure.PlotElements{
			NPCs: []adventure.NPCDefinition{
				{Name: "Cassia Morvan", NarrativeIntegration: &adventure.NPCNarrativeLink{IntroductionSession: 2}},
				{Name: "Edvar Solten", NarrativeIntegration: &adventure.NPCNarrativeLink{IntroductionSession: 10}}, // future, not due
			},
		},
	}
	adv := driftAdv(t, plan, 4)
	// Only generate Edvar (not Cassia). Cassia is due (intro 2 <= 4) and missing.
	m := npcmanager.NewManager(adv.BasePath())
	if _, err := m.AddNPC(4, &npc.NPC{Name: "Edvar Solten"}, "ctx", ""); err != nil {
		t.Fatal(err)
	}

	lr := driftReport(t, adv)
	f := findRule(lr, "npc.planned_never_used")
	if f == nil || f.Severity != SeverityWarning {
		t.Fatalf("expected warning npc.planned_never_used, got %+v", lr.Findings)
	}
	if f.Context["name"] != "Cassia Morvan" {
		t.Errorf("expected Cassia Morvan flagged, got %v", f.Context["name"])
	}
}

func TestNameMentioned(t *testing.T) {
	blob := "le groupe affronte vastren dans la forteresse"
	if !nameMentioned("Colonel Vastren", blob) {
		t.Error("should match last token 'vastren'")
	}
	if nameMentioned("Aldric Fermont", blob) {
		t.Error("should not match an absent name")
	}
	if !nameMentioned("", blob) {
		t.Error("empty name should be treated as mentioned (skip)")
	}
}

func TestNPCGenerated(t *testing.T) {
	gen := map[string]bool{"cassia morvan": true}
	if !npcGenerated("Cassia Morvan", gen) {
		t.Error("exact case-insensitive match should succeed")
	}
	if npcGenerated("Edvar Solten", gen) {
		t.Error("absent NPC should not match")
	}
}
