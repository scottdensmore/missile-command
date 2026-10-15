package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// GameState represents the high-level arcade game state machine.
type GameState int

const (
	StateAttract GameState = iota
	StateWaveStart
	StatePlaying
	StateTally
	StateBonusRebuild
	StateTheEnd
	StateHighScoreEntry
	StateOptions
)

// Game implements the ebiten.Game interface for the arcade replica.
type Game struct {
	Width, Height int

	// Rendering pipeline & High Scores
	pipeline   *Pipeline
	highScores *HighScores

	// State Machine
	State     GameState
	StateTick int

	// Crosshair
	Crosshair Crosshair

	// Game Entities
	Cities      [6]City
	Batteries   [3]Battery
	ICBMs       []*ICBM
	SmartBombs  []*SmartBomb
	ActiveFlier *Flier
	ABMs        []*ABM
	Explosions  []*Explosion

	// Wave & Scoring
	Wave               int
	Score              int
	BonusCitiesStored  int
	NextBonusThreshold int
	Palette            Palette

	// Wave Spawning Management
	ICBMsRemaining       int
	SmartBombsLeft       int
	SpawnCooldown        int
	FlierCooldown        int
	FlierSpawnedThisWave bool
	WaveEndGraceTimer    int

	// End-of-Wave Tally State
	TallyPhase       int // 0: Ammo tally, 1: City tally, 2: Complete
	TallyTimer       int
	TallyBatteryIdx  int
	TallyCityIdx     int
	TallyScoreEarned int

	// Rebuild bonus city state
	RebuildCityIdx int
	RebuildTimer   int

	// "THE END" state
	TheEndTimer int

	// High score initials entry
	Initials      [3]rune
	InitialSlot   int
	CharSelectIdx int
	Letters       []rune

	// Attract Mode & AI Demo
	AttractDemoMode bool
	AttractTimer    int
	AIDemoShootTick int

	// On-screen HUD Notification
	NotificationText  string
	NotificationTimer int

	// Mouse, Pause and Input
	PrevMouseX, PrevMouseY int
	MouseCaptured          bool
	Paused                 bool
	Tick                   int

	// Options & Settings
	Settings          *GameSettings
	OptionSelectedIdx int
	PreviousState     GameState
}

var cityXPositions = [6]float64{40, 65, 90, 150, 175, 200}
var batteryXPositions = [3]float64{18, 128, 238}

// NewGame initializes the game engine and loads scores/audio/settings.
func NewGame() *Game {
	InitAudio()
	settings := LoadSettings()
	SetSoundEffectsEnabled(settings.SoundEffectsEnabled)

	g := &Game{
		Width:              1024,
		Height:             924,
		pipeline:           NewPipeline(),
		highScores:         LoadHighScores(),
		Settings:           settings,
		State:              StateAttract,
		Wave:               1,
		NextBonusThreshold: 10000,
		Letters:            []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ.!- 0123456789"),
		AttractTimer:       600, // 10 seconds Great Scores display before AI demo
		Crosshair:          Crosshair{Pos: Point{X: SimWidth / 2, Y: SimHeight / 2}},
		Palette:            GetPaletteForWave(1),
	}
	g.pipeline.UseCRT = settings.UseCRT
	return g
}

func (g *Game) enterAttractMode() {
	g.Paused = false
	StopAllContinuousSounds()
	g.State = StateAttract
	g.AttractDemoMode = false
	g.AttractTimer = 600
}

func (g *Game) showNotification(msg string) {
	g.NotificationText = msg
	g.NotificationTimer = 90
}

// StartGame starts a fresh 1-player game.
func StartGame(g *Game) {
	g.Wave = 1
	g.Score = 0
	g.BonusCitiesStored = 0
	g.NextBonusThreshold = 10000
	g.Palette = GetPaletteForWave(1)

	g.resetCities()
	g.resetSilos()
	g.startWave(1)
}

func (g *Game) resetCities() {
	for i, x := range cityXPositions {
		g.Cities[i] = City{
			Position:  Point{X: x, Y: 216},
			Destroyed: false,
		}
	}
}

func (g *Game) resetSilos() {
	maxAmmo := 10
	if g.Settings != nil && g.Settings.MissilesPerSilo > 0 {
		maxAmmo = g.Settings.MissilesPerSilo
	}
	for i, x := range batteryXPositions {
		g.Batteries[i] = Battery{
			Index:    i,
			Position: Point{X: x, Y: 214},
			MaxAmmo:  maxAmmo,
			Ammo:     maxAmmo,
		}
	}
}

func (g *Game) clearActiveEntities() {
	g.ICBMs = nil
	g.SmartBombs = nil
	g.ActiveFlier = nil
	g.ABMs = nil
	g.Explosions = nil
	StopAllContinuousSounds()
}

func (g *Game) startWave(wave int) {
	g.Wave = wave
	g.Palette = GetPaletteForWave(wave)
	wdata := GetWaveData(wave)

	g.ICBMsRemaining = wdata.TotalICBMs
	g.SmartBombsLeft = wdata.SmartBombs
	g.SpawnCooldown = 45
	g.FlierCooldown = wdata.FlierDelay
	g.FlierSpawnedThisWave = false
	g.WaveEndGraceTimer = 0

	g.clearActiveEntities()
	g.resetSilos()

	g.State = StateWaveStart
	g.StateTick = 90 // 1.5 seconds wave start banner
	PlaySirenAlert()
}

