package combat

import (
	"testing"
)

func TestNewCombat(t *testing.T) {
	c := NewCombat()

	if c == nil {
		t.Fatal("NewCombat() returned nil")
	}
	if c.Round != 0 {
		t.Errorf("Round = %d, want 0", c.Round)
	}
	if len(c.Combatants) != 0 {
		t.Errorf("Combatants length = %d, want 0", len(c.Combatants))
	}
	if c.IsActive {
		t.Error("IsActive should be false initially")
	}
}

func TestAddCombatant(t *testing.T) {
	c := NewCombat()

	// Add a PC
	aldric := c.AddCombatant("Aldric", 1, 16, 10, 2, "1d8+2", false)

	if aldric == nil {
		t.Fatal("AddCombatant returned nil")
	}
	if aldric.Name != "Aldric" {
		t.Errorf("Name = %q, want %q", aldric.Name, "Aldric")
	}
	if aldric.DexMod != 1 {
		t.Errorf("DexMod = %d, want 1", aldric.DexMod)
	}
	if aldric.AC != 16 {
		t.Errorf("AC = %d, want 16", aldric.AC)
	}
	if aldric.HP != 10 {
		t.Errorf("HP = %d, want 10", aldric.HP)
	}
	if aldric.MaxHP != 10 {
		t.Errorf("MaxHP = %d, want 10", aldric.MaxHP)
	}
	if aldric.IsEnemy {
		t.Error("IsEnemy should be false for PC")
	}

	// Add an enemy
	goblin := c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	if !goblin.IsEnemy {
		t.Error("IsEnemy should be true for enemy")
	}

	if len(c.Combatants) != 2 {
		t.Errorf("Combatants length = %d, want 2", len(c.Combatants))
	}
}

func TestAddCombatantSimple(t *testing.T) {
	c := NewCombat()

	pc := c.AddCombatantSimple("Hero", 15, 10, false)
	enemy := c.AddCombatantSimple("Monster", 12, 8, true)

	if pc.Damage != "1d8" {
		t.Errorf("PC damage = %q, want 1d8", pc.Damage)
	}
	if enemy.Damage != "1d6" {
		t.Errorf("Enemy damage = %q, want 1d6", enemy.Damage)
	}
}

func TestGetCombatant(t *testing.T) {
	c := NewCombat()
	c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)

	found := c.GetCombatant("Aldric")
	if found == nil {
		t.Error("GetCombatant should find Aldric")
	}

	notFound := c.GetCombatant("NonExistent")
	if notFound != nil {
		t.Error("GetCombatant should return nil for non-existent")
	}
}

func TestRemoveCombatant(t *testing.T) {
	c := NewCombat()
	c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	removed := c.RemoveCombatant("Goblin")
	if !removed {
		t.Error("RemoveCombatant should return true")
	}
	if len(c.Combatants) != 1 {
		t.Errorf("Combatants length = %d, want 1", len(c.Combatants))
	}

	notRemoved := c.RemoveCombatant("NonExistent")
	if notRemoved {
		t.Error("RemoveCombatant should return false for non-existent")
	}
}

func TestRollInitiative(t *testing.T) {
	c := NewCombat()
	c.AddCombatant("Aldric", 2, 15, 10, 1, "1d8", false)
	c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	c.RollInitiative()

	if c.Round != 1 {
		t.Errorf("Round = %d, want 1", c.Round)
	}
	if !c.IsActive {
		t.Error("IsActive should be true after RollInitiative")
	}

	// Check that initiative was rolled for all combatants
	for _, combatant := range c.Combatants {
		// Initiative should be 1d6 + dexMod, so range is:
		// Aldric: 1+2=3 to 6+2=8
		// Goblin: 1+0=1 to 6+0=6
		minInit := 1 + combatant.DexMod
		maxInit := 6 + combatant.DexMod
		if combatant.Initiative < minInit || combatant.Initiative > maxInit {
			t.Errorf("%s initiative = %d, want %d-%d",
				combatant.Name, combatant.Initiative, minInit, maxInit)
		}
	}
}

