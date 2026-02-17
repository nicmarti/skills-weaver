package character

import (
	"testing"
)

func TestGetSpellSlots_FullCaster(t *testing.T) {
	tests := []struct {
		class string
		level int
		want  SpellSlotTable
	}{
		// Level 1 wizard
		{"wizard", 1, SpellSlotTable{1: 2}},
		// Level 3 sorcerer (gets level 2 slots)
		{"sorcerer", 3, SpellSlotTable{1: 4, 2: 2}},
		// Level 5 cleric (gets level 3 slots)
		{"cleric", 5, SpellSlotTable{1: 4, 2: 3, 3: 2}},
		// Level 9 druid (gets level 5 slots)
		{"druid", 9, SpellSlotTable{1: 4, 2: 3, 3: 3, 4: 3, 5: 1}},
		// Level 17 bard (gets level 9 slots)
		{"bard", 17, SpellSlotTable{1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1, 7: 1, 8: 1, 9: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := GetSpellSlots(tt.class, tt.level)
			if len(got) != len(tt.want) {
				t.Errorf("GetSpellSlots(%s, %d) returned %d slots, want %d", tt.class, tt.level, len(got), len(tt.want))
			}
			for level, slots := range tt.want {
				if got[level] != slots {
					t.Errorf("GetSpellSlots(%s, %d) level %d = %d slots, want %d", tt.class, tt.level, level, got[level], slots)
				}
			}
		})
	}
}

func TestGetSpellSlots_HalfCaster(t *testing.T) {
	tests := []struct {
		class string
		level int
		want  SpellSlotTable
	}{
		// Level 1 paladin - no spells yet
		{"paladin", 1, SpellSlotTable{}},
		// Level 2 paladin - first spell slots
		{"paladin", 2, SpellSlotTable{1: 2}},
		// Level 5 ranger - gets level 2 slots
		{"ranger", 5, SpellSlotTable{1: 4, 2: 2}},
		// Level 13 paladin - gets level 4 slots
		{"paladin", 13, SpellSlotTable{1: 4, 2: 3, 3: 3, 4: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := GetSpellSlots(tt.class, tt.level)
			if len(got) != len(tt.want) {
				t.Errorf("GetSpellSlots(%s, %d) returned %d slots, want %d", tt.class, tt.level, len(got), len(tt.want))
			}
			for level, slots := range tt.want {
				if got[level] != slots {
					t.Errorf("GetSpellSlots(%s, %d) level %d = %d slots, want %d", tt.class, tt.level, level, got[level], slots)
				}
			}
		})
	}
}

func TestGetSpellSlots_Warlock(t *testing.T) {
	tests := []struct {
		level int
		want  SpellSlotTable
	}{
		// Level 1 - 1 slot of level 1
		{1, SpellSlotTable{1: 1}},
		// Level 2 - 2 slots of level 1
		{2, SpellSlotTable{1: 2}},
		// Level 3 - 2 slots of level 2 (pact slots upgrade!)
		{3, SpellSlotTable{2: 2}},
		// Level 11 - 3 slots of level 5
		{11, SpellSlotTable{5: 3}},
		// Level 17 - 4 slots of level 5
		{17, SpellSlotTable{5: 4}},
	}

	for _, tt := range tests {
		got := GetSpellSlots("warlock", tt.level)
		if len(got) != len(tt.want) {
			t.Errorf("GetSpellSlots(warlock, %d) returned %d slots, want %d", tt.level, len(got), len(tt.want))
		}
		for level, slots := range tt.want {
			if got[level] != slots {
				t.Errorf("GetSpellSlots(warlock, %d) level %d = %d slots, want %d", tt.level, level, got[level], slots)
			}
		}
	}
}

func TestGetSpellSlots_ThirdCaster(t *testing.T) {
	tests := []struct {
		class string
		level int
		want  SpellSlotTable
	}{
		// Level 2 fighter - no spells yet (Eldritch Knight starts at 3)
		{"fighter", 2, SpellSlotTable{}},
		// Level 3 fighter - first spell slots
		{"fighter", 3, SpellSlotTable{1: 2}},
		// Level 7 rogue - gets level 2 slots
		{"rogue", 7, SpellSlotTable{1: 4, 2: 2}},
		// Level 19 fighter - gets level 4 slots
		{"fighter", 19, SpellSlotTable{1: 4, 2: 3, 3: 3, 4: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := GetSpellSlots(tt.class, tt.level)
			if len(got) != len(tt.want) {
				t.Errorf("GetSpellSlots(%s, %d) returned %d slots, want %d", tt.class, tt.level, len(got), len(tt.want))
			}
			for level, slots := range tt.want {
				if got[level] != slots {
					t.Errorf("GetSpellSlots(%s, %d) level %d = %d slots, want %d", tt.class, tt.level, level, got[level], slots)
				}
			}
		})
	}
}

func TestGetSpellSlots_NonCaster(t *testing.T) {
	nonCasters := []string{"barbarian", "monk"}

	for _, class := range nonCasters {
		got := GetSpellSlots(class, 5)
		if got != nil {
			t.Errorf("GetSpellSlots(%s, 5) = %v, want nil (non-caster)", class, got)
		}
	}
}

