package coherence

import (
	"strings"
	"testing"
)

const sampleLog = `[2026-03-01 19:00:32] TOOL CALL: roll_dice (ID: t1)
  Parameters:
  {
    "description": "Perception passive du groupe",
    "expression": "1d20+4"
  }
[2026-03-01 19:00:40] TOOL CALL: roll_dice (ID: t2)
  Parameters:
  {
    "description": "Persuasion pour convaincre le garde",
    "expression": "1d20+6"
  }
[2026-03-01 19:01:00] TOOL CALL: roll_dice (ID: t3)
  Parameters:
  {
    "description": "Jet d'attaque contre le gobelin",
    "expression": "1d20+5"
  }
[2026-03-01 19:01:10] TOOL CALL: roll_dice (ID: t4)
  Parameters:
  {
    "description": "Jet de sauvegarde de Dextérité",
    "expression": "1d20+2"
  }
[2026-03-01 19:01:20] TOOL CALL: roll_dice (ID: t5)
  Parameters:
  {
    "description": "Quelque chose d'indéterminé",
    "expression": "1d20"
  }
[2026-03-01 19:02:00] TOOL CALL: invoke_agent (ID: t6)
  Parameters:
  {
    "agent_name": "world-keeper",
    "question": "coherence?"
  }
[2026-03-01 19:03:00] TOOL CALL: log_event (ID: t7)
  Parameters:
  {
    "type": "story",
    "content": "le groupe avance"
  }
`

func TestParseSessionLog(t *testing.T) {
	sb := parseSessionLog(strings.NewReader(sampleLog), 1)

	if sb.SessionID != 1 {
		t.Errorf("SessionID = %d, want 1", sb.SessionID)
	}
	if sb.ToolCalls["roll_dice"] != 5 {
		t.Errorf("roll_dice calls = %d, want 5", sb.ToolCalls["roll_dice"])
	}
	if sb.ToolCalls["invoke_agent"] != 1 || sb.ToolCalls["log_event"] != 1 {
		t.Errorf("unexpected tool counts: %+v", sb.ToolCalls)
	}

	r := sb.Rolls
	if r.Total != 5 {
		t.Errorf("Rolls.Total = %d, want 5", r.Total)
	}
	if r.Combat != 1 {
		t.Errorf("Combat = %d, want 1", r.Combat)
	}
	if r.Social != 1 {
		t.Errorf("Social = %d, want 1", r.Social)
	}
	if r.Skill != 1 {
		t.Errorf("Skill = %d, want 1 (perception)", r.Skill)
	}
	if r.Save != 1 {
		t.Errorf("Save = %d, want 1", r.Save)
	}
	if r.Other != 1 {
		t.Errorf("Other = %d, want 1", r.Other)
	}
	if sb.AgentConsults["world-keeper"] != 1 {
		t.Errorf("world-keeper consults = %d, want 1", sb.AgentConsults["world-keeper"])
	}
}

func TestCategorizeRoll(t *testing.T) {
	cases := map[string]string{
		"Jet d'attaque contre l'orc":        "combat",
		"Dégâts de l'épée longue":           "combat",
		"Persuasion pour amadouer le maire": "social",
		"Intimidation du bandit":            "social",
		"Discrétion dans l'ombre":           "skill",
		"Perception pour repérer le piège":  "skill",
		"Jet de sauvegarde de Force":        "save",
		"Initiative":                        "combat",
		"Un truc bizarre":                   "other",
	}
	for desc, want := range cases {
		if got := categorizeRoll(desc); got != want {
			t.Errorf("categorizeRoll(%q) = %q, want %q", desc, got, want)
		}
	}
}

func TestSessionIDFromLogName(t *testing.T) {
	cases := map[string]int{
		"sw-dm-session-3.log":       3,
		"/a/b/sw-dm-session-12.log": 12,
		"sw-dm-session-1.log.2.gz":  1,
	}
	for name, want := range cases {
		got, ok := sessionIDFromLogName(name)
		if !ok || got != want {
			t.Errorf("sessionIDFromLogName(%q) = %d,%v want %d", name, got, ok, want)
		}
	}
	if _, ok := sessionIDFromLogName("journal-session-1.json"); ok {
		t.Error("non-log filename should not match")
	}
}
