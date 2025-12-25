package image

import (
	"fmt"
	"strings"

	"dungeons/internal/adventure"
	"dungeons/internal/character"
	"dungeons/internal/npc"
)

// PromptStyle defines the art style for generation.
type PromptStyle string

const (
	StyleIllustrated PromptStyle = "illustrated"
	StyleDarkFantasy PromptStyle = "dark_fantasy"
)

// StyleSuffixes maps styles to prompt suffixes.
var StyleSuffixes = map[PromptStyle]string{
	StyleIllustrated: "digital illustration, vibrant colors, detailed linework, fantasy RPG art style",
	StyleDarkFantasy: "dark fantasy art, moody lighting, dramatic shadows, gritty atmosphere",
}

// BasePromptSuffix is added to all prompts for consistent quality.
const BasePromptSuffix = "medieval fantasy setting, dungeons and dragons style, high quality"

// BuildCharacterPrompt creates a prompt for a character portrait.
func BuildCharacterPrompt(char *character.Character, style PromptStyle) string {
	var parts []string

	// If character has detailed appearance, use it
	if char.Appearance != nil && (char.Appearance.Age > 0 || char.Appearance.Build != "" || char.Appearance.HairColor != "") {
		return buildDetailedCharacterPrompt(char, style)
	}

	// Fallback to generic descriptions for characters without appearance
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

	// Build prompt - use gender if available even in minimal appearance
	gender := ""
	if char.Appearance != nil && char.Appearance.Gender != "" {
		gender = char.Appearance.Gender + " "
	}

	raceD := raceDesc[char.Race]
	if raceD == "" {
		raceD = char.Race
	}
	classD := classDesc[char.Class]
	if classD == "" {
		classD = char.Class
	}

	parts = append(parts, fmt.Sprintf("Portrait of a %s%s %s",
		gender, raceD, classD))

	parts = append(parts, fmt.Sprintf("named %s", char.Name))

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	parts = append(parts, BasePromptSuffix)

	return strings.Join(parts, ", ")
}

