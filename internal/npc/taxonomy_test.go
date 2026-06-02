package npc

import (
	"strings"
	"testing"
)

func TestNormalizeRace(t *testing.T) {
	cases := map[string]Race{
		"human":    RaceHumain, // legacy EN
		"humain":   RaceHumain,
		"Nain":     RaceNain, // case-insensitive
		"dwarf":    RaceNain,
		" elf ":    RaceElfe, // trimmed
		"elfe":     RaceElfe,
		"halfling": RaceHalfelin,
		"halfelin": RaceHalfelin,
		"":         RaceHumain, // default
		"orc":      RaceHumain, // unknown -> default
	}
	for in, want := range cases {
		if got := NormalizeRace(in); got != want {
			t.Errorf("NormalizeRace(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeRole(t *testing.T) {
	cases := map[string]Role{
		"quest_giver":      RoleDonneurDeQuete, // legacy EN
		"donneur_de_quete": RoleDonneurDeQuete,
		"Antagonist":       RoleAntagoniste, // case-insensitive
		"antagoniste":      RoleAntagoniste,
		"ally":             RoleAllie,
		"allie":            RoleAllie,
		"rival":            RoleRival,
		"informant":        RoleInformateur,
		"":                 RoleInformateur, // default
	}
	for in, want := range cases {
		if got := NormalizeRole(in); got != want {
			t.Errorf("NormalizeRole(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEnumLists(t *testing.T) {
	if len(AllRaces()) != 4 {
		t.Errorf("AllRaces() = %d races, want 4", len(AllRaces()))
	}
	if len(AllRoles()) != 5 {
		t.Errorf("AllRoles() = %d roles, want 5", len(AllRoles()))
	}
	if !strings.Contains(RaceEnumList(), `"humain"`) || !strings.Contains(RaceEnumList(), `"halfelin"`) {
		t.Errorf("RaceEnumList missing FR values: %s", RaceEnumList())
	}
	if !strings.Contains(RoleEnumList(), `"antagoniste"`) {
		t.Errorf("RoleEnumList missing FR values: %s", RoleEnumList())
	}
}
