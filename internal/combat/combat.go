// Package combat provides combat management for D&D 5e.
// It handles initiative tracking, turn order, attack resolution, and action economy.
package combat

import (
	"fmt"
	"sort"

	"dungeons/internal/dice"
)

// Combatant represents a participant in combat.
type Combatant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Initiative  int    `json:"initiative"`
	DexMod      int    `json:"dex_mod"`      // Dexterity modifier for initiative
	AC          int    `json:"ac"`           // Armor Class
	HP          int    `json:"hp"`           // Current hit points
	MaxHP       int    `json:"max_hp"`       // Maximum hit points
	AttackBonus int    `json:"attack_bonus"` // Attack bonus
	Damage      string `json:"damage"`       // Damage expression (e.g., "1d8+2")
	IsEnemy     bool   `json:"is_enemy"`     // True if monster/enemy
	IsDelaying  bool   `json:"is_delaying"`  // True if delaying action
	HasActed    bool   `json:"has_acted"`    // True if acted this round

	// D&D 5e Action Economy
	ActionUsed      bool `json:"action_used"`       // Action taken this turn
	BonusActionUsed bool `json:"bonus_action_used"` // Bonus action taken this turn
	ReactionUsed    bool `json:"reaction_used"`     // Reaction taken this round
	Movement        int  `json:"movement"`          // Movement remaining (feet)
	MaxMovement     int  `json:"max_movement"`      // Maximum movement (typically 30)
}

// ResetActions resets action economy at the start of a combatant's turn (D&D 5e).
func (c *Combatant) ResetActions() {
	c.ActionUsed = false
	c.BonusActionUsed = false
	c.Movement = c.MaxMovement
}

// ResetReaction resets reaction at the start of a round (D&D 5e).
func (c *Combatant) ResetReaction() {
	c.ReactionUsed = false
}

// UseAction marks the main action as used. Returns false if already used.
func (c *Combatant) UseAction() bool {
	if c.ActionUsed {
		return false
	}
	c.ActionUsed = true
	return true
}

// UseBonusAction marks the bonus action as used. Returns false if already used.
func (c *Combatant) UseBonusAction() bool {
	if c.BonusActionUsed {
		return false
	}
	c.BonusActionUsed = true
	return true
}

// UseReaction marks the reaction as used. Returns false if already used.
func (c *Combatant) UseReaction() bool {
	if c.ReactionUsed {
		return false
	}
	c.ReactionUsed = true
	return true
}

// UseMovement reduces available movement (in feet). Returns false if insufficient movement.
func (c *Combatant) UseMovement(feet int) bool {
	if feet > c.Movement {
		return false
	}
	c.Movement -= feet
	return true
}

// Combat tracks an ongoing combat encounter.
type Combat struct {
	Round       int          `json:"round"`
	Combatants  []*Combatant `json:"combatants"`
	TurnOrder   []*Combatant `json:"-"` // Sorted by initiative (not persisted)
	CurrentTurn int          `json:"current_turn"`
	IsActive    bool         `json:"is_active"`
	roller      *dice.Roller
}

// AttackResult represents the outcome of an attack.
type AttackResult struct {
	Attacker     string `json:"attacker"`
	Defender     string `json:"defender"`
	AttackRoll   int    `json:"attack_roll"`
	NaturalRoll  int    `json:"natural_roll"`
	TargetAC     int    `json:"target_ac"`
	Hit          bool   `json:"hit"`
	CriticalHit  bool   `json:"critical_hit"`
	CriticalMiss bool   `json:"critical_miss"`
	Damage       int    `json:"damage"`
	DefenderHP   int    `json:"defender_hp"`
	DefenderDead bool   `json:"defender_dead"`
}

// NewCombat creates a new combat encounter.
func NewCombat() *Combat {
	return &Combat{
		Round:       0,
		Combatants:  make([]*Combatant, 0),
		TurnOrder:   make([]*Combatant, 0),
		CurrentTurn: 0,
		IsActive:    false,
		roller:      dice.New(),
	}
}

