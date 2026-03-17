package image

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const googleImagenURL = "https://generativelanguage.googleapis.com/v1beta/models/imagen-4.0-generate-001:predict"

// GoogleGenerator handles image generation via Google Imagen API.
type GoogleGenerator struct {
	apiKey     string
	httpClient *http.Client
	outputDir  string
}

// googleRequest represents the request body for Google Imagen API.
type googleRequest struct {
	Instances  []googleInstance  `json:"instances"`
	Parameters googleParameters `json:"parameters"`
}

type googleInstance struct {
	Prompt string `json:"prompt"`
}

type googleParameters struct {
	SampleCount int    `json:"sampleCount"`
	AspectRatio string `json:"aspectRatio"`
}

// googleResponse represents the response from Google Imagen API.
type googleResponse struct {
	Predictions []googlePrediction `json:"predictions"`
}

type googlePrediction struct {
	BytesBase64Encoded string `json:"bytesBase64Encoded"`
	MimeType           string `json:"mimeType,omitempty"`
}

// NewGoogleGenerator creates a new Google Imagen image generator.
func NewGoogleGenerator(outputDir string) (*GoogleGenerator, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	return &GoogleGenerator{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		outputDir: outputDir,
	}, nil
}

// imageSizeToAspectRatio converts FAL.ai image size to Google Imagen aspect ratio.
func imageSizeToAspectRatio(imageSize string) string {
	switch imageSize {
	case "square_hd", "square":
		return "1:1"
	case "portrait_4_3":
		return "3:4"
	case "portrait_16_9":
		return "9:16"
	case "landscape_4_3":
		return "4:3"
	default: // landscape_16_9 and unknown sizes
		return "16:9"
	}
}

// Generate creates an image from a prompt using Google Imagen API.
// Note: model, steps, seed, and output format options are ignored (Google always uses
// imagen-4.0-generate-001 and PNG output).
func (g *GoogleGenerator) Generate(prompt string, opts ...Option) (*GeneratedImage, error) {
	cfg := &config{
		numImages:    1,
		outputFormat: "png",
		imageSize:    "landscape_16_9",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	aspectRatio := imageSizeToAspectRatio(cfg.imageSize)

	fmt.Printf("Modele: imagen-4.0-generate-001 (Google)\n")

	req := googleRequest{
		Instances: []googleInstance{
			{Prompt: prompt},
		},
		Parameters: googleParameters{
			SampleCount: 1,
			AspectRatio: aspectRatio,
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", googleImagenURL, g.apiKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
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

	var googleResp googleResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if len(googleResp.Predictions) == 0 {
		return nil, fmt.Errorf("no images returned")
	}

	prediction := googleResp.Predictions[0]
	if prediction.BytesBase64Encoded == "" {
		return nil, fmt.Errorf("empty image data in response")
	}

	// Decode base64 image data
	imageData, err := base64.StdEncoding.DecodeString(prediction.BytesBase64Encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding image data: %w", err)
	}

	// Generate filename with _imagen suffix (mirrors FAL's _<model.Short> pattern)
	var filename string
	if cfg.filenamePrefix != "" {
		filename = fmt.Sprintf("%s_imagen.png", cfg.filenamePrefix)
	} else {
		filename = fmt.Sprintf("image_%d_imagen.png", time.Now().UnixNano())
	}
	localPath := filepath.Join(g.outputDir, filename)

	if err := os.WriteFile(localPath, imageData, 0644); err != nil {
		return nil, fmt.Errorf("writing image file: %w", err)
	}

	return &GeneratedImage{
		URL:       "", // No CDN URL for Google Imagen (local only)
		LocalPath: localPath,
		Prompt:    prompt,
	}, nil
}
