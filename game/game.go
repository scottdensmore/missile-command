package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
		Score:              0,
		NextBonusThreshold: 10000,
		Letters:            []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ.!- 0123456789"),
		AttractDemoMode:    false,
		AttractTimer:       600, // 10 seconds Great Scores display before AI demo
	}
	g.pipeline.UseCRT = settings.UseCRT

	g.Crosshair.Pos = Point{X: SimWidth / 2, Y: SimHeight / 2}
	g.Palette = GetPaletteForWave(1)
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

	// Rebuild all 6 cities
	// Left valley: [31, 115] -> cities at X: 40, 65, 90 (9px margins & gaps)
	// Right valley: [141, 225] -> cities at X: 150, 175, 200 (9px margins & gaps)
	cityXs := []float64{40, 65, 90, 150, 175, 200}
	for i, x := range cityXs {
		g.Cities[i] = City{
			Position:  Point{X: x, Y: 216},
			Destroyed: false,
		}
	}

	g.resetSilos()
	g.startWave(1)
}

func (g *Game) resetSilos() {
	maxAmmo := 10
	if g.Settings != nil && g.Settings.MissilesPerSilo > 0 {
		maxAmmo = g.Settings.MissilesPerSilo
	}
	batteryXs := []float64{18, 128, 238}
	for i, x := range batteryXs {
		g.Batteries[i] = Battery{
			Index:            i,
			Position:         Point{X: x, Y: 214},
			MaxAmmo:          maxAmmo,
			Ammo:             maxAmmo,
			Destroyed:        false,
			LowWarningPlayed: false,
		}
	}
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

	g.ICBMs = nil
	g.SmartBombs = nil
	g.ActiveFlier = nil
	g.ABMs = nil
	g.Explosions = nil
	StopAllContinuousSounds()

	g.resetSilos()

	g.State = StateWaveStart
	g.StateTick = 90 // 1.5 seconds wave start banner
	PlaySirenAlert()
}

// Update handles game logic across all arcade states.
func (g *Game) Update() error {
	g.Tick++

	// Toggle Mouse Capture on Click
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !g.MouseCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		g.MouseCaptured = true
		g.PrevMouseX, g.PrevMouseY = ebiten.CursorPosition()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.MouseCaptured {
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
		g.MouseCaptured = false
	}

	// CRT Filter Toggle (F1 / Tab)
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) || inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.pipeline.UseCRT = !g.pipeline.UseCRT
		if g.Settings != nil {
			g.Settings.UseCRT = g.pipeline.UseCRT
			_ = g.Settings.Save()
		}
		if g.pipeline.UseCRT {
			g.showNotification("CRT SHADER: ON")
		} else {
			g.showNotification("CRT SHADER: OFF")
		}
	}

	// Fullscreen Toggle (F11 / Alt+Enter)
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) ||
		(inpututil.IsKeyJustPressed(ebiten.KeyEnter) && (ebiten.IsKeyPressed(ebiten.KeyAlt) || ebiten.IsKeyPressed(ebiten.KeyAltRight) || ebiten.IsKeyPressed(ebiten.KeyAltLeft))) {
		isFull := !ebiten.IsFullscreen()
		ebiten.SetFullscreen(isFull)
		if isFull {
			g.showNotification("FULLSCREEN")
		} else {
			g.showNotification("WINDOWED")
		}
	}

	// Sound Mute Toggle (M Key)
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		muted := ToggleMute()
		if g.Settings != nil {
			g.Settings.SoundEffectsEnabled = !muted
			_ = g.Settings.Save()
		}
		if muted {
			g.showNotification("SOUND: MUTED")
		} else {
			g.showNotification(fmt.Sprintf("SOUND: %d%%", int(GetMasterVolume()*100)))
		}
	}

	// Volume Adjust (- / +)
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		vol := AdjustVolume(-0.1)
		g.showNotification(fmt.Sprintf("VOL: %d%%", int(vol*100)))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		vol := AdjustVolume(0.1)
		g.showNotification(fmt.Sprintf("VOL: %d%%", int(vol*100)))
	}

	if g.NotificationTimer > 0 {
		g.NotificationTimer--
	}

	// Options Toggle (O Key or Gamepad Select/Back)
	isOptionsPressed := inpututil.IsKeyJustPressed(ebiten.KeyO)
	if !isOptionsPressed {
		ids := ebiten.AppendGamepadIDs(nil)
		for _, id := range ids {
			if ebiten.IsStandardGamepadLayoutAvailable(id) && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterLeft) {
				isOptionsPressed = true
				break
			}
		}
	}

	if isOptionsPressed {
		if g.State == StateOptions {
			g.closeOptions()
			return nil
		} else if g.State == StateAttract {
			g.openOptions(StateAttract)
			return nil
		} else if g.State == StatePlaying || g.Paused {
			if !g.Paused {
				g.Paused = true
				StopAllContinuousSounds()
			}
			g.openOptions(StatePlaying)
			return nil
		}
	}

	// If in options state, handle options updates directly
	if g.State == StateOptions {
		g.updateOptions()
		return nil
	}

	// Quit / Exit Handling (Q Key)
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if g.State == StateAttract {
			return ebiten.Termination
		}
		// Return to attract mode from game/pause
		g.enterAttractMode()
		return nil
	}

	// Pause / Resume Toggle (P Key or Gamepad Start in game)
	isPausePressed := inpututil.IsKeyJustPressed(ebiten.KeyP)
	if !isPausePressed && g.State != StateAttract && g.State != StateHighScoreEntry {
		ids := ebiten.AppendGamepadIDs(nil)
		for _, id := range ids {
			if ebiten.IsStandardGamepadLayoutAvailable(id) && inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterRight) {
				isPausePressed = true
				break
			}
		}
	}

	if isPausePressed {
		if g.State != StateAttract && g.State != StateHighScoreEntry {
			g.Paused = !g.Paused
			if g.Paused {
				StopAllContinuousSounds()
			} else {
				g.restoreContinuousSounds()
			}
		}
	}

	// If paused, halt simulation updates
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

func (g *Game) restoreContinuousSounds() {
	if g.ActiveFlier != nil && g.ActiveFlier.Active {
		if g.ActiveFlier.Type == FlierBomber {
			SetBomberSound(true)
		} else {
			SetSatelliteSound(true)
		}
	}
	hasActiveSmartBomb := false
	for _, sb := range g.SmartBombs {
		if sb.Active {
			hasActiveSmartBomb = true
			break
		}
	}
	if hasActiveSmartBomb {
		SetSmartBombSound(true)
	}
}

// --- STATE UPDATES ---

func (g *Game) checkAnyStartInput() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.Key1) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return true
	}
	ids := ebiten.AppendGamepadIDs(nil)
	for _, id := range ids {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonCenterRight) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightLeft) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightTop) {
				return true
			}
		}
	}
	return false
}

func (g *Game) updateAttract() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		simPos := ScreenToSim(float64(mx), float64(my), g.Width, g.Height)
		if simPos.Y >= 184 && simPos.Y <= 198 && simPos.X >= 20 && simPos.X <= 100 {
			g.openOptions(StateAttract)
			return
		}
	}

	if g.checkAnyStartInput() {
		StartGame(g)
		return
	}

	if g.AttractDemoMode {
		g.updateAttractDemo()
	} else {
		g.AttractTimer--
		if g.AttractTimer <= 0 {
			g.startAttractDemo()
		}
	}
}