func TestGetTurnOrder(t *testing.T) {
	c := NewCombat()
	// Add combatants with fixed initiatives (we'll set them manually)
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	goblin := c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)
	lyra := c.AddCombatant("Lyra", 0, 12, 6, 1, "1d4", false)

	// Manually set initiatives for predictable testing
	aldric.Initiative = 3
	goblin.Initiative = 5
	lyra.Initiative = 4

	order := c.GetTurnOrder()

	// Should be sorted: Goblin(5), Lyra(4), Aldric(3)
	if order[0].Name != "Goblin" {
		t.Errorf("First in order = %q, want Goblin", order[0].Name)
	}
	if order[1].Name != "Lyra" {
		t.Errorf("Second in order = %q, want Lyra", order[1].Name)
	}
	if order[2].Name != "Aldric" {
		t.Errorf("Third in order = %q, want Aldric", order[2].Name)
	}
}

func TestNextTurn(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	goblin := c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	aldric.Initiative = 5
	goblin.Initiative = 3
	c.sortTurnOrder()
	c.IsActive = true
	c.Round = 1

	// First turn should be Aldric (initiative 5)
	current := c.GetCurrentCombatant()
	if current.Name != "Aldric" {
		t.Errorf("Current = %q, want Aldric", current.Name)
	}

	// Advance to next turn
	next := c.NextTurn()
	if next == nil {
		t.Fatal("NextTurn returned nil")
	}
	if next.Name != "Goblin" {
		t.Errorf("Next = %q, want Goblin", next.Name)
	}

	// Advance again - should be nil (round over)
	endOfRound := c.NextTurn()
	if endOfRound != nil {
		t.Error("NextTurn should return nil at end of round")
	}
}

func TestAttack(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 5, "1d8", false)
	_ = c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	// Set high attack bonus to ensure hit
	aldric.AttackBonus = 15

	result, err := c.Attack("Aldric", "Goblin")
	if err != nil {
		t.Fatalf("Attack error: %v", err)
	}

	if result.Attacker != "Aldric" {
		t.Errorf("Attacker = %q, want Aldric", result.Attacker)
	}
	if result.Defender != "Goblin" {
		t.Errorf("Defender = %q, want Goblin", result.Defender)
	}
	if result.TargetAC != 13 {
		t.Errorf("TargetAC = %d, want 13", result.TargetAC)
	}

	// With +15 bonus, should almost always hit (unless natural 1)
	if result.CriticalMiss {
		// Natural 1, miss is expected
		if result.Hit {
			t.Error("Natural 1 should be a miss")
		}
	} else {
		// Should hit
		if !result.Hit {
			t.Errorf("Attack with +15 bonus vs AC 13 should hit (rolled %d)", result.NaturalRoll)
		}
		// Damage should be applied
		if result.Hit && result.Damage <= 0 {
			t.Error("Damage should be positive on hit")
		}
	}
}

func TestAttackDeadCombatants(t *testing.T) {
	c := NewCombat()
	c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	dead := c.AddCombatant("Dead", 0, 13, 0, 0, "1d6", true)
	dead.HP = 0

	_, err := c.Attack("Dead", "Aldric")
	if err == nil {
		t.Error("Attack from dead combatant should error")
	}

	_, err = c.Attack("Aldric", "Dead")
	if err == nil {
		t.Error("Attack on dead combatant should error")
	}
}

func TestHealAndDamage(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)

	// Take damage
	err := c.TakeDamage("Aldric", 3)
	if err != nil {
		t.Fatalf("TakeDamage error: %v", err)
	}
	if aldric.HP != 7 {
		t.Errorf("HP after damage = %d, want 7", aldric.HP)
	}

	// Heal
	err = c.Heal("Aldric", 2)
	if err != nil {
		t.Fatalf("Heal error: %v", err)
	}
	if aldric.HP != 9 {
		t.Errorf("HP after heal = %d, want 9", aldric.HP)
	}

	// Overheal should cap at MaxHP
	err = c.Heal("Aldric", 10)
	if err != nil {
		t.Fatalf("Heal error: %v", err)
	}
	if aldric.HP != 10 {
		t.Errorf("HP after overheal = %d, want 10 (max)", aldric.HP)
	}

	// Overkill should set HP to 0
	err = c.TakeDamage("Aldric", 20)
	if err != nil {
		t.Fatalf("TakeDamage error: %v", err)
	}
	if aldric.HP != 0 {
		t.Errorf("HP after overkill = %d, want 0", aldric.HP)
	}
}

