package npc

import "strings"

// This file is the SINGLE SOURCE OF TRUTH for the NPC race and narrative-role
// taxonomies. Canonical values are FRENCH and used everywhere (generation,
// appearance, display, importance mapping, prompt building). Normalization
// accepts the legacy English values (existing adventures, occasional LLM slips)
// and converts them to the French canonical form.

// Race is an NPC's species. Canonical values are French.
type Race string

const (
	RaceHumain   Race = "humain"
	RaceNain     Race = "nain"
	RaceElfe     Race = "elfe"
	RaceHalfelin Race = "halfelin"
)

// raceWeights gives the random-draw weight of each race (humans most common).
// The slice order is the canonical order used by AllRaces/RaceEnumList.
var raceWeights = []struct {
	Race   Race
	Weight int
}{
	{RaceHumain, 60},
	{RaceNain, 15},
	{RaceElfe, 15},
	{RaceHalfelin, 10},
}

// raceAliases maps every accepted spelling (French canonical + legacy English)
// to the canonical French race.
var raceAliases = map[string]Race{
	"humain": RaceHumain, "human": RaceHumain,
	"nain": RaceNain, "dwarf": RaceNain,
	"elfe": RaceElfe, "elf": RaceElfe,
	"halfelin": RaceHalfelin, "halfling": RaceHalfelin,
}

// AllRaces returns the canonical races in stable order.
func AllRaces() []Race {
	out := make([]Race, len(raceWeights))
	for i, rw := range raceWeights {
		out[i] = rw.Race
	}
	return out
}

// NormalizeRace converts any accepted spelling (FR or legacy EN, any case) to
// the canonical French race. Unknown/empty input defaults to RaceHumain.
func NormalizeRace(s string) Race {
	if r, ok := raceAliases[strings.ToLower(strings.TrimSpace(s))]; ok {
		return r
	}
	return RaceHumain
}

// Role is the narrative role of an NPC in the campaign plan. Canonical values
// are French.
type Role string

const (
	RoleDonneurDeQuete Role = "donneur_de_quete"
	RoleAntagoniste    Role = "antagoniste"
	RoleAllie          Role = "allie"
	RoleRival          Role = "rival"
	RoleInformateur    Role = "informateur"
)

// allRoles is the canonical role order (used by AllRoles/RoleEnumList).
var allRoles = []Role{RoleDonneurDeQuete, RoleAntagoniste, RoleAllie, RoleRival, RoleInformateur}

// roleAliases maps every accepted spelling (French canonical + legacy English)
// to the canonical French role.
var roleAliases = map[string]Role{
	"donneur_de_quete": RoleDonneurDeQuete, "quest_giver": RoleDonneurDeQuete,
	"antagoniste": RoleAntagoniste, "antagonist": RoleAntagoniste,
	"allie": RoleAllie, "ally": RoleAllie,
	"rival": RoleRival,
	"informateur": RoleInformateur, "informant": RoleInformateur,
}

// AllRoles returns the canonical roles in stable order.
func AllRoles() []Role {
	out := make([]Role, len(allRoles))
	copy(out, allRoles)
	return out
}

// NormalizeRole converts any accepted spelling (FR or legacy EN, any case) to
// the canonical French role. Unknown/empty input defaults to RoleInformateur.
func NormalizeRole(s string) Role {
	if r, ok := roleAliases[strings.ToLower(strings.TrimSpace(s))]; ok {
		return r
	}
	return RoleInformateur
}

// quotedList renders values as a prompt-ready quoted, comma-separated list.
func quotedList[T ~string](items []T) string {
	parts := make([]string, len(items))
	for i, it := range items {
		parts[i] = `"` + string(it) + `"`
	}
	return strings.Join(parts, ", ")
}

// RaceStrings returns the canonical race values as a plain []string (e.g. for
// JSON-schema "enum" fields).
func RaceStrings() []string {
	races := AllRaces()
	out := make([]string, len(races))
	for i, r := range races {
		out[i] = string(r)
	}
	return out
}

// RaceEnumList renders the races for a prompt: `"humain", "nain", "elfe", "halfelin"`.
func RaceEnumList() string { return quotedList(AllRaces()) }

// RoleEnumList renders the roles for a prompt: `"donneur_de_quete", "antagoniste", ...`.
func RoleEnumList() string { return quotedList(AllRoles()) }