func (g *Game) startAttractDemo() {
	g.AttractDemoMode = true
	g.AttractTimer = 900 // 15 seconds demo
	g.AIDemoShootTick = 0

	g.Wave = 1
	g.Score = 0
	g.Palette = GetPaletteForWave(1)
	wdata := GetWaveData(1)
	g.ICBMsRemaining = wdata.TotalICBMs
	g.SmartBombsLeft = wdata.SmartBombs
	g.SpawnCooldown = 25
	g.FlierCooldown = 120
	g.FlierSpawnedThisWave = false
	g.WaveEndGraceTimer = 0

	g.ICBMs = nil
	g.SmartBombs = nil
	g.ActiveFlier = nil
	g.ABMs = nil
	g.Explosions = nil
	StopAllContinuousSounds()

	cityXs := []float64{40, 65, 90, 150, 175, 200}
	for i, x := range cityXs {
		g.Cities[i] = City{
			Position:  Point{X: x, Y: 216},
			Destroyed: false,
		}
	}
	g.resetSilos()
}

func (g *Game) updateAttractDemo() {
	g.AttractTimer--
	if g.AttractTimer <= 0 {
		g.stopAttractDemo()
		return
	}

	// 1. Spawning
	g.updateSpawning()
	// 2. Fliers
	g.updateFlier()
	// 3. ABMs
	g.updateABMs()
	// 4. ICBMs
	g.updateICBMs()
	// 5. SmartBombs
	g.updateSmartBombs()
	// 6. Explosions
	g.updateExplosions()
	// 7. Collisions
	g.checkCollisions()

	// 8. AI Player Logic
	g.updateAIDemoPlayer()

	// Check if all cities destroyed or wave ends
	survivingCities := 0
	for _, city := range g.Cities {
		if !city.Destroyed {
			survivingCities++
		}
	}
	if survivingCities == 0 {
		g.stopAttractDemo()
		return
	}
}

func (g *Game) stopAttractDemo() {
	StopAllContinuousSounds()
	g.AttractDemoMode = false
	g.AttractTimer = 600
}

func (g *Game) updateAIDemoPlayer() {
	var bestTarget *Point
	var lowestY float64 = -1.0

	for _, icbm := range g.ICBMs {
		if icbm.Active && icbm.Curr.Y > 25 && icbm.Curr.Y < 175 && icbm.Curr.Y > lowestY {
			intercept := Point{
				X: icbm.Curr.X + (icbm.Target.X-icbm.Curr.X)*0.25,
				Y: icbm.Curr.Y + 28,
			}
			bestTarget = &intercept
			lowestY = icbm.Curr.Y
		}
	}

	for _, sb := range g.SmartBombs {
		if sb.Active && sb.Curr.Y > 25 && sb.Curr.Y < 175 && sb.Curr.Y > lowestY {
			intercept := Point{
				X: sb.Curr.X,
				Y: sb.Curr.Y + 22,
			}
			bestTarget = &intercept
			lowestY = sb.Curr.Y
		}
	}

	if bestTarget == nil && g.ActiveFlier != nil && g.ActiveFlier.Active {
		fPos := Point{X: g.ActiveFlier.X + 8 + g.ActiveFlier.Speed*20, Y: g.ActiveFlier.Y}
		bestTarget = &fPos
	}

	if bestTarget != nil {
		g.Crosshair.Pos.X += (bestTarget.X - g.Crosshair.Pos.X) * 0.18
		g.Crosshair.Pos.Y += (bestTarget.Y - g.Crosshair.Pos.Y) * 0.18

		// Clamp
		g.Crosshair.Pos.X = math.Max(4, math.Min(SimWidth-4, g.Crosshair.Pos.X))
		g.Crosshair.Pos.Y = math.Max(8, math.Min(206, g.Crosshair.Pos.Y))

		g.AIDemoShootTick++
		if g.AIDemoShootTick >= 35 {
			g.AIDemoShootTick = 0
			targetX := g.Crosshair.Pos.X
			siloIdx := 1
			if targetX < 90 && !g.Batteries[0].Destroyed && g.Batteries[0].Ammo > 0 {
				siloIdx = 0
			} else if targetX > 166 && !g.Batteries[2].Destroyed && g.Batteries[2].Ammo > 0 {
				siloIdx = 2
			} else if g.Batteries[1].Destroyed || g.Batteries[1].Ammo == 0 {
				if !g.Batteries[0].Destroyed && g.Batteries[0].Ammo > 0 {
					siloIdx = 0
				} else if !g.Batteries[2].Destroyed && g.Batteries[2].Ammo > 0 {
					siloIdx = 2
				}
			}
			g.fireABM(siloIdx)
		}
	}
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

	// 1. Spawning Enemy ICBMs & Smart Bombs
	g.updateSpawning()

	// 2. Fliers (Bombers / Satellites)
	g.updateFlier()

	// 3. Update Player ABMs
	g.updateABMs()

	// 4. Update ICBMs (including MIRV splits)
	g.updateICBMs()

	// 5. Update Smart Bombs (with avoidance AI)
	g.updateSmartBombs()

	// 6. Update Explosions
	g.updateExplosions()

	// 7. Check Collisions & Chain Reactions
	g.checkCollisions()

	// 8. Check Bonus City thresholds (Every 10,000 points)
	if g.Score >= g.NextBonusThreshold {
		g.BonusCitiesStored++
		g.NextBonusThreshold += 10000
		PlayBonusCitySound()
	}

	// 9. Check Wave Completion or Total Annihilation
	g.checkWaveCompletion()
}

func (g *Game) handleCrosshairInput() {
	currX, currY := ebiten.CursorPosition()
	if g.MouseCaptured {
		dx := float64(currX - g.PrevMouseX)
		dy := float64(currY - g.PrevMouseY)
		sensitivity := 0.45
		g.Crosshair.Pos.X += dx * sensitivity
		g.Crosshair.Pos.Y += dy * sensitivity
		g.PrevMouseX, g.PrevMouseY = currX, currY
	} else {
		g.Crosshair.Pos = ScreenToSimWithLetterbox(float64(currX), float64(currY), g.Width, g.Height)
	}

	// Keyboard arrow adjustments
	speed := 2.5
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.Crosshair.Pos.X -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.Crosshair.Pos.X += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.Crosshair.Pos.Y -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.Crosshair.Pos.Y += speed
	}

	// Gamepad analog stick & D-pad
	ids := ebiten.AppendGamepadIDs(nil)
	for _, id := range ids {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			lx := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
			ly := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
			deadzone := 0.18
			if math.Abs(lx) > deadzone {
				g.Crosshair.Pos.X += lx * 3.5
			}
			if math.Abs(ly) > deadzone {
				g.Crosshair.Pos.Y += ly * 3.5
			}

			if ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftLeft) {
				g.Crosshair.Pos.X -= speed
			}
			if ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftRight) {
				g.Crosshair.Pos.X += speed
			}
			if ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftTop) {
				g.Crosshair.Pos.Y -= speed
			}
			if ebiten.IsStandardGamepadButtonPressed(id, ebiten.StandardGamepadButtonLeftBottom) {
				g.Crosshair.Pos.Y += speed
			}
		}
	}

	// Clamp within playfield (cannot aim below ground line Y=210)
	g.Crosshair.Pos.X = math.Max(4, math.Min(SimWidth-4, g.Crosshair.Pos.X))
	g.Crosshair.Pos.Y = math.Max(8, math.Min(206, g.Crosshair.Pos.Y))
}

