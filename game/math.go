package game

import "math"

const (
	SimWidth  = 256.0
	SimHeight = 231.0
)

// Point represents a position in the 2D simulation plane.
type Point struct {
	X, Y float64
}

// Distance returns the Euclidean distance between two points.
func (p Point) Distance(other Point) float64 {
	return math.Hypot(p.X-other.X, p.Y-other.Y)
}

// Lerp performs linear interpolation between two points for t in [0.0, 1.0].
func (p Point) Lerp(other Point, t float64) Point {
	t = clamp(t, 0, 1)
	return Point{
		X: p.X + (other.X-p.X)*t,
		Y: p.Y + (other.Y-p.Y)*t,
	}
}

// ScreenToSim converts window pixel coordinates to 256x231 simulation coordinates,
// accounting for aspect-ratio letterboxing.
func ScreenToSim(screenX, screenY float64, windowW, windowH int) Point {
	scale := math.Min(float64(windowW)/SimWidth, float64(windowH)/SimHeight)
	if scale <= 0 {
		return Point{X: SimWidth / 2, Y: SimHeight / 2}
	}

	offsetX := (float64(windowW) - SimWidth*scale) / 2.0
	offsetY := (float64(windowH) - SimHeight*scale) / 2.0

	return Point{
		X: (screenX - offsetX) / scale,
		Y: (screenY - offsetY) / scale,
	}
}

func clamp(v, minVal, maxVal float64) float64 {
	return math.Max(minVal, math.Min(maxVal, v))
}
