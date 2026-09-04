package game

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestMain(m *testing.M) {
	tempSettings := filepath.Join(os.TempDir(), "missile_command_test_settings.json")
	_ = os.Remove(tempSettings)
	SetCustomSettingsPath(tempSettings)

	code := m.Run()

	SetCustomSettingsPath("")
	_ = os.Remove(tempSettings)
	os.Exit(code)
}

func TestPalettesAndMultipliers(t *testing.T) {
	tests := []struct {
		wave         int
		expectedMult int
	}{
		{wave: 1, expectedMult: 1},
		{wave: 2, expectedMult: 1},
		{wave: 3, expectedMult: 2},
		{wave: 4, expectedMult: 2},
		{wave: 5, expectedMult: 3},
		{wave: 6, expectedMult: 3},
		{wave: 7, expectedMult: 4},
		{wave: 8, expectedMult: 4},
		{wave: 9, expectedMult: 5},
		{wave: 10, expectedMult: 5},
		{wave: 11, expectedMult: 6},
		{wave: 12, expectedMult: 6},
		{wave: 20, expectedMult: 6},
		{wave: 25, expectedMult: 6},
	}

	for _, tc := range tests {
		p := GetPaletteForWave(tc.wave)
		if p.Multiplier != tc.expectedMult {
			t.Errorf("Wave %d: expected multiplier %d, got %d", tc.wave, tc.expectedMult, p.Multiplier)
		}
	}
}

func TestWaveTableProgression(t *testing.T) {
	w1 := GetWaveData(1)
	if w1.TotalICBMs != 12 || w1.SmartBombs != 0 {
		t.Errorf("Wave 1 data mismatch: %+v", w1)
	}

	w6 := GetWaveData(6)
	if w6.SmartBombs != 1 {
		t.Errorf("Wave 6 should introduce 1 smart bomb, got %d", w6.SmartBombs)
	}

	w19 := GetWaveData(19)
	if w19.TotalICBMs != 22 || w19.SmartBombs != 7 {
		t.Errorf("Wave 19 data mismatch: %+v", w19)
	}

	// Beyond wave 19 should clamp to wave 19 values
	w50 := GetWaveData(50)
	if w50.TotalICBMs != w19.TotalICBMs || w50.SmartBombs != w19.SmartBombs {
		t.Errorf("Wave 50 should match wave 19: %+v vs %+v", w50, w19)
	}
}

func TestCenterSiloSpeed(t *testing.T) {
	g := NewGame()
	StartGame(g)
	g.State = StatePlaying

	// Aim at center of screen
	g.Crosshair.Pos = Point{X: 128, Y: 100}

	// Fire Left Base (idx 0)
	g.fireABM(0)
	// Fire Center Base (idx 1)
	g.fireABM(1)

	if len(g.ABMs) != 2 {
		t.Fatalf("Expected 2 ABMs fired, got %d", len(g.ABMs))
	}

	leftABM := g.ABMs[0]
	centerABM := g.ABMs[1]

	if leftABM.Speed >= centerABM.Speed {
		t.Errorf("Center silo ABM speed (%f) should be significantly faster than side silo ABM speed (%f)",
			centerABM.Speed, leftABM.Speed)
	}

	if centerABM.Speed < 7.0 {
		t.Errorf("Expected Center silo ABM speed >= 7.0, got %f", centerABM.Speed)
	}
}

func TestSmartBombEvasion(t *testing.T) {
	g := NewGame()
	StartGame(g)
	g.State = StatePlaying

	// Place an active expanding explosion right in front of a descending smart bomb
	exp := &Explosion{
		Center:    Point{X: 100, Y: 100},
		Radius:    15.0,
		MaxRadius: 22.0,
		State:     StateExpanding,
	}
	g.Explosions = append(g.Explosions, exp)

	sb := &SmartBomb{
		Start:    Point{X: 100, Y: 75},
		Curr:     Point{X: 100, Y: 75},
		Target:   Point{X: 100, Y: 216},
		Speed:    1.5,
		Progress: 0.0,
		Active:   true,
	}
	g.SmartBombs = append(g.SmartBombs, sb)

	initialX := sb.Curr.X
	g.updateSmartBombs()

	// The smart bomb should steer sideways (dx != 0) to avoid the explosion center
	if sb.Curr.X == initialX {
		t.Errorf("Smart bomb failed to evade explosion: X remained %f", sb.Curr.X)
	}
}

