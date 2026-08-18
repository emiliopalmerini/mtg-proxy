package halftone_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/epalmerini/mtg-proxy/internal/halftone"
)

func TestApplyReturnsCorrectSize(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 80, 80))
	out := halftone.Apply(src, 8)

	if out.Bounds().Dx() != 80 || out.Bounds().Dy() != 80 {
		t.Errorf("expected 80x80, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestApplyWhiteInputStaysWhite(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			src.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	out := halftone.Apply(src, 8)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if got := color.GrayModel.Convert(out.At(x, y)).(color.Gray).Y; got != 255 {
				t.Fatalf("expected white pixel, got %d at (%d,%d)", got, x, y)
			}
		}
	}
}

func TestApplyUsesOnlyInkSavingPalette(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			src.SetGray(x, y, color.Gray{Y: uint8((x * 255) / 31)})
		}
	}

	out := halftone.Apply(src, 8)
	allowed := map[uint8]bool{255: true, 224: true, 176: true, 64: true}
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			v := color.GrayModel.Convert(out.At(x, y)).(color.Gray).Y
			if !allowed[v] {
				t.Fatalf("pixel (%d,%d) uses unexpected gray level %d", x, y, v)
			}
		}
	}
}

func TestApplyLightensMidtones(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			src.SetGray(x, y, color.Gray{Y: 128})
		}
	}

	out := halftone.Apply(src, 8)
	var total int
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			total += int(color.GrayModel.Convert(out.At(x, y)).(color.Gray).Y)
		}
	}
	average := float64(total) / (64 * 64)
	if average <= 128 {
		t.Fatalf("expected ink-saving output lighter than source, average %.1f", average)
	}
}
