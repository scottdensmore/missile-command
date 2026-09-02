package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
)

// resizeImage resamples src to width x height using area averaging / box filtering
func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := src.Bounds()
	srcW := float64(srcBounds.Dx())
	srcH := float64(srcBounds.Dy())

	xScale := srcW / float64(width)
	yScale := srcH / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX0 := float64(x) * xScale
			srcX1 := float64(x+1) * xScale
			srcY0 := float64(y) * yScale
			srcY1 := float64(y+1) * yScale

			var totR, totG, totB, totA, totWeight float64

			minX := int(math.Floor(srcX0))
			maxX := int(math.Ceil(srcX1))
			minY := int(math.Floor(srcY0))
			maxY := int(math.Ceil(srcY1))

			for sy := minY; sy < maxY; sy++ {
				for sx := minX; sx < maxX; sx++ {
					if sx < srcBounds.Min.X || sx >= srcBounds.Max.X || sy < srcBounds.Min.Y || sy >= srcBounds.Max.Y {
						continue
					}

					overlapX := math.Min(srcX1, float64(sx+1)) - math.Max(srcX0, float64(sx))
					overlapY := math.Min(srcY1, float64(sy+1)) - math.Max(srcY0, float64(sy))
					if overlapX <= 0 || overlapY <= 0 {
						continue
					}
					weight := overlapX * overlapY

					r, g, b, a := src.At(sx, sy).RGBA()
					aF := float64(a) / 65535.0
					totR += (float64(r) / 65535.0) * weight
					totG += (float64(g) / 65535.0) * weight
					totB += (float64(b) / 65535.0) * weight
					totA += aF * weight
					totWeight += weight
				}
			}

			if totWeight > 0 {
				finalA := totA / totWeight
				var finalR, finalG, finalB float64
				if finalA > 0 {
					finalR = (totR / totWeight) / finalA
					finalG = (totG / totWeight) / finalA
					finalB = (totB / totWeight) / finalA
				}
				dst.Set(x, y, color.RGBA{
					R: uint8(math.Min(255.0, finalR*255.0)),
					G: uint8(math.Min(255.0, finalG*255.0)),
					B: uint8(math.Min(255.0, finalB*255.0)),
					A: uint8(math.Min(255.0, finalA*255.0)),
				})
			}
		}
	}

	return dst
}

