package assets

import (
	"testing"
)

func TestGetWindowIcons(t *testing.T) {
	icons := GetWindowIcons()
	expectedSizes := []int{16, 32, 48, 64, 128}

	if len(icons) != len(expectedSizes) {
		t.Fatalf("expected %d icons, got %d", len(expectedSizes), len(icons))
	}

	for i, expected := range expectedSizes {
		img := icons[i]
		if img == nil {
			t.Fatalf("icon %d is nil", i)
		}
		bounds := img.Bounds()
		if bounds.Dx() != expected || bounds.Dy() != expected {
			t.Errorf("icon %d: expected %dx%d, got %dx%d", i, expected, expected, bounds.Dx(), bounds.Dy())
		}
	}
}
