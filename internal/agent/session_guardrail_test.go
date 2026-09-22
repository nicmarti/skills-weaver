package agent

import (
	"testing"

	"dungeons/internal/adventure"
)

// startSessionStub is a minimal stand-in for the real start_session tool: it starts a
// game session on the adventure and returns a recognizable hidden briefing, so the
// guard-rail's decision and briefing-injection logic can be exercised without the
// full world-keeper machinery.
type startSessionStub struct {
	adv   *adventure.Adventure
	calls int
}

func (s *startSessionStub) Name() string        { return "start_session" }
func (s *startSessionStub) Description() string { return "stub start_session" }
func (s *startSessionStub) InputSchema() map[string]interface{} {
	return map[string]interface{}{"type": "object"}
}
func (s *startSessionStub) Execute(params map[string]interface{}) (interface{}, error) {
	s.calls++
	if _, err := s.adv.StartSession(); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "system_brief": "BRIEF_MARKER"}, nil
}

func newGuardRailAgent(t *testing.T) (*Agent, *adventure.Adventure, *startSessionStub) {
	t.Helper()
	adv := adventure.New("Guard Rail Test", "")
	adv.SetBasePath(t.TempDir())

	ctx := &AdventureContext{Adventure: adv}
	reg := NewToolRegistry(ctx)
	stub := &startSessionStub{adv: adv}
	reg.Register(stub)

	return &Agent{adventureCtx: ctx, toolRegistry: reg}, adv, stub
}

// TestEnsureSessionStartedAutoStartsWhenNoneActive verifies the guard-rail starts a
// session and injects the hidden briefing when none is active.
func TestEnsureSessionStartedAutoStartsWhenNoneActive(t *testing.T) {
	a, adv, stub := newGuardRailAgent(t)

	if cur, _ := adv.GetCurrentSession(); cur != nil {
		t.Fatalf("expected no active session at start, got %v", cur)
	}

	a.ensureSessionStarted()

	if stub.calls != 1 {
		t.Errorf("expected start_session to be invoked once, got %d", stub.calls)
	}
	cur, err := adv.GetCurrentSession()
	if err != nil || cur == nil {
		t.Fatalf("expected an active session after the guard-rail, got %v (err=%v)", cur, err)
	}
	if a.systemGuidance != "BRIEF_MARKER" {
		t.Errorf("expected the hidden briefing to be injected, got %q", a.systemGuidance)
	}
}

// TestEnsureSessionStartedNoOpWhenActive verifies the guard-rail does nothing when a
// session is already active — so it never double-starts or re-injects a briefing.
func TestEnsureSessionStartedNoOpWhenActive(t *testing.T) {
	a, adv, stub := newGuardRailAgent(t)

	if _, err := adv.StartSession(); err != nil {
		t.Fatal(err)
	}

	a.ensureSessionStarted()

	if stub.calls != 0 {
		t.Errorf("guard-rail should be a no-op when a session is active, but start_session ran %d time(s)", stub.calls)
	}
	if a.systemGuidance != "" {
		t.Errorf("no briefing should be injected on a no-op, got %q", a.systemGuidance)
	}
}
