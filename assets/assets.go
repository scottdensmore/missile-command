package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"log"
)

// Embedded multi-resolution PNG icons for runtime desktop window/taskbar/dock icons.
//
//go:embed icons/icon-16.png
var Icon16PNG []byte

//go:embed icons/icon-32.png
var Icon32PNG []byte

//go:embed icons/icon-48.png
var Icon48PNG []byte

//go:embed icons/icon-64.png
var Icon64PNG []byte

//go:embed icons/icon-128.png
var Icon128PNG []byte

// GetWindowIcons decodes and returns multi-resolution window icons suitable
// for ebiten.SetWindowIcon().
func GetWindowIcons() []image.Image {
	raws := [][]byte{Icon16PNG, Icon32PNG, Icon48PNG, Icon64PNG, Icon128PNG}
	var icons []image.Image
	for _, raw := range raws {
		img, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			log.Printf("warning: failed to decode embedded icon: %v", err)
			continue
		}
		icons = append(icons, img)
	}
	return icons
}