func TestHighScores(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test_highscores.json")
	SetCustomHighScorePath(tempFile)
	defer SetCustomHighScorePath("")

	hs := DefaultHighScores()
	if len(hs.Entries) != 10 {
		t.Fatalf("Expected 10 default high score entries, got %d", len(hs.Entries))
	}

	if hs.Entries[0].Initials != "DFT" {
		t.Errorf("Expected top high score creator DFT, got %s", hs.Entries[0].Initials)
	}

	// Add a score higher than DFT
	hs.AddScore("NEW", 99999)
	if hs.Entries[0].Initials != "NEW" || hs.Entries[0].Score != 99999 {
		t.Errorf("Failed to insert top score properly: %+v", hs.Entries[0])
	}
	if len(hs.Entries) != 10 {
		t.Errorf("High score list should remain capped at 10, got %d", len(hs.Entries))
	}

	// Verify persistence reload from disk
	reloaded := LoadHighScores()
	if reloaded.Entries[0].Initials != "NEW" || reloaded.Entries[0].Score != 99999 {
		t.Errorf("Reloaded high scores mismatch: %+v", reloaded.Entries[0])
	}
}

func TestAudioBuffers(t *testing.T) {
	InitAudio()

	if len(launchPCM) == 0 {
		t.Error("launchPCM buffer is empty")
	}
	if len(explosionPCM) == 0 {
		t.Error("explosionPCM buffer is empty")
	}
	if len(tallyPCM) == 0 {
		t.Error("tallyPCM buffer is empty")
	}
	if len(sirenPCM) == 0 {
		t.Error("sirenPCM buffer is empty")
	}
	if len(siloLowPCM) == 0 {
		t.Error("siloLowPCM buffer is empty")
	}
	if len(cantFirePCM) == 0 {
		t.Error("cantFirePCM buffer is empty")
	}
	if len(bonusCityPCM) == 0 {
		t.Error("bonusCityPCM buffer is empty")
	}
	if len(gameOverPCM) == 0 {
		t.Error("gameOverPCM buffer is empty")
	}
	if len(bomberPCM) == 0 {
		t.Error("bomberPCM buffer is empty")
	}
	if len(satPCM) == 0 {
		t.Error("satPCM buffer is empty")
	}
	if len(smartBombPCM) == 0 {
		t.Error("smartBombPCM buffer is empty")
	}
}

func TestRenderPipeline(t *testing.T) {
	g := NewGame()
	screen := ebiten.NewImage(1024, 924)

	states := []GameState{
		StateAttract,
		StateWaveStart,
		StatePlaying,
		StateTally,
		StateBonusRebuild,
		StateTheEnd,
		StateHighScoreEntry,
		StateOptions,
	}

	for _, s := range states {
		g.State = s
		g.Draw(screen)
	}
}

