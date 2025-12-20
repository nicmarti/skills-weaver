package image

import (
	"fmt"
	"strings"

	"dungeons/internal/character"
	"dungeons/internal/npc"
)

// PromptStyle defines the art style for generation.
type PromptStyle string

const (
	StyleRealistic   PromptStyle = "realistic"
	StylePainted     PromptStyle = "painted"
	StyleIllustrated PromptStyle = "illustrated"
	StyleDarkFantasy PromptStyle = "dark_fantasy"
	StyleEpic        PromptStyle = "epic"
)

// StyleSuffixes maps styles to prompt suffixes.
var StyleSuffixes = map[PromptStyle]string{
	StyleRealistic:   "photorealistic, highly detailed, dramatic lighting, 8k resolution",
	StylePainted:     "oil painting style, rich colors, detailed brushstrokes, fantasy art",
	StyleIllustrated: "digital illustration, vibrant colors, detailed linework, fantasy RPG art style",
	StyleDarkFantasy: "dark fantasy art, moody lighting, dramatic shadows, gritty atmosphere",
	StyleEpic:        "epic fantasy art, cinematic composition, dramatic lighting, heroic pose",
}

// BasePromptSuffix is added to all prompts for consistent quality.
const BasePromptSuffix = "medieval fantasy setting, dungeons and dragons style, high quality"

// BuildCharacterPrompt creates a prompt for a character portrait.
func BuildCharacterPrompt(char *character.Character, style PromptStyle) string {
	var parts []string

	// Race description
	raceDesc := map[string]string{
		"human":    "human",
		"elf":      "elven, pointed ears, graceful features",
		"dwarf":    "dwarven, stocky build, strong features, thick beard",
		"halfling": "halfling, small stature, youthful face, curly hair",
	}

	// Class description
	classDesc := map[string]string{
		"fighter":    "warrior, armored, battle-ready, sword and shield",
		"cleric":     "cleric, holy symbol, religious robes, divine aura",
		"magic-user": "wizard, mystical robes, arcane staff, magical energy",
		"thief":      "rogue, hooded cloak, daggers, stealthy appearance",
	}

	// Build prompt (character doesn't have gender, so we omit it)
	raceD := raceDesc[char.Race]
	if raceD == "" {
		raceD = char.Race
	}
	classD := classDesc[char.Class]
	if classD == "" {
		classD = char.Class
	}

	parts = append(parts, fmt.Sprintf("Portrait of a %s %s",
		raceD, classD))

	parts = append(parts, fmt.Sprintf("named %s", char.Name))

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix)

	return strings.Join(parts, ", ")
}

// BuildNPCPrompt creates a prompt for an NPC portrait.
func BuildNPCPrompt(n *npc.NPC, style PromptStyle) string {
	var parts []string

	// Race description
	raceDesc := map[string]string{
		"human":    "human",
		"elf":      "elven, pointed ears",
		"dwarf":    "dwarven, stocky",
		"halfling": "halfling, small",
	}

	// Gender
	gender := "male"
	if n.Gender == "female" {
		gender = "female"
	}

	// Build base description
	parts = append(parts, fmt.Sprintf("Portrait of a %s %s %s",
		gender,
		raceDesc[n.Race],
		n.Occupation))

	// Add appearance details
	if n.Appearance.Build != "" {
		parts = append(parts, n.Appearance.Build+" build")
	}
	if n.Appearance.HairColor != "" && n.Appearance.HairStyle != "" {
		parts = append(parts, fmt.Sprintf("%s %s hair", n.Appearance.HairColor, n.Appearance.HairStyle))
	}
	if n.Appearance.EyeColor != "" {
		parts = append(parts, n.Appearance.EyeColor+" eyes")
	}
	if n.Appearance.FacialFeature != "" {
		parts = append(parts, n.Appearance.FacialFeature)
	}

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix)

	return strings.Join(parts, ", ")
}

