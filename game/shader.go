package game

// CRTShaderSrc provides subtle, authentic arcade CRT scanline and phosphor bloom effects.
var CRTShaderSrc = []byte(`
package main

var Resolution vec2
var FlashIntensity float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	orig := imageSrc0At(texCoord)
	if FlashIntensity > 0.0 {
		orig = mix(orig, vec4(1.0, 1.0, 1.0, 1.0), FlashIntensity)
	}

	// Subtle CRT scanline modulation
	scanline := sin(texCoord.y * Resolution.y * 3.14159) * 0.08
	col := orig.rgb - vec3(scanline)

	// Slight phosphor saturation boost
	col = col * 1.05

	return vec4(col, orig.a)
}
`)
