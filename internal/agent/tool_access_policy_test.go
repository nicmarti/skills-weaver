package agent

import (
	"testing"
)

func TestGetPolicyForAgent(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		wantNil   bool
	}{
		{"rules-keeper exists", "rules-keeper", false},
		{"character-creator exists", "character-creator", false},
		{"world-keeper exists", "world-keeper", false},
		{"unknown agent returns nil", "unknown-agent", true},
		{"dungeon-master has no policy", "dungeon-master", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := GetPolicyForAgent(tt.agentName)
			if tt.wantNil && policy != nil {
				t.Errorf("GetPolicyForAgent(%q) = %v, want nil", tt.agentName, policy)
			}
			if !tt.wantNil && policy == nil {
				t.Errorf("GetPolicyForAgent(%q) = nil, want non-nil", tt.agentName)
			}
		})
	}
}

func TestToolAccessPolicy_IsToolAllowed(t *testing.T) {
	policy := &ToolAccessPolicy{
		AgentName:      "test-agent",
		AllowedTools:   []string{"roll_dice", "get_monster", "get_spell"},
		ForbiddenTools: []string{"invoke_agent", "invoke_skill", "get_spell"},
		MaxIterations:  5,
	}

	tests := []struct {
		name     string
		toolName string
		want     bool
	}{
		{"allowed tool", "roll_dice", true},
		{"allowed tool 2", "get_monster", true},
		{"forbidden takes precedence over allowed", "get_spell", false},
		{"forbidden tool", "invoke_agent", false},
		{"not in any list", "generate_image", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.IsToolAllowed(tt.toolName)
			if got != tt.want {
				t.Errorf("IsToolAllowed(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestToolAccessPolicy_GetAllowedToolNames(t *testing.T) {
	policy := &ToolAccessPolicy{
		AgentName:      "test-agent",
		AllowedTools:   []string{"roll_dice", "get_monster", "invoke_agent"},
		ForbiddenTools: []string{"invoke_agent", "invoke_skill"},
		MaxIterations:  5,
	}

	allowed := policy.GetAllowedToolNames()

	// Should include roll_dice and get_monster, but NOT invoke_agent (forbidden)
	if len(allowed) != 2 {
		t.Errorf("GetAllowedToolNames() returned %d tools, want 2", len(allowed))
	}

	// Check that forbidden tools are excluded
	for _, tool := range allowed {
		if tool == "invoke_agent" {
			t.Error("GetAllowedToolNames() should not include invoke_agent (forbidden)")
		}
	}
}

func TestAlwaysForbiddenTools_ContainsExpected(t *testing.T) {
	expected := []string{
		"invoke_agent",
		"invoke_skill",
		"start_session",
		"end_session",
		"log_event",
		"add_gold",
		"add_item",
		"remove_item",
	}

	for _, tool := range expected {
		found := false
		for _, forbidden := range AlwaysForbiddenTools {
			if forbidden == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AlwaysForbiddenTools should contain %q", tool)
		}
	}
}

func TestRulesKeeperPolicy(t *testing.T) {
	policy := GetPolicyForAgent("rules-keeper")
	if policy == nil {
		t.Fatal("rules-keeper policy should not be nil")
	}

	// Should allow dice rolling
	if !policy.IsToolAllowed("roll_dice") {
		t.Error("rules-keeper should be allowed to use roll_dice")
	}

	// Should allow monster lookup
	if !policy.IsToolAllowed("get_monster") {
		t.Error("rules-keeper should be allowed to use get_monster")
	}

	// Should NOT allow agent invocation
	if policy.IsToolAllowed("invoke_agent") {
		t.Error("rules-keeper should NOT be allowed to use invoke_agent")
	}

	// Should NOT allow state modification
	if policy.IsToolAllowed("log_event") {
		t.Error("rules-keeper should NOT be allowed to use log_event")
	}
}

func TestCharacterCreatorPolicy(t *testing.T) {
	policy := GetPolicyForAgent("character-creator")
	if policy == nil {
		t.Fatal("character-creator policy should not be nil")
	}

	// Should allow dice rolling (for stat generation)
	if !policy.IsToolAllowed("roll_dice") {
		t.Error("character-creator should be allowed to use roll_dice")
	}

	// Should allow name generation
	if !policy.IsToolAllowed("generate_name") {
		t.Error("character-creator should be allowed to use generate_name")
	}

	// Should NOT allow NPC generation (persistent content)
	if policy.IsToolAllowed("generate_npc") {
		t.Error("character-creator should NOT be allowed to use generate_npc")
	}
}

func TestWorldKeeperPolicy(t *testing.T) {
	policy := GetPolicyForAgent("world-keeper")
	if policy == nil {
		t.Fatal("world-keeper policy should not be nil")
	}

	// Should allow campaign plan lookup
	if !policy.IsToolAllowed("get_campaign_plan") {
		t.Error("world-keeper should be allowed to use get_campaign_plan")
	}

	// Should allow foreshadow listing
	if !policy.IsToolAllowed("list_foreshadows") {
		t.Error("world-keeper should be allowed to use list_foreshadows")
	}

	// Should NOT allow foreshadow planting (modification)
	if policy.IsToolAllowed("plant_foreshadow") {
		t.Error("world-keeper should NOT be allowed to use plant_foreshadow")
	}

	// Should NOT allow image generation
	if policy.IsToolAllowed("generate_image") {
		t.Error("world-keeper should NOT be allowed to use generate_image")
	}
}