// encodePNG encodes an image to PNG bytes
func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// savePNG writes an image to a PNG file
func savePNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// packICO writes multiple PNG-encoded images into a Windows .ico file
func packICO(pngImages [][]byte, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	count := uint16(len(pngImages))

	// ICO Header (6 bytes)
	header := make([]byte, 6)
	binary.LittleEndian.PutUint16(header[0:2], 0)
	binary.LittleEndian.PutUint16(header[2:4], 1)
	binary.LittleEndian.PutUint16(header[4:6], count)
	if _, err := f.Write(header); err != nil {
		return err
	}

	// Directory entries (16 bytes each)
	offset := uint32(6 + 16*int(count))
	for _, pngData := range pngImages {
		cfg, err := png.DecodeConfig(bytes.NewReader(pngData))
		if err != nil {
			return err
		}

		w := uint8(cfg.Width)
		if cfg.Width >= 256 {
			w = 0 // 0 represents 256 in ICO spec
		}
		h := uint8(cfg.Height)
		if cfg.Height >= 256 {
			h = 0
		}

		entry := make([]byte, 16)
		entry[0] = w
		entry[1] = h
		entry[2] = 0 // color count
		entry[3] = 0 // reserved
		binary.LittleEndian.PutUint16(entry[4:6], 1)                        // color planes
		binary.LittleEndian.PutUint16(entry[6:8], 32)                       // bits per pixel
		binary.LittleEndian.PutUint32(entry[8:12], uint32(len(pngData)))    // data size
		binary.LittleEndian.PutUint32(entry[12:16], offset)                 // data offset

		if _, err := f.Write(entry); err != nil {
			return err
		}
		offset += uint32(len(pngData))
	}

	// Write PNG image payloads
	for _, pngData := range pngImages {
		if _, err := f.Write(pngData); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	masterPath := "assets/icon-1024.png"
	f, err := os.Open(masterPath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", masterPath, err)
		os.Exit(1)
	}
	defer f.Close()

	masterImg, _, err := image.Decode(f)
	if err != nil {
		fmt.Printf("Error decoding %s: %v\n", masterPath, err)
		os.Exit(1)
	}

	fmt.Println("🚀 Generating multi-platform icon assets from assets/icon-1024.png...")

	// 1. Standard PNG sizes for Linux & runtime embedding
	standardSizes := []int{16, 32, 48, 64, 128, 256, 512, 1024}
	scaledMap := make(map[int]image.Image)

	for _, sz := range standardSizes {
		if sz == 1024 {
			scaledMap[sz] = masterImg
		} else {
			scaledMap[sz] = resizeImage(masterImg, sz, sz)
		}
		outPng := fmt.Sprintf("assets/icons/icon-%d.png", sz)
		if err := savePNG(scaledMap[sz], outPng); err != nil {
			fmt.Printf("Failed to save %s: %v\n", outPng, err)
		} else {
			fmt.Printf("  ✓ Generated %s\n", outPng)
		}
	}

	// 2. macOS Apple Icon Image (.icns) via iconutil
	iconsetDir := "assets/icon.iconset"
	if err := os.MkdirAll(iconsetDir, 0755); err != nil {
		fmt.Printf("Failed to create iconset dir: %v\n", err)
	} else {
		appleSizes := []struct {
			filename string
			size     int
		}{
			{"icon_16x16.png", 16},
			{"icon_16x16@2x.png", 32},
			{"icon_32x32.png", 32},
			{"icon_32x32@2x.png", 64},
			{"icon_128x128.png", 128},
			{"icon_128x128@2x.png", 256},
			{"icon_256x256.png", 256},
			{"icon_256x256@2x.png", 512},
			{"icon_512x512.png", 512},
			{"icon_512x512@2x.png", 1024},
		}

		for _, item := range appleSizes {
			img, ok := scaledMap[item.size]
			if !ok {
				img = resizeImage(masterImg, item.size, item.size)
			}
			p := filepath.Join(iconsetDir, item.filename)
			_ = savePNG(img, p)
		}

		cmd := exec.Command("iconutil", "-c", "icns", iconsetDir, "-o", "assets/icon.icns")
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("  ⚠ iconutil note: %v (output: %s)\n", err, string(out))
		} else {
			fmt.Println("  ✓ Generated assets/icon.icns (Apple Icon Image)")
		}

		_ = os.RemoveAll(iconsetDir)
	}

	// 3. Windows Multi-Resolution Icon (.ico)
	icoSizes := []int{16, 32, 48, 64, 128, 256}
	var icoPayloads [][]byte
	for _, sz := range icoSizes {
		img, ok := scaledMap[sz]
		if !ok {
			img = resizeImage(masterImg, sz, sz)
		}
		data, err := encodePNG(img)
		if err != nil {
			fmt.Printf("Failed encoding ICO size %d: %v\n", sz, err)
			continue
		}
		icoPayloads = append(icoPayloads, data)
	}
	if err := packICO(icoPayloads, "assets/icon.ico"); err != nil {
		fmt.Printf("Failed to generate assets/icon.ico: %v\n", err)
	} else {
		fmt.Println("  ✓ Generated assets/icon.ico (Windows Multi-Resolution Icon)")
	}

	// 4. WebAssembly Web Assets & Favicons (Light and Dark Mode support)
	var faviconPayloads [][]byte
	for _, sz := range []int{16, 32, 48} {
		data, _ := encodePNG(scaledMap[sz])
		faviconPayloads = append(faviconPayloads, data)
	}
	_ = packICO(faviconPayloads, "web/favicon.ico")
	fmt.Println("  ✓ Generated web/favicon.ico")

	// Standard Light Mode web favicon (32x32)
	_ = savePNG(scaledMap[32], "web/favicon-light.png")
	fmt.Println("  ✓ Generated web/favicon-light.png")

	// Dark Mode web favicon (32x32)
	darkFavicon := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.Draw(darkFavicon, darkFavicon.Bounds(), scaledMap[32], image.Point{}, draw.Over)
	_ = savePNG(darkFavicon, "web/favicon-dark.png")
	fmt.Println("  ✓ Generated web/favicon-dark.png")

	// Apple touch icon (180x180)
	touchImg := resizeImage(masterImg, 180, 180)
	_ = savePNG(touchImg, "web/apple-touch-icon.png")
	fmt.Println("  ✓ Generated web/apple-touch-icon.png (iOS / WebKit Touch Icon)")

	// PWA icons
	_ = savePNG(resizeImage(masterImg, 192, 192), "web/icon-192.png")
	_ = savePNG(scaledMap[512], "web/icon-512.png")
	fmt.Println("  ✓ Generated web/icon-192.png and web/icon-512.png")

	fmt.Println("\n🎉 All icon assets generated successfully!")
}
