package game

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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

// Pipeline manages the 256x231 arcade framebuffer, aspect-ratio scaling, and CRT post-processing.
type Pipeline struct {
	frameBuffer *ebiten.Image
	crtShader   *ebiten.Shader
	UseCRT      bool
	FlashTimer  int
}

// NewPipeline creates a new rendering pipeline.
func NewPipeline() *Pipeline {
	p := &Pipeline{
		frameBuffer: ebiten.NewImage(int(SimWidth), int(SimHeight)),
		UseCRT:      true,
	}

	s, err := ebiten.NewShader(CRTShaderSrc)
	if err != nil {
		log.Printf("Warning: failed to compile CRT shader: %v", err)
	} else {
		p.crtShader = s
	}

	return p
}

// DrawFrame renders the 256x231 buffer onto the destination screen with authentic aspect ratio.
func (p *Pipeline) DrawFrame(screen *ebiten.Image, windowW, windowH int) {
	scale := math.Min(float64(windowW)/SimWidth, float64(windowH)/SimHeight)
	offsetX := (float64(windowW) - SimWidth*scale) / 2.0
	offsetY := (float64(windowH) - SimHeight*scale) / 2.0

	screen.Fill(color.Black)

	flashVal := 0.0
	if p.FlashTimer > 0 {
		flashVal = float64(p.FlashTimer) / 10.0
		p.FlashTimer--
	}

	if p.UseCRT && p.crtShader != nil {
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = p.frameBuffer
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(offsetX, offsetY)
		op.Uniforms = map[string]interface{}{
			"Resolution":     []float32{float32(SimWidth), float32(SimHeight)},
			"FlashIntensity": float32(flashVal),
		}
		screen.DrawRectShader(int(SimWidth), int(SimHeight), p.crtShader, op)
	} else {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(offsetX, offsetY)
		if flashVal > 0 {
			f := 1.0 + float32(flashVal)
			op.ColorScale.Scale(f, f, f, 1.0)
		}
		screen.DrawImage(p.frameBuffer, op)
	}
}

// TriggerFlash triggers a brief bright white screen flash (e.g. city destruction or game over).
func (p *Pipeline) TriggerFlash(frames int) {
	p.FlashTimer = frames
}
