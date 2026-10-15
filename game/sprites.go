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
var CitySprite = [][]byte{
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
var DestroyedCitySprite = [][]byte{
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
var BomberSprite = [][]byte{
	{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0},
	{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
}

// SatelliteSprite is the authentic arcade satellite (14x8).
var SatelliteSprite = [][]byte{
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
var SmartBombFrames = [2][][]byte{
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
var BonusCityIcon = [][]byte{
	{0, 1, 0, 0, 1, 0, 0, 0},
	{1, 1, 0, 1, 1, 0, 1, 0},
	{1, 1, 1, 1, 1, 0, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1},
	{1, 1, 1, 1, 1, 1, 1, 1},
}

func setSimPixel(target *ebiten.Image, px, py int, col color.RGBA) {
	if px >= 0 && px < int(SimWidth) && py >= 0 && py < int(SimHeight) {
		target.Set(px, py, col)
	}
}

func drawGrid(target *ebiten.Image, grid [][]byte, x, y int, flipX bool, col color.RGBA) {
	for r, row := range grid {
		w := len(row)
		for c := 0; c < w; c++ {
			srcCol := c
			if flipX {
				srcCol = w - 1 - c
			}
			if row[srcCol] != 0 {
				setSimPixel(target, x+c, y+r, col)
			}
		}
	}
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

func drawBitmap8x8(target *ebiten.Image, bmp Bitmap8x8, x, y int, col color.RGBA) {
	for row := 0; row < 8; row++ {
		b := bmp[row]
		for colIdx := 0; colIdx < 8; colIdx++ {
			if (b & (0x80 >> colIdx)) != 0 {
				setSimPixel(target, x+colIdx, y+row, col)
			}
		}
	}
}

// DrawCity renders a city sprite (intact or destroyed) at (x, y).
func DrawCity(target *ebiten.Image, x, y int, destroyed bool, col color.RGBA) {
	if destroyed {
		drawGrid(target, DestroyedCitySprite, x, y, false, col)
	} else {
		drawGrid(target, CitySprite, x, y, false, col)
	}
}

// DrawBomber renders the bomber plane at (x, y), flipped horizontally if moving left.
func DrawBomber(target *ebiten.Image, x, y int, movingRight bool, col color.RGBA) {
	drawGrid(target, BomberSprite, x, y, !movingRight, col)
}

// DrawSatellite renders the satellite at (x, y).
func DrawSatellite(target *ebiten.Image, x, y int, col color.RGBA) {
	drawGrid(target, SatelliteSprite, x, y, false, col)
}

// DrawSmartBomb renders the animated spinning smart bomb at (x, y).
func DrawSmartBomb(target *ebiten.Image, x, y int, frame int, col color.RGBA) {
	drawGrid(target, SmartBombFrames[frame%2], x, y, false, col)
}

// DrawBonusCityIcon renders a mini bonus city indicator icon in the lower margin.
func DrawBonusCityIcon(target *ebiten.Image, x, y int, col color.RGBA) {
	drawGrid(target, BonusCityIcon, x, y, false, col)
}

// DrawAmmoPyramid draws the 10-missile pyramid (4, 3, 2, 1) for a battery.
func DrawAmmoPyramid(target *ebiten.Image, centerX, baseY int, ammo int, col color.RGBA) {
	positions := [10][2]int{
		{centerX - 6, baseY}, {centerX - 2, baseY}, {centerX + 2, baseY}, {centerX + 6, baseY},
		{centerX - 4, baseY - 4}, {centerX, baseY - 4}, {centerX + 4, baseY - 4},
		{centerX - 2, baseY - 8}, {centerX + 2, baseY - 8},
		{centerX, baseY - 12},
	}

	for i := 0; i < ammo && i < len(positions); i++ {
		mx, my := positions[i][0], positions[i][1]
		for ox := 0; ox < 2; ox++ {
			for oy := 0; oy < 3; oy++ {
				setSimPixel(target, mx+ox, my+oy, col)
			}
		}
	}
}

// DrawSiloTicks draws the missile ammo in the silo:
// When ammo <= 10, it renders the classic 10-missile pyramid.
// When ammo > 10, it renders the numerical count inside the silo mound.
func DrawSiloTicks(target *ebiten.Image, bx, baseY, ammo, maxAmmo int, col, groundCol color.RGBA) {
	if ammo <= 0 {
		return
	}
	if ammo <= 10 {
		DrawAmmoPyramid(target, bx, 220, ammo, col)
		return
	}

	ammoStr := fmt.Sprintf("%d", ammo)
	vector.DrawFilledRect(target, float32(bx-13), 212, 26, 12, groundCol, false)
	DrawArcadeText(target, ammoStr, bx-len(ammoStr)*4, 214, col)
}
