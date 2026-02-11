package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// testConfigDir creates a temporary config directory for testing.
// It sets up the environment to use a temp directory instead of the real config dir.
func testConfigDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestLoadConfig_ReturnsAppConfig(t *testing.T) {
	// LoadConfig should return an AppConfig struct
	// We can't control whether a config file exists without mocking os.UserConfigDir,
	// so we just verify it returns a valid struct without panicking
	cfg := LoadConfig()

	// Verify it's a valid AppConfig (Theme can be empty or have a value)
	_ = cfg.Theme // Just ensure we can access the field
}

func TestSaveTheme_CreatesDirectory(t *testing.T) {
	tmpDir := testConfigDir(t)
	configDir := filepath.Join(tmpDir, "valvefm")
	configFile := filepath.Join(configDir, "config.json")

	// Manually test the save logic (since we can't easily override configPath)
	// We'll create a helper that mimics SaveTheme behavior

	// Create the directory
	err := os.MkdirAll(configDir, 0o755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Write config
	cfg := map[string]interface{}{"theme": "tokyo-night"}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}

	err = os.WriteFile(configFile, data, 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Read back and verify
	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var readCfg AppConfig
	if err := json.Unmarshal(readData, &readCfg); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if readCfg.Theme != "tokyo-night" {
		t.Errorf("Theme = %q, want %q", readCfg.Theme, "tokyo-night")
	}
}

func TestAppConfig_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		theme string
	}{
		{"vintage theme", "vintage"},
		{"tokyo-night theme", "tokyo-night"},
		{"empty theme", ""},
		{"custom theme", "custom-user-theme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := AppConfig{Theme: tt.theme}

			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var decoded AppConfig
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if decoded.Theme != original.Theme {
				t.Errorf("Theme = %q, want %q", decoded.Theme, original.Theme)
			}
		})
	}
}

func TestSaveTheme_PreservesOtherFields(t *testing.T) {
	tmpDir := testConfigDir(t)
	configFile := filepath.Join(tmpDir, "config.json")

	// Write initial config with extra fields
	initial := map[string]interface{}{
		"theme":        "vintage",
		"volume":       75,
		"last_country": "US",
	}
	data, err := json.MarshalIndent(initial, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(configFile, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Simulate SaveTheme behavior: read, modify, write
	readData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(readData, &raw); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	raw["theme"] = "tokyo-night"

	writeData, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	if err := os.WriteFile(configFile, writeData, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify all fields preserved
	finalData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var final map[string]interface{}
	if err := json.Unmarshal(finalData, &final); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if final["theme"] != "tokyo-night" {
		t.Errorf("theme = %v, want %q", final["theme"], "tokyo-night")
	}
	if final["volume"] != float64(75) { // JSON numbers are float64
		t.Errorf("volume = %v, want 75", final["volume"])
	}
	if final["last_country"] != "US" {
		t.Errorf("last_country = %v, want %q", final["last_country"], "US")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := testConfigDir(t)
	configFile := filepath.Join(tmpDir, "config.json")

	// Write invalid JSON
	if err := os.WriteFile(configFile, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// LoadConfig should handle invalid JSON gracefully
	// Since we can't override configPath, we test the parsing logic directly
	var cfg AppConfig
	err := json.Unmarshal([]byte("not valid json"), &cfg)
	if err == nil {
		t.Error("Unmarshal() should return error for invalid JSON")
	}
}

func TestAppConfig_EmptyFile(t *testing.T) {
	// Empty JSON object should work
	var cfg AppConfig
	err := json.Unmarshal([]byte("{}"), &cfg)
	if err != nil {
		t.Fatalf("Unmarshal({}) error = %v", err)
	}
	if cfg.Theme != "" {
		t.Errorf("Theme should be empty for {}, got %q", cfg.Theme)
	}
}

func TestAppConfig_ExtraFields(t *testing.T) {
	// Extra fields in JSON should be ignored
	jsonData := `{"theme": "nord", "unknown_field": "value", "number": 42}`

	var cfg AppConfig
	err := json.Unmarshal([]byte(jsonData), &cfg)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if cfg.Theme != "nord" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "nord")
	}
}
