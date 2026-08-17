package game

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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
	// Calculate letterbox / pillarbox scaling
	scaleX := float64(windowW) / SimWidth
	scaleY := float64(windowH) / SimHeight
	scale := math.Min(scaleX, scaleY)

	// Keep integer scale if close to integer, else float scale
	if scale > 1.0 {
		// e.g. 1024x924 -> scale = 4.0
	}

	renderW := SimWidth * scale
	renderH := SimHeight * scale
	offsetX := (float64(windowW) - renderW) / 2.0
	offsetY := (float64(windowH) - renderH) / 2.0

	// Clear full window to black
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
			op.ColorScale.Scale(1.0+float32(flashVal), 1.0+float32(flashVal), 1.0+float32(flashVal), 1.0)
		}
		screen.DrawImage(p.frameBuffer, op)
	}
}

// TriggerFlash triggers a brief bright white screen flash (e.g. city destruction or game over).
func (p *Pipeline) TriggerFlash(frames int) {
	p.FlashTimer = frames
}
