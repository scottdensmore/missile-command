package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// GameSettings contains user-configurable options.
type GameSettings struct {
	SoundEffectsEnabled bool `json:"sound_effects_enabled"`
	MissilesPerSilo     int  `json:"missiles_per_silo"`
	UseCRT              bool `json:"use_crt"`
}

var (
	customSettingsPath string
	settingsMu         sync.RWMutex
)

// SetCustomSettingsPath sets a custom path for loading/saving settings (useful for testing).
func SetCustomSettingsPath(p string) {
	settingsMu.Lock()
	defer settingsMu.Unlock()
	customSettingsPath = p
}

// GetSettingsFilePath resolves the target settings file path in the user config directory.
func GetSettingsFilePath() string {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	if customSettingsPath != "" {
		return customSettingsPath
	}

	configDir, err := os.UserConfigDir()
	if err == nil && configDir != "" {
		return filepath.Join(configDir, "missile-command", "settings.json")
	}

	return "settings.json"
}

// DefaultSettings returns the default game settings.
func DefaultSettings() *GameSettings {
	return &GameSettings{
		SoundEffectsEnabled: true,
		MissilesPerSilo:     10,
		UseCRT:              true,
	}
}

// Normalize ensures settings values remain within valid game bounds.
func (s *GameSettings) Normalize() {
	if s.MissilesPerSilo < 10 {
		s.MissilesPerSilo = 10
	} else if s.MissilesPerSilo > 100 {
		s.MissilesPerSilo = 100
	} else {
		// Snap to nearest multiple of 10
		remainder := s.MissilesPerSilo % 10
		if remainder >= 5 {
			s.MissilesPerSilo += 10 - remainder
		} else {
			s.MissilesPerSilo -= remainder
		}
		if s.MissilesPerSilo < 10 {
			s.MissilesPerSilo = 10
		}
	}
}

// LoadSettings loads settings from disk, falling back to defaults if missing or corrupted.
func LoadSettings() *GameSettings {
	path := GetSettingsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		s := DefaultSettings()
		_ = s.Save()
		return s
	}

	var s GameSettings
	if err := json.Unmarshal(data, &s); err != nil {
		sDef := DefaultSettings()
		_ = sDef.Save()
		return sDef
	}

	s.Normalize()
	return &s
}

// Save writes current settings to disk.
func (s *GameSettings) Save() error {
	s.Normalize()
	path := GetSettingsFilePath()
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
