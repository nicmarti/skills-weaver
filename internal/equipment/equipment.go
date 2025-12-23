package equipment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Weapon represents a weapon in BFRPG.
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

// Armor represents armor or shield in BFRPG.
type Armor struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	NameEN  string  `json:"name_en"`
	ACBonus int     `json:"ac_bonus"`
	Weight  float64 `json:"weight"`
	Cost    float64 `json:"cost"`
	Type    string  `json:"type"`
}

// Gear represents adventuring equipment.
type Gear struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	NameEN string  `json:"name_en"`
	Cost   float64 `json:"cost"`
	Weight float64 `json:"weight"`
}

// Ammunition represents ammunition items.
type Ammunition struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	NameEN string  `json:"name_en"`
	Cost   float64 `json:"cost"`
	Weight float64 `json:"weight"`
}

// StartingEquipment represents starting equipment for a class.
type StartingEquipment struct {
	Required      []string   `json:"required"`
	WeaponChoices [][]string `json:"weapon_choices"`
	ArmorChoices  []string   `json:"armor_choices"`
}

// EquipmentData holds all equipment data from JSON.
type EquipmentData struct {
	Weapons           []Weapon                     `json:"weapons"`
	Armor             []Armor                      `json:"armor"`
	AdventuringGear   []Gear                       `json:"adventuring_gear"`
	Ammunition        []Ammunition                 `json:"ammunition"`
	StartingEquipment map[string]StartingEquipment `json:"starting_equipment"`
}

// Catalog manages equipment data.
type Catalog struct {
	data    *EquipmentData
	dataDir string
}

// NewCatalog creates a new equipment catalog from the data directory.
func NewCatalog(dataDir string) (*Catalog, error) {
	path := filepath.Join(dataDir, "equipment.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading equipment.json: %w", err)
	}

	var equipmentData EquipmentData
	if err := json.Unmarshal(data, &equipmentData); err != nil {
		return nil, fmt.Errorf("parsing equipment.json: %w", err)
	}

	return &Catalog{
		data:    &equipmentData,
		dataDir: dataDir,
	}, nil
}

// GetWeapon returns a weapon by ID.
func (c *Catalog) GetWeapon(id string) (*Weapon, error) {
	id = strings.ToLower(id)
	for i := range c.data.Weapons {
		if strings.ToLower(c.data.Weapons[i].ID) == id {
			return &c.data.Weapons[i], nil
		}
	}
	return nil, fmt.Errorf("weapon not found: %s", id)
}

// GetArmor returns armor by ID.
func (c *Catalog) GetArmor(id string) (*Armor, error) {
	id = strings.ToLower(id)
	for i := range c.data.Armor {
		if strings.ToLower(c.data.Armor[i].ID) == id {
			return &c.data.Armor[i], nil
		}
	}
	return nil, fmt.Errorf("armor not found: %s", id)
}

// GetGear returns gear by ID.
func (c *Catalog) GetGear(id string) (*Gear, error) {
	id = strings.ToLower(id)
	for i := range c.data.AdventuringGear {
		if strings.ToLower(c.data.AdventuringGear[i].ID) == id {
			return &c.data.AdventuringGear[i], nil
		}
	}
	return nil, fmt.Errorf("gear not found: %s", id)
}

// GetItem returns any item by ID (weapon, armor, gear, or ammunition).
func (c *Catalog) GetItem(id string) (interface{}, string, error) {
	if w, err := c.GetWeapon(id); err == nil {
		return w, "weapon", nil
	}
	if a, err := c.GetArmor(id); err == nil {
		return a, "armor", nil
	}
	if g, err := c.GetGear(id); err == nil {
		return g, "gear", nil
	}
	// Check ammunition
	id = strings.ToLower(id)
	for i := range c.data.Ammunition {
		if strings.ToLower(c.data.Ammunition[i].ID) == id {
			return &c.data.Ammunition[i], "ammunition", nil
		}
	}
	return nil, "", fmt.Errorf("item not found: %s", id)
}

