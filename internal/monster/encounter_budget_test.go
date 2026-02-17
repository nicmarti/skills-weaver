package monster

import (
	"testing"
)

func TestGetXPThreshold(t *testing.T) {
	tests := []struct {
		level      int
		difficulty string
		wantXP     int
	}{
		{1, "easy", 25},
		{1, "medium", 50},
		{1, "hard", 75},
		{1, "deadly", 100},
		{5, "easy", 250},
		{5, "medium", 500},
		{5, "hard", 750},
		{5, "deadly", 1100},
		{20, "easy", 2800},
		{20, "medium", 5700},
		{20, "hard", 8500},
		{20, "deadly", 12700},
	}

	for _, tt := range tests {
		got := GetXPThreshold(tt.level, tt.difficulty)
		if got != tt.wantXP {
			t.Errorf("GetXPThreshold(level=%d, difficulty=%s) = %d, want %d",
				tt.level, tt.difficulty, got, tt.wantXP)
		}
	}
}

func TestCalculatePartyBudget(t *testing.T) {
	tests := []struct {
		partyLevel int
		partySize  int
		difficulty string
		wantBudget int
	}{
		{1, 4, "easy", 100},    // 25 * 4
		{1, 4, "medium", 200},  // 50 * 4
		{1, 4, "hard", 300},    // 75 * 4
		{1, 4, "deadly", 400},  // 100 * 4
		{3, 5, "medium", 750},  // 150 * 5
		{5, 4, "hard", 3000},   // 750 * 4
		{10, 6, "deadly", 16800}, // 2800 * 6
	}

	for _, tt := range tests {
		got := CalculatePartyBudget(tt.partyLevel, tt.partySize, tt.difficulty)
		if got != tt.wantBudget {
			t.Errorf("CalculatePartyBudget(level=%d, size=%d, difficulty=%s) = %d, want %d",
				tt.partyLevel, tt.partySize, tt.difficulty, got, tt.wantBudget)
		}
	}
}

func TestGetEncounterMultiplier(t *testing.T) {
	tests := []struct {
		numMonsters int
		wantMult    float64
	}{
		{1, 1.0},
		{2, 1.5},
		{3, 2.0},
		{4, 2.0},
		{5, 2.0},
		{6, 2.0},
		{7, 2.5},
		{8, 2.5},
		{9, 2.5},
		{10, 2.5},
		{11, 3.0},
		{12, 3.0},
		{13, 3.0},
		{14, 3.0},
		{15, 4.0},
		{20, 4.0},
		{100, 4.0},
	}

	for _, tt := range tests {
		got := GetEncounterMultiplier(tt.numMonsters)
		if got != tt.wantMult {
			t.Errorf("GetEncounterMultiplier(%d) = %.1f, want %.1f",
				tt.numMonsters, got, tt.wantMult)
		}
	}
}

func TestCalculateAdjustedXP(t *testing.T) {
	tests := []struct {
		totalXP     int
		numMonsters int
		wantAdjXP   int
	}{
		{100, 1, 100},   // 100 * 1.0
		{100, 2, 150},   // 100 * 1.5
		{100, 4, 200},   // 100 * 2.0
		{100, 8, 250},   // 100 * 2.5
		{100, 12, 300},  // 100 * 3.0
		{100, 20, 400},  // 100 * 4.0
		{200, 3, 400},   // 200 * 2.0
		{450, 2, 675},   // 450 * 1.5
	}

	for _, tt := range tests {
		got := CalculateAdjustedXP(tt.totalXP, tt.numMonsters)
		if got != tt.wantAdjXP {
			t.Errorf("CalculateAdjustedXP(totalXP=%d, numMonsters=%d) = %d, want %d",
				tt.totalXP, tt.numMonsters, got, tt.wantAdjXP)
		}
	}
}

