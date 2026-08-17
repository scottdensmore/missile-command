package game

import (
	"math"
)

const (
	SimWidth  = 256.0
	SimHeight = 231.0
)

// Point represents a position in the 2D plane.
type Point struct {
	X, Y float64
}

// Distance returns the Euclidean distance between two points.
func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Lerp performs linear interpolation between this point and another point based on progress t (0.0 to 1.0).
func (p Point) Lerp(other Point, t float64) Point {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return Point{
		X: p.X + (other.X-p.X)*t,
		Y: p.Y + (other.Y-p.Y)*t,
	}
}

// SimToScreen transforms a simulation point (256x231) to screen coordinates based on actual window size.
func (p Point) SimToScreen(screenWidth, screenHeight int) (float32, float32) {
	scaleX := float64(screenWidth) / SimWidth
	scaleY := float64(screenHeight) / SimHeight
	return float32(p.X * scaleX), float32(p.Y * scaleY)
}

// ScreenToSim transforms screen coordinates to the simulation space (256x231).
func ScreenToSim(screenX, screenY float64, screenWidth, screenHeight int) Point {
	return ScreenToSimWithLetterbox(screenX, screenY, screenWidth, screenHeight)
}

// ScreenToSimWithLetterbox converts window pixel coordinates to 256x231 simulation coordinates accounting for aspect-ratio letterboxing.
func ScreenToSimWithLetterbox(screenX, screenY float64, windowW, windowH int) Point {
	scaleX := float64(windowW) / SimWidth
	scaleY := float64(windowH) / SimHeight
	scale := math.Min(scaleX, scaleY)
	if scale <= 0 {
		return Point{X: 128, Y: 115}
	}

	renderW := SimWidth * scale
	renderH := SimHeight * scale
	offsetX := (float64(windowW) - renderW) / 2.0
	offsetY := (float64(windowH) - renderH) / 2.0

	simX := (screenX - offsetX) / scale
	simY := (screenY - offsetY) / scale

	return Point{
		X: simX,
		Y: simY,
	}
}
