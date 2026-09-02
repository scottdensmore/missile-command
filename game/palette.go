package game

import "image/color"

// Palette defines the colors used for rendering a specific wave.
type Palette struct {
	SkyColor        color.RGBA
	GroundColor     color.RGBA
	CityColor       color.RGBA
	SiloColor       color.RGBA
	ICBMColor       color.RGBA
	ABMColor        color.RGBA
	TextColor       color.RGBA
	MultiplierColor color.RGBA
	Multiplier      int
}

// WavePalettes contains the 10 authentic Atari 1980 arcade wave color combinations.
// They cycle every 20 waves (2 waves per palette).
var WavePalettes = [10]Palette{
	// Waves 1-2: 1X (Black Sky, Yellow Ground, Blue Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 0, G: 0, B: 0, A: 255},
		GroundColor:     color.RGBA{R: 228, G: 196, B: 36, A: 255},
		CityColor:       color.RGBA{R: 50, G: 120, B: 245, A: 255},
		SiloColor:       color.RGBA{R: 50, G: 120, B: 245, A: 255},
		ICBMColor:       color.RGBA{R: 240, G: 32, B: 32, A: 255},
		ABMColor:        color.RGBA{R: 50, G: 200, B: 255, A: 255},
		TextColor:       color.RGBA{R: 240, G: 240, B: 240, A: 255},
		MultiplierColor: color.RGBA{R: 240, G: 240, B: 240, A: 255},
		Multiplier:      1,
	},
	// Waves 3-4: 2X (Black Sky, Blue Ground, Cyan Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 0, G: 0, B: 0, A: 255},
		GroundColor:     color.RGBA{R: 36, G: 100, B: 240, A: 255},
		CityColor:       color.RGBA{R: 36, G: 224, B: 180, A: 255},
		SiloColor:       color.RGBA{R: 36, G: 224, B: 180, A: 255},
		ICBMColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		ABMColor:        color.RGBA{R: 240, G: 240, B: 240, A: 255},
		TextColor:       color.RGBA{R: 36, G: 224, B: 180, A: 255},
		MultiplierColor: color.RGBA{R: 36, G: 224, B: 180, A: 255},
		Multiplier:      2,
	},
	// Waves 5-6: 3X (Black Sky, Red Ground, Yellow Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 0, G: 0, B: 0, A: 255},
		GroundColor:     color.RGBA{R: 224, G: 36, B: 36, A: 255},
		CityColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		SiloColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		ICBMColor:       color.RGBA{R: 36, G: 224, B: 224, A: 255},
		ABMColor:        color.RGBA{R: 240, G: 220, B: 32, A: 255},
		TextColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		MultiplierColor: color.RGBA{R: 240, G: 220, B: 32, A: 255},
		Multiplier:      3,
	},
	// Waves 7-8: 4X (Black Sky, Red Ground, Cyan Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 0, G: 0, B: 0, A: 255},
		GroundColor:     color.RGBA{R: 224, G: 36, B: 36, A: 255},
		CityColor:       color.RGBA{R: 36, G: 224, B: 224, A: 255},
		SiloColor:       color.RGBA{R: 36, G: 224, B: 224, A: 255},
		ICBMColor:       color.RGBA{R: 240, G: 240, B: 240, A: 255},
		ABMColor:        color.RGBA{R: 36, G: 224, B: 224, A: 255},
		TextColor:       color.RGBA{R: 36, G: 224, B: 224, A: 255},
		MultiplierColor: color.RGBA{R: 36, G: 224, B: 224, A: 255},
		Multiplier:      4,
	},
	// Waves 9-10: 5X (Dark Blue Sky, Yellow Ground, Red Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 16, G: 24, B: 140, A: 255},
		GroundColor:     color.RGBA{R: 228, G: 196, B: 36, A: 255},
		CityColor:       color.RGBA{R: 224, G: 36, B: 36, A: 255},
		SiloColor:       color.RGBA{R: 224, G: 36, B: 36, A: 255},
		ICBMColor:       color.RGBA{R: 240, G: 40, B: 40, A: 255},
		ABMColor:        color.RGBA{R: 240, G: 240, B: 240, A: 255},
		TextColor:       color.RGBA{R: 228, G: 196, B: 36, A: 255},
		MultiplierColor: color.RGBA{R: 228, G: 196, B: 36, A: 255},
		Multiplier:      5,
	},
	// Waves 11-12: 6X (Light Blue Sky, Yellow Ground, Dark Blue Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 60, G: 120, B: 224, A: 255},
		GroundColor:     color.RGBA{R: 228, G: 196, B: 36, A: 255},
		CityColor:       color.RGBA{R: 16, G: 24, B: 140, A: 255},
		SiloColor:       color.RGBA{R: 16, G: 24, B: 140, A: 255},
		ICBMColor:       color.RGBA{R: 20, G: 20, B: 20, A: 255},
		ABMColor:        color.RGBA{R: 240, G: 240, B: 240, A: 255},
		TextColor:       color.RGBA{R: 240, G: 240, B: 240, A: 255},
		MultiplierColor: color.RGBA{R: 240, G: 240, B: 240, A: 255},
		Multiplier:      6,
	},
	// Waves 13-14: 6X (Purple Sky, Green Ground, Yellow Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 140, G: 24, B: 140, A: 255},
		GroundColor:     color.RGBA{R: 36, G: 180, B: 48, A: 255},
		CityColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		SiloColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		ICBMColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		ABMColor:        color.RGBA{R: 36, G: 224, B: 224, A: 255},
		TextColor:       color.RGBA{R: 240, G: 220, B: 32, A: 255},
		MultiplierColor: color.RGBA{R: 240, G: 220, B: 32, A: 255},
		Multiplier:      6,
	},
	// Waves 15-16: 6X (Yellow Sky, Green Ground, Red Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 220, G: 190, B: 24, A: 255},
		GroundColor:     color.RGBA{R: 36, G: 180, B: 48, A: 255},
		CityColor:       color.RGBA{R: 224, G: 36, B: 36, A: 255},
		SiloColor:       color.RGBA{R: 224, G: 36, B: 36, A: 255},
		ICBMColor:       color.RGBA{R: 36, G: 36, B: 200, A: 255},
		ABMColor:        color.RGBA{R: 224, G: 36, B: 36, A: 255},
		TextColor:       color.RGBA{R: 36, G: 36, B: 36, A: 255},
		MultiplierColor: color.RGBA{R: 36, G: 36, B: 36, A: 255},
		Multiplier:      6,
	},
	// Waves 17-18: 6X (White Sky, Red Ground, Blue Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 230, G: 230, B: 230, A: 255},
		GroundColor:     color.RGBA{R: 224, G: 36, B: 36, A: 255},
		CityColor:       color.RGBA{R: 24, G: 48, B: 200, A: 255},
		SiloColor:       color.RGBA{R: 24, G: 48, B: 200, A: 255},
		ICBMColor:       color.RGBA{R: 20, G: 20, B: 20, A: 255},
		ABMColor:        color.RGBA{R: 24, G: 48, B: 200, A: 255},
		TextColor:       color.RGBA{R: 24, G: 48, B: 200, A: 255},
		MultiplierColor: color.RGBA{R: 24, G: 48, B: 200, A: 255},
		Multiplier:      6,
	},
	// Waves 19-20: 6X (Red Sky, Yellow Ground, Black Cities/Ammo)
	{
		SkyColor:        color.RGBA{R: 210, G: 32, B: 32, A: 255},
		GroundColor:     color.RGBA{R: 228, G: 196, B: 36, A: 255},
		CityColor:       color.RGBA{R: 20, G: 20, B: 20, A: 255},
		SiloColor:       color.RGBA{R: 20, G: 20, B: 20, A: 255},
		ICBMColor:       color.RGBA{R: 20, G: 20, B: 20, A: 255},
		ABMColor:        color.RGBA{R: 228, G: 196, B: 36, A: 255},
		TextColor:       color.RGBA{R: 228, G: 196, B: 36, A: 255},
		MultiplierColor: color.RGBA{R: 228, G: 196, B: 36, A: 255},
		Multiplier:      6,
	},
}

