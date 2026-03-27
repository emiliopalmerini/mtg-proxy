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

func TestApplyBlackInputProducesLargeDots(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 16, 16))
	// Fill black
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			src.SetGray(x, y, color.Gray{Y: 0})
		}
	}

	out := halftone.Apply(src, 8)

	// Count black pixels — should be high for a black input
	black := countBlack(out)
	total := 16 * 16
	ratio := float64(black) / float64(total)
	if ratio < 0.4 {
		t.Errorf("expected high black ratio for black input, got %.2f", ratio)
	}
}

func TestApplyWhiteInputProducesNoDots(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 16, 16))
	// Fill white
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			src.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	out := halftone.Apply(src, 8)

	black := countBlack(out)
	if black != 0 {
		t.Errorf("expected 0 black pixels for white input, got %d", black)
	}
}

func TestApplyOutputIsBinaryBW(t *testing.T) {
	src := image.NewGray(image.Rect(0, 0, 24, 24))
	// Fill mid-gray
	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			src.SetGray(x, y, color.Gray{Y: 128})
		}
	}

	out := halftone.Apply(src, 8)

	for y := 0; y < 24; y++ {
		for x := 0; x < 24; x++ {
			r, g, b, _ := out.At(x, y).RGBA()
			isBlack := r == 0 && g == 0 && b == 0
			isWhite := r == 0xffff && g == 0xffff && b == 0xffff
			if !isBlack && !isWhite {
				t.Fatalf("pixel (%d,%d) is neither black nor white", x, y)
			}
		}
	}
}

func countBlack(img image.Image) int {
	b := img.Bounds()
	count := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && bb == 0 {
				count++
			}
		}
	}
	return count
}