func TestEvaluateEncounter(t *testing.T) {
	tests := []struct {
		name        string
		totalXP     int
		numMonsters int
		partyLevel  int
		partySize   int
		wantDiff    string
		wantBalance bool
	}{
		{
			name:        "4 goblins (CR 1/4) vs level 1 party of 4",
			totalXP:     200,  // 4 * 50 XP
			numMonsters: 4,
			partyLevel:  1,
			partySize:   4,
			wantDiff:    "deadly", // Adjusted: 200 * 2.0 = 400, deadly threshold = 400
			wantBalance: true,
		},
		{
			name:        "1 goblin (CR 1/4) vs level 1 party of 4",
			totalXP:     50,
			numMonsters: 1,
			partyLevel:  1,
			partySize:   4,
			wantDiff:    "trivial", // Adjusted: 50 * 1.0 = 50, less than easy threshold = 100
			wantBalance: true,
		},
		{
			name:        "2 orcs (CR 1/2) vs level 2 party of 4",
			totalXP:     200,  // 2 * 100 XP
			numMonsters: 2,
			partyLevel:  2,
			partySize:   4,
			wantDiff:    "easy", // Adjusted: 200 * 1.5 = 300, medium threshold = 400
			wantBalance: true,
		},
		{
			name:        "1 ogre (CR 2) vs level 3 party of 4",
			totalXP:     450,
			numMonsters: 1,
			partyLevel:  3,
			partySize:   4,
			wantDiff:    "easy", // Adjusted: 450 * 1.0 = 450, medium threshold = 600
			wantBalance: true,
		},
		{
			name:        "8 skeletons (CR 1/4) vs level 2 party of 4",
			totalXP:     400,  // 8 * 50 XP
			numMonsters: 8,
			partyLevel:  2,
			partySize:   4,
			wantDiff:    "deadly", // Adjusted: 400 * 2.5 = 1000, deadly threshold = 800
			wantBalance: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateEncounter(tt.totalXP, tt.numMonsters, tt.partyLevel, tt.partySize)

			if result.TotalXP != tt.totalXP {
				t.Errorf("TotalXP = %d, want %d", result.TotalXP, tt.totalXP)
			}

			if result.Difficulty != tt.wantDiff {
				t.Errorf("Difficulty = %s, want %s (AdjustedXP=%d, PartyBudget=%d)",
					result.Difficulty, tt.wantDiff, result.AdjustedXP, result.PartyBudget)
			}

			if result.IsBalanced != tt.wantBalance {
				t.Errorf("IsBalanced = %v, want %v", result.IsBalanced, tt.wantBalance)
			}
		})
	}
}

func TestEvaluateEncounterOverBudget(t *testing.T) {
	// 20 dire wolves (CR 1) vs level 1 party of 4
	// Total XP: 20 * 200 = 4000
	// Adjusted XP: 4000 * 4.0 = 16000
	// Level 1 deadly threshold for party of 4: 400
	// This is way over budget (deadly+ should trigger)
	result := EvaluateEncounter(4000, 20, 1, 4)

	if result.Difficulty != "deadly+" {
		t.Errorf("Difficulty = %s, want deadly+ (AdjustedXP=%d, PartyBudget=%d)",
			result.Difficulty, result.AdjustedXP, result.PartyBudget)
	}

	if result.IsBalanced {
		t.Error("IsBalanced should be false for deadly+ encounter")
	}
}

func TestEvaluateEncounterTrivial(t *testing.T) {
	// 1 goblin (CR 1/4) vs level 5 party of 4
	// Total XP: 50
	// Adjusted XP: 50 * 1.0 = 50
	// Level 5 easy threshold for party of 4: 1000
	// This should be trivial
	result := EvaluateEncounter(50, 1, 5, 4)

	if result.Difficulty != "trivial" {
		t.Errorf("Difficulty = %s, want trivial (AdjustedXP=%d, PartyBudget=%d)",
			result.Difficulty, result.AdjustedXP, result.PartyBudget)
	}

	if !result.IsBalanced {
		t.Error("IsBalanced should be true for trivial encounter")
	}
}
