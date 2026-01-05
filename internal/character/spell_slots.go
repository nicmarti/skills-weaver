// Package character provides D&D 5e spell slot progression tables.
package character

// SpellSlotTable maps spell level to number of spell slots available.
// Key: spell level (1-9), Value: number of slots
type SpellSlotTable map[int]int

// fullCasterSlots defines spell slot progression for full casters.
// Classes: Wizard, Sorcerer, Cleric, Druid, Bard
// These classes get spell slots from level 1 and progress fully.
var fullCasterSlots = map[int]SpellSlotTable{
	1:  {1: 2},
	2:  {1: 3},
	3:  {1: 4, 2: 2},
	4:  {1: 4, 2: 3},
	5:  {1: 4, 2: 3, 3: 2},
	6:  {1: 4, 2: 3, 3: 3},
	7:  {1: 4, 2: 3, 3: 3, 4: 1},
	8:  {1: 4, 2: 3, 3: 3, 4: 2},
	9:  {1: 4, 2: 3, 3: 3, 4: 3, 5: 1},
	10: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2},
	11: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1},
	12: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1},
	13: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1, 7: 1},
	14: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1, 7: 1},
	15: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1, 7: 1, 8: 1},
	16: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1, 7: 1, 8: 1},
	17: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2, 6: 1, 7: 1, 8: 1, 9: 1},
	18: {1: 4, 2: 3, 3: 3, 4: 3, 5: 3, 6: 1, 7: 1, 8: 1, 9: 1},
	19: {1: 4, 2: 3, 3: 3, 4: 3, 5: 3, 6: 2, 7: 1, 8: 1, 9: 1},
	20: {1: 4, 2: 3, 3: 3, 4: 3, 5: 3, 6: 2, 7: 2, 8: 1, 9: 1},
}

// halfCasterSlots defines spell slot progression for half casters.
// Classes: Paladin, Ranger
// These classes start getting spell slots at level 2 with slower progression.
var halfCasterSlots = map[int]SpellSlotTable{
	1:  {},                 // No spells at level 1
	2:  {1: 2},
	3:  {1: 3},
	4:  {1: 3},
	5:  {1: 4, 2: 2},
	6:  {1: 4, 2: 2},
	7:  {1: 4, 2: 3},
	8:  {1: 4, 2: 3},
	9:  {1: 4, 2: 3, 3: 2},
	10: {1: 4, 2: 3, 3: 2},
	11: {1: 4, 2: 3, 3: 3},
	12: {1: 4, 2: 3, 3: 3},
	13: {1: 4, 2: 3, 3: 3, 4: 1},
	14: {1: 4, 2: 3, 3: 3, 4: 1},
	15: {1: 4, 2: 3, 3: 3, 4: 2},
	16: {1: 4, 2: 3, 3: 3, 4: 2},
	17: {1: 4, 2: 3, 3: 3, 4: 3, 5: 1},
	18: {1: 4, 2: 3, 3: 3, 4: 3, 5: 1},
	19: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2},
	20: {1: 4, 2: 3, 3: 3, 4: 3, 5: 2},
}

// warlockSlots defines special Pact Magic slots for Warlocks.
// Warlock slots are different: all slots are of the same level (Pact slot level).
// They restore on short rest instead of long rest.
var warlockSlots = map[int]SpellSlotTable{
	1:  {1: 1},  // 1 slot, max spell level 1
	2:  {1: 2},  // 2 slots, max spell level 1
	3:  {2: 2},  // 2 slots, max spell level 2 (pact slots upgrade)
	4:  {2: 2},
	5:  {3: 2},  // 2 slots, max spell level 3
	6:  {3: 2},
	7:  {4: 2},  // 2 slots, max spell level 4
	8:  {4: 2},
	9:  {5: 2},  // 2 slots, max spell level 5
	10: {5: 2},
	11: {5: 3},  // 3 slots, max spell level 5
	12: {5: 3},
	13: {5: 3},
	14: {5: 3},
	15: {5: 3},
	16: {5: 3},
	17: {5: 4},  // 4 slots, max spell level 5
	18: {5: 4},
	19: {5: 4},
	20: {5: 4},
}

