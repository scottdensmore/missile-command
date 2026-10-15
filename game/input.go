package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func isStandardGamepadButtonJustPressed(buttons ...ebiten.StandardGamepadButton) bool {
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			for _, b := range buttons {
				if inpututil.IsStandardGamepadButtonJustPressed(id, b) {
					return true
				}
			}
		}
	}
	return false
}

// handleGlobalInput processes window/audio/options/pause hotkeys and returns true if Update should return early.
func (g *Game) handleGlobalInput() (bool, error) {
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
	if inpututil.IsKeyJustPressed(ebiten.KeyO) || isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonCenterLeft) {
		switch {
		case g.State == StateOptions:
			g.closeOptions()
			return true, nil
		case g.State == StateAttract:
			g.openOptions(StateAttract)
			return true, nil
		case g.State == StatePlaying || g.Paused:
			if !g.Paused {
				g.Paused = true
				StopAllContinuousSounds()
			}
			g.openOptions(StatePlaying)
			return true, nil
		}
	}

	if g.State == StateOptions {
		g.updateOptions()
		return true, nil
	}

	// Quit / Exit Handling (Q Key)
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if g.State == StateAttract {
			return true, ebiten.Termination
		}
		g.enterAttractMode()
		return true, nil
	}

	// Pause / Resume Toggle (P Key or Gamepad Start in game)
	if g.State != StateAttract && g.State != StateHighScoreEntry {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) || isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonCenterRight) {
			g.Paused = !g.Paused
			if g.Paused {
				StopAllContinuousSounds()
			} else {
				g.restoreContinuousSounds()
			}
		}
	}

	return false, nil
}

func (g *Game) checkAnyStartInput() bool {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.Key1) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return true
	}
	return isStandardGamepadButtonJustPressed(
		ebiten.StandardGamepadButtonCenterRight,
		ebiten.StandardGamepadButtonRightBottom,
		ebiten.StandardGamepadButtonRightRight,
		ebiten.StandardGamepadButtonRightLeft,
		ebiten.StandardGamepadButtonRightTop,
	)
}

func (g *Game) handleCrosshairInput() {
	currX, currY := ebiten.CursorPosition()
	if g.MouseCaptured {
		dx := float64(currX - g.PrevMouseX)
		dy := float64(currY - g.PrevMouseY)
		g.Crosshair.Pos.X += dx * 0.45
		g.Crosshair.Pos.Y += dy * 0.45
		g.PrevMouseX, g.PrevMouseY = currX, currY
	} else {
		g.Crosshair.Pos = ScreenToSim(float64(currX), float64(currY), g.Width, g.Height)
	}

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

	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if ebiten.IsStandardGamepadLayoutAvailable(id) {
			lx := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
			ly := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
			if math.Abs(lx) > 0.18 {
				g.Crosshair.Pos.X += lx * 3.5
			}
			if math.Abs(ly) > 0.18 {
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

	g.Crosshair.Pos.X = clamp(g.Crosshair.Pos.X, 4, SimWidth-4)
	g.Crosshair.Pos.Y = clamp(g.Crosshair.Pos.Y, 8, 206)
}

func (g *Game) handleFiringInput() {
	firedIdx := -1

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.Key1):
		firedIdx = 0
	case inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.Key2) || inpututil.IsKeyJustPressed(ebiten.KeySpace):
		firedIdx = 1
	case inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.Key3):
		firedIdx = 2
	}

	if g.MouseCaptured {
		switch {
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
			firedIdx = 0
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle):
			firedIdx = 1
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
			firedIdx = 2
		}
	}

	switch {
	case isStandardGamepadButtonJustPressed(
		ebiten.StandardGamepadButtonFrontTopLeft,
		ebiten.StandardGamepadButtonFrontBottomLeft,
		ebiten.StandardGamepadButtonRightBottom,
	):
		firedIdx = 0
	case isStandardGamepadButtonJustPressed(
		ebiten.StandardGamepadButtonRightLeft,
		ebiten.StandardGamepadButtonRightTop,
	):
		firedIdx = 1
	case isStandardGamepadButtonJustPressed(
		ebiten.StandardGamepadButtonFrontTopRight,
		ebiten.StandardGamepadButtonFrontBottomRight,
		ebiten.StandardGamepadButtonRightRight,
	):
		firedIdx = 2
	}

	if firedIdx != -1 {
		g.fireABM(firedIdx)
	}
}

// --- ATTRACT MODE & AI DEMO ---

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

	g.clearActiveEntities()
	g.resetCities()
	g.resetSilos()
}

func (g *Game) updateAttractDemo() {
	g.AttractTimer--
	if g.AttractTimer <= 0 {
		g.stopAttractDemo()
		return
	}

	g.stepSimulation()
	g.updateAIDemoPlayer()

	if g.survivingCitiesCount() == 0 {
		g.stopAttractDemo()
	}
}

func (g *Game) stopAttractDemo() {
	StopAllContinuousSounds()
	g.AttractDemoMode = false
	g.AttractTimer = 600
}

func (g *Game) updateAIDemoPlayer() {
	var bestTarget *Point
	lowestY := -1.0

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
			intercept := Point{X: sb.Curr.X, Y: sb.Curr.Y + 22}
			bestTarget = &intercept
			lowestY = sb.Curr.Y
		}
	}

	if bestTarget == nil && g.ActiveFlier != nil && g.ActiveFlier.Active {
		fPos := Point{X: g.ActiveFlier.X + 8 + g.ActiveFlier.Speed*20, Y: g.ActiveFlier.Y}
		bestTarget = &fPos
	}

	if bestTarget != nil {
		g.Crosshair.Pos.X = clamp(g.Crosshair.Pos.X+(bestTarget.X-g.Crosshair.Pos.X)*0.18, 4, SimWidth-4)
		g.Crosshair.Pos.Y = clamp(g.Crosshair.Pos.Y+(bestTarget.Y-g.Crosshair.Pos.Y)*0.18, 8, 206)

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
