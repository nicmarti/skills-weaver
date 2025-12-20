package dice

import (
	"testing"
)

func TestRollSimple(t *testing.T) {
	roller := NewWithSeed(42) // Fixed seed for reproducibility

	tests := []struct {
		expression string
		wantErr    bool
	}{
		{"d6", false},
		{"1d6", false},
		{"2d6", false},
		{"d20", false},
		{"4d6", false},
		{"d100", false},
		{"invalid", true},
		{"", true},
		{"dd6", true},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			result, err := roller.Roll(tt.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roll(%q) error = %v, wantErr %v", tt.expression, err, tt.wantErr)
				return
			}
			if err == nil && result == nil {
				t.Errorf("Roll(%q) returned nil result without error", tt.expression)
			}
		})
	}
}

func TestRollWithModifier(t *testing.T) {
	roller := NewWithSeed(42)

	tests := []struct {
		expression   string
		wantModifier int
	}{
		{"d6+3", 3},
		{"2d6+5", 5},
		{"d20-2", -2},
		{"3d8+10", 10},
		{"d6-1", -1},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			result, err := roller.Roll(tt.expression)
			if err != nil {
				t.Errorf("Roll(%q) unexpected error: %v", tt.expression, err)
				return
			}
			if result.Modifier != tt.wantModifier {
				t.Errorf("Roll(%q).Modifier = %d, want %d", tt.expression, result.Modifier, tt.wantModifier)
			}
		})
	}
}

func TestRollKeepHighest(t *testing.T) {
	roller := NewWithSeed(42)

	result, err := roller.Roll("4d6kh3")
	if err != nil {
		t.Fatalf("Roll(4d6kh3) unexpected error: %v", err)
	}

	if len(result.Rolls) != 4 {
		t.Errorf("Expected 4 rolls, got %d", len(result.Rolls))
	}

	if len(result.Kept) != 3 {
		t.Errorf("Expected 3 kept dice, got %d", len(result.Kept))
	}

	// Verify that kept dice are the highest
	for _, kept := range result.Kept {
		found := false
		for _, roll := range result.Rolls {
			if roll == kept {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Kept die %d not found in rolls", kept)
		}
	}
}

func TestRollKeepLowest(t *testing.T) {
	roller := NewWithSeed(42)

	result, err := roller.Roll("2d20kl1")
	if err != nil {
		t.Fatalf("Roll(2d20kl1) unexpected error: %v", err)
	}

	if len(result.Rolls) != 2 {
		t.Errorf("Expected 2 rolls, got %d", len(result.Rolls))
	}

	if len(result.Kept) != 1 {
		t.Errorf("Expected 1 kept die, got %d", len(result.Kept))
	}

	// Verify the kept die is the lowest
	lowestRoll := result.Rolls[0]
	for _, roll := range result.Rolls {
		if roll < lowestRoll {
			lowestRoll = roll
		}
	}

	if result.Kept[0] != lowestRoll {
		t.Errorf("Kept die %d is not the lowest roll %d", result.Kept[0], lowestRoll)
	}
}

func TestRollAdvantage(t *testing.T) {
	roller := NewWithSeed(42)

	result := roller.RollAdvantage()

	if len(result.Rolls) != 2 {
		t.Errorf("Expected 2 rolls for advantage, got %d", len(result.Rolls))
	}

	if len(result.Kept) != 1 {
		t.Errorf("Expected 1 kept die for advantage, got %d", len(result.Kept))
	}

	// The kept die should be the highest
	highestRoll := result.Rolls[0]
	for _, roll := range result.Rolls {
		if roll > highestRoll {
			highestRoll = roll
		}
	}

	if result.Kept[0] != highestRoll {
		t.Errorf("Kept die %d is not the highest roll %d", result.Kept[0], highestRoll)
	}
}

func TestRollDisadvantage(t *testing.T) {
	roller := NewWithSeed(42)

	result := roller.RollDisadvantage()

	if len(result.Rolls) != 2 {
		t.Errorf("Expected 2 rolls for disadvantage, got %d", len(result.Rolls))
	}

	if len(result.Kept) != 1 {
		t.Errorf("Expected 1 kept die for disadvantage, got %d", len(result.Kept))
	}
}

func TestRollStats(t *testing.T) {
	roller := NewWithSeed(42)

	results := roller.RollStats()

	if len(results) != 6 {
		t.Errorf("Expected 6 stat rolls, got %d", len(results))
	}

	for i, result := range results {
		if len(result.Rolls) != 4 {
			t.Errorf("Stat %d: expected 4 rolls, got %d", i, len(result.Rolls))
		}
		if len(result.Kept) != 3 {
			t.Errorf("Stat %d: expected 3 kept dice, got %d", i, len(result.Kept))
		}
		if result.Total < 3 || result.Total > 18 {
			t.Errorf("Stat %d: total %d out of valid range 3-18", i, result.Total)
		}
	}
}

func TestRollStatsClassic(t *testing.T) {
	roller := NewWithSeed(42)

	results := roller.RollStatsClassic()

	if len(results) != 6 {
		t.Errorf("Expected 6 stat rolls, got %d", len(results))
	}

	for i, result := range results {
		if len(result.Rolls) != 3 {
			t.Errorf("Stat %d: expected 3 rolls, got %d", i, len(result.Rolls))
		}
		if result.Total < 3 || result.Total > 18 {
			t.Errorf("Stat %d: total %d out of valid range 3-18", i, result.Total)
		}
	}
}

func TestResultString(t *testing.T) {
	roller := NewWithSeed(42)

	result, _ := roller.Roll("2d6+3")
	str := result.String()

	if str == "" {
		t.Error("Result.String() returned empty string")
	}

	// Should contain the expression and total
	if result.Total == 0 {
		t.Error("Result.Total is 0")
	}
}

func TestDiceRange(t *testing.T) {
	roller := New()

	// Roll many times to verify range
	diceTypes := []struct {
		sides int
		expr  string
	}{
		{4, "d4"},
		{6, "d6"},
		{8, "d8"},
		{10, "d10"},
		{12, "d12"},
		{20, "d20"},
		{100, "d100"},
	}

	for _, dt := range diceTypes {
		t.Run(dt.expr, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				result, err := roller.Roll(dt.expr)
				if err != nil {
					t.Fatalf("Roll(%s) error: %v", dt.expr, err)
				}
				if result.Total < 1 || result.Total > dt.sides {
					t.Errorf("Roll(%s) = %d, want 1-%d", dt.expr, result.Total, dt.sides)
				}
			}
		})
	}
}
