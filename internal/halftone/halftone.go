package halftone

import (
	"image"
	"image/color"
)

// Apply converts an image to an ink-saving, four-level grayscale palette using
// Floyd-Steinberg error-diffusion dithering. The deliberately light palette
// keeps the art recognizable while reducing average toner/ink coverage.
func Apply(src image.Image, _ int) *image.Gray {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewGray(image.Rect(0, 0, w, h))

	// Light-biased quantization palette. Pure black is retained for the darkest
	// details, while most pixels are pushed toward paper white.
	palette := [...]float64{255, 224, 176, 64}
	errCurr := make([]float64, w+2)
	errNext := make([]float64, w+2)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := src.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 257.0
			v := clamp(lum+errCurr[x+1], 0, 255)
			q := nearest(v, palette[:])
			dst.SetGray(x, y, color.Gray{Y: uint8(q)})

			e := v - q
			errCurr[x+2] += e * 7 / 16
			errNext[x] += e * 3 / 16
			errNext[x+1] += e * 5 / 16
			errNext[x+2] += e * 1 / 16
		}
		errCurr, errNext = errNext, errCurr
		clear(errNext)
	}

	return dst
}

func nearest(v float64, palette []float64) float64 {
	best := palette[0]
	bestDistance := abs(v - best)
	for _, candidate := range palette[1:] {
		distance := abs(v - candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