func (g *Game) handleFiringInput() {
	var firedIdx = -1

	// Keyboard: A (Left), S / Space (Center), D (Right) or 1, 2, 3
	if inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.Key1) {
		firedIdx = 0
	} else if inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.Key2) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		firedIdx = 1
	} else if inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.Key3) {
		firedIdx = 2
	}

	// Mouse buttons: Left (Left), Middle (Center), Right (Right)
	if g.MouseCaptured {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			firedIdx = 0
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
			firedIdx = 1
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			firedIdx = 2
		}
	}

	// Gamepad buttons (Left Silo: LB/LT/A, Center Silo: X/Y, Right Silo: RB/RT/B)
	ids := ebiten.AppendGamepadIDs(nil)
	for _, id := range ids {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontTopLeft) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontBottomLeft) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
				firedIdx = 0
			} else if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightLeft) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightTop) {
				firedIdx = 1
			} else if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontTopRight) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonFrontBottomRight) ||
				inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightRight) {
				firedIdx = 2
			}
		}
	}

	if firedIdx != -1 {
		g.fireABM(firedIdx)
	}
}

func (g *Game) fireABM(siloIdx int) {
	// Authentic arcade limit: Max 8 simultaneous ABMs on screen
	if len(g.ABMs) >= 8 {
		return
	}

	bat := &g.Batteries[siloIdx]
	if bat.Destroyed || bat.Ammo <= 0 {
		PlayCantFireSound()
		return
	}

	// Deduct 1 ammo
	bat.Ammo--

	// Play silo low warning if drops to 3 or fewer
	if bat.Ammo <= 3 && !bat.LowWarningPlayed {
		PlaySiloLowSound()
		bat.LowWarningPlayed = true
	}

	// Authentic speed: Center silo (Delta) is FAST (~7.5 units/frame), Side silos ~3.2 units/frame
	launchSpeed := 3.2
	if siloIdx == 1 {
		launchSpeed = 7.5
	}

	abm := &ABM{
		Start:     bat.Position,
		Curr:      bat.Position,
		Target:    g.Crosshair.Pos,
		SiloIndex: siloIdx,
		Speed:     launchSpeed,
		Progress:  0.0,
		Active:    true,
	}
	g.ABMs = append(g.ABMs, abm)
	PlayLaunchSound()
}

func (g *Game) updateSpawning() {
	// Limit simultaneous active attacks to 8 (Smart bombs count as 2)
	activeAttacks := len(g.ICBMs) + len(g.SmartBombs)*2
	if activeAttacks >= 8 {
		return
	}

	if g.ICBMsRemaining <= 0 && g.SmartBombsLeft <= 0 {
		return
	}

	// Authentic arcade pacing: The game does not launch any new attacks while
	// the highest ICBM or smart bomb is above (202 - 2 * wave_number), with a floor of 180.
	// In screen coordinates (Y=0 at top): thresholdY = min(51.0, 29.0 + float64(g.Wave)*2.0).
	// Attacks are blocked until all active threats have descended past this altitude.
	thresholdY := math.Min(51.0, 29.0+float64(g.Wave)*2.0)
	for _, icbm := range g.ICBMs {
		if icbm.Active && icbm.Curr.Y < thresholdY {
			return
		}
	}
	for _, sb := range g.SmartBombs {
		if sb.Active && sb.Curr.Y < thresholdY {
			return
		}
	}

	g.SpawnCooldown--
	if g.SpawnCooldown > 0 {
		return
	}

	wdata := GetWaveData(g.Wave)

	// Determine whether to spawn a Smart Bomb or ICBM.
	// Arcade limit: maximum 3 active smart bombs on screen simultaneously.
	canSpawnSB := g.SmartBombsLeft > 0 && len(g.SmartBombs) < 3
	if canSpawnSB && (g.ICBMsRemaining <= 0 || rand.Float64() < 0.35) {
		g.SmartBombsLeft--
		spawnX := 20.0 + rand.Float64()*(SimWidth-40.0)
		target := g.pickRandomTarget()

		sb := &SmartBomb{
			Start:    Point{X: spawnX, Y: 0},
			Curr:     Point{X: spawnX, Y: 0},
			Target:   target,
			Speed:    wdata.Speed, // Authentic: same speed as ICBMs
			Progress: 0.0,
			Active:   true,
		}
		g.SmartBombs = append(g.SmartBombs, sb)
		SetSmartBombSound(true)
	} else if g.ICBMsRemaining > 0 {
		g.ICBMsRemaining--
		spawnX := rand.Float64() * SimWidth
		target := g.pickRandomTarget()

		// Wave >= 2 splitters (MIRVs): requires at least 2 remaining missiles for sisters
		isSplit := false
		if g.Wave >= 2 && g.ICBMsRemaining >= 2 && rand.Float64() < 0.25 {
			isSplit = true
		}

		// Arcade split altitude is between 128 and 159 (Y in 72..103 in screen coords)
		splitAlt := 72.0 + rand.Float64()*31.0

		icbm := &ICBM{
			Start:         Point{X: spawnX, Y: 0},
			Curr:          Point{X: spawnX, Y: 0},
			Target:        target,
			Speed:         wdata.Speed,
			Progress:      0.0,
			IsSplitter:    isSplit,
			SplitAltitude: splitAlt,
			Active:        true,
		}
		g.ICBMs = append(g.ICBMs, icbm)
	}

	// Set next spawn cooldown (paced naturally alongside altitude threshold)
	minCD := math.Max(20, 55-float64(g.Wave)*2.5)
	maxCD := math.Max(40, 95-float64(g.Wave)*4.0)
	g.SpawnCooldown = int(minCD) + rand.Intn(int(maxCD-minCD+1))
}

func (g *Game) pickRandomTarget() Point {
	targets := []Point{}
	for _, city := range g.Cities {
		if !city.Destroyed {
			targets = append(targets, Point{X: city.Position.X + 8, Y: 216})
		}
	}
	for _, bat := range g.Batteries {
		if !bat.Destroyed {
			targets = append(targets, bat.Position)
		}
	}
	if len(targets) > 0 {
		return targets[rand.Intn(len(targets))]
	}
	return Point{X: 20 + rand.Float64()*(SimWidth-40), Y: 216}
}

