package adventure

import (
	"testing"
)

func newTestAdventure(t *testing.T) *Adventure {
	t.Helper()
	a := &Adventure{}
	a.SetBasePath(t.TempDir())
	return a
}

func TestCombatState_DefaultIsInactive(t *testing.T) {
	a := newTestAdventure(t)
	if a.IsCombatActive() {
		t.Fatal("expected combat to be inactive by default (no combat-state.json)")
	}
}

func TestCombatState_StartEndRoundTrip(t *testing.T) {
	a := newTestAdventure(t)

	cs, err := a.StartCombat()
	if err != nil {
		t.Fatalf("StartCombat: %v", err)
	}
	if !cs.Active {
		t.Fatal("StartCombat should return an active state")
	}
	if cs.StartedAt.IsZero() {
		t.Fatal("StartCombat should stamp StartedAt")
	}
	if !a.IsCombatActive() {
		t.Fatal("IsCombatActive should be true after StartCombat (persisted)")
	}

	if _, err := a.EndCombat(); err != nil {
		t.Fatalf("EndCombat: %v", err)
	}
	if a.IsCombatActive() {
		t.Fatal("IsCombatActive should be false after EndCombat")
	}
}

func TestCombatState_StartIsIdempotent(t *testing.T) {
	a := newTestAdventure(t)

	first, err := a.StartCombat()
	if err != nil {
		t.Fatalf("StartCombat: %v", err)
	}
	second, err := a.StartCombat()
	if err != nil {
		t.Fatalf("StartCombat (again): %v", err)
	}
	if !second.StartedAt.Equal(first.StartedAt) {
		t.Fatalf("re-StartCombat should preserve original StartedAt: %v != %v", second.StartedAt, first.StartedAt)
	}
}

func TestCombatState_PersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	a := &Adventure{}
	a.SetBasePath(dir)
	if _, err := a.StartCombat(); err != nil {
		t.Fatalf("StartCombat: %v", err)
	}

	// A fresh Adventure pointed at the same dir must read the persisted flag.
	b := &Adventure{}
	b.SetBasePath(dir)
	if !b.IsCombatActive() {
		t.Fatal("combat state should persist on disk and be readable by another instance")
	}
}
