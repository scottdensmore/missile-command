package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Draw renders the arcade game visuals into the 256x231 framebuffer and outputs via the pipeline.
func (g *Game) Draw(screen *ebiten.Image) {
	fb := g.pipeline.frameBuffer
	fb.Clear()

	skyCol := g.Palette.SkyColor
	if g.State == StateAttract {
		skyCol = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	fb.Fill(skyCol)

	g.renderPlayfield(fb)
	g.pipeline.DrawFrame(screen, g.Width, g.Height)
}

func (g *Game) renderPlayfield(target *ebiten.Image) {
	g.drawHUD(target)

	switch g.State {
	case StateAttract:
		g.drawAttractScreen(target)
		return
	case StateOptions:
		g.drawOptionsScreen(target)
		return
	}

	g.drawEntities(target, g.State == StatePlaying || g.State == StateWaveStart)

	switch g.State {
	case StateWaveStart:
		DrawArcadeText(target, fmt.Sprintf("START WAVE %d", g.Wave), 80, 80, g.Palette.TextColor)
		DrawArcadeText(target, fmt.Sprintf("%dX BONUS POINTS", g.Palette.Multiplier), 68, 100, g.Palette.MultiplierColor)
	case StateTally:
		DrawArcadeText(target, "BONUS POINTS", 80, 60, ColorYellow)
		DrawArcadeText(target, fmt.Sprintf("+%d", g.TallyScoreEarned), 108, 80, ColorWhite)
	case StateTheEnd:
		g.drawTheEndScreen(target)
	case StateHighScoreEntry:
		g.drawHighScoreEntry(target)
	}

	if g.Paused {
		vector.DrawFilledRect(target, 36, 72, 184, 68, color.RGBA{R: 0, G: 0, B: 0, A: 230}, false)
		vector.StrokeRect(target, 36, 72, 184, 68, 1.0, ColorYellow, false)

		pauseCol := ColorWhite
		if (g.Tick/20)%2 == 0 {
			pauseCol = ColorYellow
		}
		DrawArcadeText(target, "GAME PAUSED", 84, 80, pauseCol)
		DrawArcadeText(target, "PRESS P TO RESUME", 60, 96, ColorWhite)
		DrawArcadeText(target, "PRESS O FOR OPTIONS", 52, 110, ColorCyan)
		DrawArcadeText(target, "PRESS Q TO QUIT", 68, 124, g.Palette.TextColor)
	}
}

// drawEntities renders terrain, cities, silos, ordnance, explosions, and the crosshair.
// Shared between active gameplay and the attract mode AI demo.
func (g *Game) drawEntities(target *ebiten.Image, showCrosshair bool) {
	g.drawTerrain(target)

	for _, city := range g.Cities {
		DrawCity(target, int(city.Position.X), int(city.Position.Y), city.Destroyed, g.Palette.CityColor)
	}

	for _, bat := range g.Batteries {
		g.drawSilo(target, bat)
	}

	for i := 0; i < g.BonusCitiesStored && i < 12; i++ {
		DrawBonusCityIcon(target, 4+i*10, int(SimHeight)-7, g.Palette.CityColor)
	}

	for _, icbm := range g.ICBMs {
		if !icbm.Active {
			continue
		}
		vector.StrokeLine(target, float32(icbm.Start.X), float32(icbm.Start.Y), float32(icbm.Curr.X), float32(icbm.Curr.Y), 1.0, g.Palette.ICBMColor, false)
		target.Set(int(icbm.Curr.X), int(icbm.Curr.Y), ColorWhite)
	}

	for _, sb := range g.SmartBombs {
		if !sb.Active {
			continue
		}
		vector.StrokeLine(target, float32(sb.Start.X), float32(sb.Start.Y), float32(sb.Curr.X), float32(sb.Curr.Y), 1.0, ColorRed, false)
		DrawSmartBomb(target, int(sb.Curr.X)-4, int(sb.Curr.Y)-4, sb.AnimTick/6, ColorYellow)
	}

	if g.ActiveFlier != nil && g.ActiveFlier.Active {
		flier := g.ActiveFlier
		if flier.Type == FlierBomber {
			DrawBomber(target, int(flier.X), int(flier.Y), flier.Speed > 0, g.Palette.ICBMColor)
		} else {
			DrawSatellite(target, int(flier.X), int(flier.Y), g.Palette.ICBMColor)
		}
	}

	for _, abm := range g.ABMs {
		if !abm.Active {
			continue
		}
		vector.StrokeLine(target, float32(abm.Start.X), float32(abm.Start.Y), float32(abm.Curr.X), float32(abm.Curr.Y), 1.0, g.Palette.ABMColor, false)
		DrawArcadeText(target, "X", int(abm.Target.X)-4, int(abm.Target.Y)-4, ColorWhite)
	}

	for _, exp := range g.Explosions {
		if exp.State == StateDead {
			continue
		}
		col := GetExplosionColor(exp.Tick)
		vector.DrawFilledCircle(target, float32(exp.Center.X), float32(exp.Center.Y), float32(exp.Radius), col, false)
		if exp.Radius > 5.0 {
			vector.DrawFilledCircle(target, float32(exp.Center.X), float32(exp.Center.Y), float32(exp.Radius*0.4), ColorWhite, false)
		}
	}

	if showCrosshair {
		DrawArcadeText(target, "X", int(g.Crosshair.Pos.X)-4, int(g.Crosshair.Pos.Y)-4, ColorWhite)
	}
}

func (g *Game) drawHUD(target *ebiten.Image) {
	DrawArcadeText(target, fmt.Sprintf("%d", g.Score), 24, 8, ColorWhite)
	DrawArcadeText(target, fmt.Sprintf("%d", g.highScores.GetTopScore()), 116, 8, ColorYellow)

	if g.State == StatePlaying {
		DrawArcadeText(target, fmt.Sprintf("%dX", g.Palette.Multiplier), 220, 8, g.Palette.MultiplierColor)
	}

	if g.NotificationTimer > 0 {
		DrawArcadeText(target, g.NotificationText, 128-len(g.NotificationText)*4, 22, ColorYellow)
	}
}

func (g *Game) drawTerrain(target *ebiten.Image) {
	groundCol := g.Palette.GroundColor
	vector.DrawFilledRect(target, 0, 224, SimWidth, 7, groundCol, false)

	for _, sx := range batteryXPositions {
		vector.DrawFilledRect(target, float32(sx-13), 219, 26, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-10), 215, 20, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-7), 211, 14, 5, groundCol, false)
		vector.DrawFilledRect(target, float32(sx-4), 207, 8, 5, groundCol, false)
	}
}