// Update handles game logic across all arcade states.
func (g *Game) Update() error {
	g.Tick++

	if done, err := g.handleGlobalInput(); done || err != nil {
		return err
	}

	if g.Paused {
		return nil
	}

	switch g.State {
	case StateAttract:
		g.updateAttract()
	case StateWaveStart:
		g.updateWaveStart()
	case StatePlaying:
		g.updatePlaying()
	case StateTally:
		g.updateTally()
	case StateBonusRebuild:
		g.updateBonusRebuild()
	case StateTheEnd:
		g.updateTheEnd()
	case StateHighScoreEntry:
		g.updateHighScoreEntry()
	case StateOptions:
		g.updateOptions()
	}

	return nil
}

func (g *Game) updateWaveStart() {
	g.StateTick--
	if g.StateTick <= 0 {
		g.State = StatePlaying
	}
}

func (g *Game) updatePlaying() {
	g.handleCrosshairInput()
	g.handleFiringInput()
	g.stepSimulation()

	// Bonus City threshold (Every 10,000 points)
	if g.Score >= g.NextBonusThreshold {
		g.BonusCitiesStored++
		g.NextBonusThreshold += 10000
		PlayBonusCitySound()
	}

	g.checkWaveCompletion()
}

func (g *Game) updateTally() {
	g.TallyTimer--
	if g.TallyTimer > 0 {
		return
	}
	g.TallyTimer = 8

	switch g.TallyPhase {
	case 0: // Tally Unused ABM Ammo (+5 * Multiplier each)
		for g.TallyBatteryIdx < len(g.Batteries) {
			bat := &g.Batteries[g.TallyBatteryIdx]
			if bat.Ammo > 0 {
				decrement := 1
				if bat.Ammo > 20 {
					decrement = 5
				} else if bat.Ammo > 10 {
					decrement = 2
				}
				decrement = min(decrement, bat.Ammo)
				bat.Ammo -= decrement
				pts := decrement * 5 * g.Palette.Multiplier
				g.Score += pts
				g.TallyScoreEarned += pts
				PlayTallySound()
				if bat.MaxAmmo > 10 {
					g.TallyTimer = 3
				}
				return
			}
			g.TallyBatteryIdx++
		}
		g.TallyPhase = 1

	case 1: // Tally Surviving Cities (+100 * Multiplier each)
		for g.TallyCityIdx < len(g.Cities) {
			idx := g.TallyCityIdx
			g.TallyCityIdx++
			if !g.Cities[idx].Destroyed {
				pts := 100 * g.Palette.Multiplier
				g.Score += pts
				g.TallyScoreEarned += pts
				PlayTallySound()
				return
			}
		}
		g.TallyPhase = 2
		g.TallyTimer = 55

	case 2: // Tally complete, check bonus city deployment
		g.State = StateBonusRebuild
		g.RebuildCityIdx = 0
		g.RebuildTimer = 20
	}
}

func (g *Game) updateBonusRebuild() {
	g.RebuildTimer--
	if g.RebuildTimer > 0 {
		return
	}

	if g.BonusCitiesStored > 0 {
		for i := range g.Cities {
			if g.Cities[i].Destroyed {
				g.Cities[i].Destroyed = false
				g.BonusCitiesStored--
				PlayBonusCitySound()
				g.RebuildTimer = 45
				return
			}
		}
	}

	g.startWave(g.Wave + 1)
}

func (g *Game) updateTheEnd() {
	g.TheEndTimer--
	if g.TheEndTimer <= 0 {
		if g.highScores.IsHighScore(g.Score) {
			g.State = StateHighScoreEntry
			g.Initials = [3]rune{'A', 'A', 'A'}
			g.InitialSlot = 0
			g.CharSelectIdx = 0
		} else {
			g.State = StateAttract
		}
	}
}

func (g *Game) updateHighScoreEntry() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.CharSelectIdx = (g.CharSelectIdx - 1 + len(g.Letters)) % len(g.Letters)
		g.Initials[g.InitialSlot] = g.Letters[g.CharSelectIdx]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.CharSelectIdx = (g.CharSelectIdx + 1) % len(g.Letters)
		g.Initials[g.InitialSlot] = g.Letters[g.CharSelectIdx]
	}

	for _, r := range g.Letters {
		if r >= 'A' && r <= 'Z' {
			k := ebiten.Key(int(ebiten.KeyA) + int(r-'A'))
			if inpututil.IsKeyJustPressed(k) {
				g.Initials[g.InitialSlot] = r
				g.advanceInitialSlot()
				return
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.advanceInitialSlot()
	}
}

func (g *Game) advanceInitialSlot() {
	PlayTallySound()
	g.InitialSlot++
	if g.InitialSlot >= 3 {
		g.highScores.AddScore(string(g.Initials[:]), g.Score)
		g.State = StateAttract
	} else {
		g.CharSelectIdx = 0
		g.Initials[g.InitialSlot] = g.Letters[0]
	}
}

// Layout maintains the native window dimensions for letterboxing onto the 256x231 buffer.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.Width = outsideWidth
	g.Height = outsideHeight
	return outsideWidth, outsideHeight
}