// thirdCasterSlots defines spell slot progression for 1/3 casters.
// Subclasses: Eldritch Knight (Fighter), Arcane Trickster (Rogue)
// These start getting spells at level 3 with very slow progression.
var thirdCasterSlots = map[int]SpellSlotTable{
	1:  {},                 // No spells
	2:  {},                 // No spells
	3:  {1: 2},
	4:  {1: 3},
	5:  {1: 3},
	6:  {1: 3},
	7:  {1: 4, 2: 2},
	8:  {1: 4, 2: 2},
	9:  {1: 4, 2: 2},
	10: {1: 4, 2: 3},
	11: {1: 4, 2: 3},
	12: {1: 4, 2: 3},
	13: {1: 4, 2: 3, 3: 2},
	14: {1: 4, 2: 3, 3: 2},
	15: {1: 4, 2: 3, 3: 2},
	16: {1: 4, 2: 3, 3: 3},
	17: {1: 4, 2: 3, 3: 3},
	18: {1: 4, 2: 3, 3: 3},
	19: {1: 4, 2: 3, 3: 3, 4: 1},
	20: {1: 4, 2: 3, 3: 3, 4: 1},
}

// GetSpellSlots returns the spell slot table for a given class and level.
// Returns nil if the class is a non-caster at that level.
func GetSpellSlots(classID string, level int) SpellSlotTable {
	switch classID {
	case "wizard", "sorcerer", "cleric", "druid", "bard":
		if slots, ok := fullCasterSlots[level]; ok {
			return slots
		}
	case "paladin", "ranger":
		if slots, ok := halfCasterSlots[level]; ok {
			return slots
		}
	case "warlock":
		if slots, ok := warlockSlots[level]; ok {
			return slots
		}
	case "fighter", "rogue":
		// Eldritch Knight (Fighter 3+), Arcane Trickster (Rogue 3+)
		// These are subclasses, but we'll allow spell slots for simplicity
		// Real implementation would check subclass
		if slots, ok := thirdCasterSlots[level]; ok {
			return slots
		}
	}
	return nil // Non-caster or invalid level
}

// cantripsKnown defines how many cantrips each caster class knows at different levels.
// The map structure is: class -> level -> number of cantrips
var cantripsKnown = map[string]map[int]int{
	"wizard": {
		1:  3,
		4:  4,
		10: 5,
	},
	"sorcerer": {
		1:  4,
		4:  5,
		10: 6,
	},
	"bard": {
		1:  2,
		4:  3,
		10: 4,
	},
	"cleric": {
		1:  3,
		4:  4,
		10: 5,
	},
	"druid": {
		1:  2,
		4:  3,
		10: 4,
	},
	"warlock": {
		1:  2,
		4:  3,
		10: 4,
	},
	"fighter": {
		// Eldritch Knight
		3:  2,
		10: 3,
	},
	"rogue": {
		// Arcane Trickster
		3:  3,
		10: 4,
	},
}

// GetCantripsKnown returns the number of cantrips known for a class at a given level.
// Returns 0 if the class doesn't have cantrips or the level is too low.
func GetCantripsKnown(classID string, level int) int {
	table, ok := cantripsKnown[classID]
	if !ok {
		return 0 // Class doesn't have cantrips
	}

	// Find the highest level threshold that applies
	// e.g., for wizard level 7, we want the "4" entry (level threshold 4 gives 4 cantrips)
	cantrips := 0
	highestThreshold := 0
	for threshold, count := range table {
		if level >= threshold && threshold > highestThreshold {
			highestThreshold = threshold
			cantrips = count
		}
	}

	return cantrips
}

// IsCaster returns true if the class can cast spells at the given level.
func IsCaster(classID string, level int) bool {
	slots := GetSpellSlots(classID, level)
	return slots != nil && len(slots) > 0
}

// GetMaxSpellLevel returns the highest spell level the character can cast.
// Returns 0 if they can only cast cantrips or are not a caster.
func GetMaxSpellLevel(classID string, level int) int {
	slots := GetSpellSlots(classID, level)
	if slots == nil {
		return 0
	}

	maxLevel := 0
	for spellLevel := range slots {
		if spellLevel > maxLevel {
			maxLevel = spellLevel
		}
	}
	return maxLevel
}