func (g *Game) drawSilo(target *ebiten.Image, bat Battery) {
	bx, by := int(bat.Position.X), int(bat.Position.Y)
	if bat.Destroyed {
		DrawArcadeText(target, "...", bx-12, by-4, ColorRed)
		return
	}

	DrawSiloTicks(target, bx, 222, bat.Ammo, bat.MaxAmmo, g.Palette.SiloColor, g.Palette.GroundColor)

	if (g.Tick/15)%2 == 0 {
		if bat.Ammo == 0 {
			DrawArcadeText(target, "OUT", bx-12, 196, ColorRed)
		} else if bat.Ammo <= 3 {
			DrawArcadeText(target, "LOW", bx-12, 196, ColorYellow)
		}
	}
}

func (g *Game) drawAttractScreen(target *ebiten.Image) {
	if g.AttractDemoMode {
		g.drawEntities(target, true)
		DrawArcadeText(target, "GAME OVER", 92, 40, ColorRed)
		DrawArcadeText(target, "DEFEND CITIES", 76, 55, ColorYellow)
		DrawArcadeText(target, "DEMO MODE", 92, 70, ColorCyan)
		if (g.Tick/30)%2 == 0 {
			DrawArcadeText(target, "PRESS SPACE OR CLICK", 48, 90, ColorWhite)
		}
		return
	}

	DrawArcadeText(target, "MISSILE COMMAND", 68, 30, ColorYellow)
	DrawArcadeText(target, "DEFEND CITIES", 76, 45, ColorRed)

	DrawArcadeText(target, "GREAT SCORES", 80, 68, ColorCyan)
	for i, entry := range g.highScores.Entries {
		if i >= 8 {
			break
		}
		line := fmt.Sprintf("%2d  %5d  %3s", i+1, entry.Score, entry.Initials)
		DrawArcadeText(target, line, 64, 84+i*11, ColorWhite)
	}

	DrawArcadeText(target, "© 1980 ATARI INC", 64, 176, ColorYellow)
	DrawArcadeText(target, "O:OPTIONS  P:PAUSE  Q:QUIT", 24, 190, ColorCyan)
	if (g.Tick/30)%2 == 0 {
		DrawArcadeText(target, "PRESS SPACE OR CLICK", 48, 208, ColorWhite)
	}
}