func TestAmmoPyramidRenderingAndReduction(t *testing.T) {
	g := NewGame()
	StartGame(g)
	g.State = StatePlaying

	// Silos initially have 10 ammo
	for i := 0; i < 3; i++ {
		if g.Batteries[i].Ammo != 10 {
			t.Errorf("Silo %d should start with 10 ammo, got %d", i, g.Batteries[i].Ammo)
		}
	}

	// SiloColor must NOT match GroundColor in any wave palette for visibility
	for wave := 1; wave <= 20; wave++ {
		p := GetPaletteForWave(wave)
		if p.SiloColor == p.GroundColor {
			t.Errorf("Wave %d: SiloColor (%v) matches GroundColor (%v), making missiles invisible!",
				wave, p.SiloColor, p.GroundColor)
		}
	}

	// Verify DrawAmmoPyramid runs cleanly for every ammo count from 0 to 10
	img := ebiten.NewImage(256, 231)
	for ammo := 0; ammo <= 10; ammo++ {
		DrawAmmoPyramid(img, 128, 220, ammo, g.Palette.SiloColor)
	}

	// Test firing reduces ammo and triggers LOW/OUT properly
	g.Crosshair.Pos = Point{X: 128, Y: 100}
	for i := 0; i < 10; i++ {
		expectedAmmo := 10 - i
		if g.Batteries[0].Ammo != expectedAmmo {
			t.Errorf("Step %d: Expected Silo 0 ammo %d, got %d", i, expectedAmmo, g.Batteries[0].Ammo)
		}
		// Clear ABMs to stay under the 8 active ABM arcade limit
		g.ABMs = nil
		g.fireABM(0)
	}

	if g.Batteries[0].Ammo != 0 {
		t.Errorf("Expected Silo 0 to have 0 ammo after 10 shots, got %d", g.Batteries[0].Ammo)
	}
}

func TestPauseAndAudioMuting(t *testing.T) {
	g := NewGame()
	StartGame(g)
	g.State = StatePlaying

	// Spawn a descending ICBM
	g.ICBMs = append(g.ICBMs, &ICBM{
		Start:  Point{X: 100, Y: 0},
		Curr:   Point{X: 100, Y: 50},
		Target: Point{X: 100, Y: 216},
		Speed:  1.0,
		Active: true,
	})

	// Add an active bomber and start continuous sound
	g.ActiveFlier = &Flier{
		Type:   FlierBomber,
		X:      100,
		Y:      70,
		Speed:  0.8,
		Active: true,
	}
	SetBomberSound(true)

	// Pause game
	g.Paused = true
	StopAllContinuousSounds()

	// Capture ICBM position
	initialY := g.ICBMs[0].Curr.Y

	// Run Update while paused
	_ = g.Update()

	// Simulation should NOT advance while paused
	if g.ICBMs[0].Curr.Y != initialY {
		t.Errorf("ICBM moved while paused: expected Y=%f, got Y=%f", initialY, g.ICBMs[0].Curr.Y)
	}

	// Verify continuous sounds can be restored upon unpause
	g.Paused = false
	g.restoreContinuousSounds()

	// Unpaused Update should advance simulation
	_ = g.Update()
	if g.ICBMs[0].Curr.Y == initialY {
		t.Errorf("ICBM failed to advance after unpause")
	}

	StopAllContinuousSounds()
}

func TestWaveCompletionAndProgression(t *testing.T) {
	g := NewGame()
	StartGame(g)
	g.State = StatePlaying

	// Set remaining threats to 0 and clear entities
	g.ICBMsRemaining = 0
	g.SmartBombsLeft = 0
	g.ICBMs = nil
	g.SmartBombs = nil
	g.ActiveFlier = nil
	g.ABMs = nil
	g.Explosions = nil

	// Check wave completion
	g.checkWaveCompletion()

	if g.State != StateTally {
		t.Errorf("Expected state to transition to StateTally on wave end, got %v", g.State)
	}

	StopAllContinuousSounds()
	// Step through Tally state (30 ammo * 8 frames + 6 cities * 8 frames + delays ~ 400 frames)
	for i := 0; i < 500 && g.State == StateTally; i++ {
		g.updateTally()
	}

	if g.State != StateBonusRebuild {
		t.Errorf("Expected state to transition to StateBonusRebuild, got %v", g.State)
	}

	// Step through BonusRebuild to advance to Wave 2
	for i := 0; i < 100 && g.State == StateBonusRebuild; i++ {
		g.updateBonusRebuild()
	}

	if g.Wave != 2 {
		t.Errorf("Expected wave to advance to 2, got %d", g.Wave)
	}
	if g.State != StateWaveStart {
		t.Errorf("Expected state to be StateWaveStart for wave 2, got %v", g.State)
	}
}