// AddCombatant adds a participant to the combat.
func (c *Combat) AddCombatant(name string, dexMod, ac, hp, attackBonus int, damage string, isEnemy bool) *Combatant {
	combatant := &Combatant{
		ID:          fmt.Sprintf("%s-%d", name, len(c.Combatants)+1),
		Name:        name,
		DexMod:      dexMod,
		AC:          ac,
		HP:          hp,
		MaxHP:       hp,
		AttackBonus: attackBonus,
		Damage:      damage,
		IsEnemy:     isEnemy,
		HasActed:    false,
		IsDelaying:  false,
		// D&D 5e action economy defaults
		MaxMovement: 30, // Standard D&D 5e speed (30 feet)
		Movement:    30,
	}
	c.Combatants = append(c.Combatants, combatant)
	return combatant
}

// AddCombatantSimple adds a combatant with minimal info (for quick setup).
func (c *Combat) AddCombatantSimple(name string, ac, hp int, isEnemy bool) *Combatant {
	damage := "1d6"
	if isEnemy {
		damage = "1d6" // Default monster damage
	} else {
		damage = "1d8" // Default PC damage (longsword)
	}
	return c.AddCombatant(name, 0, ac, hp, 0, damage, isEnemy)
}

// RemoveCombatant removes a participant from combat by name.
func (c *Combat) RemoveCombatant(name string) bool {
	for i, combatant := range c.Combatants {
		if combatant.Name == name {
			c.Combatants = append(c.Combatants[:i], c.Combatants[i+1:]...)
			return true
		}
	}
	return false
}

// GetCombatant finds a combatant by name.
func (c *Combat) GetCombatant(name string) *Combatant {
	for _, combatant := range c.Combatants {
		if combatant.Name == name {
			return combatant
		}
	}
	return nil
}

// RollInitiative rolls initiative for all combatants and sorts turn order.
// D&D 5e: 1d20 + DEX modifier, higher acts first, ties act simultaneously.
func (c *Combat) RollInitiative() {
	for _, combatant := range c.Combatants {
		result := c.roller.Initiative(combatant.DexMod)
		combatant.Initiative = result.Total
		combatant.HasActed = false
		combatant.IsDelaying = false
		// Reset action economy for first round
		combatant.ResetActions()
		combatant.ResetReaction()
	}
	c.sortTurnOrder()
	c.Round = 1
	c.CurrentTurn = 0
	c.IsActive = true
}

// sortTurnOrder sorts combatants by initiative (highest first).
func (c *Combat) sortTurnOrder() {
	c.TurnOrder = make([]*Combatant, len(c.Combatants))
	copy(c.TurnOrder, c.Combatants)

	sort.SliceStable(c.TurnOrder, func(i, j int) bool {
		// Higher initiative acts first
		return c.TurnOrder[i].Initiative > c.TurnOrder[j].Initiative
	})
}

// GetTurnOrder returns combatants sorted by initiative.
func (c *Combat) GetTurnOrder() []*Combatant {
	if len(c.TurnOrder) == 0 {
		c.sortTurnOrder()
	}
	return c.TurnOrder
}

// GetCurrentCombatant returns the combatant whose turn it is.
func (c *Combat) GetCurrentCombatant() *Combatant {
	order := c.GetTurnOrder()
	if c.CurrentTurn < len(order) {
		return order[c.CurrentTurn]
	}
	return nil
}

// NextTurn advances to the next combatant's turn.
// Returns the new current combatant, or nil if round is over.
func (c *Combat) NextTurn() *Combatant {
	current := c.GetCurrentCombatant()
	if current != nil {
		current.HasActed = true
	}

	c.CurrentTurn++

	// Skip dead combatants
	order := c.GetTurnOrder()
	for c.CurrentTurn < len(order) && order[c.CurrentTurn].HP <= 0 {
		c.CurrentTurn++
	}

	if c.CurrentTurn >= len(order) {
		return nil // Round over
	}

	// Reset actions for the new combatant's turn (D&D 5e)
	next := order[c.CurrentTurn]
	next.ResetActions()

	return next
}

