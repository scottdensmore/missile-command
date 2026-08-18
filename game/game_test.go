package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

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
