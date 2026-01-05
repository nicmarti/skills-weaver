package monster

// XPThreshold represents XP budgets for a single character at a given level (D&D 5e).
type XPThreshold struct {
	Easy   int
	Medium int
	Hard   int
	Deadly int
}

// xpThresholdsPerLevel contains XP thresholds for levels 1-20 (D&D 5e).
var xpThresholdsPerLevel = map[int]XPThreshold{
	1:  {25, 50, 75, 100},
	2:  {50, 100, 150, 200},
	3:  {75, 150, 225, 400},
	4:  {125, 250, 375, 500},
	5:  {250, 500, 750, 1100},
	6:  {300, 600, 900, 1400},
	7:  {350, 750, 1100, 1700},
	8:  {450, 900, 1400, 2100},
	9:  {550, 1100, 1600, 2400},
	10: {600, 1200, 1900, 2800},
	11: {800, 1600, 2400, 3600},
	12: {1000, 2000, 3000, 4500},
	13: {1100, 2200, 3400, 5100},
	14: {1250, 2500, 3800, 5700},
	15: {1400, 2800, 4300, 6400},
	16: {1600, 3200, 4800, 7200},
	17: {2000, 3900, 5900, 8800},
	18: {2100, 4200, 6300, 9500},
	19: {2400, 4900, 7300, 10900},
	20: {2800, 5700, 8500, 12700},
}

// EncounterMultiplier adjusts XP based on number of monsters (D&D 5e).
type EncounterMultiplier struct {
	MinMonsters int
	MaxMonsters int
	Multiplier  float64
}

// encounterMultipliers for groups of monsters (D&D 5e DMG p.82).
var encounterMultipliers = []EncounterMultiplier{
	{1, 1, 1.0},      // 1 monster
	{2, 2, 1.5},      // 2 monsters
	{3, 6, 2.0},      // 3-6 monsters
	{7, 10, 2.5},     // 7-10 monsters
	{11, 14, 3.0},    // 11-14 monsters
	{15, 999, 4.0},   // 15+ monsters
}

// GetXPThreshold returns the XP threshold for a given level and difficulty.
func GetXPThreshold(level int, difficulty string) int {
	threshold, ok := xpThresholdsPerLevel[level]
	if !ok {
		// Default to level 1 if out of range
		threshold = xpThresholdsPerLevel[1]
	}

	switch difficulty {
	case "easy":
		return threshold.Easy
	case "medium":
		return threshold.Medium
	case "hard":
		return threshold.Hard
	case "deadly":
		return threshold.Deadly
	default:
		return threshold.Medium
	}
}

// CalculatePartyBudget calculates the total XP budget for a party (D&D 5e).
// partySize is the number of PCs, difficulty is "easy", "medium", "hard", or "deadly".
func CalculatePartyBudget(partyLevel, partySize int, difficulty string) int {
	perCharacter := GetXPThreshold(partyLevel, difficulty)
	return perCharacter * partySize
}

// GetEncounterMultiplier returns the multiplier based on number of monsters.
func GetEncounterMultiplier(numMonsters int) float64 {
	for _, em := range encounterMultipliers {
		if numMonsters >= em.MinMonsters && numMonsters <= em.MaxMonsters {
			return em.Multiplier
		}
	}
	return 1.0
}

// CalculateAdjustedXP calculates the adjusted XP for an encounter (D&D 5e).
// This accounts for the action economy advantage of multiple monsters.
func CalculateAdjustedXP(totalXP int, numMonsters int) int {
	multiplier := GetEncounterMultiplier(numMonsters)
	return int(float64(totalXP) * multiplier)
}

// EncounterDifficulty represents the calculated difficulty of an encounter.
type EncounterDifficulty struct {
	TotalXP        int    // Raw XP from all monsters
	AdjustedXP     int    // XP adjusted for number of monsters
	PartyBudget    int    // XP budget for the party
	Difficulty     string // "trivial", "easy", "medium", "hard", "deadly"
	IsBalanced     bool   // True if within party budget
}

// EvaluateEncounter evaluates an encounter's difficulty against a party (D&D 5e).
func EvaluateEncounter(totalXP, numMonsters, partyLevel, partySize int) *EncounterDifficulty {
	adjustedXP := CalculateAdjustedXP(totalXP, numMonsters)

	// Get thresholds
	threshold := xpThresholdsPerLevel[partyLevel]
	if threshold.Easy == 0 {
		threshold = xpThresholdsPerLevel[1]
	}

	// Calculate party budgets
	partyEasy := threshold.Easy * partySize
	partyMedium := threshold.Medium * partySize
	partyHard := threshold.Hard * partySize
	partyDeadly := threshold.Deadly * partySize

	// Determine difficulty
	var difficulty string
	var partyBudget int
	var isBalanced bool

	switch {
	case adjustedXP < partyEasy:
		difficulty = "trivial"
		partyBudget = partyEasy
		isBalanced = true
	case adjustedXP < partyMedium:
		difficulty = "easy"
		partyBudget = partyMedium
		isBalanced = true
	case adjustedXP < partyHard:
		difficulty = "medium"
		partyBudget = partyHard
		isBalanced = true
	case adjustedXP < partyDeadly:
		difficulty = "hard"
		partyBudget = partyDeadly
		isBalanced = true
	case adjustedXP < partyDeadly*2:
		difficulty = "deadly"
		partyBudget = partyDeadly
		isBalanced = true
	default:
		difficulty = "deadly+"
		partyBudget = partyDeadly
		isBalanced = false // Way over budget
	}

	return &EncounterDifficulty{
		TotalXP:     totalXP,
		AdjustedXP:  adjustedXP,
		PartyBudget: partyBudget,
		Difficulty:  difficulty,
		IsBalanced:  isBalanced,
	}
}