func TestFullGameLifecycle(t *testing.T) {
	g := NewGame()
	StartGame(g)

	// Simulate Wave Start
	g.State = StateWaveStart
	for g.State == StateWaveStart {
		_ = g.Update()
	}

	if g.State != StatePlaying {
		t.Fatalf("Expected state to transition from WaveStart to Playing, got %v", g.State)
	}

	// Aim and fire across multiple silos
	g.Crosshair.Pos = Point{X: 128, Y: 100}
	g.fireABM(0) // Left
	g.fireABM(1) // Center
	g.fireABM(2) // Right

	if len(g.ABMs) != 3 {
		t.Errorf("Expected 3 in-flight ABMs, got %d", len(g.ABMs))
	}

	// Advance frames until ABMs detonate into explosions
	for i := 0; i < 60 && len(g.ABMs) > 0; i++ {
		_ = g.Update()
	}

	if len(g.Explosions) == 0 {
		t.Errorf("Expected active explosions from detonated ABMs, got 0")
	}

	// Test Annihilation -> The End -> High Score Entry
	for i := range g.Cities {
		g.Cities[i].Destroyed = true
	}
	g.checkWaveCompletion()

	if g.State != StateTheEnd {
		t.Fatalf("Expected state to transition to StateTheEnd on annihilation, got %v", g.State)
	}

	// Step through The End sequence
	for i := 0; i < 250 && g.State == StateTheEnd; i++ {
		_ = g.Update()
	}

	if g.State != StateHighScoreEntry && g.State != StateAttract {
		t.Errorf("Expected state to transition from The End, got %v", g.State)
	}
}

func TestAudioSynthesisIntegrity(t *testing.T) {
	InitAudio()

	pcms := map[string][]byte{
		"launch":    launchPCM,
		"explosion": explosionPCM,
		"tally":     tallyPCM,
		"siren":     sirenPCM,
		"siloLow":   siloLowPCM,
		"cantFire":  cantFirePCM,
		"bonusCity": bonusCityPCM,
		"gameOver":  gameOverPCM,
		"bomber":    bomberPCM,
		"sat":       satPCM,
		"smartBomb": smartBombPCM,
	}

	for name, buf := range pcms {
		if len(buf) == 0 {
			t.Errorf("Audio PCM buffer %s is empty!", name)
		}
		if len(buf)%2 != 0 {
			t.Errorf("Audio PCM buffer %s length (%d) is not an even number of 16-bit bytes!", name, len(buf))
		}
	}
}

func TestCrosshairAndInputClamping(t *testing.T) {
	g := NewGame()
	StartGame(g)
	g.State = StatePlaying

	// Test extreme out-of-bounds crosshair positions
	g.Crosshair.Pos = Point{X: -500, Y: -500}
	g.handleCrosshairInput()

	if g.Crosshair.Pos.X < 4 || g.Crosshair.Pos.Y < 8 {
		t.Errorf("Crosshair min clamping failed: pos = %+v", g.Crosshair.Pos)
	}

	g.Crosshair.Pos = Point{X: 1000, Y: 1000}
	g.handleCrosshairInput()

	if g.Crosshair.Pos.X > SimWidth-4 || g.Crosshair.Pos.Y > 206 {
		t.Errorf("Crosshair max clamping failed: pos = %+v", g.Crosshair.Pos)
	}
}

