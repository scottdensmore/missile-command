package game

import (
	"encoding/json"
	"os"
	"sort"
)

// ScoreEntry represents an arcade high score entry with 3-letter initials.
type ScoreEntry struct {
	Initials string `json:"initials"`
	Score    int    `json:"score"`
}

// HighScores manages the top 10 arcade high scores.
type HighScores struct {
	Entries []ScoreEntry `json:"entries"`
}

const HighScoreFile = "highscores.json"

// DefaultHighScores returns the classic Atari Missile Command default high score table.
func DefaultHighScores() *HighScores {
	return &HighScores{
		Entries: []ScoreEntry{
			{Initials: "DFT", Score: 75000}, // Dave Theurer (Game Creator)
			{Initials: "SRC", Score: 65000},
			{Initials: "EPL", Score: 55000},
			{Initials: "GMD", Score: 45000},
			{Initials: "MOS", Score: 35000},
			{Initials: "JRH", Score: 25000},
			{Initials: "LSD", Score: 20000},
			{Initials: "TOM", Score: 15000},
			{Initials: "GUY", Score: 10000},
			{Initials: "ACE", Score: 5000},
		},
	}
}

// LoadHighScores loads scores from disk, falling back to defaults if not found.
func LoadHighScores() *HighScores {
	data, err := os.ReadFile(HighScoreFile)
	if err != nil {
		hs := DefaultHighScores()
		_ = hs.Save()
		return hs
	}

	var hs HighScores
	if err := json.Unmarshal(data, &hs); err != nil || len(hs.Entries) == 0 {
		hs := DefaultHighScores()
		_ = hs.Save()
		return hs
	}

	return &hs
}

// Save writes high scores to disk.
func (hs *HighScores) Save() error {
	data, err := json.MarshalIndent(hs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(HighScoreFile, data, 0644)
}

// IsHighScore checks if a given score qualifies for the top 10 list.
func (hs *HighScores) IsHighScore(score int) bool {
	if score <= 0 {
		return false
	}
	if len(hs.Entries) < 10 {
		return true
	}
	return score > hs.Entries[len(hs.Entries)-1].Score
}

// AddScore inserts a score in sorted order, truncating to 10 entries.
func (hs *HighScores) AddScore(initials string, score int) {
	if len(initials) > 3 {
		initials = initials[:3]
	}
	hs.Entries = append(hs.Entries, ScoreEntry{
		Initials: initials,
		Score:    score,
	})

	sort.SliceStable(hs.Entries, func(i, j int) bool {
		return hs.Entries[i].Score > hs.Entries[j].Score
	})

	if len(hs.Entries) > 10 {
		hs.Entries = hs.Entries[:10]
	}

	_ = hs.Save()
}

// GetTopScore returns the highest score currently recorded.
func (hs *HighScores) GetTopScore() int {
	if len(hs.Entries) > 0 {
		return hs.Entries[0].Score
	}
	return 75000
}