func (g *Game) updateFlier() {
	if g.ActiveFlier == nil {
		if !g.FlierSpawnedThisWave && g.ICBMsRemaining > 2 {
			g.FlierCooldown--
			if g.FlierCooldown <= 0 {
				g.FlierSpawnedThisWave = true
				ft := FlierBomber
				alt := 65.0 + rand.Float64()*25.0
				// Authentic arcade flier speeds (constant across all waves):
				// Bombers move at 1 pixel every 3 frames (~0.333 px/frame).
				// Satellites move at 1 pixel every 2 frames (0.500 px/frame).
				speed := 1.0 / 3.0
				if rand.Float64() < 0.45 {
					ft = FlierSatellite
					alt = 28.0 + rand.Float64()*18.0
					speed = 0.50
				}

				movingRight := rand.Float64() < 0.5
				startX := -16.0
				if !movingRight {
					startX = SimWidth + 16.0
					speed = -speed
				}

				g.ActiveFlier = &Flier{
					Type:           ft,
					X:              startX,
					Y:              alt,
					Speed:          speed,
					DropCooldown:   40 + rand.Intn(50),
					BombsRemaining: 2,
					Active:         true,
				}

				if ft == FlierBomber {
					SetBomberSound(true)
				} else {
					SetSatelliteSound(true)
				}
			}
		}
		return
	}

	flier := g.ActiveFlier
	flier.AnimTick++
	flier.X += flier.Speed

	// Check if flier left screen
	if (flier.Speed > 0 && flier.X > SimWidth+20) || (flier.Speed < 0 && flier.X < -20) {
		flier.Active = false
		g.ActiveFlier = nil
		SetBomberSound(false)
		SetSatelliteSound(false)
		return
	}

	// Dropping bombs from flier
	if flier.BombsRemaining > 0 && flier.X >= 15 && flier.X <= SimWidth-15 {
		flier.DropCooldown--
		if flier.DropCooldown <= 0 {
			flier.BombsRemaining--
			flier.DropCooldown = 60 + rand.Intn(40)

			// Decrement wave missile quota if remaining
			if g.ICBMsRemaining > 0 {
				g.ICBMsRemaining--
			}

			target := g.pickRandomTarget()
			wdata := GetWaveData(g.Wave)
			icbm := &ICBM{
				Start:    Point{X: flier.X, Y: flier.Y},
				Curr:     Point{X: flier.X, Y: flier.Y},
				Target:   target,
				Speed:    wdata.Speed, // Standard wave speed, no artificial boost
				Progress: 0.0,
				Active:   true,
			}
			g.ICBMs = append(g.ICBMs, icbm)
		}
	}
}

func (g *Game) updateABMs() {
	active := []*ABM{}
	for _, abm := range g.ABMs {
		if !abm.Active {
			continue
		}
		dist := abm.Start.Distance(abm.Target)
		if dist <= 0 {
			abm.Active = false
			continue
		}

		abm.Progress += abm.Speed / dist
		if abm.Progress >= 1.0 {
			abm.Active = false
			// Detonate at target
			g.Explosions = append(g.Explosions, &Explosion{
				Center:    abm.Target,
				Radius:    0.0,
				MaxRadius: 22.0,
				HoldTimer: 18,
				State:     StateExpanding,
			})
			PlayExplosionSound()
		} else {
			abm.Curr = abm.Start.Lerp(abm.Target, abm.Progress)
			active = append(active, abm)
		}
	}
	g.ABMs = active
}

func (g *Game) updateICBMs() {
	active := []*ICBM{}
	for _, icbm := range g.ICBMs {
		if !icbm.Active {
			continue
		}

		// MIRV Split Logic: in the authentic arcade game, sister warheads count toward the wave total.
		// Missiles only split if quota remains.
		if icbm.IsSplitter && !icbm.Splitted && icbm.Curr.Y >= icbm.SplitAltitude {
			icbm.Splitted = true
			sistersToSpawn := 0
			if g.ICBMsRemaining >= 2 {
				sistersToSpawn = 2
			} else if g.ICBMsRemaining == 1 {
				sistersToSpawn = 1
			}
			g.ICBMsRemaining -= sistersToSpawn

			for s := 0; s < sistersToSpawn; s++ {
				sisterTarget := g.pickRandomTarget()
				sister := &ICBM{
					Start:      icbm.Curr,
					Curr:       icbm.Curr,
					Target:     sisterTarget,
					Speed:      icbm.Speed,
					Progress:   0.0,
					IsSplitter: false,
					Active:     true,
				}
				active = append(active, sister)
			}
		}

		dist := icbm.Start.Distance(icbm.Target)
		if dist <= 0 {
			icbm.Active = false
			continue
		}

		icbm.Progress += icbm.Speed / dist
		if icbm.Progress >= 1.0 || icbm.Curr.Y >= 216 {
			icbm.Active = false
			g.detonateThreatOnGround(icbm.Target)
		} else {
			icbm.Curr = icbm.Start.Lerp(icbm.Target, icbm.Progress)
			active = append(active, icbm)
		}
	}
	g.ICBMs = active
}

func (g *Game) updateSmartBombs() {
	active := []*SmartBomb{}
	for _, sb := range g.SmartBombs {
		if !sb.Active {
			continue
		}
		sb.AnimTick++

		// Smart Bomb Evasion AI: Scan for nearby expanding explosions
		evading := false
		for _, exp := range g.Explosions {
			if exp.State == StateDead {
				continue
			}
			dist := sb.Curr.Distance(exp.Center)
			dangerZone := exp.Radius + 18.0
			if dist < dangerZone {
				evading = true
				// Calculate vector away from explosion center
				dx := sb.Curr.X - exp.Center.X
				dy := sb.Curr.Y - exp.Center.Y
				if math.Abs(dx) < 0.1 {
					if rand.Float64() < 0.5 {
						dx = 1.0
					} else {
						dx = -1.0
					}
				}
				length := math.Hypot(dx, dy)
				if length > 0 {
					steerX := (dx / length) * sb.Speed * 1.25
					sb.Curr.X += steerX
					// Always ensure downward progress so it never gets stuck or floats upward
					sb.Curr.Y += sb.Speed * 0.40
				}
				break
			}
		}

		if !evading {
			// Normal ballistic tracking toward target
			dist := sb.Curr.Distance(sb.Target)
			if dist <= sb.Speed {
				sb.Active = false
				g.detonateThreatOnGround(sb.Target)
				continue
			} else {
				dirX := (sb.Target.X - sb.Curr.X) / dist
				dirY := (sb.Target.Y - sb.Curr.Y) / dist
				sb.Curr.X += dirX * sb.Speed
				sb.Curr.Y += dirY * sb.Speed
			}
		}

		// Clamp within playfield width
		sb.Curr.X = math.Max(4, math.Min(SimWidth-4, sb.Curr.X))

		// Detonate if hit ground
		if sb.Curr.Y >= 216 {
			sb.Active = false
			g.detonateThreatOnGround(sb.Curr)
		} else {
			active = append(active, sb)
		}
	}
	g.SmartBombs = active

	if len(g.SmartBombs) == 0 {
		SetSmartBombSound(false)
	}
}

func (g *Game) detonateThreatOnGround(pos Point) {
	// Check City Hits
	for i := range g.Cities {
		if !g.Cities[i].Destroyed {
			cx := g.Cities[i].Position.X + 8
			if math.Abs(pos.X-cx) < 12.0 && math.Abs(pos.Y-216) < 8.0 {
				g.Cities[i].Destroyed = true
				g.pipeline.TriggerFlash(8)
				break
			}
		}
	}

	// Check Base Hits
	for i := range g.Batteries {
		if !g.Batteries[i].Destroyed {
			bx := g.Batteries[i].Position.X
			if math.Abs(pos.X-bx) < 14.0 && math.Abs(pos.Y-214) < 8.0 {
				g.Batteries[i].Destroyed = true
				g.Batteries[i].Ammo = 0
				g.pipeline.TriggerFlash(6)
				break
			}
		}
	}

	// Ground detonation explosion
	g.Explosions = append(g.Explosions, &Explosion{
		Center:    pos,
		Radius:    0.0,
		MaxRadius: 18.0,
		HoldTimer: 14,
		State:     StateExpanding,
	})
	PlayExplosionSound()
}

