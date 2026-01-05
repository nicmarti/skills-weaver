// Package dice provides dice rolling functionality for tabletop RPGs.
// It supports standard notation like "2d6+3", "4d6kh3" (keep highest 3),
// and advantage/disadvantage mechanics.
package dice

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Roller handles dice rolling with optional seeded randomness.
type Roller struct {
	rng *rand.Rand
}

// Result represents the outcome of a dice roll.
type Result struct {
	Expression   string // Original expression (e.g., "2d6+3")
	Rolls        []int  // Individual die results
	Kept         []int  // Dice kept after filtering (for kh/kl notation)
	KeptIndices  []int  // Indices of kept dice in Rolls
	Modifier     int    // Added/subtracted modifier
	Total        int    // Final total
}

// New creates a new Roller with a random seed.
func New() *Roller {
	return &Roller{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewWithSeed creates a new Roller with a specific seed for reproducibility.
func NewWithSeed(seed int64) *Roller {
	return &Roller{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Roll parses a dice expression and returns the result.
// Supported formats:
//   - "d20" or "1d20" - roll one d20
//   - "2d6" - roll two d6
//   - "2d6+3" - roll 2d6 and add 3
//   - "2d6-1" - roll 2d6 and subtract 1
//   - "4d6kh3" - roll 4d6, keep highest 3
//   - "2d20kl1" - roll 2d20, keep lowest 1 (disadvantage)
func (r *Roller) Roll(expression string) (*Result, error) {
	expression = strings.ToLower(strings.TrimSpace(expression))

	// Parse the expression
	parsed, err := parseExpression(expression)
	if err != nil {
		return nil, err
	}

	// Roll the dice
	rolls := make([]int, parsed.numDice)
	for i := 0; i < parsed.numDice; i++ {
		rolls[i] = r.rng.Intn(parsed.sides) + 1
	}

	// Determine which dice to keep
	kept, keptIndices := keepDice(rolls, parsed.keepHighest, parsed.keepLowest)

	// Calculate total
	total := 0
	for _, v := range kept {
		total += v
	}
	total += parsed.modifier

	return &Result{
		Expression:  expression,
		Rolls:       rolls,
		Kept:        kept,
		KeptIndices: keptIndices,
		Modifier:    parsed.modifier,
		Total:       total,
	}, nil
}

// RollAdvantage rolls 2d20 and keeps the highest.
func (r *Roller) RollAdvantage() *Result {
	result, _ := r.Roll("2d20kh1")
	result.Expression = "d20 (advantage)"
	return result
}

// RollDisadvantage rolls 2d20 and keeps the lowest.
func (r *Roller) RollDisadvantage() *Result {
	result, _ := r.Roll("2d20kl1")
	result.Expression = "d20 (disadvantage)"
	return result
}

// RollStats rolls 4d6 and keeps the highest 3, repeated 6 times for character stats.
func (r *Roller) RollStats() []Result {
	results := make([]Result, 6)
	for i := 0; i < 6; i++ {
		result, _ := r.Roll("4d6kh3")
		results[i] = *result
	}
	return results
}

// RollStatsClassic rolls 3d6 six times for classic stat generation.
func (r *Roller) RollStatsClassic() []Result {
	results := make([]Result, 6)
	for i := 0; i < 6; i++ {
		result, _ := r.Roll("3d6")
		results[i] = *result
	}
	return results
}

// Initiative rolls 1d20 and adds the dexterity modifier for combat initiative.
// D&D 5e rule: Each combatant rolls 1d20 + DEX modifier. Higher acts first.
// Ties are resolved by DEX score, then simultaneous.
func (r *Roller) Initiative(dexMod int) *Result {
	result, _ := r.Roll("1d20")
	result.Modifier = dexMod
	result.Total = result.Rolls[0] + dexMod
	result.Expression = fmt.Sprintf("Initiative (1d20%+d)", dexMod)
	return result
}

// AttackRoll rolls d20 + attack bonus for a combat attack.
// Returns the result with the natural roll preserved for critical hit detection.
// BFRPG rule: d20 + attack bonus >= target AC to hit.
// Natural 20 is always a hit, natural 1 is always a miss.
func (r *Roller) AttackRoll(attackBonus int) *Result {
	result, _ := r.Roll("1d20")
	result.Modifier = attackBonus
	result.Total = result.Rolls[0] + attackBonus
	result.Expression = fmt.Sprintf("Attack (d20%+d)", attackBonus)
	return result
}

// NaturalRoll returns the unmodified die result (for critical detection).
func (res *Result) NaturalRoll() int {
	if len(res.Rolls) > 0 {
		return res.Rolls[0]
	}
	return 0
}

// IsCriticalHit returns true if the natural roll is 20.
func (res *Result) IsCriticalHit() bool {
	return res.NaturalRoll() == 20
}

// IsCriticalMiss returns true if the natural roll is 1.
func (res *Result) IsCriticalMiss() bool {
	return res.NaturalRoll() == 1
}

// String returns a human-readable representation of the result.
func (res *Result) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: ", res.Expression))

	// Show individual rolls
	sb.WriteString("[")
	for i, roll := range res.Rolls {
		if i > 0 {
			sb.WriteString(", ")
		}
		// Check if this index is in the kept indices
		kept := false
		for _, idx := range res.KeptIndices {
			if idx == i {
				kept = true
				break
			}
		}
		if !kept && len(res.Kept) < len(res.Rolls) {
			sb.WriteString(fmt.Sprintf("~%d~", roll))
		} else {
			sb.WriteString(fmt.Sprintf("%d", roll))
		}
	}
	sb.WriteString("]")

	// Show modifier
	if res.Modifier > 0 {
		sb.WriteString(fmt.Sprintf(" + %d", res.Modifier))
	} else if res.Modifier < 0 {
		sb.WriteString(fmt.Sprintf(" - %d", -res.Modifier))
	}

	sb.WriteString(fmt.Sprintf(" = %d", res.Total))
	return sb.String()
}

// parsedExpression holds the components of a dice expression.
type parsedExpression struct {
	numDice     int
	sides       int
	modifier    int
	keepHighest int
	keepLowest  int
}

// parseExpression parses a dice notation string.
func parseExpression(expr string) (*parsedExpression, error) {
	// Regex pattern for dice notation: [num]d<sides>[kh/kl<keep>][+/-<mod>]
	pattern := regexp.MustCompile(`^(\d*)d(\d+)(kh(\d+)|kl(\d+))?([+-]\d+)?$`)
	matches := pattern.FindStringSubmatch(expr)

	if matches == nil {
		return nil, fmt.Errorf("invalid dice expression: %s", expr)
	}

	parsed := &parsedExpression{
		numDice: 1,
	}

	// Number of dice (default 1)
	if matches[1] != "" {
		parsed.numDice, _ = strconv.Atoi(matches[1])
	}

	// Number of sides
	parsed.sides, _ = strconv.Atoi(matches[2])
	if parsed.sides < 1 {
		return nil, fmt.Errorf("invalid number of sides: %d", parsed.sides)
	}

	// Keep highest
	if matches[4] != "" {
		parsed.keepHighest, _ = strconv.Atoi(matches[4])
	}

	// Keep lowest
	if matches[5] != "" {
		parsed.keepLowest, _ = strconv.Atoi(matches[5])
	}

	// Modifier
	if matches[6] != "" {
		parsed.modifier, _ = strconv.Atoi(matches[6])
	}

	return parsed, nil
}

// keepDice filters dice based on keep highest/lowest rules.
// Returns the kept dice values and their indices in the original rolls.
func keepDice(rolls []int, keepHighest, keepLowest int) ([]int, []int) {
	n := len(rolls)
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}

	if keepHighest == 0 && keepLowest == 0 {
		// Keep all dice
		result := make([]int, n)
		copy(result, rolls)
		return result, indices
	}

	// Create pairs of (value, original index) and sort by value
	type pair struct {
		value int
		index int
	}
	pairs := make([]pair, n)
	for i, v := range rolls {
		pairs[i] = pair{value: v, index: i}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value < pairs[j].value
	})

	var keptPairs []pair
	if keepHighest > 0 {
		start := n - keepHighest
		if start < 0 {
			start = 0
		}
		keptPairs = pairs[start:]
	} else if keepLowest > 0 {
		end := keepLowest
		if end > n {
			end = n
		}
		keptPairs = pairs[:end]
	}

	kept := make([]int, len(keptPairs))
	keptIndices := make([]int, len(keptPairs))
	for i, p := range keptPairs {
		kept[i] = p.value
		keptIndices[i] = p.index
	}

	return kept, keptIndices
}
