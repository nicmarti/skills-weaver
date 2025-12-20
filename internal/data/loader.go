// Package data provides loading and access to BFRPG game data.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AbilityModifiers represents racial ability score modifiers.
type AbilityModifiers map[string]int

// SpecialAbility represents a racial or class special ability.
type SpecialAbility struct {
	Name          string `json:"name"`
	NameEN        string `json:"name_en"`
	Description   string `json:"description"`
	LevelRequired int    `json:"level_required,omitempty"`
}

// Race represents a playable race in BFRPG.
type Race struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	NameEN           string           `json:"name_en"`
	Description      string           `json:"description"`
	AbilityModifiers AbilityModifiers `json:"ability_modifiers"`
	SpecialAbilities []SpecialAbility `json:"special_abilities"`
	AllowedClasses   []string         `json:"allowed_classes"`
	LevelLimits      map[string]any   `json:"level_limits"`
	Languages        []string         `json:"languages"`
	BaseMovement     int              `json:"base_movement"`
	Size             string           `json:"size"`
}

// SavingThrows represents saving throw values.
type SavingThrows struct {
	DeathRayPoison  int `json:"death_ray_poison"`
	MagicWands      int `json:"magic_wands"`
	ParalysisPetrif int `json:"paralysis_petrify"`
	DragonBreath    int `json:"dragon_breath"`
	Spells          int `json:"spells"`
}

// Class represents a playable class in BFRPG.
type Class struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	NameEN           string                   `json:"name_en"`
	Description      string                   `json:"description"`
	HitDie           string                   `json:"hit_die"`
	HitDieSides      int                      `json:"hit_die_sides"`
	PrimeRequisite   string                   `json:"prime_requisite"`
	XPBonusThreshold int                      `json:"xp_bonus_threshold"`
	ArmorAllowed     []string                 `json:"armor_allowed"`
	ShieldsAllowed   bool                     `json:"shields_allowed"`
	WeaponsAllowed   []string                 `json:"weapons_allowed"`
	SpecialAbilities []SpecialAbility         `json:"special_abilities"`
	SpellsPerLevel   map[string][]int         `json:"spells_per_level,omitempty"`
	SavingThrows     map[string]SavingThrows  `json:"saving_throws"`
	AttackBonus      map[string]int           `json:"attack_bonus"`
	XPTable          map[string]int           `json:"xp_table"`
	ThiefSkills      map[string]map[string]int `json:"thief_skills,omitempty"`
	TurnUndead       map[string]map[string]any `json:"turn_undead,omitempty"`
	StartingGold     string                   `json:"starting_gold"`
}

// Weapon represents a weapon item.
type Weapon struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	NameEN     string   `json:"name_en"`
	Damage     string   `json:"damage"`
	Weight     float64  `json:"weight"`
	Cost       float64  `json:"cost"`
	Type       string   `json:"type"`
	Properties []string `json:"properties"`
	Range      string   `json:"range,omitempty"`
}

// Armor represents an armor item.
type Armor struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	NameEN  string  `json:"name_en"`
	ACBonus int     `json:"ac_bonus"`
	Weight  float64 `json:"weight"`
	Cost    float64 `json:"cost"`
	Type    string  `json:"type"`
}

// Gear represents adventuring gear.
type Gear struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	NameEN string  `json:"name_en"`
	Cost   float64 `json:"cost"`
	Weight float64 `json:"weight"`
}

// StartingEquipment represents class-specific starting equipment options.
type StartingEquipment struct {
	Required      []string   `json:"required"`
	WeaponChoices [][]string `json:"weapon_choices"`
	ArmorChoices  []string   `json:"armor_choices"`
}

// GameData holds all loaded game data.
type GameData struct {
	dataDir           string
	Races             map[string]*Race
	Classes           map[string]*Class
	Weapons           map[string]*Weapon
	Armor             map[string]*Armor
	Gear              map[string]*Gear
	StartingEquipment map[string]*StartingEquipment
}

// racesFile represents the JSON structure for races.
type racesFile struct {
	Races []Race `json:"races"`
}

// classesFile represents the JSON structure for classes.
type classesFile struct {
	Classes []Class `json:"classes"`
}

// equipmentFile represents the JSON structure for equipment.
type equipmentFile struct {
	Weapons           []Weapon                     `json:"weapons"`
	Armor             []Armor                      `json:"armor"`
	AdventuringGear   []Gear                       `json:"adventuring_gear"`
	StartingEquipment map[string]StartingEquipment `json:"starting_equipment"`
	Ammunition        []Gear                       `json:"ammunition"`
}

