package image

import (
	"encoding/json"
	"testing"
)

func TestFalRequestWithSeed(t *testing.T) {
	seed := 1024
	req := FalRequest{
		Prompt:              "test prompt",
		NumImages:           1,
		EnableSafetyChecker: true,
		OutputFormat:        "png",
		ImageSize:           "square_hd",
		NumInferenceSteps:   28,
		Seed:                &seed,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Unmarshal to verify seed is present
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["seed"] == nil {
		t.Error("seed field is missing from JSON")
	}

	seedValue, ok := result["seed"].(float64)
	if !ok {
		t.Errorf("seed is not a number, got type %T", result["seed"])
	}

	if int(seedValue) != 1024 {
		t.Errorf("Expected seed=1024, got %d", int(seedValue))
	}

	t.Logf("✓ Seed parameter correctly included: %s", string(jsonData))
}

func TestFalRequestWithoutSeed(t *testing.T) {
	req := FalRequest{
		Prompt:              "test prompt without seed",
		NumImages:           1,
		EnableSafetyChecker: true,
		OutputFormat:        "png",
		ImageSize:           "square_hd",
		NumInferenceSteps:   8,
		Seed:                nil,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify seed is omitted when nil
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["seed"] != nil {
		t.Error("seed field should be omitted when nil")
	}

	t.Logf("✓ Seed parameter correctly omitted: %s", string(jsonData))
}

func TestSeedreamRequestOmitsInferenceSteps(t *testing.T) {
	// Seedream v4 doesn't use num_inference_steps, so it should be omitted when 0
	seed := 1024
	req := FalRequest{
		Prompt:              "test seedream prompt",
		NumImages:           1,
		EnableSafetyChecker: true,
		OutputFormat:        "png",
		ImageSize:           "square_hd",
		NumInferenceSteps:   0, // Seedream doesn't use this
		Seed:                &seed,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify num_inference_steps is omitted when 0
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["num_inference_steps"] != nil {
		t.Errorf("num_inference_steps should be omitted when 0, but got: %v", result["num_inference_steps"])
	}

	if result["seed"] == nil {
		t.Error("seed should be present for seedream")
	}

	t.Logf("✓ Seedream request correctly omits num_inference_steps: %s", string(jsonData))
}
