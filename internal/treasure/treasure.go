package treasure

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"dungeons/internal/dice"
)

// CoinType represents a type of coin.
type CoinType struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	NameFR string `json:"name_fr"`
	Value  int    `json:"value"` // Value in copper pieces
}

// GemValue represents a gem value tier.
type GemValue struct {
	Value    int      `json:"value"`
	Examples []string `json:"examples"`
}

// JewelryValue represents a jewelry value tier.
type JewelryValue struct {
	Min      int      `json:"min"`
	Max      int      `json:"max"`
	Examples []string `json:"examples"`
}

// Potion represents a magical potion.
type Potion struct {
	ID     string `json:"id"`
	NameFR string `json:"name_fr"`
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

// Scroll represents a magical scroll.
type Scroll struct {
	ID     string `json:"id"`
	NameFR string `json:"name_fr"`
	Value  int    `json:"value"`
}

// Ring represents a magical ring.
type Ring struct {
	ID     string `json:"id"`
	NameFR string `json:"name_fr"`
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

// Weapon represents a magical weapon.
type Weapon struct {
	ID      string `json:"id"`
	NameFR  string `json:"name_fr"`
	Bonus   int    `json:"bonus"`
	Special string `json:"special,omitempty"`
	Value   int    `json:"value"`
}

// Armor represents magical armor.
type Armor struct {
	ID     string `json:"id"`
	NameFR string `json:"name_fr"`
	Bonus  int    `json:"bonus"`
	Value  int    `json:"value"`
}

// Wand represents a magical wand.
type Wand struct {
	ID      string `json:"id"`
	NameFR  string `json:"name_fr"`
	Charges string `json:"charges"`
	Value   int    `json:"value"`
}

// MiscItem represents a miscellaneous magic item.
type MiscItem struct {
	ID     string `json:"id"`
	NameFR string `json:"name_fr"`
	Effect string `json:"effect"`
	Value  int    `json:"value"`
}

// CoinEntry represents a coin entry in a treasure type.
type CoinEntry struct {
	Type   string `json:"type"`
	Chance int    `json:"chance"`
	Amount string `json:"amount"`
}

// TreasureComponent represents gems, jewelry, or magic entries.
type TreasureComponent struct {
	Chance    int  `json:"chance"`
	Amount    string `json:"amount,omitempty"`
	Items     int    `json:"items,omitempty"`
	NoWeapons bool   `json:"no_weapons,omitempty"`
}

// TreasureType represents a treasure type definition.
type TreasureType struct {
	Description string             `json:"description"`
	Coins       []CoinEntry        `json:"coins,omitempty"`
	Gems        *TreasureComponent `json:"gems,omitempty"`
	Jewelry     *TreasureComponent `json:"jewelry,omitempty"`
	Magic       *TreasureComponent `json:"magic,omitempty"`
	Potions     *TreasureComponent `json:"potions,omitempty"`
	Scrolls     *TreasureComponent `json:"scrolls,omitempty"`
}

// TreasureData holds all treasure data from JSON.
type TreasureData struct {
	CoinTypes      []CoinType              `json:"coin_types"`
	GemValues      []GemValue              `json:"gem_values"`
	JewelryValues  []JewelryValue          `json:"jewelry_values"`
	Potions        []Potion                `json:"potions"`
	Scrolls        []Scroll                `json:"scrolls"`
	Rings          []Ring                  `json:"rings"`
	Weapons        []Weapon                `json:"weapons"`
	Armor          []Armor                 `json:"armor"`
	Wands          []Wand                  `json:"wands"`
	MiscItems      []MiscItem              `json:"misc_items"`
	TreasureTypes  map[string]TreasureType `json:"treasure_types"`
}

// GeneratedCoin represents coins in generated treasure.
type GeneratedCoin struct {
	Type   string
	NameFR string
	Amount int
	ValueGP int // Total value in gold pieces
}

// GeneratedGem represents a gem in generated treasure.
type GeneratedGem struct {
	Name  string
	Value int
}

// GeneratedJewelry represents jewelry in generated treasure.
type GeneratedJewelry struct {
	Name  string
	Value int
}

// GeneratedMagicItem represents a magic item in generated treasure.
type GeneratedMagicItem struct {
	Category string // potion, scroll, ring, weapon, armor, wand, misc
	Name     string
	Effect   string
	Value    int
	Charges  int // For wands
}

// GeneratedTreasure represents all generated treasure.
type GeneratedTreasure struct {
	TreasureType string
	Description  string
	Coins        []GeneratedCoin
	Gems         []GeneratedGem
	Jewelry      []GeneratedJewelry
	MagicItems   []GeneratedMagicItem
	TotalValueGP int
}

// Generator handles treasure generation.
type Generator struct {
	data    *TreasureData
	rng     *rand.Rand
	roller  *dice.Roller
	dataDir string
}

// NewGenerator creates a new treasure generator from the data directory.
func NewGenerator(dataDir string) (*Generator, error) {
	path := filepath.Join(dataDir, "treasure.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading treasure.json: %w", err)
	}

	var treasureData TreasureData
	if err := json.Unmarshal(data, &treasureData); err != nil {
		return nil, fmt.Errorf("parsing treasure.json: %w", err)
	}

	return &Generator{
		data:    &treasureData,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		roller:  dice.New(),
		dataDir: dataDir,
	}, nil
}

// GetTreasureTypes returns all available treasure type codes.
func (g *Generator) GetTreasureTypes() []string {
	types := make([]string, 0, len(g.data.TreasureTypes))
	for t := range g.data.TreasureTypes {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// GetTreasureType returns information about a treasure type.
func (g *Generator) GetTreasureType(code string) (*TreasureType, error) {
	code = strings.ToUpper(code)
	tt, ok := g.data.TreasureTypes[code]
	if !ok {
		return nil, fmt.Errorf("treasure type not found: %s", code)
	}
	return &tt, nil
}

// rollAmount parses and rolls a dice expression like "1d6x1000" or "2d4".
func (g *Generator) rollAmount(expr string) int {
	// Handle multiplier expressions like "1d6x1000"
	if strings.Contains(expr, "x") {
		parts := strings.Split(expr, "x")
		if len(parts) != 2 {
			return 0
		}
		result, err := g.roller.Roll(parts[0])
		if err != nil {
			return 0
		}
		multiplier, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0
		}
		return result.Total * multiplier
	}

	// Simple dice roll
	result, err := g.roller.Roll(expr)
	if err != nil {
		return 0
	}
	return result.Total
}

// checkChance returns true if a percentage chance succeeds.
func (g *Generator) checkChance(chance int) bool {
	return g.rng.Intn(100) < chance
}

// GenerateTreasure generates treasure for a given treasure type.
func (g *Generator) GenerateTreasure(code string) (*GeneratedTreasure, error) {
	code = strings.ToUpper(code)
	tt, ok := g.data.TreasureTypes[code]
	if !ok {
		return nil, fmt.Errorf("treasure type not found: %s", code)
	}

	treasure := &GeneratedTreasure{
		TreasureType: code,
		Description:  tt.Description,
		Coins:        []GeneratedCoin{},
		Gems:         []GeneratedGem{},
		Jewelry:      []GeneratedJewelry{},
		MagicItems:   []GeneratedMagicItem{},
	}

	// Generate coins
	for _, coinEntry := range tt.Coins {
		if g.checkChance(coinEntry.Chance) {
			amount := g.rollAmount(coinEntry.Amount)
			if amount > 0 {
				coinType := g.getCoinType(coinEntry.Type)
				valueGP := (amount * coinType.Value) / 100 // Convert to GP
				treasure.Coins = append(treasure.Coins, GeneratedCoin{
					Type:    coinEntry.Type,
					NameFR:  coinType.NameFR,
					Amount:  amount,
					ValueGP: valueGP,
				})
				treasure.TotalValueGP += valueGP
			}
		}
	}

	// Generate gems
	if tt.Gems != nil && g.checkChance(tt.Gems.Chance) {
		count := g.rollAmount(tt.Gems.Amount)
		for i := 0; i < count; i++ {
			gem := g.generateGem()
			treasure.Gems = append(treasure.Gems, gem)
			treasure.TotalValueGP += gem.Value
		}
	}

	// Generate jewelry
	if tt.Jewelry != nil && g.checkChance(tt.Jewelry.Chance) {
		count := g.rollAmount(tt.Jewelry.Amount)
		for i := 0; i < count; i++ {
			jewelry := g.generateJewelry()
			treasure.Jewelry = append(treasure.Jewelry, jewelry)
			treasure.TotalValueGP += jewelry.Value
		}
	}

	// Generate potions
	if tt.Potions != nil && g.checkChance(tt.Potions.Chance) {
		count := g.rollAmount(tt.Potions.Amount)
		for i := 0; i < count; i++ {
			potion := g.generatePotion()
			treasure.MagicItems = append(treasure.MagicItems, potion)
			treasure.TotalValueGP += potion.Value
		}
	}

	// Generate scrolls
	if tt.Scrolls != nil && g.checkChance(tt.Scrolls.Chance) {
		count := g.rollAmount(tt.Scrolls.Amount)
		for i := 0; i < count; i++ {
			scroll := g.generateScroll()
			treasure.MagicItems = append(treasure.MagicItems, scroll)
			treasure.TotalValueGP += scroll.Value
		}
	}

	// Generate magic items
	if tt.Magic != nil && g.checkChance(tt.Magic.Chance) {
		for i := 0; i < tt.Magic.Items; i++ {
			item := g.generateMagicItem(tt.Magic.NoWeapons)
			treasure.MagicItems = append(treasure.MagicItems, item)
			treasure.TotalValueGP += item.Value
		}
	}

	return treasure, nil
}

// getCoinType returns information about a coin type.
func (g *Generator) getCoinType(id string) CoinType {
	for _, ct := range g.data.CoinTypes {
		if ct.ID == id {
			return ct
		}
	}
	return CoinType{ID: id, NameFR: id, Value: 1}
}

// generateGem generates a random gem.
func (g *Generator) generateGem() GeneratedGem {
	// Roll 1d100 to determine gem value
	roll := g.rng.Intn(100) + 1
	var gemValue GemValue
	switch {
	case roll <= 20:
		gemValue = g.data.GemValues[0] // 10 gp
	case roll <= 45:
		gemValue = g.data.GemValues[1] // 50 gp
	case roll <= 75:
		gemValue = g.data.GemValues[2] // 100 gp
	case roll <= 90:
		gemValue = g.data.GemValues[3] // 500 gp
	case roll <= 98:
		gemValue = g.data.GemValues[4] // 1000 gp
	default:
		gemValue = g.data.GemValues[5] // 5000 gp
	}

	name := gemValue.Examples[g.rng.Intn(len(gemValue.Examples))]
	return GeneratedGem{
		Name:  name,
		Value: gemValue.Value,
	}
}

// generateJewelry generates random jewelry.
func (g *Generator) generateJewelry() GeneratedJewelry {
	// Roll 1d100 to determine jewelry value tier
	roll := g.rng.Intn(100) + 1
	var jv JewelryValue
	switch {
	case roll <= 20:
		jv = g.data.JewelryValues[0]
	case roll <= 45:
		jv = g.data.JewelryValues[1]
	case roll <= 70:
		jv = g.data.JewelryValues[2]
	case roll <= 90:
		jv = g.data.JewelryValues[3]
	default:
		jv = g.data.JewelryValues[4]
	}

	// Roll actual value within range
	value := jv.Min + g.rng.Intn(jv.Max-jv.Min+1)
	name := jv.Examples[g.rng.Intn(len(jv.Examples))]

	return GeneratedJewelry{
		Name:  name,
		Value: value,
	}
}

// generatePotion generates a random potion.
func (g *Generator) generatePotion() GeneratedMagicItem {
	potion := g.data.Potions[g.rng.Intn(len(g.data.Potions))]
	return GeneratedMagicItem{
		Category: "potion",
		Name:     potion.NameFR,
		Effect:   potion.Effect,
		Value:    potion.Value,
	}
}

// generateScroll generates a random scroll.
func (g *Generator) generateScroll() GeneratedMagicItem {
	scroll := g.data.Scrolls[g.rng.Intn(len(g.data.Scrolls))]
	return GeneratedMagicItem{
		Category: "scroll",
		Name:     scroll.NameFR,
		Value:    scroll.Value,
	}
}

// generateMagicItem generates a random magic item.
func (g *Generator) generateMagicItem(noWeapons bool) GeneratedMagicItem {
	// Roll for item type
	roll := g.rng.Intn(100) + 1
	var category string
	switch {
	case roll <= 25:
		category = "potion"
	case roll <= 40:
		category = "scroll"
	case roll <= 50:
		category = "ring"
	case roll <= 60:
		category = "wand"
	case roll <= 75:
		category = "misc"
	case roll <= 85:
		if noWeapons {
			category = "misc"
		} else {
			category = "armor"
		}
	default:
		if noWeapons {
			category = "misc"
		} else {
			category = "weapon"
		}
	}

	switch category {
	case "potion":
		return g.generatePotion()
	case "scroll":
		return g.generateScroll()
	case "ring":
		ring := g.data.Rings[g.rng.Intn(len(g.data.Rings))]
		return GeneratedMagicItem{
			Category: "ring",
			Name:     ring.NameFR,
			Effect:   ring.Effect,
			Value:    ring.Value,
		}
	case "wand":
		wand := g.data.Wands[g.rng.Intn(len(g.data.Wands))]
		charges := g.rollAmount(wand.Charges)
		return GeneratedMagicItem{
			Category: "wand",
			Name:     wand.NameFR,
			Charges:  charges,
			Value:    wand.Value,
		}
	case "armor":
		armor := g.data.Armor[g.rng.Intn(len(g.data.Armor))]
		return GeneratedMagicItem{
			Category: "armor",
			Name:     armor.NameFR,
			Value:    armor.Value,
		}
	case "weapon":
		weapon := g.data.Weapons[g.rng.Intn(len(g.data.Weapons))]
		effect := ""
		if weapon.Special != "" {
			effect = weapon.Special
		}
		return GeneratedMagicItem{
			Category: "weapon",
			Name:     weapon.NameFR,
			Effect:   effect,
			Value:    weapon.Value,
		}
	default: // misc
		misc := g.data.MiscItems[g.rng.Intn(len(g.data.MiscItems))]
		return GeneratedMagicItem{
			Category: "misc",
			Name:     misc.NameFR,
			Effect:   misc.Effect,
			Value:    misc.Value,
		}
	}
}

// GetPotions returns all available potions.
func (g *Generator) GetPotions() []Potion {
	return g.data.Potions
}

// GetScrolls returns all available scrolls.
func (g *Generator) GetScrolls() []Scroll {
	return g.data.Scrolls
}

// GetRings returns all available rings.
func (g *Generator) GetRings() []Ring {
	return g.data.Rings
}

// GetWeapons returns all available magic weapons.
func (g *Generator) GetWeapons() []Weapon {
	return g.data.Weapons
}

// GetArmor returns all available magic armor.
func (g *Generator) GetArmor() []Armor {
	return g.data.Armor
}

// GetWands returns all available wands.
func (g *Generator) GetWands() []Wand {
	return g.data.Wands
}

// GetMiscItems returns all miscellaneous magic items.
func (g *Generator) GetMiscItems() []MiscItem {
	return g.data.MiscItems
}

// ToMarkdown returns a formatted markdown description of the treasure.
func (t *GeneratedTreasure) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Trésor (Type %s)\n\n", t.TreasureType))
	sb.WriteString(fmt.Sprintf("*%s*\n\n", t.Description))

	// Coins
	if len(t.Coins) > 0 {
		sb.WriteString("### Pièces\n\n")
		for _, coin := range t.Coins {
			sb.WriteString(fmt.Sprintf("- **%s** : %d (%d po)\n", coin.NameFR, coin.Amount, coin.ValueGP))
		}
		sb.WriteString("\n")
	}

	// Gems
	if len(t.Gems) > 0 {
		sb.WriteString("### Gemmes\n\n")
		for _, gem := range t.Gems {
			sb.WriteString(fmt.Sprintf("- %s (%d po)\n", gem.Name, gem.Value))
		}
		sb.WriteString("\n")
	}

	// Jewelry
	if len(t.Jewelry) > 0 {
		sb.WriteString("### Bijoux\n\n")
		for _, jewelry := range t.Jewelry {
			sb.WriteString(fmt.Sprintf("- %s (%d po)\n", jewelry.Name, jewelry.Value))
		}
		sb.WriteString("\n")
	}

	// Magic items
	if len(t.MagicItems) > 0 {
		sb.WriteString("### Objets Magiques\n\n")
		for _, item := range t.MagicItems {
			sb.WriteString(fmt.Sprintf("- **%s**", item.Name))
			if item.Effect != "" {
				sb.WriteString(fmt.Sprintf(" - %s", item.Effect))
			}
			if item.Charges > 0 {
				sb.WriteString(fmt.Sprintf(" (%d charges)", item.Charges))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Empty treasure
	if len(t.Coins) == 0 && len(t.Gems) == 0 && len(t.Jewelry) == 0 && len(t.MagicItems) == 0 {
		sb.WriteString("*Aucun trésor trouvé.*\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Valeur totale estimée** : %d po\n", t.TotalValueGP))

	return sb.String()
}

// ToJSON returns the treasure as JSON string.
func (t *GeneratedTreasure) ToJSON() (string, error) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