// buildDetailedCharacterPrompt builds a detailed prompt using character appearance.
func buildDetailedCharacterPrompt(char *character.Character, style PromptStyle) string {
	var parts []string
	app := char.Appearance

	// Start with character name and basic info
	parts = append(parts, fmt.Sprintf("Portrait of %s", char.Name))

	// Gender prefix
	genderPrefix := ""
	if app.Gender != "" {
		genderPrefix = app.Gender + " "
	}

	// Age and race
	if app.Age > 0 {
		parts = append(parts, fmt.Sprintf("%d-year-old %s%s", app.Age, genderPrefix, char.Race))
	} else {
		parts = append(parts, genderPrefix+char.Race)
	}

	// Class
	parts = append(parts, char.Class)

	// Physical build and height
	if app.Height != "" && app.Build != "" {
		parts = append(parts, fmt.Sprintf("%s and %s build", app.Height, app.Build))
	} else if app.Height != "" {
		parts = append(parts, app.Height)
	} else if app.Build != "" {
		parts = append(parts, app.Build+" build")
	}

	// Hair
	if app.HairStyle != "" && app.HairColor != "" {
		parts = append(parts, fmt.Sprintf("%s %s hair", app.HairColor, app.HairStyle))
	} else if app.HairColor != "" {
		parts = append(parts, app.HairColor+" hair")
	} else if app.HairStyle != "" {
		parts = append(parts, app.HairStyle+" hair")
	}

	// Eyes
	if app.EyeColor != "" {
		parts = append(parts, app.EyeColor+" eyes")
	}

	// Skin tone
	if app.SkinTone != "" {
		parts = append(parts, app.SkinTone+" skin")
	}

	// Facial features
	if app.FacialFeature != "" {
		parts = append(parts, app.FacialFeature)
	}

	// Distinctive features
	if app.DistinctiveFeature != "" {
		parts = append(parts, app.DistinctiveFeature)
	}

	// Armor/clothing
	if app.ArmorDescription != "" {
		parts = append(parts, "wearing "+app.ArmorDescription)
	}

	// Weapon (visible in portrait)
	if app.WeaponDescription != "" {
		parts = append(parts, "holding "+app.WeaponDescription)
	}

	// Accessories
	if app.Accessories != "" {
		parts = append(parts, "with "+app.Accessories)
	}

	// Add style suffix
	if suffix, ok := StyleSuffixes[style]; ok {
		parts = append(parts, suffix)
	}

	// Add base suffix
	parts = append(parts, BasePromptSuffix)
	parts = append(parts, "professional portrait photography, detailed character art")

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
		"goblin":   "Goblin creature, small green-skinned humanoid, pointy ears, wicked grin, crude weapons",
		"orc":      "Orc warrior, muscular green-skinned humanoid, tusks, battle armor, fierce expression",
		"skeleton": "Animated skeleton, undead warrior, ancient armor, glowing eye sockets, rusty sword",
		"zombie":   "Shambling zombie, decaying flesh, tattered clothes, mindless hunger",
		"dragon":   "Massive dragon, scales glinting, powerful wings, breathing fire, ancient and wise",
		"troll":    "Cave troll, massive and ugly, regenerating flesh, club weapon, dim-witted",
		"ogre":     "Ogre brute, huge humanoid, crude clothing, massive club, hungry expression",
		"wolf":     "Dire wolf, massive wolf, glowing eyes, bared fangs, pack hunter",
		"spider":   "Giant spider, massive arachnid, multiple eyes, venomous fangs, web silk",
		"rat":      "Giant rat, oversized rodent, disease-ridden, red eyes, sharp teeth",
		"bat":      "Giant bat, massive wingspan, echolocation screech, cave dwelling",
		"slime":    "Gelatinous ooze, translucent blob, dissolving debris inside, acidic",
		"ghost":    "Spectral ghost, translucent apparition, flowing ethereal form, haunting presence",
		"vampire":  "Vampire lord, pale aristocrat, red eyes, fangs, elegant dark clothing",
		"werewolf": "Werewolf, half-man half-wolf, muscular, fur-covered, savage claws",
		"minotaur": "Minotaur, bull-headed humanoid, massive horns, labyrinth dweller, battle axe",
		"basilisk": "Basilisk, serpentine creature, deadly gaze, scales, venomous",
		"chimera":  "Chimera, lion body, goat head, serpent tail, fire breathing",
		"hydra":    "Hydra, multiple serpent heads, regenerating, water dwelling, massive",
		"lich":     "Lich king, skeletal mage, dark robes, phylactery, necromantic power",
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
		"weapon":   "Magical weapon, glowing runes, ancient craftsmanship",
		"armor":    "Enchanted armor, gleaming metal, protective runes",
		"potion":   "Magical potion bottle, glowing liquid, alchemical symbols",
		"scroll":   "Ancient scroll, magical writing, glowing text, arcane symbols",
		"ring":     "Magical ring, precious metal, embedded gem, mystical glow",
		"amulet":   "Enchanted amulet, pendant on chain, magical symbols, glowing",
		"staff":    "Wizard staff, carved wood or crystal, magical focus, glowing tip",
		"wand":     "Magic wand, elegant design, arcane core, casting sparks",
		"book":     "Spellbook, leather bound, magical lock, glowing pages",
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
		"city":       "Medieval fantasy city, high walls, towers, bustling streets, market square",
		"town":       "Fantasy town, wooden buildings, town square, church steeple",
		"village":    "Small medieval village, thatched cottages, farmland, peaceful",
		"castle":     "Grand fantasy castle, stone walls, towers with flags, moat and drawbridge",
		"dungeon":    "Underground dungeon complex, stone corridors, torch sconces, mysterious",
		"forest":     "Enchanted forest, ancient trees, magical creatures, hidden paths",
		"mountain":   "Mountain fortress, carved into rock, winding paths, dramatic peaks",
		"swamp":      "Murky swamp, twisted trees, fog, dangerous waters, hidden dangers",
		"desert":     "Desert oasis, sand dunes, ancient ruins, mysterious temples",
		"coast":      "Coastal town, harbor with ships, lighthouse, fishing boats",
		"island":     "Mysterious island, tropical vegetation, hidden coves, ancient secrets",
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
		StyleIllustrated,
		StyleDarkFantasy,
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

// JournalEntryPrompt represents a prompt generated from a journal entry.
type JournalEntryPrompt struct {
	EntryID   int
	EntryType string
	Content   string
	Prompt    string
	Style     PromptStyle
	ImageSize string
}

// IllustratableTypes returns journal entry types that can be illustrated.
func IllustratableTypes() []string {
	return []string{
		"combat",
		"exploration",
		"story",    // Major story events
		"note",
		"discovery",
		"loot",
		"session", // Only session end summaries
	}
}

// IsIllustratable checks if a journal entry type can be illustrated.
func IsIllustratable(entryType string) bool {
	for _, t := range IllustratableTypes() {
		if t == entryType {
			return true
		}
	}
	return false
}

// BuildJournalEntryPrompt creates a prompt for a journal entry.
// BuildJournalEntryPromptWithCharacters builds prompt with character context for consistency.
func BuildJournalEntryPromptWithCharacters(
	entry adventure.JournalEntry,
	characters []*character.Character,
) *JournalEntryPrompt {
	// Build base prompt without character context using the legacy function
	baseResult := BuildJournalEntryPromptWithoutCharacters(entry)
	if baseResult == nil {
		return nil
	}

	// If no characters, return base result
	if len(characters) == 0 {
		return baseResult
	}

	// Build character references string
	charRefs := buildCharacterReferences(characters)

	// Inject character references into the prompt
	if charRefs != "" {
		baseResult.Prompt = injectCharacterContext(baseResult.Prompt, charRefs, entry.Type)
	}

	return baseResult
}

// buildDetailedCharacterReference creates a rich description for a character in journal context
func buildDetailedCharacterReference(char *character.Character) string {
	if char.Appearance == nil {
		// Fallback to short snippet if no appearance data
		return char.GetImagePromptSnippet()
	}

	a := char.Appearance
	var parts []string

	// Gender prefix
	genderPrefix := ""
	if a.Gender != "" {
		genderPrefix = a.Gender + " "
	}

	// Name and age/class
	if a.Age > 0 {
		parts = append(parts, fmt.Sprintf("%s, a %d-year-old %s%s %s", char.Name, a.Age, genderPrefix, char.Race, char.Class))
	} else {
		parts = append(parts, fmt.Sprintf("%s, a %s%s %s", char.Name, genderPrefix, char.Race, char.Class))
	}

	// Physical traits
	if a.Height != "" || a.Build != "" {
		traits := []string{}
		if a.Height != "" {
			traits = append(traits, a.Height)
		}
		if a.Build != "" {
			traits = append(traits, a.Build)
		}
		parts = append(parts, strings.Join(traits, " and "))
	}

	// Hair and eyes
	if a.HairColor != "" || a.EyeColor != "" {
		features := []string{}
		if a.HairColor != "" && a.HairStyle != "" {
			features = append(features, fmt.Sprintf("%s %s", a.HairColor, a.HairStyle))
		} else if a.HairColor != "" {
			features = append(features, a.HairColor)
		}
		if a.EyeColor != "" {
			features = append(features, fmt.Sprintf("%s eyes", a.EyeColor))
		}
		if len(features) > 0 {
			parts = append(parts, strings.Join(features, " and "))
		}
	}

	// Skin tone
	if a.SkinTone != "" {
		parts = append(parts, a.SkinTone)
	}

	// Distinctive features
	if a.DistinctiveFeature != "" {
		parts = append(parts, fmt.Sprintf("with %s", a.DistinctiveFeature))
	}

	// Equipment
	equipment := []string{}
	if a.ArmorDescription != "" {
		equipment = append(equipment, a.ArmorDescription)
	}
	if a.WeaponDescription != "" {
		equipment = append(equipment, a.WeaponDescription)
	}
	if a.Accessories != "" {
		equipment = append(equipment, a.Accessories)
	}
	if len(equipment) > 0 {
		parts = append(parts, fmt.Sprintf("carrying %s", strings.Join(equipment, " and ")))
	}

	return strings.Join(parts, ", ")
}

// buildCharacterReferences creates detailed character list for journal prompts.
func buildCharacterReferences(characters []*character.Character) string {
	if len(characters) == 0 {
		return ""
	}

	refs := make([]string, 0, len(characters))
	for _, c := range characters {
		refs = append(refs, buildDetailedCharacterReference(c))
	}

	return strings.Join(refs, "; ")
}

// injectCharacterContext adds character references to the prompt.
func injectCharacterContext(basePrompt, charRefs, entryType string) string {
	// Find the first period or colon to inject characters
	parts := strings.SplitN(basePrompt, ":", 2)
	if len(parts) == 2 {
		// Format: "Prefix: Description. Suffix"
		// Inject: "Prefix featuring Characters: Description. Suffix"
		return parts[0] + " featuring " + charRefs + ":" + parts[1]
	}

	// Fallback: just prepend character info
	return "Featuring " + charRefs + ". " + basePrompt
}

// BuildJournalEntryPromptWithoutCharacters is the original logic without character context.
func BuildJournalEntryPromptWithoutCharacters(entry adventure.JournalEntry) *JournalEntryPrompt {
	var prompt string
	var style PromptStyle
	var imageSize string

	switch entry.Type {
	case "combat":
		// Combat scenes are epic and action-packed
		style = StyleIllustrated
		imageSize = "landscape_16_9"
		prompt = buildCombatPrompt(entry)

	case "exploration":
		// Exploration scenes are atmospheric
		style = StyleIllustrated
		imageSize = "landscape_16_9"
		prompt = buildExplorationPrompt(entry)

	case "discovery":
		// Discoveries focus on items or revelations
		style = StyleIllustrated
		imageSize = "landscape_16_9"
		prompt = buildDiscoveryPrompt(entry)

	case "loot":
		// Treasure scenes
		style = StyleIllustrated
		imageSize = "landscape_16_9"
		prompt = buildLootPrompt(entry)

	case "note":
		// Notes are narrative moments, conversations, or observations
		style = StyleIllustrated
		imageSize = "landscape_16_9"
		prompt = buildNotePrompt(entry)

	case "session":
		// Session summaries get epic treatment if they contain victory/defeat
		style = StyleIllustrated
		imageSize = "landscape_16_9"
		prompt = buildSessionPrompt(entry)

	default:
		return nil
	}

	return &JournalEntryPrompt{
		EntryID:   entry.ID,
		EntryType: entry.Type,
		Content:   entry.Content,
		Prompt:    prompt,
		Style:     style,
		ImageSize: imageSize,
	}
}

// BuildJournalEntryPrompt creates a prompt for a journal entry (legacy, without character context).
func BuildJournalEntryPrompt(entry adventure.JournalEntry) *JournalEntryPrompt {
	return BuildJournalEntryPromptWithCharacters(entry, nil)
}

// getEntryDescription returns the best description available (fallback: Description > DescriptionFr > Content).
func getEntryDescription(entry adventure.JournalEntry) string {
	if entry.Description != "" {
		return entry.Description
	}
	if entry.DescriptionFr != "" {
		return entry.DescriptionFr
	}
	return entry.Content
}

// buildCombatPrompt creates a prompt for combat entries.
func buildCombatPrompt(entry adventure.JournalEntry) string {
	baseText := getEntryDescription(entry)
	parts := []string{
		"Epic fantasy battle scene",
		baseText,
		StyleSuffixes[StyleIllustrated],
		BasePromptSuffix,
		"dynamic action, dramatic lighting",
	}
	return strings.Join(parts, ", ")
}

// buildExplorationPrompt creates a prompt for exploration entries.
func buildExplorationPrompt(entry adventure.JournalEntry) string {
	baseText := getEntryDescription(entry)
	parts := []string{
		"",
		baseText,
		StyleSuffixes[StyleIllustrated],
		BasePromptSuffix,
		"",
	}
	return strings.Join(parts, ", ")
}

// buildDiscoveryPrompt creates a prompt for discovery entries.
func buildDiscoveryPrompt(entry adventure.JournalEntry) string {
	baseText := getEntryDescription(entry)
	parts := []string{
		"",
		baseText,
		StyleSuffixes[StyleIllustrated],
		BasePromptSuffix,
		"",
	}
	return strings.Join(parts, ", ")
}

// buildLootPrompt creates a prompt for loot entries.
func buildLootPrompt(entry adventure.JournalEntry) string {
	baseText := getEntryDescription(entry)
	parts := []string{
		"",
		baseText,
		StyleSuffixes[StyleIllustrated],
		BasePromptSuffix,
		"",
	}
	return strings.Join(parts, ", ")
}

// buildNotePrompt creates a prompt for note entries.
func buildNotePrompt(entry adventure.JournalEntry) string {
	baseText := getEntryDescription(entry)
	parts := []string{
		"",
		baseText,
		StyleSuffixes[StyleIllustrated],
		BasePromptSuffix,
		"",
	}
	return strings.Join(parts, ", ")
}

// buildSessionPrompt creates a prompt for session summary entries.
func buildSessionPrompt(entry adventure.JournalEntry) string {
	baseText := getEntryDescription(entry)
	// Check for victory/defeat keywords
	lowerContent := strings.ToLower(baseText)
	var mood string
	if strings.Contains(lowerContent, "victoire") || strings.Contains(lowerContent, "victory") {
		mood = "triumphant heroes, celebration, victory"
	} else if strings.Contains(lowerContent, "défaite") || strings.Contains(lowerContent, "defeat") || strings.Contains(lowerContent, "mort") {
		mood = "somber scene, aftermath of battle, fallen heroes"
	} else {
		mood = "adventurers resting, campfire, end of journey"
	}

	parts := []string{
		"Epic fantasy scene",
		baseText,
		mood,
		StyleSuffixes[StyleIllustrated],
		BasePromptSuffix,
	}
	return strings.Join(parts, ", ")
}