// BuildScenePrompt creates a prompt for a scene illustration.
func BuildScenePrompt(description string, sceneType string, style PromptStyle) string {
	var parts []string

	// Scene type prefixes
	scenePrefix := map[string]string{
		"tavern":   "Interior of a medieval tavern, warm candlelight, wooden beams",
		"dungeon":  "Dark dungeon corridor, stone walls, torchlight, mysterious atmosphere",
		"forest":   "Deep fantasy forest, ancient trees, dappled sunlight, mystical",
		"castle":   "Medieval castle interior, grand hall, tapestries, noble atmosphere",
		"village":  "Medieval village street, thatched roofs, cobblestones, bustling",
		"cave":     "Dark cave entrance, stalactites, mysterious glow, adventure awaits",
		"battle":   "Epic battle scene, warriors clashing, dramatic action",
		"treasure": "Ancient treasure chamber, gold coins, magical artifacts, glowing",
		"camp":     "Adventurer campsite, campfire, tents, night sky with stars",
		"ruins":    "Ancient ruins, crumbling stones, overgrown with vines, mysterious",
	}

	// Add scene prefix if available
	if prefix, ok := scenePrefix[sceneType]; ok {
		parts = append(parts, prefix)
	}

	// Add user description
	parts = append(parts, description)

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix)

	return strings.Join(parts, ", ")
}

// BuildMonsterPrompt creates a prompt for a monster illustration.
func BuildMonsterPrompt(monsterType string, style PromptStyle) string {
	var parts []string

	// Monster descriptions
	monsterDesc := map[string]string{
		"goblin":    "Goblin creature, small green-skinned humanoid, pointy ears, wicked grin, crude weapons",
		"orc":       "Orc warrior, muscular green-skinned humanoid, tusks, battle armor, fierce expression",
		"skeleton":  "Animated skeleton, undead warrior, ancient armor, glowing eye sockets, rusty sword",
		"zombie":    "Shambling zombie, decaying flesh, tattered clothes, mindless hunger",
		"dragon":    "Massive dragon, scales glinting, powerful wings, breathing fire, ancient and wise",
		"troll":     "Cave troll, massive and ugly, regenerating flesh, club weapon, dim-witted",
		"ogre":      "Ogre brute, huge humanoid, crude clothing, massive club, hungry expression",
		"wolf":      "Dire wolf, massive wolf, glowing eyes, bared fangs, pack hunter",
		"spider":    "Giant spider, massive arachnid, multiple eyes, venomous fangs, web silk",
		"rat":       "Giant rat, oversized rodent, disease-ridden, red eyes, sharp teeth",
		"bat":       "Giant bat, massive wingspan, echolocation screech, cave dwelling",
		"slime":     "Gelatinous ooze, translucent blob, dissolving debris inside, acidic",
		"ghost":     "Spectral ghost, translucent apparition, flowing ethereal form, haunting presence",
		"vampire":   "Vampire lord, pale aristocrat, red eyes, fangs, elegant dark clothing",
		"werewolf":  "Werewolf, half-man half-wolf, muscular, fur-covered, savage claws",
		"minotaur":  "Minotaur, bull-headed humanoid, massive horns, labyrinth dweller, battle axe",
		"basilisk":  "Basilisk, serpentine creature, deadly gaze, scales, venomous",
		"chimera":   "Chimera, lion body, goat head, serpent tail, fire breathing",
		"hydra":     "Hydra, multiple serpent heads, regenerating, water dwelling, massive",
		"lich":      "Lich king, skeletal mage, dark robes, phylactery, necromantic power",
	}

	// Get monster description or use generic
	desc := monsterType + ", fantasy creature, menacing"
	if d, ok := monsterDesc[strings.ToLower(monsterType)]; ok {
		desc = d
	}

	parts = append(parts, desc)

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix)

	return strings.Join(parts, ", ")
}