func (g *Game) updateExplosions() {
	active := []*Explosion{}
	for _, exp := range g.Explosions {
		if exp.State == StateDead {
			continue
		}
		exp.Tick++

		switch exp.State {
		case StateExpanding:
			exp.Radius += 0.65
			if exp.Radius >= exp.MaxRadius {
				exp.State = StateHolding
			}
		case StateHolding:
			exp.HoldTimer--
			if exp.HoldTimer <= 0 {
				exp.State = StateContracting
			}
		case StateContracting:
			exp.Radius -= 0.55
			if exp.Radius <= 0 {
				exp.State = StateDead
			}
		}

		if exp.State != StateDead {
			active = append(active, exp)
		}
	}
	g.Explosions = active
}

func (g *Game) checkCollisions() {
	// Check ICBMs against active explosions
	for _, icbm := range g.ICBMs {
		if !icbm.Active {
			continue
		}
		for _, exp := range g.Explosions {
			if exp.State == StateDead {
				continue
			}
			if icbm.Curr.Distance(exp.Center) <= exp.Radius {
				icbm.Active = false
				g.Score += 25 * g.Palette.Multiplier
				// Chain reaction
				g.Explosions = append(g.Explosions, &Explosion{
					Center:    icbm.Curr,
					Radius:    0.0,
					MaxRadius: 18.0,
					HoldTimer: 14,
					State:     StateExpanding,
				})
				PlayExplosionSound()
				break
			}
		}
	}

	// Check Smart Bombs against active explosions
	for _, sb := range g.SmartBombs {
		if !sb.Active {
			continue
		}
		for _, exp := range g.Explosions {
			if exp.State == StateDead {
				continue
			}
			if sb.Curr.Distance(exp.Center) <= exp.Radius {
				sb.Active = false
				g.Score += 125 * g.Palette.Multiplier
				g.Explosions = append(g.Explosions, &Explosion{
					Center:    sb.Curr,
					Radius:    0.0,
					MaxRadius: 18.0,
					HoldTimer: 14,
					State:     StateExpanding,
				})
				PlayExplosionSound()
				break
			}
		}
	}

	// Check Flier against active explosions
	if g.ActiveFlier != nil && g.ActiveFlier.Active {
		flier := g.ActiveFlier
		fPos := Point{X: flier.X + 8, Y: flier.Y + 4}
		for _, exp := range g.Explosions {
			if exp.State == StateDead {
				continue
			}
			if fPos.Distance(exp.Center) <= exp.Radius+6.0 {
				flier.Active = false
				g.ActiveFlier = nil
				SetBomberSound(false)
				SetSatelliteSound(false)
				g.Score += 100 * g.Palette.Multiplier
				g.Explosions = append(g.Explosions, &Explosion{
					Center:    fPos,
					Radius:    0.0,
					MaxRadius: 22.0,
					HoldTimer: 16,
					State:     StateExpanding,
				})
				PlayExplosionSound()
				break
			}
		}
	}
}

func (g *Game) checkWaveCompletion() {
	// Check if all 6 cities are destroyed
	survivingCities := 0
	for _, city := range g.Cities {
		if !city.Destroyed {
			survivingCities++
		}
	}

	if survivingCities == 0 {
		// Annihilation!
		StopAllContinuousSounds()
		g.State = StateTheEnd
		g.TheEndTimer = 190
		g.pipeline.TriggerFlash(25)
		PlayGameOverSound()
		return
	}

	// Wave completes when no threats remain
	noThreatsRemaining := g.ICBMsRemaining <= 0 && g.SmartBombsLeft <= 0 && len(g.ICBMs) == 0 &&
		len(g.SmartBombs) == 0 && g.ActiveFlier == nil

	if noThreatsRemaining {
		g.WaveEndGraceTimer++
		// Transition once in-flight munitions finish, or force transition after 45 frames
		if (len(g.ABMs) == 0 && len(g.Explosions) == 0) || g.WaveEndGraceTimer >= 45 {
			StopAllContinuousSounds()
			g.State = StateTally
			g.TallyPhase = 0
			g.TallyTimer = 30
			g.TallyBatteryIdx = 0
			g.TallyCityIdx = 0
			g.TallyScoreEarned = 0
			g.WaveEndGraceTimer = 0
		}
	} else {
		g.WaveEndGraceTimer = 0
	}
}

func (g *Game) updateTally() {
	g.TallyTimer--
	if g.TallyTimer > 0 {
		return
	}

	// 8 frames interval between tally blips
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
				if decrement > bat.Ammo {
					decrement = bat.Ammo
				}
				bat.Ammo -= decrement
				pts := decrement * 5 * g.Palette.Multiplier
				g.Score += pts
				g.TallyScoreEarned += pts
				PlayTallySound()
				if bat.MaxAmmo > 10 {
					g.TallyTimer = 3
				} else {
					g.TallyTimer = 8
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
		g.TallyTimer = 55 // wait before bonus rebuild

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

	// Check if player has stored bonus cities and destroyed cities to rebuild
	if g.BonusCitiesStored > 0 {
		for i := range g.Cities {
			if g.Cities[i].Destroyed {
				g.Cities[i].Destroyed = false
				g.BonusCitiesStored--
				PlayBonusCitySound()
				g.RebuildTimer = 45 // delay to show rebuilt city
				return
			}
		}
	}

	// Advance to next wave
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
	// Scroll letters with Arrow Keys or Mouse Wheel / Y movement
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.CharSelectIdx = (g.CharSelectIdx - 1 + len(g.Letters)) % len(g.Letters)
		g.Initials[g.InitialSlot] = g.Letters[g.CharSelectIdx]
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.CharSelectIdx = (g.CharSelectIdx + 1) % len(g.Letters)
		g.Initials[g.InitialSlot] = g.Letters[g.CharSelectIdx]
	}

	// Direct keyboard character typing
	for _, r := range g.Letters {
		keyName := string(r)
		if len(keyName) == 1 {
			// Check standard letter keys
			if r >= 'A' && r <= 'Z' {
				k := ebiten.Key(int(ebiten.KeyA) + int(r-'A'))
				if inpututil.IsKeyJustPressed(k) {
					g.Initials[g.InitialSlot] = r
					g.advanceInitialSlot()
					return
				}
			}
		}
	}

	// Confirm letter on Enter, Space, or Left Click
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
		// Save High Score!
		initialStr := string(g.Initials[:])
		g.highScores.AddScore(initialStr, g.Score)
		g.State = StateAttract
	} else {
		g.CharSelectIdx = 0
		g.Initials[g.InitialSlot] = g.Letters[0]
	}
}

func (g *Game) openOptions(fromState GameState) {
	g.PreviousState = fromState
	g.OptionSelectedIdx = 0
	g.State = StateOptions
}

