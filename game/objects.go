package game

import "math"

type Crosshair struct {
	Pos Point
}

type ICBM struct {
	Start, Curr, Target Point
	Speed               float64
	Progress            float64 // 0.0 to 1.0
	IsSplitter          bool
	Splitted            bool
	SplitAltitude       float64
	Active              bool
}

type SmartBomb struct {
	Start, Curr, Target Point
	Speed               float64
	Progress            float64
	EvasionVelocity     Point
	Active              bool
	AnimTick            int
}

type FlierType int

const (
	FlierNone FlierType = iota
	FlierBomber
	FlierSatellite
)

type Flier struct {
	Type            FlierType
	X, Y            float64
	Speed           float64 // positive: moves right, negative: moves left
	DropCooldown    int
	BombsRemaining  int
	Active          bool
	AnimTick        int
}

type ABM struct {
	Start, Curr, Target Point
	SiloIndex           int     // 0: Left, 1: Center (Fast), 2: Right
	Speed               float64 // 3.2 for side silos, 7.5 for center silo
	Progress            float64
	Active              bool
}

type ExplosionState int

const (
	StateExpanding ExplosionState = iota
	StateHolding
	StateContracting
	StateDead
)

type Explosion struct {
	Center    Point
	Radius    float64
	MaxRadius float64
	HoldTimer int
	State     ExplosionState
	Tick      int
}

type City struct {
	Position     Point
	Destroyed    bool
	Rebuilding   bool
	RebuildTimer int
}

type Battery struct {
	Index             int // 0: Left, 1: Center, 2: Right
	Position          Point
	MaxAmmo           int
	Ammo              int
	Destroyed         bool
	LowWarningPlayed  bool
	CanFireCooldown   int
}

// WaveData defines enemy assault parameters for a specific wave.
type WaveData struct {
	TotalICBMs  int
	Speed       float64
	SmartBombs  int
	FlierDelay  int
}

// ArcadeWaveTable matches the authentic 1980 Atari arcade missile command progression.
// Speeds are derived directly from the arcade ROM ICBM frame delay table: speed = 1 / (1 + delay).
// Wave 15+ reaches 0 delay (updates every frame, 1.00 px/frame at 60 FPS).
var ArcadeWaveTable = []WaveData{
	{TotalICBMs: 12, Speed: 0.172, SmartBombs: 0, FlierDelay: 320}, // Wave 1 (delay 4.8125)
	{TotalICBMs: 15, Speed: 0.258, SmartBombs: 0, FlierDelay: 280}, // Wave 2 (delay 2.875)
	{TotalICBMs: 18, Speed: 0.364, SmartBombs: 0, FlierDelay: 260}, // Wave 3 (delay 1.75)
	{TotalICBMs: 12, Speed: 0.492, SmartBombs: 0, FlierDelay: 240}, // Wave 4 (delay 1.031)
	{TotalICBMs: 16, Speed: 0.615, SmartBombs: 0, FlierDelay: 220}, // Wave 5 (delay 0.625)
	{TotalICBMs: 14, Speed: 0.727, SmartBombs: 1, FlierDelay: 200}, // Wave 6 (delay 0.375)
	{TotalICBMs: 17, Speed: 0.800, SmartBombs: 1, FlierDelay: 190}, // Wave 7 (delay 0.25)
	{TotalICBMs: 10, Speed: 0.889, SmartBombs: 2, FlierDelay: 180}, // Wave 8 (delay 0.125)
	{TotalICBMs: 13, Speed: 0.941, SmartBombs: 3, FlierDelay: 170}, // Wave 9 (delay 0.0625)
	{TotalICBMs: 16, Speed: 0.962, SmartBombs: 4, FlierDelay: 160}, // Wave 10 (delay 0.04)
	{TotalICBMs: 19, Speed: 0.980, SmartBombs: 4, FlierDelay: 150}, // Wave 11 (delay 0.02)
	{TotalICBMs: 12, Speed: 0.984, SmartBombs: 5, FlierDelay: 140}, // Wave 12 (delay 0.016)
	{TotalICBMs: 14, Speed: 0.992, SmartBombs: 5, FlierDelay: 130}, // Wave 13 (delay 0.008)
	{TotalICBMs: 16, Speed: 0.996, SmartBombs: 6, FlierDelay: 120}, // Wave 14 (delay 0.004)
	{TotalICBMs: 18, Speed: 1.000, SmartBombs: 6, FlierDelay: 110}, // Wave 15 (delay 0.0)
	{TotalICBMs: 14, Speed: 1.000, SmartBombs: 7, FlierDelay: 100}, // Wave 16
	{TotalICBMs: 17, Speed: 1.000, SmartBombs: 7, FlierDelay: 100}, // Wave 17
	{TotalICBMs: 19, Speed: 1.000, SmartBombs: 7, FlierDelay: 90},  // Wave 18
	{TotalICBMs: 22, Speed: 1.000, SmartBombs: 7, FlierDelay: 80},  // Wave 19+
}

// GetWaveData returns the WaveData configuration for any wave index (1-based).
func GetWaveData(wave int) WaveData {
	if wave < 1 {
		wave = 1
	}
	idx := wave - 1
	if idx >= len(ArcadeWaveTable) {
		idx = len(ArcadeWaveTable) - 1
	}
	return ArcadeWaveTable[idx]
}

// DistanceToPoint calculates Euclidean distance.
func DistanceToPoint(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}