// ListWeapons returns all weapons, optionally filtered by type.
func (c *Catalog) ListWeapons(weaponType string) []*Weapon {
	var results []*Weapon
	weaponType = strings.ToLower(weaponType)

	for i := range c.data.Weapons {
		w := &c.data.Weapons[i]
		if weaponType == "" || strings.ToLower(w.Type) == weaponType {
			results = append(results, w)
		}
	}

	// Sort by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// ListArmor returns all armor, optionally filtered by type.
func (c *Catalog) ListArmor(armorType string) []*Armor {
	var results []*Armor
	armorType = strings.ToLower(armorType)

	for i := range c.data.Armor {
		a := &c.data.Armor[i]
		if armorType == "" || strings.ToLower(a.Type) == armorType {
			results = append(results, a)
		}
	}

	// Sort by AC bonus
	sort.Slice(results, func(i, j int) bool {
		return results[i].ACBonus < results[j].ACBonus
	})

	return results
}

// ListGear returns all adventuring gear.
func (c *Catalog) ListGear() []*Gear {
	results := make([]*Gear, len(c.data.AdventuringGear))
	for i := range c.data.AdventuringGear {
		results[i] = &c.data.AdventuringGear[i]
	}

	// Sort by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

// ListAmmunition returns all ammunition.
func (c *Catalog) ListAmmunition() []*Ammunition {
	results := make([]*Ammunition, len(c.data.Ammunition))
	for i := range c.data.Ammunition {
		results[i] = &c.data.Ammunition[i]
	}
	return results
}

// SearchItems searches all items by name (FR or EN).
func (c *Catalog) SearchItems(query string) []interface{} {
	query = strings.ToLower(query)
	var results []interface{}

	for i := range c.data.Weapons {
		w := &c.data.Weapons[i]
		if strings.Contains(strings.ToLower(w.Name), query) ||
			strings.Contains(strings.ToLower(w.NameEN), query) ||
			strings.Contains(strings.ToLower(w.ID), query) {
			results = append(results, w)
		}
	}

	for i := range c.data.Armor {
		a := &c.data.Armor[i]
		if strings.Contains(strings.ToLower(a.Name), query) ||
			strings.Contains(strings.ToLower(a.NameEN), query) ||
			strings.Contains(strings.ToLower(a.ID), query) {
			results = append(results, a)
		}
	}

	for i := range c.data.AdventuringGear {
		g := &c.data.AdventuringGear[i]
		if strings.Contains(strings.ToLower(g.Name), query) ||
			strings.Contains(strings.ToLower(g.NameEN), query) ||
			strings.Contains(strings.ToLower(g.ID), query) {
			results = append(results, g)
		}
	}

	for i := range c.data.Ammunition {
		a := &c.data.Ammunition[i]
		if strings.Contains(strings.ToLower(a.Name), query) ||
			strings.Contains(strings.ToLower(a.NameEN), query) ||
			strings.Contains(strings.ToLower(a.ID), query) {
			results = append(results, a)
		}
	}

	return results
}

// GetStartingEquipment returns the starting equipment for a class.
func (c *Catalog) GetStartingEquipment(class string) (*StartingEquipment, error) {
	class = strings.ToLower(class)
	if se, ok := c.data.StartingEquipment[class]; ok {
		return &se, nil
	}
	return nil, fmt.Errorf("class not found: %s", class)
}

// GetClasses returns the list of classes with starting equipment.
func (c *Catalog) GetClasses() []string {
	classes := make([]string, 0, len(c.data.StartingEquipment))
	for class := range c.data.StartingEquipment {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	return classes
}

// ToMarkdown returns a formatted markdown description of a weapon.
func (w *Weapon) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", w.Name, w.NameEN))

	sb.WriteString("| Caractéristique | Valeur |\n")
	sb.WriteString("|-----------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Dégâts** | %s |\n", w.Damage))
	sb.WriteString(fmt.Sprintf("| **Type** | %s |\n", w.Type))
	sb.WriteString(fmt.Sprintf("| **Poids** | %.1f po |\n", w.Weight))
	sb.WriteString(fmt.Sprintf("| **Coût** | %.0f po |\n", w.Cost))

	if len(w.Properties) > 0 {
		sb.WriteString(fmt.Sprintf("| **Propriétés** | %s |\n", strings.Join(w.Properties, ", ")))
	}

	if w.Range != "" {
		sb.WriteString(fmt.Sprintf("| **Portée** | %s |\n", w.Range))
	}

	return sb.String()
}

// ToShortDescription returns a one-line description of a weapon.
func (w *Weapon) ToShortDescription() string {
	props := ""
	if len(w.Properties) > 0 {
		props = fmt.Sprintf(" [%s]", strings.Join(w.Properties, ", "))
	}
	return fmt.Sprintf("%s - %s, %.0f po%s", w.Name, w.Damage, w.Cost, props)
}

// ToJSON returns the weapon as JSON string.
func (w *Weapon) ToJSON() (string, error) {
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMarkdown returns a formatted markdown description of armor.
func (a *Armor) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", a.Name, a.NameEN))

	sb.WriteString("| Caractéristique | Valeur |\n")
	sb.WriteString("|-----------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Bonus CA** | +%d |\n", a.ACBonus))
	sb.WriteString(fmt.Sprintf("| **Type** | %s |\n", a.Type))
	sb.WriteString(fmt.Sprintf("| **Poids** | %.1f po |\n", a.Weight))
	sb.WriteString(fmt.Sprintf("| **Coût** | %.0f po |\n", a.Cost))

	return sb.String()
}