func TestCityPlacementAndSiloSeparation(t *testing.T) {
	g := NewGame()
	StartGame(g)

	siloBases := [][2]float64{
		{18 - 13, 18 + 13},   // Left silo mound: [5, 31]
		{128 - 13, 128 + 13}, // Middle silo mound: [115, 141]
		{238 - 13, 238 + 13}, // Right silo mound: [225, 251]
	}

	cityWidth := 16.0

	// Check each city does not overlap with any silo
	for i, city := range g.Cities {
		cityLeft := city.Position.X
		cityRight := city.Position.X + cityWidth

		for sIdx, silo := range siloBases {
			siloLeft := silo[0]
			siloRight := silo[1]

			if cityLeft < siloRight && cityRight > siloLeft {
				t.Errorf("City %d [%.1f, %.1f] overlaps with silo %d [%.1f, %.1f]",
					i, cityLeft, cityRight, sIdx, siloLeft, siloRight)
			}
		}
	}

	// Verify equal 9px spacing and centering in valleys
	expectedXs := []float64{40, 65, 90, 150, 175, 200}
	for i, expX := range expectedXs {
		if g.Cities[i].Position.X != expX {
			t.Errorf("City %d X position = %.1f, expected %.1f", i, g.Cities[i].Position.X, expX)
		}
	}

	// Verify symmetry across screen center (128.0)
	for i := 0; i < 3; i++ {
		leftCity := g.Cities[i]
		rightCity := g.Cities[5-i]

		leftDist := 128.0 - (leftCity.Position.X + cityWidth/2.0)
		rightDist := (rightCity.Position.X + cityWidth/2.0) - 128.0

		if math.Abs(leftDist-rightDist) > 1e-6 {
			t.Errorf("Asymmetry between city %d (center %.1f) and city %d (center %.1f)",
				i, leftCity.Position.X+cityWidth/2.0, 5-i, rightCity.Position.X+cityWidth/2.0)
		}
	}
}

func TestAttractModeAndDemoTransitions(t *testing.T) {
	g := NewGame()
	if g.State != StateAttract {
		t.Fatalf("Expected initial state to be StateAttract, got %v", g.State)
	}
	if g.AttractDemoMode {
		t.Error("Attract mode should start in high scores screen, not demo mode")
	}

	// Trigger demo mode transition
	g.AttractTimer = 1
	g.updateAttract()

	if !g.AttractDemoMode {
		t.Error("Expected AttractDemoMode to become true after AttractTimer expires")
	}
	if g.Wave != 1 || g.Batteries[0].Ammo != 10 {
		t.Errorf("Demo mode did not initialize wave 1: wave=%d, ammo=%d", g.Wave, g.Batteries[0].Ammo)
	}

	// Step a few demo simulation frames
	for i := 0; i < 30; i++ {
		g.updateAttractDemo()
	}

	// Verify Draw handles demo mode without panic
	screen := ebiten.NewImage(256, 231)
	g.Draw(screen)

	// Stop demo mode
	g.stopAttractDemo()
	if g.AttractDemoMode {
		t.Error("Expected AttractDemoMode to become false after stopAttractDemo")
	}
}

func TestMasterVolumeAndMute(t *testing.T) {
	InitAudio()

	SetMasterVolume(1.0)
	if math.Abs(GetMasterVolume()-1.0) > 1e-6 {
		t.Errorf("Expected volume 1.0, got %f", GetMasterVolume())
	}

	AdjustVolume(-0.2)
	if math.Abs(GetMasterVolume()-0.8) > 1e-6 {
		t.Errorf("Expected volume 0.8, got %f", GetMasterVolume())
	}

	// Test bounds clamping
	AdjustVolume(-2.0)
	if GetMasterVolume() < 0.0 {
		t.Errorf("Volume below 0: %f", GetMasterVolume())
	}

	AdjustVolume(2.0)
	if GetMasterVolume() > 1.0 {
		t.Errorf("Volume above 1.0: %f", GetMasterVolume())
	}

	// Test mute toggle
	if IsMuted() {
		ToggleMute() // ensure unmuted
	}

	muted := ToggleMute()
	if !muted || !IsMuted() {
		t.Error("Expected mute to be active")
	}

	unmuted := ToggleMute()
	if unmuted || IsMuted() {
		t.Error("Expected mute to be inactive")
	}
}

