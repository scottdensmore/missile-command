use serde::{Deserialize, Serialize};
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::{OnceLock, RwLock};

/// Represents an arcade high score entry with 3-letter initials.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct ScoreEntry {
    pub initials: String,
    pub score: i32,
}

/// Manages the top 10 arcade high scores.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct HighScores {
    pub entries: Vec<ScoreEntry>,
}

static CUSTOM_HIGH_SCORE_PATH: OnceLock<RwLock<Option<PathBuf>>> = OnceLock::new();

fn custom_path_lock() -> &'static RwLock<Option<PathBuf>> {
    CUSTOM_HIGH_SCORE_PATH.get_or_init(|| RwLock::new(None))
}

/// Sets a custom path for loading/saving high scores (useful for isolated testing).
pub fn set_custom_high_score_path<P: AsRef<Path>>(path: Option<P>) {
    let mut guard = custom_path_lock()
        .write()
        .unwrap_or_else(|e| e.into_inner());
    *guard = path.map(|p| p.as_ref().to_path_buf());
}

/// Resolves the target high score file path in the user config directory.
pub fn get_high_score_file_path() -> PathBuf {
    let guard = custom_path_lock().read().unwrap_or_else(|e| e.into_inner());
    if let Some(ref p) = *guard {
        if !p.as_os_str().is_empty() {
            return p.clone();
        }
    }
    resolve_user_config_path("highscores.json")
}

/// Returns the classic Atari Missile Command default high score table.
pub fn default_high_scores() -> HighScores {
    HighScores {
        entries: vec![
            ScoreEntry {
                initials: "DFT".to_string(),
                score: 75000,
            }, // Dave Theurer (Game Creator)
            ScoreEntry {
                initials: "SRC".to_string(),
                score: 65000,
            },
            ScoreEntry {
                initials: "EPL".to_string(),
                score: 55000,
            },
            ScoreEntry {
                initials: "GMD".to_string(),
                score: 45000,
            },
            ScoreEntry {
                initials: "MOS".to_string(),
                score: 35000,
            },
            ScoreEntry {
                initials: "JRH".to_string(),
                score: 25000,
            },
            ScoreEntry {
                initials: "LSD".to_string(),
                score: 20000,
            },
            ScoreEntry {
                initials: "TOM".to_string(),
                score: 15000,
            },
            ScoreEntry {
                initials: "GUY".to_string(),
                score: 10000,
            },
            ScoreEntry {
                initials: "ACE".to_string(),
                score: 5000,
            },
        ],
    }
}

/// Loads scores from disk, falling back to defaults if not found.
pub fn load_high_scores() -> HighScores {
    let path = get_high_score_file_path();
    if let Ok(data) = fs::read_to_string(&path) {
        if let Ok(hs) = serde_json::from_str::<HighScores>(&data) {
            if !hs.entries.is_empty() {
                return hs;
            }
        }
    }

    let hs = default_high_scores();
    let _ = hs.save();
    hs
}

impl HighScores {
    /// Writes high scores to disk.
    pub fn save(&self) -> std::io::Result<()> {
        write_json_file(&get_high_score_file_path(), self)
    }

    /// Checks if a given score qualifies for the top 10 list.
    pub fn is_high_score(&self, score: i32) -> bool {
        if score <= 0 {
            return false;
        }
        self.entries.len() < 10 || score > self.entries.last().map(|e| e.score).unwrap_or(0)
    }

    /// Inserts a score in sorted order, truncating to 10 entries.
    pub fn add_score(&mut self, initials: &str, score: i32) {
        let trimmed: String = initials.chars().take(3).collect();
        self.entries.push(ScoreEntry {
            initials: trimmed,
            score,
        });
        self.entries.sort_by_key(|a| std::cmp::Reverse(a.score));
        if self.entries.len() > 10 {
            self.entries.truncate(10);
        }
        let _ = self.save();
    }

    /// Returns the highest score currently recorded.
    pub fn get_top_score(&self) -> i32 {
        self.entries.first().map(|e| e.score).unwrap_or(75000)
    }
}

pub(crate) fn resolve_user_config_path(filename: &str) -> PathBuf {
    #[cfg(target_os = "macos")]
    if let Ok(home) = std::env::var("HOME") {
        if !home.is_empty() {
            return PathBuf::from(home)
                .join("Library")
                .join("Application Support")
                .join("missile-command")
                .join(filename);
        }
    }

    #[cfg(target_os = "windows")]
    if let Ok(appdata) = std::env::var("APPDATA") {
        if !appdata.is_empty() {
            return PathBuf::from(appdata)
                .join("missile-command")
                .join(filename);
        }
    }

    #[cfg(not(any(target_os = "macos", target_os = "windows")))]
    {
        if let Ok(xdg) = std::env::var("XDG_CONFIG_HOME") {
            if !xdg.is_empty() {
                return PathBuf::from(xdg).join("missile-command").join(filename);
            }
        }
        if let Ok(home) = std::env::var("HOME") {
            if !home.is_empty() {
                return PathBuf::from(home)
                    .join(".config")
                    .join("missile-command")
                    .join(filename);
            }
        }
    }

    PathBuf::from(filename)
}

pub(crate) fn write_json_file<T: Serialize>(path: &Path, value: &T) -> std::io::Result<()> {
    if let Some(parent) = path.parent() {
        if !parent.as_os_str().is_empty() && parent != Path::new(".") {
            let _ = fs::create_dir_all(parent);
        }
    }
    let data = serde_json::to_string_pretty(value)
        .map_err(|e| std::io::Error::new(std::io::ErrorKind::InvalidData, e))?;
    fs::write(path, data)
}
