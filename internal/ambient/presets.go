package ambient

// Fixed Lyria scene presets for the highest-stakes cues.
//
// Combat and aftermath music bypass the Claude prompt generator entirely, for
// two reasons:
//   - Latency: combat must land the instant initiative is rolled. Skipping the
//     Claude round-trip removes a hop so the cue is not late to the moment.
//   - Safety: a fixed prompt can never be steered by player-controlled text
//     (no prompt-injection surface), unlike a scene description fed to Claude.
//
// These presets are triggered by the canonical combat state (start_combat /
// end_combat), not by free-form prose.

// CombatPreset returns the stable Lyria parameters for an active combat.
func CombatPreset() *LyriaSceneParams {
	return &LyriaSceneParams{
		Prompt:      "epic battle orchestra, fast war drums, heroic strings, intense combat fantasy, driving rhythm",
		BPM:         140,
		Temperature: 1.0,
		DisplayName: "Combat",
	}
}

// AftermathPreset returns the stable parameters used when combat ends, easing
// back toward calm before the DM sets a location-specific ambiance.
func AftermathPreset() *LyriaSceneParams {
	return &LyriaSceneParams{
		Prompt:      "aftermath of battle, relief, slow strings, tense calm, fading tension",
		BPM:         70,
		Temperature: 0.9,
		DisplayName: "Après le combat",
	}
}
