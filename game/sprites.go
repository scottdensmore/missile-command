package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Bitmap8x8 represents an 8x8 single-bit pixel matrix.
type Bitmap8x8 [8]byte

// ArcadeFont8x8 contains the 1980 Atari arcade 8x8 character set.
var ArcadeFont8x8 = map[rune]Bitmap8x8{
	'0': {0x3C, 0x66, 0x6E, 0x76, 0x66, 0x66, 0x3C, 0x00},
	'1': {0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00},
	'2': {0x3C, 0x66, 0x06, 0x0C, 0x18, 0x30, 0x7E, 0x00},
	'3': {0x3C, 0x66, 0x06, 0x1C, 0x06, 0x66, 0x3C, 0x00},
	'4': {0x0C, 0x1C, 0x34, 0x64, 0x7E, 0x04, 0x04, 0x00},
	'5': {0x7E, 0x60, 0x7C, 0x06, 0x06, 0x66, 0x3C, 0x00},
	'6': {0x1C, 0x30, 0x60, 0x7C, 0x66, 0x66, 0x3C, 0x00},
	'7': {0x7E, 0x06, 0x0C, 0x18, 0x18, 0x18, 0x18, 0x00},
	'8': {0x3C, 0x66, 0x66, 0x3C, 0x66, 0x66, 0x3C, 0x00},
	'9': {0x3C, 0x66, 0x66, 0x3E, 0x06, 0x0C, 0x38, 0x00},

	'A': {0x18, 0x3C, 0x66, 0x66, 0x7E, 0x66, 0x66, 0x00},
	'B': {0x7C, 0x66, 0x66, 0x7C, 0x66, 0x66, 0x7C, 0x00},
	'C': {0x3C, 0x66, 0x60, 0x60, 0x60, 0x66, 0x3C, 0x00},
	'D': {0x78, 0x6C, 0x66, 0x66, 0x66, 0x6C, 0x78, 0x00},
	'E': {0x7E, 0x60, 0x60, 0x7C, 0x60, 0x60, 0x7E, 0x00},
	'F': {0x7E, 0x60, 0x60, 0x7C, 0x60, 0x60, 0x60, 0x00},
	'G': {0x3C, 0x66, 0x60, 0x6E, 0x66, 0x66, 0x3C, 0x00},
	'H': {0x66, 0x66, 0x66, 0x7E, 0x66, 0x66, 0x66, 0x00},
	'I': {0x3C, 0x18, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00},
	'J': {0x0E, 0x06, 0x06, 0x06, 0x06, 0x66, 0x3C, 0x00},
	'K': {0x66, 0x6C, 0x78, 0x70, 0x78, 0x6C, 0x66, 0x00},
	'L': {0x60, 0x60, 0x60, 0x60, 0x60, 0x60, 0x7E, 0x00},
	'M': {0x63, 0x77, 0x7F, 0x6B, 0x63, 0x63, 0x63, 0x00},
	'N': {0x66, 0x76, 0x7E, 0x7E, 0x6E, 0x66, 0x66, 0x00},
	'O': {0x3C, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00},
	'P': {0x7C, 0x66, 0x66, 0x7C, 0x60, 0x60, 0x60, 0x00},
	'Q': {0x3C, 0x66, 0x66, 0x66, 0x6A, 0x64, 0x3A, 0x00},
	'R': {0x7C, 0x66, 0x66, 0x7C, 0x78, 0x6C, 0x66, 0x00},
	'S': {0x3C, 0x66, 0x60, 0x3C, 0x06, 0x66, 0x3C, 0x00},
	'T': {0x7E, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18, 0x00},
	'U': {0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x00},
	'V': {0x66, 0x66, 0x66, 0x66, 0x66, 0x3C, 0x18, 0x00},
	'W': {0x63, 0x63, 0x63, 0x6B, 0x7F, 0x77, 0x63, 0x00},
	'X': {0x66, 0x66, 0x3C, 0x18, 0x3C, 0x66, 0x66, 0x00},
	'Y': {0x66, 0x66, 0x66, 0x3C, 0x18, 0x18, 0x18, 0x00},
	'Z': {0x7E, 0x06, 0x0C, 0x18, 0x30, 0x60, 0x7E, 0x00},

	' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	':': {0x00, 0x18, 0x18, 0x00, 0x18, 0x18, 0x00, 0x00},
	'-': {0x00, 0x00, 0x00, 0x7E, 0x00, 0x00, 0x00, 0x00},
	'+': {0x00, 0x18, 0x18, 0x7E, 0x18, 0x18, 0x00, 0x00},
	'.': {0x00, 0x00, 0x00, 0x00, 0x00, 0x18, 0x18, 0x00},
	'!': {0x18, 0x18, 0x18, 0x18, 0x00, 0x18, 0x18, 0x00},
	'?': {0x3C, 0x66, 0x06, 0x1C, 0x18, 0x00, 0x18, 0x00},
	'/': {0x06, 0x0C, 0x18, 0x30, 0x60, 0x40, 0x00, 0x00},
	'*': {0x00, 0x66, 0x3C, 0xFF, 0x3C, 0x66, 0x00, 0x00},
	'©': {0x3C, 0x42, 0x99, 0xA1, 0x99, 0x42, 0x3C, 0x00},
}

