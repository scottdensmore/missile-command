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
	ICBMsRemaining int
	SmartBombsLeft int
	SpawnCooldown  int
	FlierCooldown  int

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

	// Mouse and Input
	PrevMouseX, PrevMouseY int
	MouseCaptured          bool
	Tick                   int
}

// NewGame initializes the game engine and loads scores/audio.
func NewGame() *Game {
	InitAudio()

	g := &Game{
		Width:              1024,
		Height:             924,
		pipeline:           NewPipeline(),
		highScores:         LoadHighScores(),
		State:              StateAttract,
		Wave:               1,
		Score:              0,
		NextBonusThreshold: 10000,
		Letters:            []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ.!- 0123456789"),
	}

	g.Crosshair.Pos = Point{X: SimWidth / 2, Y: SimHeight / 2}
	g.Palette = GetPaletteForWave(1)
	return g
}

// StartGame starts a fresh 1-player game.
func StartGame(g *Game) {
	g.Wave = 1
	g.Score = 0
	g.BonusCitiesStored = 0
	g.NextBonusThreshold = 10000
	g.Palette = GetPaletteForWave(1)

	// Rebuild all 6 cities
	cityXs := []float64{42, 64, 86, 170, 192, 214}
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
	batteryXs := []float64{18, 128, 238}
	for i, x := range batteryXs {
		g.Batteries[i] = Battery{
			Index:            i,
			Position:         Point{X: x, Y: 214},
			MaxAmmo:          10,
			Ammo:             10,
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
	}

	return nil
}

// --- STATE UPDATES ---

func (g *Game) updateAttract() {
	// Start game on Left Click, Space, 1, or Enter
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.Key1) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		StartGame(g)
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

	g.SpawnCooldown--
	if g.SpawnCooldown > 0 {
		return
	}

	wdata := GetWaveData(g.Wave)

	// Determine whether to spawn a Smart Bomb or ICBM
	if g.SmartBombsLeft > 0 && (g.ICBMsRemaining <= 0 || rand.Float64() < 0.35) {
		g.SmartBombsLeft--
		spawnX := 20.0 + rand.Float64()*(SimWidth-40.0)
		target := g.pickRandomTarget()

		sb := &SmartBomb{
			Start:    Point{X: spawnX, Y: 0},
			Curr:     Point{X: spawnX, Y: 0},
			Target:   target,
			Speed:    wdata.Speed * 0.95,
			Progress: 0.0,
			Active:   true,
		}
		g.SmartBombs = append(g.SmartBombs, sb)
		SetSmartBombSound(true)
	} else if g.ICBMsRemaining > 0 {
		g.ICBMsRemaining--
		spawnX := rand.Float64() * SimWidth
		target := g.pickRandomTarget()

		// Wave >= 2 splitters (MIRVs)
		isSplit := false
		if g.Wave >= 2 && rand.Float64() < 0.25 {
			isSplit = true
		}

		splitAlt := 50.0 + rand.Float64()*80.0

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

	// Set next spawn cooldown
	minCD := math.Max(15, 60-float64(g.Wave)*3.5)
	maxCD := math.Max(30, 110-float64(g.Wave)*6.0)
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
		g.FlierCooldown--
		if g.FlierCooldown <= 0 && (g.ICBMsRemaining > 0 || len(g.ICBMs) > 0) {
			// Spawn Flier
			ft := FlierBomber
			alt := 65.0 + rand.Float64()*25.0
			speed := 0.65 + float64(g.Wave)*0.06
			if rand.Float64() < 0.45 {
				ft = FlierSatellite
				alt = 28.0 + rand.Float64()*18.0
				speed = 0.85 + float64(g.Wave)*0.08
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
				DropCooldown:   40 + rand.Intn(60),
				BombsRemaining: 2 + rand.Intn(3),
				Active:         true,
			}

			if ft == FlierBomber {
				SetBomberSound(true)
			} else {
				SetSatelliteSound(true)
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
		wdata := GetWaveData(g.Wave)
		g.FlierCooldown = wdata.FlierDelay + rand.Intn(120)
		return
	}

	// Dropping bombs from flier
	if flier.BombsRemaining > 0 && flier.X >= 10 && flier.X <= SimWidth-10 {
		flier.DropCooldown--
		if flier.DropCooldown <= 0 {
			flier.BombsRemaining--
			flier.DropCooldown = 50 + rand.Intn(50)

			target := g.pickRandomTarget()
			wdata := GetWaveData(g.Wave)
			icbm := &ICBM{
				Start:    Point{X: flier.X, Y: flier.Y},
				Curr:     Point{X: flier.X, Y: flier.Y},
				Target:   target,
				Speed:    wdata.Speed * 1.1,
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

		// MIRV Split Logic
		if icbm.IsSplitter && !icbm.Splitted && icbm.Curr.Y >= icbm.SplitAltitude {
			icbm.Splitted = true
			// Spawn 2 sister warheads
			for s := 0; s < 2; s++ {
				sisterTarget := g.pickRandomTarget()
				sister := &ICBM{
					Start:      icbm.Curr,
					Curr:       icbm.Curr,
					Target:     sisterTarget,
					Speed:      icbm.Speed * 1.05,
					Progress:   0.0,
					IsSplitter: false,
					Active:     true,
				}
				g.ICBMs = append(g.ICBMs, sister)
			}
		}

		dist := icbm.Start.Distance(icbm.Target)
		if dist <= 0 {
			icbm.Active = false
			continue
		}

		icbm.Progress += icbm.Speed / dist
		if icbm.Progress >= 1.0 {
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
					steerY := (dy / length) * sb.Speed * 0.40
					sb.Curr.X += steerX
					sb.Curr.Y += steerY
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

	// Wave completes when no threats and no active player munitions remain
	if g.ICBMsRemaining <= 0 && g.SmartBombsLeft <= 0 && len(g.ICBMs) == 0 &&
		len(g.SmartBombs) == 0 && g.ActiveFlier == nil && len(g.ABMs) == 0 && len(g.Explosions) == 0 {

		StopAllContinuousSounds()
		g.State = StateTally
		g.TallyPhase = 0
		g.TallyTimer = 30
		g.TallyBatteryIdx = 0
		g.TallyCityIdx = 0
		g.TallyScoreEarned = 0
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
				bat.Ammo--
				pts := 5 * g.Palette.Multiplier
				g.Score += pts
				g.TallyScoreEarned += pts
				PlayTallySound()
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

	// 1. Draw HUD
	g.drawHUD(target)

	if g.State == StateAttract {
		g.drawAttractScreen(target)
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
}

func (g *Game) drawTerrain(target *ebiten.Image) {
	groundCol := g.Palette.GroundColor

	// Draw base ground fill
	vector.DrawFilledRect(target, 0, 224, SimWidth, 7, groundCol, false)

	// Draw 3 Silo terrain mounds
	siloXs := []float64{18, 128, 238}
	for _, sx := range siloXs {
		vector.DrawFilledRect(target, float32(sx-12), 220, 24, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-8), 216, 16, 5, groundCol, false)
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

	// Draw Ammo Pyramid
	DrawAmmoPyramid(target, bx, by+6, bat.Ammo, g.Palette.SiloColor)

	// Status text indicators
	if bat.Ammo <= 3 && bat.Ammo > 0 {
		// Flash LOW
		if (g.Tick/15)%2 == 0 {
			DrawArcadeText(target, "LOW", bx-12, by-10, yellow)
		}
	} else if bat.Ammo == 0 {
		// Flash OUT
		if (g.Tick/15)%2 == 0 {
			DrawArcadeText(target, "OUT", bx-12, by-10, red)
		}
	}
}

func (g *Game) drawAttractScreen(target *ebiten.Image) {
	white := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	yellow := color.RGBA{R: 240, G: 220, B: 32, A: 255}
	cyan := color.RGBA{R: 40, G: 220, B: 240, A: 255}
	red := color.RGBA{R: 240, G: 40, B: 40, A: 255}

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
	DrawArcadeText(target, "© 1980 ATARI INC", 64, 178, yellow)
	DrawArcadeText(target, "PRESS SPACE OR CLICK", 48, 208, white)
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