func TestSoundEffectsToggle(t *testing.T) {
	SetSoundEffectsEnabled(true)
	if !IsSoundEffectsEnabled() {
		t.Errorf("Expected SoundEffectsEnabled to be true")
	}
	if IsMuted() {
		t.Errorf("Expected IsMuted to be false when sound effects enabled")
	}

	SetSoundEffectsEnabled(false)
	if IsSoundEffectsEnabled() {
		t.Errorf("Expected SoundEffectsEnabled to be false")
	}
	if !IsMuted() {
		t.Errorf("Expected IsMuted to be true when sound effects disabled")
	}

	// Restore sound
	SetSoundEffectsEnabled(true)
}

func TestMissilesPerSiloConfiguration(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test_settings.json")
	SetCustomSettingsPath(tempFile)
	defer SetCustomSettingsPath("")

	g := NewGame()
	g.Settings.MissilesPerSilo = 50
	g.resetSilos()

	for i := 0; i < 3; i++ {
		if g.Batteries[i].MaxAmmo != 50 || g.Batteries[i].Ammo != 50 {
			t.Errorf("Battery %d expected MaxAmmo=50 and Ammo=50, got Max=%d Ammo=%d",
				i, g.Batteries[i].MaxAmmo, g.Batteries[i].Ammo)
		}
	}

	// Test 100 missiles
	g.Settings.MissilesPerSilo = 100
	g.resetSilos()
	for i := 0; i < 3; i++ {
		if g.Batteries[i].MaxAmmo != 100 || g.Batteries[i].Ammo != 100 {
			t.Errorf("Battery %d expected MaxAmmo=100 and Ammo=100, got Max=%d Ammo=%d",
				i, g.Batteries[i].MaxAmmo, g.Batteries[i].Ammo)
		}
	}
}

func TestCalculateSiloNeedles(t *testing.T) {
	// For maxAmmo 10
	if CalculateSiloNeedles(0, 10) != 0 {
		t.Errorf("Expected 0 needles for 0 ammo")
	}
	if CalculateSiloNeedles(5, 10) != 5 {
		t.Errorf("Expected 5 needles for 5 ammo out of 10")
	}
	if CalculateSiloNeedles(10, 10) != 10 {
		t.Errorf("Expected 10 needles for 10 ammo out of 10")
	}

	// For maxAmmo 100
	if CalculateSiloNeedles(100, 100) != 10 {
		t.Errorf("Expected 10 needles for 100/100 ammo")
	}
	if CalculateSiloNeedles(90, 100) != 9 {
		t.Errorf("Expected 9 needles for 90/100 ammo")
	}
	if CalculateSiloNeedles(50, 100) != 5 {
		t.Errorf("Expected 5 needles for 50/100 ammo")
	}
	if CalculateSiloNeedles(1, 100) != 1 {
		t.Errorf("Expected 1 needle for 1/100 ammo")
	}
	if CalculateSiloNeedles(0, 100) != 0 {
		t.Errorf("Expected 0 needles for 0/100 ammo")
	}
}