// NewRound starts a new combat round.
// Re-rolls initiative for all living combatants (D&D 5e optional rule).
func (c *Combat) NewRound() {
	c.Round++
	c.CurrentTurn = 0

	// Re-roll initiative for living combatants
	for _, combatant := range c.Combatants {
		if combatant.HP > 0 {
			result := c.roller.Initiative(combatant.DexMod)
			combatant.Initiative = result.Total
			combatant.HasActed = false
			combatant.IsDelaying = false
			// Reset reactions at start of new round (D&D 5e)
			combatant.ResetReaction()
		}
	}
	c.sortTurnOrder()
}

// DelayAction marks the current combatant as delaying their action.
func (c *Combat) DelayAction() bool {
	current := c.GetCurrentCombatant()
	if current == nil {
		return false
	}
	current.IsDelaying = true
	return true
}

// ActOnInitiative allows a delaying combatant to act at a specific initiative.
func (c *Combat) ActOnInitiative(name string, targetInit int) bool {
	combatant := c.GetCombatant(name)
	if combatant == nil || !combatant.IsDelaying {
		return false
	}
	combatant.Initiative = targetInit
	combatant.IsDelaying = false
	c.sortTurnOrder()
	return true
}

// Attack performs an attack from one combatant to another.
// D&D 5e: d20 + attack bonus >= target AC to hit.
// Natural 20 always hits (critical), natural 1 always misses.
func (c *Combat) Attack(attackerName, defenderName string) (*AttackResult, error) {
	attacker := c.GetCombatant(attackerName)
	if attacker == nil {
		return nil, fmt.Errorf("attacker not found: %s", attackerName)
	}

	defender := c.GetCombatant(defenderName)
	if defender == nil {
		return nil, fmt.Errorf("defender not found: %s", defenderName)
	}

	if attacker.HP <= 0 {
		return nil, fmt.Errorf("attacker is dead: %s", attackerName)
	}

	if defender.HP <= 0 {
		return nil, fmt.Errorf("defender is already dead: %s", defenderName)
	}

	// Roll attack
	roll := c.roller.AttackRoll(attacker.AttackBonus)
	result := &AttackResult{
		Attacker:     attackerName,
		Defender:     defenderName,
		AttackRoll:   roll.Total,
		NaturalRoll:  roll.NaturalRoll(),
		TargetAC:     defender.AC,
		CriticalHit:  roll.IsCriticalHit(),
		CriticalMiss: roll.IsCriticalMiss(),
	}

	// Determine hit
	if roll.IsCriticalMiss() {
		result.Hit = false
	} else if roll.IsCriticalHit() {
		result.Hit = true
	} else {
		result.Hit = roll.Total >= defender.AC
	}

	// Roll damage if hit
	if result.Hit {
		damageRoll, err := c.roller.Roll(attacker.Damage)
		if err != nil {
			damageRoll, _ = c.roller.Roll("1d6") // Fallback
		}
		result.Damage = damageRoll.Total

		// Double damage on critical hit (simplified D&D 5e rule)
		if result.CriticalHit {
			result.Damage *= 2
		}

		// Apply damage
		defender.HP -= result.Damage
		if defender.HP < 0 {
			defender.HP = 0
		}
	}

	result.DefenderHP = defender.HP
	result.DefenderDead = defender.HP <= 0

	return result, nil
}

// Heal restores hit points to a combatant.
func (c *Combat) Heal(name string, amount int) error {
	combatant := c.GetCombatant(name)
	if combatant == nil {
		return fmt.Errorf("combatant not found: %s", name)
	}

	combatant.HP += amount
	if combatant.HP > combatant.MaxHP {
		combatant.HP = combatant.MaxHP
	}
	return nil
}

