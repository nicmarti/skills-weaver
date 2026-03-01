// Package ambient provides ambient music generation via Google Lyria RealTime API.
package ambient

// SetupMessage is sent to Lyria to initialize the session.
type SetupMessage struct {
	Setup LyriaSetup `json:"setup"`
}

// LyriaSetup contains the model to use.
type LyriaSetup struct {
	Model string `json:"model"`
}

// WeightedPrompt is a text prompt with a weight for music generation.
type WeightedPrompt struct {
	Text   string  `json:"text"`
	Weight float64 `json:"weight"`
}

// MusicGenerationConfig controls tempo and creativity.
type MusicGenerationConfig struct {
	BPM         int     `json:"bpm,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// ClientContentMessage wraps weighted prompts — matches SDK: {"clientContent": {"weightedPrompts": [...]}}
type ClientContentMessage struct {
	ClientContent ClientContent `json:"clientContent"`
}

// ClientContent holds the weighted prompts for the current scene.
type ClientContent struct {
	WeightedPrompts []WeightedPrompt `json:"weightedPrompts"`
}

// MusicGenerationConfigMessage sets BPM and temperature — matches SDK: {"musicGenerationConfig": {...}}
type MusicGenerationConfigMessage struct {
	MusicGenerationConfig MusicGenerationConfig `json:"musicGenerationConfig"`
}

// PlaybackControlMessage starts or stops music — matches SDK: {"playbackControl": "PLAY"}
type PlaybackControlMessage struct {
	PlaybackControl string `json:"playbackControl"`
}

// ServerMessage is received from Lyria.
type ServerMessage struct {
	SetupComplete *struct{}      `json:"setupComplete"`
	ServerContent *ServerContent `json:"serverContent"`
}

// ServerContent holds audio data chunks.
type ServerContent struct {
	AudioChunks []AudioChunk `json:"audioChunks"`
}

// AudioChunk contains base64-encoded PCM16 audio data.
type AudioChunk struct {
	Data string `json:"data"`
}
