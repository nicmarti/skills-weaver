package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Model represents a fal.ai image generation model.
type Model struct {
	ID       string // Full model ID (e.g., "fal-ai/flux/schnell")
	Short    string // Short name for filenames (e.g., "schnell")
	SyncURL  string // Synchronous API URL
	QueueURL string // Queue API URL
	MaxSteps int    // Maximum inference steps
	DefSteps int    // Default inference steps
}

// Available models
var (
	// ModelSchnell is the fast FLUX.1 schnell model (~$0.003/image)
	ModelSchnell = Model{
		ID:       "fal-ai/flux/schnell",
		Short:    "schnell",
		SyncURL:  "https://fal.run/fal-ai/flux/schnell",
		QueueURL: "https://queue.fal.run/fal-ai/flux/schnell",
		MaxSteps: 4,
		DefSteps: 4,
	}
	// ModelNanoBanana is Google's Nano Banana model with better quality (~$0.039/image)
	ModelNanoBanana = Model{
		ID:       "fal-ai/nano-banana",
		Short:    "banana",
		SyncURL:  "https://fal.run/fal-ai/nano-banana",
		QueueURL: "https://queue.fal.run/fal-ai/nano-banana",
		MaxSteps: 1,
		DefSteps: 1,
	}
)

// AvailableModels returns all available models.
func AvailableModels() map[string]Model {
	return map[string]Model{
		"schnell": ModelSchnell,
		"banana":  ModelNanoBanana,
	}
}

// GetModel returns a model by name, defaulting to schnell.
func GetModel(name string) Model {
	if m, ok := AvailableModels()[name]; ok {
		return m
	}
	return ModelSchnell
}

// Generator handles image generation via fal.ai API.
type Generator struct {
	apiKey     string
	httpClient *http.Client
	outputDir  string
}

// FalRequest represents the request body for fal.ai API.
type FalRequest struct {
	Prompt              string `json:"prompt"`
	NumImages           int    `json:"num_images,omitempty"`
	EnableSafetyChecker bool   `json:"enable_safety_checker"`
	OutputFormat        string `json:"output_format,omitempty"`
	ImageSize           string `json:"image_size,omitempty"`
	NumInferenceSteps   int    `json:"num_inference_steps,omitempty"`
}