func TestOptionsNavigationAndAdjustment(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "test_settings.json")
	SetCustomSettingsPath(tempFile)
	defer SetCustomSettingsPath("")

	g := NewGame()
	g.openOptions(StateAttract)

	if g.State != StateOptions {
		t.Fatalf("Expected StateOptions, got %v", g.State)
	}
	if g.OptionSelectedIdx != 0 {
		t.Errorf("Expected initial option idx 0, got %d", g.OptionSelectedIdx)
	}

	// Navigate down to Missiles/Silo (idx 1)
	g.OptionSelectedIdx = 1

	// Increase missiles by 10
	initMissiles := g.Settings.MissilesPerSilo
	g.adjustOption(1)
	if g.Settings.MissilesPerSilo != initMissiles+10 {
		t.Errorf("Expected missiles %d, got %d", initMissiles+10, g.Settings.MissilesPerSilo)
	}

	// Decrease missiles by 10
	g.adjustOption(-1)
	if g.Settings.MissilesPerSilo != initMissiles {
		t.Errorf("Expected missiles %d, got %d", initMissiles, g.Settings.MissilesPerSilo)
	}

	// Cycle forward via activateOption
	g.Settings.MissilesPerSilo = 100
	g.activateOption()
	if g.Settings.MissilesPerSilo != 10 {
		t.Errorf("Expected wrap from 100 to 10, got %d", g.Settings.MissilesPerSilo)
	}

	// Navigate to Sound Effects (idx 0) and toggle
	g.OptionSelectedIdx = 0
	initSound := g.Settings.SoundEffectsEnabled
	g.activateOption()
	if g.Settings.SoundEffectsEnabled == initSound {
		t.Errorf("Expected sound effects to toggle from %v", initSound)
	}

	// Close options
	g.closeOptions()
	if g.State != StateAttract {
		t.Errorf("Expected return to StateAttract, got %v", g.State)
	}
}

func TestSiloDrawingWithVariousAmmoLevels(t *testing.T) {
	g := NewGame()
	img := ebiten.NewImage(256, 231)

	// Test drawing silo for ammo levels 0..100 with maxAmmo 100
	for ammo := 0; ammo <= 100; ammo += 10 {
		bat := Battery{
			Index:    1,
			Position: Point{X: 128, Y: 214},
			MaxAmmo:  100,
			Ammo:     ammo,
		}
		g.drawSilo(img, bat)
	}

	// Test LOW flashing state (ammo 1, 2, 3)
	for ammo := 1; ammo <= 3; ammo++ {
		bat := Battery{
			Index:    0,
			Position: Point{X: 18, Y: 214},
			MaxAmmo:  50,
			Ammo:     ammo,
		}
		g.drawSilo(img, bat)
	}

	// Test OUT state (ammo 0)
	batZero := Battery{
		Index:    2,
		Position: Point{X: 238, Y: 214},
		MaxAmmo:  30,
		Ammo:     0,
	}
	g.drawSilo(img, batZero)
}

func TestDrawSiloTicks(t *testing.T) {
	g := NewGame()
	img := ebiten.NewImage(256, 231)

	// Test maxAmmo <= 10 (classic pyramid)
	for ammo := 0; ammo <= 10; ammo++ {
		DrawSiloTicks(img, 128, 222, ammo, 10, g.Palette.SiloColor, g.Palette.GroundColor)
	}

	// Test maxAmmo = 20..100 in steps of 10
	for maxAmmo := 20; maxAmmo <= 100; maxAmmo += 10 {
		for ammo := 0; ammo <= maxAmmo; ammo += 5 {
			DrawSiloTicks(img, 128, 222, ammo, maxAmmo, g.Palette.SiloColor, g.Palette.GroundColor)
		}
	}

	// Test edge cases
	DrawSiloTicks(img, 128, 222, -5, 100, g.Palette.SiloColor, g.Palette.GroundColor)
	DrawSiloTicks(img, 128, 222, 120, 100, g.Palette.SiloColor, g.Palette.GroundColor)
}

func TestSiloAmmoTicksAndNumbersTransition(t *testing.T) {
	g := NewGame()

	// Test rendering across the transition threshold (ammo > 10 down to 0)
	for ammo := 15; ammo >= 0; ammo-- {
		img := ebiten.NewImage(256, 231)
		bat := Battery{
			Index:    1,
			Position: Point{X: 128, Y: 214},
			MaxAmmo:  50,
			Ammo:     ammo,
		}
		g.drawSilo(img, bat)

		// Check pixel in the mound center where pyramid apex / number is drawn
		// For ammo > 10: number is drawn at Y=214
		// For ammo <= 10: pyramid ticks are drawn
		// Should execute cleanly without panic
	}
}