func (g *Game) closeOptions() {
	if g.Settings != nil {
		_ = g.Settings.Save()
		// If in active playing game, update battery max ammo and clamp current ammo
		if g.PreviousState == StatePlaying {
			for i := range g.Batteries {
				g.Batteries[i].MaxAmmo = g.Settings.MissilesPerSilo
				if g.Batteries[i].Ammo > g.Settings.MissilesPerSilo {
					g.Batteries[i].Ammo = g.Settings.MissilesPerSilo
				}
			}
		}
	}
	if g.PreviousState == StatePlaying {
		g.Paused = true
		g.State = StatePlaying
	} else if g.PreviousState != 0 {
		g.State = g.PreviousState
	} else {
		g.enterAttractMode()
	}
}

func (g *Game) updateOptions() {
	// Navigation: Up / Down
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.OptionSelectedIdx = (g.OptionSelectedIdx + 3) % 4
		PlayTallySound()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.OptionSelectedIdx = (g.OptionSelectedIdx + 1) % 4
		PlayTallySound()
	}

	// Gamepad navigation
	ids := ebiten.AppendGamepadIDs(nil)
	for _, id := range ids {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftTop) {
				g.OptionSelectedIdx = (g.OptionSelectedIdx + 3) % 4
				PlayTallySound()
			}
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftBottom) {
				g.OptionSelectedIdx = (g.OptionSelectedIdx + 1) % 4
				PlayTallySound()
			}
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftLeft) {
				g.adjustOption(-1)
			}
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonLeftRight) {
				g.adjustOption(1)
			}
			if inpututil.IsStandardGamepadButtonJustPressed(id, ebiten.StandardGamepadButtonRightBottom) {
				g.activateOption()
			}
		}
	}

	// Left / Right value changes
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.adjustOption(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.adjustOption(1)
	}

	// Enter / Space activates/cycles
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.activateOption()
	}

	// Escape / Q exits
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.closeOptions()
		return
	}

	// Mouse click handling
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		simPos := ScreenToSim(float64(mx), float64(my), g.Width, g.Height)
		startY := 64
		rowHeight := 24
		for i := 0; i < 4; i++ {
			y := startY + i*rowHeight
			if simPos.Y >= float64(y-4) && simPos.Y <= float64(y+16) && simPos.X >= 16 && simPos.X <= 240 {
				g.OptionSelectedIdx = i
				if i == 0 {
					g.toggleSoundOption()
				} else if i == 1 {
					if simPos.X < 192 {
						g.adjustMissilesOption(-10)
					} else {
						g.adjustMissilesOption(10)
					}
				} else if i == 2 {
					g.toggleCRTOption()
				} else if i == 3 {
					g.closeOptions()
				}
				return
			}
		}
	}
}

func (g *Game) adjustOption(dir int) {
	switch g.OptionSelectedIdx {
	case 0:
		g.toggleSoundOption()
	case 1:
		g.adjustMissilesOption(dir * 10)
	case 2:
		g.toggleCRTOption()
	case 3:
		// Back row - no directional adjust
	}
}

func (g *Game) activateOption() {
	switch g.OptionSelectedIdx {
	case 0:
		g.toggleSoundOption()
	case 1:
		newVal := g.Settings.MissilesPerSilo + 10
		if newVal > 100 {
			newVal = 10
		}
		g.Settings.MissilesPerSilo = newVal
		_ = g.Settings.Save()
		PlayTallySound()
	case 2:
		g.toggleCRTOption()
	case 3:
		g.closeOptions()
	}
}

func (g *Game) toggleSoundOption() {
	g.Settings.SoundEffectsEnabled = !g.Settings.SoundEffectsEnabled
	SetSoundEffectsEnabled(g.Settings.SoundEffectsEnabled)
	_ = g.Settings.Save()
	if g.Settings.SoundEffectsEnabled {
		PlayTallySound()
	}
}

func (g *Game) adjustMissilesOption(delta int) {
	newVal := g.Settings.MissilesPerSilo + delta
	if newVal < 10 {
		newVal = 10
	} else if newVal > 100 {
		newVal = 100
	}
	if newVal != g.Settings.MissilesPerSilo {
		g.Settings.MissilesPerSilo = newVal
		_ = g.Settings.Save()
		PlayTallySound()
	}
}

func (g *Game) toggleCRTOption() {
	g.pipeline.UseCRT = !g.pipeline.UseCRT
	g.Settings.UseCRT = g.pipeline.UseCRT
	_ = g.Settings.Save()
	PlayTallySound()
}

