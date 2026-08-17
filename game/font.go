package game

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Stroke represents a sequence of relative points that should be drawn connected.
type Stroke []Point

// StrokeCharacter is a slice of Strokes that define a single symbol.
type StrokeCharacter []Stroke

var fontMap = map[rune]StrokeCharacter{
	'0': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 5}, {X: -3, Y: 5}, {X: -3, Y: -5}},
		{{X: -3, Y: 5}, {X: 3, Y: -5}}, // slash
	},
	'1': {
		{{X: 0, Y: -5}, {X: 0, Y: 5}},
		{{X: -1.5, Y: -3.5}, {X: 0, Y: -5}},
		{{X: -2, Y: 5}, {X: 2, Y: 5}},
	},
	'2': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 0}, {X: -3, Y: 0}, {X: -3, Y: 5}, {X: 3, Y: 5}},
	},
	'3': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 5}, {X: -3, Y: 5}},
		{{X: -3, Y: 0}, {X: 3, Y: 0}},
	},
	'4': {
		{{X: -3, Y: -5}, {X: -3, Y: 0}, {X: 3, Y: 0}},
		{{X: 3, Y: -5}, {X: 3, Y: 5}},
	},
	'5': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 0}, {X: 3, Y: 0}, {X: 3, Y: 5}, {X: -3, Y: 5}},
	},
	'6': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: 0}, {X: -3, Y: 0}},
	},
	'7': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: -1, Y: 5}},
	},
	'8': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 5}, {X: -3, Y: 5}, {X: -3, Y: -5}},
		{{X: -3, Y: 0}, {X: 3, Y: 0}},
	},
	'9': {
		{{X: -3, Y: 0}, {X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 5}, {X: -3, Y: 5}},
		{{X: -3, Y: 0}, {X: 3, Y: 0}},
	},
	'A': {
		{{X: -3, Y: 5}, {X: -3, Y: -1}, {X: 0, Y: -5}, {X: 3, Y: -1}, {X: 3, Y: 5}},
		{{X: -3, Y: 0}, {X: 3, Y: 0}},
	},
	'B': {
		{{X: -3, Y: -5}, {X: 1, Y: -5}, {X: 3, Y: -2.5}, {X: 1, Y: 0}, {X: 3, Y: 2.5}, {X: 1, Y: 5}, {X: -3, Y: 5}, {X: -3, Y: -5}},
		{{X: -3, Y: 0}, {X: 1, Y: 0}},
	},
	'C': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}},
	},
	'D': {
		{{X: -3, Y: -5}, {X: 1, Y: -5}, {X: 3, Y: -2.5}, {X: 3, Y: 2.5}, {X: 1, Y: 5}, {X: -3, Y: 5}, {X: -3, Y: -5}},
	},
	'E': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}},
		{{X: -3, Y: 0}, {X: 1, Y: 0}},
	},
	'F': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 5}},
		{{X: -3, Y: 0}, {X: 1, Y: 0}},
	},
	'G': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: 0}, {X: 0, Y: 0}},
	},
	'H': {
		{{X: -3, Y: -5}, {X: -3, Y: 5}},
		{{X: 3, Y: -5}, {X: 3, Y: 5}},
		{{X: -3, Y: 0}, {X: 3, Y: 0}},
	},
	'I': {
		{{X: 0, Y: -5}, {X: 0, Y: 5}},
		{{X: -2, Y: -5}, {X: 2, Y: -5}},
		{{X: -2, Y: 5}, {X: 2, Y: 5}},
	},
	'J': {
		{{X: -3, Y: 2}, {X: -3, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: -5}},
	},
	'K': {
		{{X: -3, Y: -5}, {X: -3, Y: 5}},
		{{X: 3, Y: -5}, {X: -3, Y: 0}, {X: 3, Y: 5}},
	},
	'L': {
		{{X: -3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}},
	},
	'M': {
		{{X: -3, Y: 5}, {X: -3, Y: -5}, {X: 0, Y: -1}, {X: 3, Y: -5}, {X: 3, Y: 5}},
	},
	'N': {
		{{X: -3, Y: 5}, {X: -3, Y: -5}, {X: 3, Y: 5}, {X: 3, Y: -5}},
	},
	'O': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 5}, {X: -3, Y: 5}, {X: -3, Y: -5}},
	},
	'P': {
		{{X: -3, Y: 5}, {X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 0}, {X: -3, Y: 0}},
	},
	'Q': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 5}, {X: -3, Y: 5}, {X: -3, Y: -5}},
		{{X: 1, Y: 1}, {X: 3.5, Y: 5}},
	},
	'R': {
		{{X: -3, Y: 5}, {X: -3, Y: -5}, {X: 3, Y: -5}, {X: 3, Y: 0}, {X: -3, Y: 0}},
		{{X: 0, Y: 0}, {X: 3, Y: 5}},
	},
	'S': {
		{{X: 3, Y: -5}, {X: -3, Y: -5}, {X: -3, Y: 0}, {X: 3, Y: 0}, {X: 3, Y: 5}, {X: -3, Y: 5}},
	},
	'T': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}},
		{{X: 0, Y: -5}, {X: 0, Y: 5}},
	},
	'U': {
		{{X: -3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}, {X: 3, Y: -5}},
	},
	'V': {
		{{X: -3, Y: -5}, {X: 0, Y: 5}, {X: 3, Y: -5}},
	},
	'W': {
		{{X: -3, Y: -5}, {X: -3, Y: 5}, {X: 0, Y: 1}, {X: 3, Y: 5}, {X: 3, Y: -5}},
	},
	'X': {
		{{X: -3, Y: -5}, {X: 3, Y: 5}},
		{{X: -3, Y: 5}, {X: 3, Y: -5}},
	},
	'Y': {
		{{X: -3, Y: -5}, {X: 0, Y: 0}, {X: 3, Y: -5}},
		{{X: 0, Y: 0}, {X: 0, Y: 5}},
	},
	'Z': {
		{{X: -3, Y: -5}, {X: 3, Y: -5}, {X: -3, Y: 5}, {X: 3, Y: 5}},
	},
	'-': {
		{{X: -2, Y: 0}, {X: 2, Y: 0}},
	},
	'+': {
		{{X: -2, Y: 0}, {X: 2, Y: 0}},
		{{X: 0, Y: -2}, {X: 0, Y: 2}},
	},
	':': {
		{{X: 0, Y: -2.5}, {X: 0, Y: -1.5}},
		{{X: 0, Y: 1.5}, {X: 0, Y: 2.5}},
	},
	'.': {
		{{X: -0.5, Y: 4.5}, {X: 0.5, Y: 4.5}, {X: 0.5, Y: 5.5}, {X: -0.5, Y: 5.5}, {X: -0.5, Y: 4.5}},
	},
}

// DrawStrokeText renders beautiful retro glowing vector text character-by-character.
func DrawStrokeText(target *ebiten.Image, g *Game, text string, x, y float64, scale float64, col color.Color) {
	upperText := strings.ToUpper(text)
	currX := x

	// Character cell width is 6 * scale, spacing is 4 * scale
	charWidth := 7.0 * scale
	spacing := 3.0 * scale

	for _, r := range upperText {
		if r == ' ' {
			currX += charWidth + spacing
			continue
		}

		char, ok := fontMap[r]
		if ok {
			// Draw each stroke of the character
			for _, stroke := range char {
				for i := 0; i < len(stroke)-1; i++ {
					p1 := Point{X: currX + stroke[i].X*scale, Y: y + stroke[i].Y*scale}
					p2 := Point{X: currX + stroke[i+1].X*scale, Y: y + stroke[i+1].Y*scale}

					x1, y1 := p1.SimToScreen(g.Width, g.Height)
					x2, y2 := p2.SimToScreen(g.Width, g.Height)

					// Draw stroke using high-quality line vector rendering
					vector.StrokeLine(target, x1, y1, x2, y2, float32(1.2*scale), col, true)
				}
			}
		}
		currX += charWidth + spacing
	}
}