// TakeDamage applies damage to a combatant.
func (c *Combat) TakeDamage(name string, amount int) error {
	combatant := c.GetCombatant(name)
	if combatant == nil {
		return fmt.Errorf("combatant not found: %s", name)
	}

	combatant.HP -= amount
	if combatant.HP < 0 {
		combatant.HP = 0
	}
	return nil
}

// IsOver returns true if combat has ended (all enemies or all PCs dead).
func (c *Combat) IsOver() bool {
	enemiesAlive := false
	pcsAlive := false

	for _, combatant := range c.Combatants {
		if combatant.HP > 0 {
			if combatant.IsEnemy {
				enemiesAlive = true
			} else {
				pcsAlive = true
			}
		}
	}

	return !enemiesAlive || !pcsAlive
}

// GetWinner returns "party" if PCs won, "enemies" if enemies won, "" if ongoing.
func (c *Combat) GetWinner() string {
	if !c.IsOver() {
		return ""
	}

	for _, combatant := range c.Combatants {
		if combatant.HP > 0 {
			if combatant.IsEnemy {
				return "enemies"
			}
			return "party"
		}
	}
	return "" // Everyone dead?
}

// GetLivingCombatants returns all combatants with HP > 0.
func (c *Combat) GetLivingCombatants() []*Combatant {
	living := make([]*Combatant, 0)
	for _, combatant := range c.Combatants {
		if combatant.HP > 0 {
			living = append(living, combatant)
		}
	}
	return living
}

// GetLivingEnemies returns all living enemy combatants.
func (c *Combat) GetLivingEnemies() []*Combatant {
	enemies := make([]*Combatant, 0)
	for _, combatant := range c.Combatants {
		if combatant.HP > 0 && combatant.IsEnemy {
			enemies = append(enemies, combatant)
		}
	}
	return enemies
}

// GetLivingParty returns all living party members.
func (c *Combat) GetLivingParty() []*Combatant {
	party := make([]*Combatant, 0)
	for _, combatant := range c.Combatants {
		if combatant.HP > 0 && !combatant.IsEnemy {
			party = append(party, combatant)
		}
	}
	return party
}

// Status returns a summary of the current combat state.
func (c *Combat) Status() string {
	if !c.IsActive {
		return "Combat not started"
	}

	status := fmt.Sprintf("Round %d\n", c.Round)
	status += fmt.Sprintf("Turn: %d/%d\n\n", c.CurrentTurn+1, len(c.GetTurnOrder()))

	status += "Initiative Order:\n"
	for i, combatant := range c.GetTurnOrder() {
		marker := " "
		if i == c.CurrentTurn {
			marker = ">"
		}
		deadMarker := ""
		if combatant.HP <= 0 {
			deadMarker = " [DEAD]"
		}
		delayMarker := ""
		if combatant.IsDelaying {
			delayMarker = " [DELAY]"
		}
		side := "PC"
		if combatant.IsEnemy {
			side = "Enemy"
		}

		// Base status line
		line := fmt.Sprintf("%s %2d: %-15s (%s) HP:%d/%d AC:%d%s%s",
			marker, combatant.Initiative, combatant.Name, side,
			combatant.HP, combatant.MaxHP, combatant.AC, deadMarker, delayMarker)

		// Add action economy for current turn (D&D 5e)
		if i == c.CurrentTurn && combatant.HP > 0 {
			actions := ""
			if !combatant.ActionUsed {
				actions += " A"
			}
			if !combatant.BonusActionUsed {
				actions += " B"
			}
			if !combatant.ReactionUsed {
				actions += " R"
			}
			if combatant.Movement > 0 {
				actions += fmt.Sprintf(" M:%dft", combatant.Movement)
			}
			if actions != "" {
				line += " [" + actions[1:] + "]" // Remove leading space
			}
		}

		status += line + "\n"
	}

	return status
}