// Draw renders the arcade game visuals into the 256x231 framebuffer and outputs via the pipeline.
func (g *Game) Draw(screen *ebiten.Image) {
	fb := g.pipeline.frameBuffer
	fb.Clear()

	// Fill background with wave sky color
	skyCol := g.Palette.SkyColor
	if g.State == StateAttract {
		skyCol = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	fb.Fill(skyCol)

	// Render terrain and entities
	g.renderPlayfield(fb)

	// Output frame to window with aspect ratio scaling and CRT filter
	g.pipeline.DrawFrame(screen, g.Width, g.Height)
}

func (g *Game) renderPlayfield(target *ebiten.Image) {
	white := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	red := color.RGBA{R: 240, G: 32, B: 32, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}
	cyan := color.RGBA{R: 40, G: 220, B: 240, A: 255}

	// 1. Draw HUD
	g.drawHUD(target)

	if g.State == StateAttract {
		g.drawAttractScreen(target)
		return
	}
	if g.State == StateOptions {
		g.drawOptionsScreen(target)
		return
	}

	// 2. Draw Terrain Ground Profile
	g.drawTerrain(target)

	// 3. Draw Cities
	for _, city := range g.Cities {
		DrawCity(target, int(city.Position.X), int(city.Position.Y), city.Destroyed, g.Palette.CityColor)
	}

	// 4. Draw Missile Batteries / Silos
	for _, bat := range g.Batteries {
		g.drawSilo(target, bat)
	}

	// 5. Draw Bonus City Reserve Icons
	for i := 0; i < g.BonusCitiesStored && i < 12; i++ {
		DrawBonusCityIcon(target, 4+i*10, int(SimHeight)-7, g.Palette.CityColor)
	}

	// 6. Draw ICBM Trails
	for _, icbm := range g.ICBMs {
		if !icbm.Active {
			continue
		}
		vector.StrokeLine(target, float32(icbm.Start.X), float32(icbm.Start.Y), float32(icbm.Curr.X), float32(icbm.Curr.Y), 1.0, g.Palette.ICBMColor, false)
		target.Set(int(icbm.Curr.X), int(icbm.Curr.Y), white)
	}

	// 7. Draw Smart Bombs
	for _, sb := range g.SmartBombs {
		if !sb.Active {
			continue
		}
		vector.StrokeLine(target, float32(sb.Start.X), float32(sb.Start.Y), float32(sb.Curr.X), float32(sb.Curr.Y), 1.0, red, false)
		DrawSmartBomb(target, int(sb.Curr.X)-4, int(sb.Curr.Y)-4, sb.AnimTick/6, yellow)
	}

	// 8. Draw Fliers (Bombers & Satellites)
	if g.ActiveFlier != nil && g.ActiveFlier.Active {
		flier := g.ActiveFlier
		if flier.Type == FlierBomber {
			DrawBomber(target, int(flier.X), int(flier.Y), flier.Speed > 0, g.Palette.ICBMColor)
		} else {
			DrawSatellite(target, int(flier.X), int(flier.Y), g.Palette.ICBMColor)
		}
	}

	// 9. Draw ABM Trails & Targets
	for _, abm := range g.ABMs {
		if !abm.Active {
			continue
		}
		vector.StrokeLine(target, float32(abm.Start.X), float32(abm.Start.Y), float32(abm.Curr.X), float32(abm.Curr.Y), 1.0, g.Palette.ABMColor, false)
		// Draw 'X' target marker
		tx := int(abm.Target.X)
		ty := int(abm.Target.Y)
		DrawArcadeText(target, "X", tx-4, ty-4, white)
	}

	// 10. Draw Color-Cycling Explosions
	for _, exp := range g.Explosions {
		if exp.State == StateDead {
			continue
		}
		col := GetExplosionColor(exp.Tick)
		// Draw filled explosion circle
		vector.DrawFilledCircle(target, float32(exp.Center.X), float32(exp.Center.Y), float32(exp.Radius), col, false)
		// Inner white core
		if exp.Radius > 5.0 {
			vector.DrawFilledCircle(target, float32(exp.Center.X), float32(exp.Center.Y), float32(exp.Radius*0.4), white, false)
		}
	}

	// 11. Draw Crosshair
	if g.State == StatePlaying || g.State == StateWaveStart {
		cx := int(g.Crosshair.Pos.X)
		cy := int(g.Crosshair.Pos.Y)
		DrawArcadeText(target, "X", cx-4, cy-4, white)
	}

	// 12. State-Specific Overlays
	switch g.State {
	case StateWaveStart:
		DrawArcadeText(target, fmt.Sprintf("START WAVE %d", g.Wave), 80, 80, g.Palette.TextColor)
		DrawArcadeText(target, fmt.Sprintf("%dX BONUS POINTS", g.Palette.Multiplier), 68, 100, g.Palette.MultiplierColor)

	case StateTally:
		DrawArcadeText(target, "BONUS POINTS", 80, 60, yellow)
		DrawArcadeText(target, fmt.Sprintf("+%d", g.TallyScoreEarned), 108, 80, white)

	case StateTheEnd:
		g.drawTheEndScreen(target)

	case StateHighScoreEntry:
		g.drawHighScoreEntry(target)
	}

	// 13. Draw Pause Overlay
	if g.Paused {
		// Semi-opaque dark backdrop box
		vector.DrawFilledRect(target, 36, 72, 184, 68, color.RGBA{R: 0, G: 0, B: 0, A: 230}, false)
		vector.StrokeRect(target, 36, 72, 184, 68, 1.0, yellow, false)

		if (g.Tick/20)%2 == 0 {
			DrawArcadeText(target, "GAME PAUSED", 84, 80, yellow)
		} else {
			DrawArcadeText(target, "GAME PAUSED", 84, 80, white)
		}
		DrawArcadeText(target, "PRESS P TO RESUME", 60, 96, white)
		DrawArcadeText(target, "PRESS O FOR OPTIONS", 52, 110, cyan)
		DrawArcadeText(target, "PRESS Q TO QUIT", 68, 124, g.Palette.TextColor)
	}
}

func (g *Game) drawHUD(target *ebiten.Image) {
	white := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}

	scoreStr := fmt.Sprintf("%d", g.Score)
	topScoreStr := fmt.Sprintf("%d", g.highScores.GetTopScore())

	DrawArcadeText(target, scoreStr, 24, 8, white)
	DrawArcadeText(target, topScoreStr, 116, 8, yellow)

	if g.State == StatePlaying {
		multStr := fmt.Sprintf("%dX", g.Palette.Multiplier)
		DrawArcadeText(target, multStr, 220, 8, g.Palette.MultiplierColor)
	}

	// On-screen notification (e.g. CRT ON/OFF, SOUND 80%, FULLSCREEN)
	if g.NotificationTimer > 0 {
		DrawArcadeText(target, g.NotificationText, 128-len(g.NotificationText)*4, 22, yellow)
	}
}

func (g *Game) drawTerrain(target *ebiten.Image) {
	groundCol := g.Palette.GroundColor

	// Draw base ground fill
	vector.DrawFilledRect(target, 0, 224, SimWidth, 7, groundCol, false)

	// Draw 3 Silo stepped terrain mounds
	siloXs := []float64{18, 128, 238}
	for _, sx := range siloXs {
		vector.DrawFilledRect(target, float32(sx-13), 219, 26, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-10), 215, 20, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-7), 211, 14, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-4), 207, 8, 5, groundCol, false)
	}
}

func (g *Game) drawSilo(target *ebiten.Image, bat Battery) {
	red := color.RGBA{R: 240, G: 32, B: 32, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}

	bx := int(bat.Position.X)
	by := int(bat.Position.Y)

	if bat.Destroyed {
		// Draw crater
		DrawArcadeText(target, "...", bx-12, by-4, red)
		return
	}

	// Draw missile ammo as ticks in the silo
	DrawSiloTicks(target, bx, 222, bat.Ammo, bat.MaxAmmo, g.Palette.SiloColor, g.Palette.GroundColor)

	// Status text indicators above the silo (Y = 196) - only when OUT or LOW
	if bat.Ammo == 0 {
		// Flash OUT
		if (g.Tick/15)%2 == 0 {
			DrawArcadeText(target, "OUT", bx-12, 196, red)
		}
	} else if bat.Ammo <= 3 {
		// Flash LOW
		if (g.Tick/15)%2 == 0 {
			DrawArcadeText(target, "LOW", bx-12, 196, yellow)
		}
	}
}

