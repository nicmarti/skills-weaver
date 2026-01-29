package dmtools

import (
	"testing"
)

func TestLevelForXP(t *testing.T) {
	tests := []struct {
		xp            int
		expectedLevel int
	}{
		{0, 1},
		{299, 1},
		{300, 2},
		{899, 2},
		{900, 3},
		{2699, 3},
		{2700, 4},
		{6500, 5},
		{14000, 6},
		{64000, 10},
		{85000, 11},
		{355000, 20},
		{500000, 20}, // Beyond max level
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := LevelForXP(tt.xp)
			if got != tt.expectedLevel {
				t.Errorf("LevelForXP(%d) = %d, want %d", tt.xp, got, tt.expectedLevel)
			}
		})
	}
}

func TestXPForLevel(t *testing.T) {
	tests := []struct {
		level      int
		expectedXP int
	}{
		{1, 0},
		{2, 300},
		{3, 900},
		{4, 2700},
		{5, 6500},
		{10, 64000},
		{20, 355000},
		{0, 0},   // Below min
		{21, 355000}, // Above max
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := XPForLevel(tt.level)
			if got != tt.expectedXP {
				t.Errorf("XPForLevel(%d) = %d, want %d", tt.level, got, tt.expectedXP)
			}
		})
	}
}

func TestXPToNextLevel(t *testing.T) {
	tests := []struct {
		currentXP  int
		expectedXP int
	}{
		{0, 300},      // Level 1 → 2 needs 300
		{150, 150},    // Level 1 with 150 XP → needs 150 more
		{300, 600},    // Level 2 → 3 needs 600 more
		{500, 400},    // Level 2 with 500 XP → needs 400 more to reach 900
		{355000, 0},   // Max level
		{400000, 0},   // Beyond max level
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := XPToNextLevel(tt.currentXP)
			if got != tt.expectedXP {
				t.Errorf("XPToNextLevel(%d) = %d, want %d", tt.currentXP, got, tt.expectedXP)
			}
		})
	}
}

func TestLevelForXPAllLevels(t *testing.T) {
	// Test that each threshold gives exactly that level
	thresholds := map[int]int{
		1: 0, 2: 300, 3: 900, 4: 2700, 5: 6500,
		6: 14000, 7: 23000, 8: 34000, 9: 48000, 10: 64000,
		11: 85000, 12: 100000, 13: 120000, 14: 140000, 15: 165000,
		16: 195000, 17: 225000, 18: 265000, 19: 305000, 20: 355000,
	}

	for level, xp := range thresholds {
		got := LevelForXP(xp)
		if got != level {
			t.Errorf("LevelForXP(%d) = %d, want %d (exact threshold)", xp, got, level)
		}
	}
}
