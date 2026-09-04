package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsLoadDefaultAndNormalize(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "settings.json")
	SetCustomSettingsPath(testPath)
	defer SetCustomSettingsPath("")

	// Test default loading when file doesn't exist
	s := LoadSettings()
	if !s.SoundEffectsEnabled {
		t.Errorf("Expected SoundEffectsEnabled default to be true")
	}
	if s.MissilesPerSilo != 10 {
		t.Errorf("Expected MissilesPerSilo default to be 10, got %d", s.MissilesPerSilo)
	}
	if !s.UseCRT {
		t.Errorf("Expected UseCRT default to be true")
	}

	// Verify file was saved on default creation
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Fatalf("Expected settings file to be created at %s", testPath)
	}

	// Test normalization
	s.MissilesPerSilo = 5
	s.Normalize()
	if s.MissilesPerSilo != 10 {
		t.Errorf("Expected MissilesPerSilo clamped to min 10, got %d", s.MissilesPerSilo)
	}

	s.MissilesPerSilo = 150
	s.Normalize()
	if s.MissilesPerSilo != 100 {
		t.Errorf("Expected MissilesPerSilo clamped to max 100, got %d", s.MissilesPerSilo)
	}

	s.MissilesPerSilo = 24
	s.Normalize()
	if s.MissilesPerSilo != 20 {
		t.Errorf("Expected MissilesPerSilo 24 to round to 20, got %d", s.MissilesPerSilo)
	}

	s.MissilesPerSilo = 26
	s.Normalize()
	if s.MissilesPerSilo != 30 {
		t.Errorf("Expected MissilesPerSilo 26 to round to 30, got %d", s.MissilesPerSilo)
	}

	// Test saving and reloading
	s.SoundEffectsEnabled = false
	s.MissilesPerSilo = 70
	s.UseCRT = false
	if err := s.Save(); err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	reloaded := LoadSettings()
	if reloaded.SoundEffectsEnabled != false {
		t.Errorf("Expected SoundEffectsEnabled false, got true")
	}
	if reloaded.MissilesPerSilo != 70 {
		t.Errorf("Expected MissilesPerSilo 70, got %d", reloaded.MissilesPerSilo)
	}
	if reloaded.UseCRT != false {
		t.Errorf("Expected UseCRT false, got true")
	}
}