// ToShortDescription returns a one-line description of armor.
func (a *Armor) ToShortDescription() string {
	return fmt.Sprintf("%s - CA +%d, %.0f po (%s)", a.Name, a.ACBonus, a.Cost, a.Type)
}

// ToJSON returns the armor as JSON string.
func (a *Armor) ToJSON() (string, error) {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMarkdown returns a formatted markdown description of gear.
func (g *Gear) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", g.Name, g.NameEN))

	sb.WriteString("| Caractéristique | Valeur |\n")
	sb.WriteString("|-----------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Poids** | %.1f po |\n", g.Weight))
	sb.WriteString(fmt.Sprintf("| **Coût** | %.2f po |\n", g.Cost))

	return sb.String()
}

// ToShortDescription returns a one-line description of gear.
func (g *Gear) ToShortDescription() string {
	return fmt.Sprintf("%s - %.2f po, %.1f po poids", g.Name, g.Cost, g.Weight)
}

// ToJSON returns the gear as JSON string.
func (g *Gear) ToJSON() (string, error) {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMarkdown returns a formatted markdown description of ammunition.
func (a *Ammunition) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s (%s)\n\n", a.Name, a.NameEN))

	sb.WriteString("| Caractéristique | Valeur |\n")
	sb.WriteString("|-----------------|--------|\n")
	sb.WriteString(fmt.Sprintf("| **Poids** | %.1f po |\n", a.Weight))
	sb.WriteString(fmt.Sprintf("| **Coût** | %.0f po |\n", a.Cost))

	return sb.String()
}

// ToShortDescription returns a one-line description of ammunition.
func (a *Ammunition) ToShortDescription() string {
	return fmt.Sprintf("%s - %.0f po", a.Name, a.Cost)
}

// ItemToMarkdown converts any item to markdown.
func ItemToMarkdown(item interface{}) string {
	switch v := item.(type) {
	case *Weapon:
		return v.ToMarkdown()
	case *Armor:
		return v.ToMarkdown()
	case *Gear:
		return v.ToMarkdown()
	case *Ammunition:
		return v.ToMarkdown()
	default:
		return ""
	}
}

// ItemToShortDescription converts any item to short description.
func ItemToShortDescription(item interface{}) string {
	switch v := item.(type) {
	case *Weapon:
		return v.ToShortDescription()
	case *Armor:
		return v.ToShortDescription()
	case *Gear:
		return v.ToShortDescription()
	case *Ammunition:
		return v.ToShortDescription()
	default:
		return ""
	}
}