// Load reads all game data from JSON files in the specified directory.
// If dataDir is empty, it defaults to "./data".
func Load(dataDir string) (*GameData, error) {
	if dataDir == "" {
		dataDir = "data"
	}

	gd := &GameData{
		dataDir:           dataDir,
		Races:             make(map[string]*Race),
		Classes:           make(map[string]*Class),
		Weapons:           make(map[string]*Weapon),
		Armor:             make(map[string]*Armor),
		Gear:              make(map[string]*Gear),
		StartingEquipment: make(map[string]*StartingEquipment),
	}

	// Load races
	if err := gd.loadRaces(); err != nil {
		return nil, fmt.Errorf("loading races: %w", err)
	}

	// Load classes
	if err := gd.loadClasses(); err != nil {
		return nil, fmt.Errorf("loading classes: %w", err)
	}

	// Load equipment
	if err := gd.loadEquipment(); err != nil {
		return nil, fmt.Errorf("loading equipment: %w", err)
	}

	return gd, nil
}

func (gd *GameData) loadRaces() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "races.json"))
	if err != nil {
		return err
	}

	var rf racesFile
	if err := json.Unmarshal(data, &rf); err != nil {
		return err
	}

	for i := range rf.Races {
		race := &rf.Races[i]
		gd.Races[race.ID] = race
	}

	return nil
}

func (gd *GameData) loadClasses() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "classes.json"))
	if err != nil {
		return err
	}

	var cf classesFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return err
	}

	for i := range cf.Classes {
		class := &cf.Classes[i]
		gd.Classes[class.ID] = class
	}

	return nil
}

func (gd *GameData) loadEquipment() error {
	data, err := os.ReadFile(filepath.Join(gd.dataDir, "equipment.json"))
	if err != nil {
		return err
	}

	var ef equipmentFile
	if err := json.Unmarshal(data, &ef); err != nil {
		return err
	}

	for i := range ef.Weapons {
		w := &ef.Weapons[i]
		gd.Weapons[w.ID] = w
	}

	for i := range ef.Armor {
		a := &ef.Armor[i]
		gd.Armor[a.ID] = a
	}

	for i := range ef.AdventuringGear {
		g := &ef.AdventuringGear[i]
		gd.Gear[g.ID] = g
	}

	for i := range ef.Ammunition {
		g := &ef.Ammunition[i]
		gd.Gear[g.ID] = g
	}

	for classID, se := range ef.StartingEquipment {
		seCopy := se
		gd.StartingEquipment[classID] = &seCopy
	}

	return nil
}

// GetRace returns a race by ID.
func (gd *GameData) GetRace(id string) (*Race, bool) {
	r, ok := gd.Races[id]
	return r, ok
}

// GetClass returns a class by ID.
func (gd *GameData) GetClass(id string) (*Class, bool) {
	c, ok := gd.Classes[id]
	return c, ok
}

// GetWeapon returns a weapon by ID.
func (gd *GameData) GetWeapon(id string) (*Weapon, bool) {
	w, ok := gd.Weapons[id]
	return w, ok
}

// GetArmor returns an armor by ID.
func (gd *GameData) GetArmor(id string) (*Armor, bool) {
	a, ok := gd.Armor[id]
	return a, ok
}

// ListRaces returns all available races.
func (gd *GameData) ListRaces() []*Race {
	races := make([]*Race, 0, len(gd.Races))
	for _, r := range gd.Races {
		races = append(races, r)
	}
	return races
}

// ListClasses returns all available classes.
func (gd *GameData) ListClasses() []*Class {
	classes := make([]*Class, 0, len(gd.Classes))
	for _, c := range gd.Classes {
		classes = append(classes, c)
	}
	return classes
}

// CanPlayClass checks if a race can play a specific class.
func (gd *GameData) CanPlayClass(raceID, classID string) bool {
	race, ok := gd.GetRace(raceID)
	if !ok {
		return false
	}

	for _, allowed := range race.AllowedClasses {
		if allowed == classID {
			return true
		}
	}
	return false
}

// GetLevelLimit returns the level limit for a race/class combination.
// Returns 0 if unlimited, -1 if not allowed.
func (gd *GameData) GetLevelLimit(raceID, classID string) int {
	race, ok := gd.GetRace(raceID)
	if !ok {
		return -1
	}

	limit, ok := race.LevelLimits[classID]
	if !ok {
		return -1
	}

	switch v := limit.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		// Multi-class like "6/9"
		return 0
	default:
		return -1
	}
}