// BuildItemPrompt creates a prompt for a magical item illustration.
func BuildItemPrompt(itemType string, description string, style PromptStyle) string {
	var parts []string

	// Item type prefixes
	itemPrefix := map[string]string{
		"weapon":  "Magical weapon, glowing runes, ancient craftsmanship",
		"armor":   "Enchanted armor, gleaming metal, protective runes",
		"potion":  "Magical potion bottle, glowing liquid, alchemical symbols",
		"scroll":  "Ancient scroll, magical writing, glowing text, arcane symbols",
		"ring":    "Magical ring, precious metal, embedded gem, mystical glow",
		"amulet":  "Enchanted amulet, pendant on chain, magical symbols, glowing",
		"staff":   "Wizard staff, carved wood or crystal, magical focus, glowing tip",
		"wand":    "Magic wand, elegant design, arcane core, casting sparks",
		"book":    "Spellbook, leather bound, magical lock, glowing pages",
		"artifact": "Ancient artifact, mysterious origin, powerful aura, legendary",
	}

	// Add item prefix if available
	if prefix, ok := itemPrefix[strings.ToLower(itemType)]; ok {
		parts = append(parts, prefix)
	}

	// Add description
	if description != "" {
		parts = append(parts, description)
	}

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix)

	return strings.Join(parts, ", ")
}

// BuildLocationPrompt creates a prompt for a location map or overview.
func BuildLocationPrompt(locationType string, name string, style PromptStyle) string {
	var parts []string

	// Location descriptions
	locationDesc := map[string]string{
		"city":      "Medieval fantasy city, high walls, towers, bustling streets, market square",
		"town":      "Fantasy town, wooden buildings, town square, church steeple",
		"village":   "Small medieval village, thatched cottages, farmland, peaceful",
		"castle":    "Grand fantasy castle, stone walls, towers with flags, moat and drawbridge",
		"dungeon":   "Underground dungeon complex, stone corridors, torch sconces, mysterious",
		"forest":    "Enchanted forest, ancient trees, magical creatures, hidden paths",
		"mountain":  "Mountain fortress, carved into rock, winding paths, dramatic peaks",
		"swamp":     "Murky swamp, twisted trees, fog, dangerous waters, hidden dangers",
		"desert":    "Desert oasis, sand dunes, ancient ruins, mysterious temples",
		"coast":     "Coastal town, harbor with ships, lighthouse, fishing boats",
		"island":    "Mysterious island, tropical vegetation, hidden coves, ancient secrets",
		"underworld": "Underground cavern city, bioluminescent, stalactites, dark elf architecture",
	}

	// Get location description
	desc := locationType + ", fantasy location, detailed"
	if d, ok := locationDesc[strings.ToLower(locationType)]; ok {
		desc = d
	}

	parts = append(parts, desc)

	if name != "" {
		parts = append(parts, fmt.Sprintf("called %s", name))
	}

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix, "bird's eye view, map-like perspective")

	return strings.Join(parts, ", ")
}

// GetAvailableStyles returns the list of available art styles.
func GetAvailableStyles() []PromptStyle {
	return []PromptStyle{
		StyleRealistic,
		StylePainted,
		StyleIllustrated,
		StyleDarkFantasy,
		StyleEpic,
	}
}

// GetAvailableSceneTypes returns the list of available scene types.
func GetAvailableSceneTypes() []string {
	return []string{
		"tavern", "dungeon", "forest", "castle", "village",
		"cave", "battle", "treasure", "camp", "ruins",
	}
}

// GetAvailableMonsterTypes returns the list of available monster types.
func GetAvailableMonsterTypes() []string {
	return []string{
		"goblin", "orc", "skeleton", "zombie", "dragon",
		"troll", "ogre", "wolf", "spider", "rat",
		"bat", "slime", "ghost", "vampire", "werewolf",
		"minotaur", "basilisk", "chimera", "hydra", "lich",
	}
}

// GetAvailableItemTypes returns the list of available item types.
func GetAvailableItemTypes() []string {
	return []string{
		"weapon", "armor", "potion", "scroll", "ring",
		"amulet", "staff", "wand", "book", "artifact",
	}
}

// GetAvailableLocationTypes returns the list of available location types.
func GetAvailableLocationTypes() []string {
	return []string{
		"city", "town", "village", "castle", "dungeon",
		"forest", "mountain", "swamp", "desert", "coast",
		"island", "underworld",
	}
}