func (g *Game) drawAttractScreen(target *ebiten.Image) {
	white := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}
	cyan := color.RGBA{R: 40, G: 220, B: 240, A: 255}
	red := color.RGBA{R: 240, G: 40, B: 40, A: 255}

	if g.AttractDemoMode {
		// 1. Draw Terrain Ground Profile
		g.drawTerrain(target)

		// 2. Draw Cities
		for _, city := range g.Cities {
			DrawCity(target, int(city.Position.X), int(city.Position.Y), city.Destroyed, g.Palette.CityColor)
		}

		// 3. Draw Silos
		for _, bat := range g.Batteries {
			g.drawSilo(target, bat)
		}

		// 4. Draw ICBM Trails
		for _, icbm := range g.ICBMs {
			if !icbm.Active {
				continue
			}
			vector.StrokeLine(target, float32(icbm.Start.X), float32(icbm.Start.Y), float32(icbm.Curr.X), float32(icbm.Curr.Y), 1.0, g.Palette.ICBMColor, false)
			target.Set(int(icbm.Curr.X), int(icbm.Curr.Y), white)
		}

		// 5. Draw Smart Bombs
		for _, sb := range g.SmartBombs {
			if !sb.Active {
				continue
			}
			vector.StrokeLine(target, float32(sb.Start.X), float32(sb.Start.Y), float32(sb.Curr.X), float32(sb.Curr.Y), 1.0, red, false)
			DrawSmartBomb(target, int(sb.Curr.X)-4, int(sb.Curr.Y)-4, sb.AnimTick/6, yellow)
		}

		// 6. Draw Fliers
		if g.ActiveFlier != nil && g.ActiveFlier.Active {
			flier := g.ActiveFlier
			if flier.Type == FlierBomber {
				DrawBomber(target, int(flier.X), int(flier.Y), flier.Speed > 0, g.Palette.ICBMColor)
			} else {
				DrawSatellite(target, int(flier.X), int(flier.Y), g.Palette.ICBMColor)
			}
		}

		// 7. Draw ABMs
		for _, abm := range g.ABMs {
			if !abm.Active {
				continue
			}
			vector.StrokeLine(target, float32(abm.Start.X), float32(abm.Start.Y), float32(abm.Curr.X), float32(abm.Curr.Y), 1.0, g.Palette.ABMColor, false)
			DrawArcadeText(target, "X", int(abm.Target.X)-4, int(abm.Target.Y)-4, white)
		}

		// 8. Draw Explosions
		for _, exp := range g.Explosions {
			if exp.State == StateDead {
				continue
			}
			col := GetExplosionColor(exp.Tick)
			vector.DrawFilledCircle(target, float32(exp.Center.X), float32(exp.Center.Y), float32(exp.Radius), col, false)
			if exp.Radius > 5.0 {
				vector.DrawFilledCircle(target, float32(exp.Center.X), float32(exp.Center.Y), float32(exp.Radius*0.4), white, false)
			}
		}

		// 9. Draw Crosshair
		cx := int(g.Crosshair.Pos.X)
		cy := int(g.Crosshair.Pos.Y)
		DrawArcadeText(target, "X", cx-4, cy-4, white)

		// 10. Authentic Arcade Demo Overlays
		DrawArcadeText(target, "GAME OVER", 92, 40, red)
		DrawArcadeText(target, "DEFEND CITIES", 76, 55, yellow)
		DrawArcadeText(target, "DEMO MODE", 92, 70, cyan)
		if (g.Tick/30)%2 == 0 {
			DrawArcadeText(target, "PRESS SPACE OR CLICK", 48, 90, white)
		}
		return
	}

	DrawArcadeText(target, "MISSILE COMMAND", 68, 30, yellow)
	DrawArcadeText(target, "DEFEND CITIES", 76, 45, red)

	// High Scores List
	DrawArcadeText(target, "GREAT SCORES", 80, 68, cyan)
	for i, entry := range g.highScores.Entries {
		if i >= 8 {
			break
		}
		line := fmt.Sprintf("%2d  %5d  %3s", i+1, entry.Score, entry.Initials)
		DrawArcadeText(target, line, 64, 84+i*11, white)
	}

	// Copyright & Start prompt
	DrawArcadeText(target, "© 1980 ATARI INC", 64, 176, yellow)
	DrawArcadeText(target, "O:OPTIONS  P:PAUSE  Q:QUIT", 24, 190, cyan)
	if (g.Tick/30)%2 == 0 {
		DrawArcadeText(target, "PRESS SPACE OR CLICK", 48, 208, white)
	}
}

func (g *Game) drawOptionsScreen(target *ebiten.Image) {
	white := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}
	cyan := color.RGBA{R: 40, G: 220, B: 240, A: 255}
	red := color.RGBA{R: 240, G: 40, B: 40, A: 255}
	green := color.RGBA{R: 40, G: 240, B: 60, A: 255}

	// Title
	DrawArcadeText(target, "GAME OPTIONS", 80, 28, yellow)

	backLabel := "BACK TO TITLE"
	if g.PreviousState == StatePlaying {
		backLabel = "BACK TO GAME"
	}

	soundVal := "[ OFF ]"
	soundCol := red
	if g.Settings != nil && g.Settings.SoundEffectsEnabled {
		soundVal = "[ ON  ]"
		soundCol = green
	}

	missilesVal := 10
	if g.Settings != nil {
		missilesVal = g.Settings.MissilesPerSilo
	}

	crtVal := "[ OFF ]"
	crtCol := red
	if g.pipeline.UseCRT {
		crtVal = "[ ON  ]"
		crtCol = green
	}

	type optRow struct {
		label  string
		val    string
		valCol color.RGBA
	}

	rows := []optRow{
		{label: "SOUND EFFECTS", val: soundVal, valCol: soundCol},
		{label: "MISSILES/SILO", val: fmt.Sprintf("< %3d >", missilesVal), valCol: yellow},
		{label: "CRT SCANLINES", val: crtVal, valCol: crtCol},
		{label: backLabel, val: "", valCol: white},
	}

	startY := 64
	rowHeight := 24

	for i, r := range rows {
		y := startY + i*rowHeight
		isSel := (i == g.OptionSelectedIdx)

		lblCol := cyan
		if isSel {
			lblCol = yellow
			if (g.Tick/15)%2 == 0 {
				DrawArcadeText(target, ">", 16, y, yellow)
			} else {
				DrawArcadeText(target, ">", 16, y, white)
			}
		}

		DrawArcadeText(target, r.label, 28, y, lblCol)
		if r.val != "" {
			DrawArcadeText(target, r.val, 164, y, r.valCol)
		}
	}

	// Bottom instructions
	DrawArcadeText(target, "UP/DOWN : SELECT", 64, 178, cyan)
	DrawArcadeText(target, "LEFT/RIGHT/CLICK : CHANGE", 28, 192, white)
	DrawArcadeText(target, "ESC/ENTER : BACK", 64, 206, yellow)
}

func (g *Game) drawTheEndScreen(target *ebiten.Image) {
	red := color.RGBA{R: 240, G: 32, B: 32, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}

	// Expanding nuclear blast
	blastRadius := float32(math.Min(SimWidth, float64(200-g.TheEndTimer)*2.0))
	vector.DrawFilledCircle(target, SimWidth/2, SimHeight/2+20, blastRadius, red, false)

	// Giant "THE END" typography
	DrawArcadeText(target, "THE END", 100, 100, yellow)
	DrawArcadeText(target, "GAME OVER", 92, 120, red)
}

func (g *Game) drawHighScoreEntry(target *ebiten.Image) {
	white := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}
	cyan := color.RGBA{R: 40, G: 220, B: 240, A: 255}

	DrawArcadeText(target, "GREAT SCORE!", 80, 40, yellow)
	DrawArcadeText(target, fmt.Sprintf("SCORE: %d", g.Score), 76, 60, white)

	DrawArcadeText(target, "ENTER YOUR INITIALS", 52, 90, cyan)

	// Initials display
	initialsStr := fmt.Sprintf("%c %c %c", g.Initials[0], g.Initials[1], g.Initials[2])
	DrawArcadeText(target, initialsStr, 108, 120, yellow)

	// Underline cursor for active slot
	cursorX := 108 + g.InitialSlot*16
	DrawArcadeText(target, "-", cursorX, 130, white)

	DrawArcadeText(target, "UP/DOWN TO CHANGE", 60, 160, white)
	DrawArcadeText(target, "CLICK/SPACE TO SELECT", 44, 175, white)
}

// Layout maintains the native 256x231 simulation coordinates.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.Width = outsideWidth
	g.Height = outsideHeight
	return outsideWidth, outsideHeight
}
