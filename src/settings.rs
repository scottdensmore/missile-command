use crate::scores::{resolve_user_config_path, write_json_file};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::{OnceLock, RwLock};

/// User-configurable game options.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct GameSettings {
    pub sound_effects_enabled: bool,
    pub missiles_per_silo: i32,
    pub use_crt: bool,
}

static CUSTOM_SETTINGS_PATH: OnceLock<RwLock<Option<PathBuf>>> = OnceLock::new();

fn custom_settings_lock() -> &'static RwLock<Option<PathBuf>> {
    CUSTOM_SETTINGS_PATH.get_or_init(|| RwLock::new(None))
}

/// Sets a custom path for loading/saving settings (useful for isolated testing).
pub fn set_custom_settings_path<P: AsRef<Path>>(path: Option<P>) {
    let mut guard = custom_settings_lock()
        .write()
        .unwrap_or_else(|e| e.into_inner());
    *guard = path.map(|p| p.as_ref().to_path_buf());
}

/// Resolves the target settings file path in the user config directory.
pub fn get_settings_file_path() -> PathBuf {
    let guard = custom_settings_lock()
        .read()
        .unwrap_or_else(|e| e.into_inner());
    if let Some(ref p) = *guard {
        if !p.as_os_str().is_empty() {
            return p.clone();
        }
    }
    resolve_user_config_path("settings.json")
}

/// Returns the default game settings.
pub fn default_settings() -> GameSettings {
    GameSettings {
        sound_effects_enabled: true,
        missiles_per_silo: 10,
        use_crt: true,
    }
}

impl Default for GameSettings {
    fn default() -> Self {
        default_settings()
    }
}

impl GameSettings {
    /// Ensures settings values remain within valid game bounds (10..=100 in steps of 10).
    pub fn normalize(&mut self) {
        let rounded = ((self.missiles_per_silo + 5) / 10) * 10;
        self.missiles_per_silo = rounded.clamp(10, 100);
    }

    /// Writes current settings to disk.
    pub fn save(&mut self) -> std::io::Result<()> {
        self.normalize();
        write_json_file(&get_settings_file_path(), self)
    }
}

/// Loads settings from disk, falling back to defaults if missing or corrupted.
pub fn load_settings() -> GameSettings {
    let path = get_settings_file_path();
    if let Ok(data) = fs::read_to_string(&path) {
        if let Ok(mut s) = serde_json::from_str::<GameSettings>(&data) {
            s.normalize();
            return s;
        }
    }

    let mut s = default_settings();
    let _ = s.save();
    s
}