func (g *Game) drawOptionsScreen(target *ebiten.Image) {
	DrawArcadeText(target, "GAME OPTIONS", 80, 28, ColorYellow)

	backLabel := "BACK TO TITLE"
	if g.PreviousState == StatePlaying {
		backLabel = "BACK TO GAME"
	}

	soundVal, soundCol := "[ OFF ]", ColorRed
	if g.Settings != nil && g.Settings.SoundEffectsEnabled {
		soundVal, soundCol = "[ ON  ]", ColorGreen
	}

	missilesVal := 10
	if g.Settings != nil {
		missilesVal = g.Settings.MissilesPerSilo
	}

	crtVal, crtCol := "[ OFF ]", ColorRed
	if g.pipeline.UseCRT {
		crtVal, crtCol = "[ ON  ]", ColorGreen
	}

	type optRow struct {
		label  string
		val    string
		valCol color.RGBA
	}

	rows := []optRow{
		{label: "SOUND EFFECTS", val: soundVal, valCol: soundCol},
		{label: "MISSILES/SILO", val: fmt.Sprintf("< %3d >", missilesVal), valCol: ColorYellow},
		{label: "CRT SCANLINES", val: crtVal, valCol: crtCol},
		{label: backLabel, val: "", valCol: ColorWhite},
	}

	const startY, rowHeight = 64, 24
	for i, r := range rows {
		y := startY + i*rowHeight
		lblCol := ColorCyan
		if i == g.OptionSelectedIdx {
			lblCol = ColorYellow
			cursorCol := ColorWhite
			if (g.Tick/15)%2 == 0 {
				cursorCol = ColorYellow
			}
			DrawArcadeText(target, ">", 16, y, cursorCol)
		}

		DrawArcadeText(target, r.label, 28, y, lblCol)
		if r.val != "" {
			DrawArcadeText(target, r.val, 164, y, r.valCol)
		}
	}

	DrawArcadeText(target, "UP/DOWN : SELECT", 64, 178, ColorCyan)
	DrawArcadeText(target, "LEFT/RIGHT/CLICK : CHANGE", 28, 192, ColorWhite)
	DrawArcadeText(target, "ESC/ENTER : BACK", 64, 206, ColorYellow)
}

func (g *Game) drawTheEndScreen(target *ebiten.Image) {
	blastRadius := float32(math.Min(SimWidth, float64(200-g.TheEndTimer)*2.0))
	vector.DrawFilledCircle(target, SimWidth/2, SimHeight/2+20, blastRadius, ColorRed, false)

	DrawArcadeText(target, "THE END", 100, 100, ColorYellow)
	DrawArcadeText(target, "GAME OVER", 92, 120, ColorRed)
}

func (g *Game) drawHighScoreEntry(target *ebiten.Image) {
	DrawArcadeText(target, "GREAT SCORE!", 80, 40, ColorYellow)
	DrawArcadeText(target, fmt.Sprintf("SCORE: %d", g.Score), 76, 60, ColorWhite)
	DrawArcadeText(target, "ENTER YOUR INITIALS", 52, 90, ColorCyan)

	initialsStr := fmt.Sprintf("%c %c %c", g.Initials[0], g.Initials[1], g.Initials[2])
	DrawArcadeText(target, initialsStr, 108, 120, ColorYellow)
	DrawArcadeText(target, "-", 108+g.InitialSlot*16, 130, ColorWhite)

	DrawArcadeText(target, "UP/DOWN TO CHANGE", 60, 160, ColorWhite)
	DrawArcadeText(target, "CLICK/SPACE TO SELECT", 44, 175, ColorWhite)
}
