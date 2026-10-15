package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) openOptions(fromState GameState) {
	g.PreviousState = fromState
	g.OptionSelectedIdx = 0
	g.State = StateOptions
}

func (g *Game) closeOptions() {
	if g.Settings != nil {
		_ = g.Settings.Save()
		if g.PreviousState == StatePlaying {
			for i := range g.Batteries {
				g.Batteries[i].MaxAmmo = g.Settings.MissilesPerSilo
				if g.Batteries[i].Ammo > g.Settings.MissilesPerSilo {
					g.Batteries[i].Ammo = g.Settings.MissilesPerSilo
				}
			}
		}
	}
	switch {
	case g.PreviousState == StatePlaying:
		g.Paused = true
		g.State = StatePlaying
	case g.PreviousState != 0:
		g.State = g.PreviousState
	default:
		g.enterAttractMode()
	}
}

func (g *Game) updateOptions() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonLeftTop) {
		g.OptionSelectedIdx = (g.OptionSelectedIdx + 3) % 4
		PlayTallySound()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) ||
		isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonLeftBottom) {
		g.OptionSelectedIdx = (g.OptionSelectedIdx + 1) % 4
		PlayTallySound()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) ||
		isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonLeftLeft) {
		g.adjustOption(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) ||
		isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonLeftRight) {
		g.adjustOption(1)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		isStandardGamepadButtonJustPressed(ebiten.StandardGamepadButtonRightBottom) {
		g.activateOption()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		g.closeOptions()
		return
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		simPos := ScreenToSim(float64(mx), float64(my), g.Width, g.Height)
		const startY, rowHeight = 64, 24
		for i := 0; i < 4; i++ {
			y := startY + i*rowHeight
			if simPos.Y >= float64(y-4) && simPos.Y <= float64(y+16) && simPos.X >= 16 && simPos.X <= 240 {
				g.OptionSelectedIdx = i
				switch i {
				case 0:
					g.toggleSoundOption()
				case 1:
					if simPos.X < 192 {
						g.adjustMissilesOption(-10)
					} else {
						g.adjustMissilesOption(10)
					}
				case 2:
					g.toggleCRTOption()
				case 3:
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
	newVal := int(clamp(float64(g.Settings.MissilesPerSilo+delta), 10, 100))
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
