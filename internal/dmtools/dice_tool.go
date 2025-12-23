package dmtools

import (
	"fmt"

	"dungeons/internal/dice"
)

// DiceRollerTool wraps the dice roller package.
type DiceRollerTool struct {
	roller *dice.Roller
}

// NewDiceRollerTool creates a new dice roller tool.
func NewDiceRollerTool() *DiceRollerTool {
	return &DiceRollerTool{
		roller: dice.New(),
	}
}

// Name returns the tool name.
func (t *DiceRollerTool) Name() string {
	return "roll_dice"
}

// Description returns the tool description.
func (t *DiceRollerTool) Description() string {
	return "Roll dice using RPG notation (e.g., '2d6+3', 'd20', '4d6kh3'). Use this for all random rolls in the game: initiative, attacks, damage, saving throws, ability checks, etc."
}

// InputSchema returns the JSON schema for tool input.
func (t *DiceRollerTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "Dice notation (e.g., '2d6+3', 'd20', '4d6kh3', '1d6')",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Optional description of what the roll is for (e.g., 'Attack roll', 'Initiative', 'Damage')",
			},
		},
		"required": []string{"expression"},
	}
}

// Execute executes the tool with the given parameters.
func (t *DiceRollerTool) Execute(params map[string]interface{}) (interface{}, error) {
	// Extract expression
	expr, ok := params["expression"].(string)
	if !ok {
		return map[string]interface{}{
			"success": false,
			"error":   "expression parameter is required and must be a string",
		}, nil
	}

	// Extract optional description
	desc := ""
	if d, ok := params["description"].(string); ok {
		desc = d
	}

	// Roll dice
	result, err := t.roller.Roll(expr)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to roll dice: %s", err.Error()),
		}, nil
	}

	// Format display
	display := result.String()
	if desc != "" {
		display = fmt.Sprintf("%s: %s", desc, display)
	}

	return map[string]interface{}{
		"success":    true,
		"expression": result.Expression,
		"rolls":      result.Rolls,
		"total":      result.Total,
		"display":    display,
	}, nil
}
