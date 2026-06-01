package narrativeai

import (
	"errors"
	"strings"
	"testing"

	"dungeons/internal/coherence"
)

// fakeInvoker records calls and returns canned responses keyed by agent name.
type fakeInvoker struct {
	calls     []string // agent names in call order
	responses map[string]string
	failOn    string // agent name to return an error for
}

func (f *fakeInvoker) InvokeAgentSilent(agentName, question string, depth int) (string, error) {
	f.calls = append(f.calls, agentName)
	if agentName == f.failOn {
		return "", errors.New("boom")
	}
	if r, ok := f.responses[agentName]; ok {
		return r, nil
	}
	return "analyse de " + agentName, nil
}

func sampleBrief() *coherence.NarrativeBrief {
	return &coherence.NarrativeBrief{
		PlayedSessions: 3,
		TypeTotals:     map[string]int{"combat": 5, "story": 10},
		Sessions: []coherence.SessionDigest{
			{ID: 1, EntryCount: 18, CombatLogLines: 0, ProgressionMarkers: 0, Summary: "découverte du meurtre"},
			{ID: 2, EntryCount: 2, CombatLogLines: 0, ProgressionMarkers: 0, Summary: "le groupe reste à l'auberge"},
			{ID: 3, EntryCount: 19, CombatLogLines: 2, ProgressionMarkers: 0, Summary: "exploration de la crypte"},
		},
		OpenThreads:        []string{"enquete_meurtre", "culte_souterrain"},
		PendingForeshadows: []coherence.ForeshadowDigest{{ID: "fsh_1", Description: "trahison", Importance: "critical", TargetPayoffSession: 2}},
		Notes:              []string{"combat loggé coup par coup"},
	}
}

func TestJudgeAssemblesThreeLensesAndSynthesis(t *testing.T) {
	inv := &fakeInvoker{responses: map[string]string{
		"world-keeper":    "  verdict monde  ",
		"rules-keeper":    "verdict règles",
		"scenario-critic": "verdict scénario",
	}}

	j, err := Judge(sampleBrief(), inv)
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	if len(j.Lenses) != 3 {
		t.Fatalf("expected 3 lens verdicts, got %d", len(j.Lenses))
	}
	if j.PlayedSessions != 3 {
		t.Errorf("PlayedSessions = %d, want 3", j.PlayedSessions)
	}
	if j.Lenses[0].Agent != "world-keeper" || j.Lenses[1].Agent != "rules-keeper" || j.Lenses[2].Agent != "scenario-critic" {
		t.Errorf("unexpected lens agents: %+v", j.Lenses)
	}
	if j.Lenses[0].Assessment != "verdict monde" {
		t.Errorf("response should be trimmed, got %q", j.Lenses[0].Assessment)
	}
	if j.Synthesis == "" {
		t.Error("expected a synthesis")
	}
	// 3 lenses + 1 synthesis call, and synthesis goes to scenario-critic.
	if len(inv.calls) != 4 {
		t.Fatalf("expected 4 calls, got %d: %v", len(inv.calls), inv.calls)
	}
	if inv.calls[3] != "scenario-critic" {
		t.Errorf("synthesis should call scenario-critic, got %q", inv.calls[3])
	}
}

func TestJudgeReportsProgress(t *testing.T) {
	inv := &fakeInvoker{}
	var msgs []string
	_, err := Judge(sampleBrief(), inv, func(m string) { msgs = append(msgs, m) })
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	// start + (start/done per 3 lenses) + synthesis + done = 9 messages.
	if len(msgs) < 8 {
		t.Errorf("expected progress messages for each step, got %d: %v", len(msgs), msgs)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range []string{"world-keeper", "rules-keeper", "scenario-critic", "Synthèse", "terminé"} {
		if !strings.Contains(joined, want) {
			t.Errorf("progress missing %q in:\n%s", want, joined)
		}
	}
}

func TestJudgeNilProgressIsSafe(t *testing.T) {
	if _, err := Judge(sampleBrief(), &fakeInvoker{}, nil); err != nil {
		t.Errorf("nil progress callback should be safe, got %v", err)
	}
}

func TestJudgePropagatesLensError(t *testing.T) {
	inv := &fakeInvoker{failOn: "rules-keeper"}
	_, err := Judge(sampleBrief(), inv)
	if err == nil {
		t.Fatal("expected error when a lens fails")
	}
	if !strings.Contains(err.Error(), "rules-keeper") {
		t.Errorf("error should name the failing lens, got %v", err)
	}
}

func TestJudgeNilBrief(t *testing.T) {
	if _, err := Judge(nil, &fakeInvoker{}); err == nil {
		t.Error("expected error for nil brief")
	}
}

func TestBriefDigestContainsEvidence(t *testing.T) {
	d := briefDigest(sampleBrief())
	for _, want := range []string{
		"Sessions jouées: 3",
		"le groupe reste à l'auberge", // the stagnation summary must reach the agent
		"Session 2",
		"enquete_meurtre",
		"fsh_1",
		"combat loggé coup par coup", // data caveat must be passed along
	} {
		if !strings.Contains(d, want) {
			t.Errorf("briefDigest missing %q\n---\n%s", want, d)
		}
	}
}

func TestPromptsAreLensSpecific(t *testing.T) {
	b := sampleBrief()
	if !strings.Contains(worldPrompt(b), "COHÉRENCE DU MONDE") {
		t.Error("world prompt should frame the world lens")
	}
	if !strings.Contains(rulesPrompt(b), "D&D 5e") {
		t.Error("rules prompt should frame the rules lens")
	}
	if !strings.Contains(scenarioPrompt(b), "stagnation") {
		t.Error("scenario prompt should frame the scenario lens")
	}
}