// FalImage represents an image in the response.
type FalImage struct {
	URL         string `json:"url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ContentType string `json:"content_type"`
}

// FalResponse represents the response from fal.ai API.
type FalResponse struct {
	Images  []FalImage `json:"images"`
	Timings struct {
		Inference float64 `json:"inference"`
	} `json:"timings"`
	Seed      int64       `json:"seed"`
	HasNSFW   interface{} `json:"has_nsfw_concepts"` // Can be bool or []bool depending on API version
	RequestID string      `json:"request_id,omitempty"`
}

// FalQueueResponse represents the queue submission response.
type FalQueueResponse struct {
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// FalStatusResponse represents the status check response.
type FalStatusResponse struct {
	Status string `json:"status"`
}

// GeneratedImage holds information about a generated image.
type GeneratedImage struct {
	URL       string
	LocalPath string
	Width     int
	Height    int
	Prompt    string
}

// NewGenerator creates a new image generator.
func NewGenerator(outputDir string) (*Generator, error) {
	apiKey := os.Getenv("FAL_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("FAL_KEY environment variable not set")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	return &Generator{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		outputDir: outputDir,
	}, nil
}

// Generate creates an image from a prompt using the synchronous API.
func (g *Generator) Generate(prompt string, opts ...Option) (*GeneratedImage, error) {
	cfg := &config{
		numImages:     1,
		outputFormat:  "png",
		imageSize:     "landscape_16_9",
		safetyChecker: true,
		model:         ModelSchnell, // Default model
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Use model's default steps if not explicitly set
	if cfg.steps == 0 {
		cfg.steps = cfg.model.DefSteps
	}
	// Clamp steps to model's maximum
	if cfg.steps > cfg.model.MaxSteps {
		cfg.steps = cfg.model.MaxSteps
	}

	req := FalRequest{
		Prompt:              prompt,
		NumImages:           cfg.numImages,
		EnableSafetyChecker: cfg.safetyChecker,
		OutputFormat:        cfg.outputFormat,
		ImageSize:           cfg.imageSize,
		NumInferenceSteps:   cfg.steps,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	// Use model-specific URL
	httpReq, err := http.NewRequest("POST", cfg.model.SyncURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Key "+g.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var falResp FalResponse
	if err := json.Unmarshal(body, &falResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(falResp.Images) == 0 {
		return nil, fmt.Errorf("no images returned")
	}

	img := falResp.Images[0]

	// Build filename prefix with model name if prefix is provided
	filenamePrefix := cfg.filenamePrefix
	if filenamePrefix != "" {
		filenamePrefix = filenamePrefix + "_" + cfg.model.Short
	}

	// Download the image
	localPath, err := g.downloadImage(img.URL, cfg.outputFormat, filenamePrefix)
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}

	return &GeneratedImage{
		URL:       img.URL,
		LocalPath: localPath,
		Width:     img.Width,
		Height:    img.Height,
		Prompt:    prompt,
	}, nil
}

// GenerateAsync submits an image generation request to the queue.
func (g *Generator) GenerateAsync(prompt string, opts ...Option) (string, error) {
	cfg := &config{
		numImages:     1,
		outputFormat:  "png",
		imageSize:     "landscape_16_9",
		safetyChecker: true,
		model:         ModelSchnell, // Default model
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Use model's default steps if not explicitly set
	if cfg.steps == 0 {
		cfg.steps = cfg.model.DefSteps
	}

	req := FalRequest{
		Prompt:              prompt,
		NumImages:           cfg.numImages,
		EnableSafetyChecker: cfg.safetyChecker,
		OutputFormat:        cfg.outputFormat,
		ImageSize:           cfg.imageSize,
		NumInferenceSteps:   cfg.steps,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	// Use model-specific queue URL
	httpReq, err := http.NewRequest("POST", cfg.model.QueueURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Key "+g.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var queueResp FalQueueResponse
	if err := json.Unmarshal(body, &queueResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return queueResp.RequestID, nil
}

// CheckStatus checks the status of an async request.
func (g *Generator) CheckStatus(requestID string) (string, error) {
	url := fmt.Sprintf("https://queue.fal.run/fal-ai/flux/schnell/requests/%s/status", requestID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Key "+g.apiKey)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var statusResp FalStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	return statusResp.Status, nil
}

// GetResult retrieves the result of a completed async request.
func (g *Generator) GetResult(requestID string, outputFormat string) (*GeneratedImage, error) {
	url := fmt.Sprintf("https://queue.fal.run/fal-ai/flux/schnell/requests/%s", requestID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Key "+g.apiKey)

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var falResp FalResponse
	if err := json.Unmarshal(body, &falResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(falResp.Images) == 0 {
		return nil, fmt.Errorf("no images returned")
	}

	img := falResp.Images[0]

	// Download the image (no prefix for async results)
	localPath, err := g.downloadImage(img.URL, outputFormat, "")
	if err != nil {
		return nil, fmt.Errorf("downloading image: %w", err)
	}

	return &GeneratedImage{
		URL:       img.URL,
		LocalPath: localPath,
		Width:     img.Width,
		Height:    img.Height,
	}, nil
}

// downloadImage downloads an image from URL and saves it locally.
// If filenamePrefix is provided, the file will be named "prefix.format",
// otherwise it uses a timestamp-based name.
func (g *Generator) downloadImage(url string, format string, filenamePrefix string) (string, error) {
	resp, err := g.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetching image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image fetch failed with status %d", resp.StatusCode)
	}

	// Generate filename: use prefix if provided, otherwise use timestamp
	var filename string
	if filenamePrefix != "" {
		filename = fmt.Sprintf("%s.%s", filenamePrefix, format)
	} else {
		filename = fmt.Sprintf("image_%d.%s", time.Now().UnixNano(), format)
	}
	localPath := filepath.Join(g.outputDir, filename)

	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return localPath, nil
}

// Options

type config struct {
	numImages      int
	outputFormat   string
	imageSize      string
	safetyChecker  bool
	steps          int
	filenamePrefix string
	model          Model
}

// Option is a functional option for image generation.
type Option func(*config)

// WithNumImages sets the number of images to generate.
func WithNumImages(n int) Option {
	return func(c *config) {
		c.numImages = n
	}
}

// WithOutputFormat sets the output format (png, jpeg, webp).
func WithOutputFormat(format string) Option {
	return func(c *config) {
		c.outputFormat = format
	}
}

// WithImageSize sets the image size.
// Options: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9
func WithImageSize(size string) Option {
	return func(c *config) {
		c.imageSize = size
	}
}

// WithSafetyChecker enables or disables the safety checker.
func WithSafetyChecker(enabled bool) Option {
	return func(c *config) {
		c.safetyChecker = enabled
	}
}

// WithSteps sets the number of inference steps (1-4 for schnell).
func WithSteps(steps int) Option {
	return func(c *config) {
		if steps < 1 {
			steps = 1
		}
		if steps > 4 {
			steps = 4
		}
		c.steps = steps
	}
}

// WithFilenamePrefix sets a custom prefix for the generated image filename.
// The final filename will be: prefix.format (e.g., "journal_008_combat.png")
func WithFilenamePrefix(prefix string) Option {
	return func(c *config) {
		c.filenamePrefix = prefix
	}
}

// WithModel sets the fal.ai model to use.
// Available: "schnell" (fast), "dev" (quality), "pro", "pro11"
func WithModel(modelName string) Option {
	return func(c *config) {
		c.model = GetModel(modelName)
	}
}

// GetAvailableImageSizes returns the list of available image sizes.
func GetAvailableImageSizes() []string {
	return []string{
		"square_hd",      // 1024x1024
		"square",         // 512x512
		"portrait_4_3",   // 768x1024
		"portrait_16_9",  // 576x1024
		"landscape_4_3",  // 1024x768
		"landscape_16_9", // 1024x576
	}
}
