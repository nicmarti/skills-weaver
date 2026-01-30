// Package agent implements the Dungeon Master agent loop using Anthropic API.
package agent

import "slices"

// ToolAccessPolicy defines which tools an agent can access.
type ToolAccessPolicy struct {
	AgentName      string   // Name of the agent this policy applies to
	AllowedTools   []string // Whitelist of tools the agent can use
	ForbiddenTools []string // Tools explicitly forbidden (takes precedence over allowed)
	MaxIterations  int      // Maximum number of tool-use iterations for this agent
}

// AlwaysForbiddenTools are tools that nested agents can NEVER access.
// These prevent recursion and state modification.
var AlwaysForbiddenTools = []string{
	// Recursion prevention - nested agents cannot invoke other agents or skills
	"invoke_agent",
	"invoke_skill",

	// Session management - only main DM can control sessions
	"start_session",
	"end_session",

	// State modification - nested agents are read-only consultants
	"log_event",
	"add_gold",
	"add_item",
	"remove_item",
	"add_xp",
	"update_location",

	// Foreshadowing - only DM can plant/resolve narrative elements
	"plant_foreshadow",
	"resolve_foreshadow",

	// NPC management - only DM can modify NPC importance
	"update_npc_importance",

	// Campaign modifications - only DM can modify campaign state
	"update_campaign_progress",
	"add_narrative_thread",
	"remove_narrative_thread",

	// Content generation that persists - nested agents shouldn't create persistent content
	"generate_image",
	"generate_map",
	"generate_npc",
	"generate_treasure",
}

// AgentToolPolicies defines the tool access policy for each nested agent.
var AgentToolPolicies = map[string]*ToolAccessPolicy{
	"rules-keeper": {
		AgentName: "rules-keeper",
		AllowedTools: []string{
			// Dice rolling for rule demonstrations
			"roll_dice",
			// Monster lookup for combat rules
			"get_monster",
			// Spell lookup for magic rules
			"get_spell",
			// Equipment lookup for item rules
			"get_equipment",
			// Character info for ability checks
			"get_party_info",
			"get_character_info",
			// Encounter generation for CR calculations
			"generate_encounter",
			"roll_monster_hp",
		},
		ForbiddenTools: AlwaysForbiddenTools,
		MaxIterations:  5,
	},
	"character-creator": {
		AgentName: "character-creator",
		AllowedTools: []string{
			// Dice rolling for stat generation
			"roll_dice",
			// Character info lookup
			"get_party_info",
			"get_character_info",
			// Equipment for starting gear
			"get_equipment",
			// Spells for caster classes
			"get_spell",
			// Name generation (allowed - doesn't persist alone)
			"generate_name",
		},
		ForbiddenTools: AlwaysForbiddenTools,
		MaxIterations:  5,
	},
	"world-keeper": {
		AgentName: "world-keeper",
		AllowedTools: []string{
			// Character/party lookup for context
			"get_party_info",
			"get_character_info",
			// Inventory lookup
			"get_inventory",
			// NPC history lookup (read-only)
			"get_npc_history",
			// Campaign plan lookup (read-only)
			"get_campaign_plan",
			// Foreshadow lookup (read-only)
			"list_foreshadows",
			"get_stale_foreshadows",
			// Session info (read-only)
			"get_session_info",
		},
		ForbiddenTools: AlwaysForbiddenTools,
		MaxIterations:  5,
	},
}

// GetPolicyForAgent returns the tool access policy for a given agent name.
// Returns nil if no policy is defined for the agent.
func GetPolicyForAgent(agentName string) *ToolAccessPolicy {
	return AgentToolPolicies[agentName]
}

// IsToolAllowed checks if a tool is allowed for a given agent.
func (p *ToolAccessPolicy) IsToolAllowed(toolName string) bool {
	// Check forbidden list first (takes precedence)
	if slices.Contains(p.ForbiddenTools, toolName) {
		return false
	}

	// Check allowed list
	return slices.Contains(p.AllowedTools, toolName)
}

// GetAllowedToolNames returns a slice of all allowed tool names for this policy.
func (p *ToolAccessPolicy) GetAllowedToolNames() []string {
	// Filter out any tools that are in the forbidden list
	allowed := make([]string, 0, len(p.AllowedTools))
	for _, tool := range p.AllowedTools {
		if !slices.Contains(p.ForbiddenTools, tool) {
			allowed = append(allowed, tool)
		}
	}
	return allowed
}
