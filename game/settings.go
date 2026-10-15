package game

import (
	"encoding/json"
	"os"
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
	return resolveUserConfigPath("settings.json")
}

// DefaultSettings returns the default game settings.
func DefaultSettings() *GameSettings {
	return &GameSettings{
		SoundEffectsEnabled: true,
		MissilesPerSilo:     10,
		UseCRT:              true,
	}
}

// Normalize ensures settings values remain within valid game bounds (10..100 in steps of 10).
func (s *GameSettings) Normalize() {
	s.MissilesPerSilo = ((s.MissilesPerSilo + 5) / 10) * 10
	if s.MissilesPerSilo < 10 {
		s.MissilesPerSilo = 10
	} else if s.MissilesPerSilo > 100 {
		s.MissilesPerSilo = 100
	}
}

// LoadSettings loads settings from disk, falling back to defaults if missing or corrupted.
func LoadSettings() *GameSettings {
	data, err := os.ReadFile(GetSettingsFilePath())
	if err == nil {
		var s GameSettings
		if json.Unmarshal(data, &s) == nil {
			s.Normalize()
			return &s
		}
	}

	s := DefaultSettings()
	_ = s.Save()
	return s
}

// Save writes current settings to disk.
func (s *GameSettings) Save() error {
	s.Normalize()
	return writeJSONFile(GetSettingsFilePath(), s)
}