func TestIsOver(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	goblin := c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	if c.IsOver() {
		t.Error("Combat should not be over with living combatants on both sides")
	}

	// Kill the goblin
	goblin.HP = 0
	if !c.IsOver() {
		t.Error("Combat should be over when all enemies are dead")
	}
	if c.GetWinner() != "party" {
		t.Errorf("Winner = %q, want party", c.GetWinner())
	}

	// Reset and kill the PC
	goblin.HP = 4
	aldric.HP = 0
	if !c.IsOver() {
		t.Error("Combat should be over when all PCs are dead")
	}
	if c.GetWinner() != "enemies" {
		t.Errorf("Winner = %q, want enemies", c.GetWinner())
	}
}

func TestGetLivingCombatants(t *testing.T) {
	c := NewCombat()
	c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	dead := c.AddCombatant("Dead", 0, 13, 0, 0, "1d6", true)
	dead.HP = 0
	c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	living := c.GetLivingCombatants()
	if len(living) != 2 {
		t.Errorf("Living combatants = %d, want 2", len(living))
	}

	enemies := c.GetLivingEnemies()
	if len(enemies) != 1 {
		t.Errorf("Living enemies = %d, want 1", len(enemies))
	}

	party := c.GetLivingParty()
	if len(party) != 1 {
		t.Errorf("Living party = %d, want 1", len(party))
	}
}

func TestNewRound(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	_ = c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	c.RollInitiative()

	c.NewRound()

	if c.Round != 2 {
		t.Errorf("Round = %d, want 2", c.Round)
	}
	if c.CurrentTurn != 0 {
		t.Errorf("CurrentTurn = %d, want 0", c.CurrentTurn)
	}
	// Initiative should be re-rolled (may or may not change)
	// Just verify it's in valid range
	if aldric.Initiative < 1 || aldric.Initiative > 6 {
		t.Errorf("Initiative = %d, want 1-6", aldric.Initiative)
	}
}

func TestDelayAction(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 0, 15, 10, 1, "1d8", false)
	goblin := c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	aldric.Initiative = 5
	goblin.Initiative = 3
	c.sortTurnOrder()
	c.IsActive = true
	c.Round = 1

	// Delay Aldric's action
	c.DelayAction()

	current := c.GetCurrentCombatant()
	if !current.IsDelaying {
		t.Error("Current combatant should be delaying")
	}

	// Act on Goblin's initiative
	c.ActOnInitiative("Aldric", 3)
	if aldric.Initiative != 3 {
		t.Errorf("Aldric initiative = %d, want 3", aldric.Initiative)
	}
	if aldric.IsDelaying {
		t.Error("Aldric should no longer be delaying")
	}
}

func TestStatus(t *testing.T) {
	c := NewCombat()
	aldric := c.AddCombatant("Aldric", 1, 15, 10, 1, "1d8", false)
	goblin := c.AddCombatant("Goblin", 0, 13, 4, 0, "1d6", true)

	aldric.Initiative = 5
	goblin.Initiative = 3
	c.sortTurnOrder()
	c.IsActive = true
	c.Round = 1

	status := c.Status()

	if status == "" {
		t.Error("Status should not be empty")
	}
	// Check it contains expected info
	if !containsString(status, "Round 1") {
		t.Error("Status should contain Round info")
	}
	if !containsString(status, "Aldric") {
		t.Error("Status should contain Aldric")
	}
	if !containsString(status, "Goblin") {
		t.Error("Status should contain Goblin")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