// GetPaletteForWave returns the authentic arcade palette for any wave index (1-based).
func GetPaletteForWave(wave int) Palette {
	if wave < 1 {
		wave = 1
	}
	idx := ((wave - 1) / 2) % len(WavePalettes)
	p := WavePalettes[idx]

	// Scoring multiplier progression: increases by 1 every 2 waves, capped at 6x
	mult := 1 + (wave-1)/2
	if mult > 6 {
		mult = 6
	}
	p.Multiplier = mult
	return p
}

// ExplosionColors holds the authentic arcade palette cycling rainbow sequence.
var ExplosionColors = []color.RGBA{
	{R: 255, G: 255, B: 255, A: 255}, // White
	{R: 255, G: 240, B: 40, A: 255},  // Yellow
	{R: 40, G: 240, B: 255, A: 255},  // Cyan
	{R: 255, G: 60, B: 220, A: 255},  // Magenta
	{R: 255, G: 130, B: 20, A: 255},  // Orange
	{R: 50, G: 240, B: 60, A: 255},   // Green
	{R: 240, G: 40, B: 40, A: 255},   // Red
	{R: 80, G: 120, B: 255, A: 255},  // Blue
}

// GetExplosionColor returns a pulsing color from the color-cycling explosion table based on frame tick.
func GetExplosionColor(tick int) color.RGBA {
	idx := (tick / 3) % len(ExplosionColors)
	return ExplosionColors[idx]
}