// CitySprite is the intact 16x8 arcade city silhouette.
var CitySprite = [8][16]byte{
	{0, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0},
	{0, 0, 1, 1, 0, 0, 1, 1, 0, 1, 1, 0, 0, 1, 0, 0},
	{1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 1, 1, 0, 1, 1, 0},
	{1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 1, 1, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

// DestroyedCitySprite is the rubble/crater when a city is destroyed.
var DestroyedCitySprite = [8][16]byte{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0},
	{0, 1, 1, 0, 1, 0, 1, 1, 0, 1, 0, 1, 1, 0, 1, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

// BomberSprite is the authentic arcade bomber airplane (16x6).
var BomberSprite = [6][16]byte{
	{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
}

// SatelliteSprite is the authentic arcade satellite (14x8).
var SatelliteSprite = [8][14]byte{
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0},
	{1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1},
	{1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1},
	{1, 1, 1, 0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1},
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0},
}

// SmartBombFrames holds the 2 rotating frames of the smart bomb (8x8 diamond/star).
var SmartBombFrames = [2][8][8]byte{
	{
		{0, 0, 0, 1, 1, 0, 0, 0},
		{0, 0, 1, 1, 1, 1, 0, 0},
		{0, 1, 1, 0, 0, 1, 1, 0},
		{1, 1, 0, 1, 1, 0, 1, 1},
		{1, 1, 0, 1, 1, 0, 1, 1},
		{0, 1, 1, 0, 0, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 0, 0},
		{0, 0, 0, 1, 1, 0, 0, 0},
	},
	{
		{0, 1, 0, 0, 0, 0, 1, 0},
		{1, 1, 1, 0, 0, 1, 1, 1},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{0, 0, 1, 1, 1, 1, 0, 0},
		{0, 0, 1, 1, 1, 1, 0, 0},
		{0, 1, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 0, 0, 1, 1, 1},
		{0, 1, 0, 0, 0, 0, 1, 0},
	},
}

// BonusCityIcon is the mini city icon drawn in the reserve area (8x5).
var BonusCityIcon = [5][8]byte{
	{0, 1, 0, 0, 1, 0, 0, 0},
	{1, 1, 0, 1, 1, 0, 1, 0},
	{1, 1, 1, 1, 1, 0, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1},
}

// DrawArcadeText renders string text on the target bitmap using the 8x8 font.
func DrawArcadeText(target *ebiten.Image, text string, x, y int, col color.RGBA) {
	for i, r := range text {
		bmp, ok := ArcadeFont8x8[r]
		if !ok {
			bmp, ok = ArcadeFont8x8['?']
		}
		if ok {
			drawBitmap8x8(target, bmp, x+i*8, y, col)
		}
	}
}

// drawBitmap8x8 draws an 8x8 single-bit bitmap at (x, y) with color col.
func drawBitmap8x8(target *ebiten.Image, bmp Bitmap8x8, x, y int, col color.RGBA) {
	for row := 0; row < 8; row++ {
		b := bmp[row]
		for colIdx := 0; colIdx < 8; colIdx++ {
			if (b & (0x80 >> colIdx)) != 0 {
				px := x + colIdx
				py := y + row
				if px >= 0 && px < 256 && py >= 0 && py < 231 {
					target.Set(px, py, col)
				}
			}
		}
	}
}

// DrawCity renders a city sprite (intact or destroyed) at (x, y).
func DrawCity(target *ebiten.Image, x, y int, destroyed bool, col color.RGBA) {
	if destroyed {
		for row := 0; row < 8; row++ {
			for colIdx := 0; colIdx < 16; colIdx++ {
				if DestroyedCitySprite[row][colIdx] != 0 {
					px := x + colIdx
					py := y + row
					if px >= 0 && px < 256 && py >= 0 && py < 231 {
						target.Set(px, py, col)
					}
				}
			}
		}
	} else {
		for row := 0; row < 8; row++ {
			for colIdx := 0; colIdx < 16; colIdx++ {
				if CitySprite[row][colIdx] != 0 {
					px := x + colIdx
					py := y + row
					if px >= 0 && px < 256 && py >= 0 && py < 231 {
						target.Set(px, py, col)
					}
				}
			}
		}
	}
}

// DrawBomber renders the bomber plane at (x, y), flipped horizontally if moving left.
func DrawBomber(target *ebiten.Image, x, y int, movingRight bool, col color.RGBA) {
	for row := 0; row < 6; row++ {
		for colIdx := 0; colIdx < 16; colIdx++ {
			srcCol := colIdx
			if !movingRight {
				srcCol = 15 - colIdx
			}
			if BomberSprite[row][srcCol] != 0 {
				px := x + colIdx
				py := y + row
				if px >= 0 && px < 256 && py >= 0 && py < 231 {
					target.Set(px, py, col)
				}
			}
		}
	}
}

// DrawSatellite renders the satellite at (x, y).
func DrawSatellite(target *ebiten.Image, x, y int, col color.RGBA) {
	for row := 0; row < 8; row++ {
		for colIdx := 0; colIdx < 14; colIdx++ {
			if SatelliteSprite[row][colIdx] != 0 {
				px := x + colIdx
				py := y + row
				if px >= 0 && px < 256 && py >= 0 && py < 231 {
					target.Set(px, py, col)
				}
			}
		}
	}
}

// DrawSmartBomb renders the animated spinning smart bomb at (x, y).
func DrawSmartBomb(target *ebiten.Image, x, y int, frame int, col color.RGBA) {
	f := frame % 2
	for row := 0; row < 8; row++ {
		for colIdx := 0; colIdx < 8; colIdx++ {
			if SmartBombFrames[f][row][colIdx] != 0 {
				px := x + colIdx
				py := y + row
				if px >= 0 && px < 256 && py >= 0 && py < 231 {
					target.Set(px, py, col)
				}
			}
		}
	}
}

// DrawBonusCityIcon renders a mini bonus city indicator icon in the lower margin.
func DrawBonusCityIcon(target *ebiten.Image, x, y int, col color.RGBA) {
	for row := 0; row < 5; row++ {
		for colIdx := 0; colIdx < 8; colIdx++ {
			if BonusCityIcon[row][colIdx] != 0 {
				px := x + colIdx
				py := y + row
				if px >= 0 && px < 256 && py >= 0 && py < 231 {
					target.Set(px, py, col)
				}
			}
		}
	}
}

// DrawAmmoPyramid draws the 10-missile pyramid (4, 3, 2, 1) for a battery.
// centerX is the center pixel column of the battery, baseY is the bottom row Y.
func DrawAmmoPyramid(target *ebiten.Image, centerX, baseY int, ammo int, col color.RGBA) {
	type missilePos struct {
		x, y int
	}
	// 10 missile positions ordered from bottom tier to top apex:
	// Row 0 (bottom 4): Y = baseY
	// Row 1 (middle 3): Y = baseY - 4
	// Row 2 (upper 2):  Y = baseY - 8
	// Row 3 (apex 1):   Y = baseY - 12
	positions := []missilePos{
		// Row 0: Bottom 4 missiles (indices 0..3)
		{centerX - 6, baseY},
		{centerX - 2, baseY},
		{centerX + 2, baseY},
		{centerX + 6, baseY},
		// Row 1: Middle 3 missiles (indices 4..6)
		{centerX - 4, baseY - 4},
		{centerX, baseY - 4},
		{centerX + 4, baseY - 4},
		// Row 2: Upper 2 missiles (indices 7..8)
		{centerX - 2, baseY - 8},
		{centerX + 2, baseY - 8},
		// Row 3: Apex 1 missile (index 9)
		{centerX, baseY - 12},
	}

	for i := 0; i < ammo && i < len(positions); i++ {
		mx := positions[i].x
		my := positions[i].y
		// Draw 2x3 missile sprite (needle)
		for ox := 0; ox < 2; ox++ {
			for oy := 0; oy < 3; oy++ {
				target.Set(mx+ox, my+oy, col)
			}
		}
	}
}

// CalculateSiloNeedles computes how many pyramid needle sprites (0 to 10) to draw for a given ammo and maxAmmo.
func CalculateSiloNeedles(ammo, maxAmmo int) int {
	if ammo <= 0 || maxAmmo <= 0 {
		return 0
	}
	needles := (ammo*10 + maxAmmo - 1) / maxAmmo
	if needles > 10 {
		needles = 10
	}
	return needles
}

// DrawSiloTicks draws the missile ammo in the silo:
// When ammo <= 10 (the number of ticks that can fit in a silo), it renders the ammo as ticks in the classic pyramid.
// When ammo > 10 (above the number of ticks that can fit in a silo), it renders the number inside the silo mound.
func DrawSiloTicks(target *ebiten.Image, bx, baseY, ammo, maxAmmo int, col, groundCol color.RGBA) {
	if ammo <= 0 {
		return
	}
	if ammo <= 10 {
		DrawAmmoPyramid(target, bx, 220, ammo, col)
		return
	}

	// When ammo > 10, render the number inside the silo mound
	ammoStr := fmt.Sprintf("%d", ammo)
	// Fill the mound behind the number with groundCol so digits are cleanly framed
	vector.DrawFilledRect(target, float32(bx-13), 212, 26, 12, groundCol, false)
	DrawArcadeText(target, ammoStr, bx-len(ammoStr)*4, 214, col)
}