func TestGetCantripsKnown(t *testing.T) {
	tests := []struct {
		class string
		level int
		want  int
	}{
		// Wizard progression
		{"wizard", 1, 3},
		{"wizard", 3, 3}, // Still 3
		{"wizard", 4, 4}, // Increases to 4
		{"wizard", 9, 4}, // Still 4
		{"wizard", 10, 5}, // Increases to 5
		{"wizard", 20, 5}, // Stays at 5

		// Sorcerer (starts with more)
		{"sorcerer", 1, 4},
		{"sorcerer", 4, 5},
		{"sorcerer", 10, 6},

		// Bard (starts with fewer)
		{"bard", 1, 2},
		{"bard", 4, 3},
		{"bard", 10, 4},

		// Fighter Eldritch Knight (starts at level 3)
		{"fighter", 1, 0}, // Not a caster yet
		{"fighter", 2, 0}, // Not a caster yet
		{"fighter", 3, 2}, // Becomes Eldritch Knight
		{"fighter", 10, 3},

		// Non-caster
		{"barbarian", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := GetCantripsKnown(tt.class, tt.level)
			if got != tt.want {
				t.Errorf("GetCantripsKnown(%s, %d) = %d, want %d", tt.class, tt.level, got, tt.want)
			}
		})
	}
}

func TestIsCaster(t *testing.T) {
	tests := []struct {
		class string
		level int
		want  bool
	}{
		// Full casters
		{"wizard", 1, true},
		{"sorcerer", 1, true},

		// Half casters (start at level 2)
		{"paladin", 1, false},
		{"paladin", 2, true},
		{"ranger", 1, false},
		{"ranger", 2, true},

		// Warlock (starts at level 1)
		{"warlock", 1, true},

		// 1/3 casters (start at level 3)
		{"fighter", 2, false},
		{"fighter", 3, true},
		{"rogue", 2, false},
		{"rogue", 3, true},

		// Non-casters
		{"barbarian", 1, false},
		{"monk", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := IsCaster(tt.class, tt.level)
			if got != tt.want {
				t.Errorf("IsCaster(%s, %d) = %v, want %v", tt.class, tt.level, got, tt.want)
			}
		})
	}
}

func TestGetMaxSpellLevel(t *testing.T) {
	tests := []struct {
		class string
		level int
		want  int
	}{
		// Full casters
		{"wizard", 1, 1},
		{"wizard", 3, 2},
		{"wizard", 5, 3},
		{"wizard", 17, 9},

		// Half casters (slower progression)
		{"paladin", 2, 1},
		{"paladin", 5, 2},
		{"paladin", 9, 3},
		{"ranger", 17, 5}, // Max is level 5 for half casters

		// Warlock (special - maxes at 5)
		{"warlock", 1, 1},
		{"warlock", 9, 5},
		{"warlock", 20, 5}, // Stays at 5

		// 1/3 casters (very slow)
		{"fighter", 3, 1},
		{"fighter", 7, 2},
		{"fighter", 19, 4}, // Max is level 4 for 1/3 casters

		// Non-caster
		{"barbarian", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			got := GetMaxSpellLevel(tt.class, tt.level)
			if got != tt.want {
				t.Errorf("GetMaxSpellLevel(%s, %d) = %d, want %d", tt.class, tt.level, got, tt.want)
			}
		})
	}
}

func TestSpellSlotProgression_Consistency(t *testing.T) {
	// Test that spell slots never decrease as level increases
	// Note: Warlock is special - their pact slots upgrade in level
	classes := []string{"wizard", "sorcerer", "cleric", "paladin", "ranger", "fighter"}

	for _, class := range classes {
		for level := 1; level < 20; level++ {
			slots1 := GetSpellSlots(class, level)
			slots2 := GetSpellSlots(class, level+1)

			if slots1 == nil {
				continue // Skip if not a caster at this level
			}
			if slots2 == nil {
				t.Errorf("%s has slots at level %d but not at level %d", class, level, level+1)
				continue
			}

			// Check that slots never decrease
			for spellLevel, count1 := range slots1 {
				if count2, ok := slots2[spellLevel]; ok {
					if count2 < count1 {
						t.Errorf("%s level %d->%d: spell level %d slots decreased from %d to %d",
							class, level, level+1, spellLevel, count1, count2)
					}
				} else {
					t.Errorf("%s level %d->%d: spell level %d slots disappeared",
						class, level, level+1, spellLevel)
				}
			}
		}
	}
}

func TestWarlockPactSlots(t *testing.T) {
	// Test Warlock's special pact slot behavior
	// Warlock slots upgrade in level rather than accumulating
	tests := []struct {
		level         int
		wantSlotLevel int
		wantCount     int
	}{
		{1, 1, 1},
		{2, 1, 2},
		{3, 2, 2},  // Slots upgrade from level 1 to level 2
		{5, 3, 2},  // Slots upgrade to level 3
		{7, 4, 2},  // Slots upgrade to level 4
		{9, 5, 2},  // Slots upgrade to level 5
		{11, 5, 3}, // Get a 3rd slot (still level 5)
		{17, 5, 4}, // Get a 4th slot (still level 5)
	}

	for _, tt := range tests {
		slots := GetSpellSlots("warlock", tt.level)
		if slots == nil {
			t.Errorf("Warlock level %d has no slots", tt.level)
			continue
		}

		// Warlock should only have one spell level in their slot table
		if len(slots) != 1 {
			t.Errorf("Warlock level %d has %d spell levels, want 1 (pact slots)", tt.level, len(slots))
		}

		count, ok := slots[tt.wantSlotLevel]
		if !ok {
			t.Errorf("Warlock level %d doesn't have level %d slots", tt.level, tt.wantSlotLevel)
			continue
		}

		if count != tt.wantCount {
			t.Errorf("Warlock level %d has %d slots of level %d, want %d", tt.level, count, tt.wantSlotLevel, tt.wantCount)
		}
	}
}
