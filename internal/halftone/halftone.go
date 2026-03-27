package halftone

import (
	"image"
	"image/color"
	"math"
)

// Apply converts an image to a black-and-white circular halftone pattern.
// gridSize controls dot spacing: smaller = more detail, larger = more stylized.
func Apply(src image.Image, gridSize int) *image.Gray {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	dst := image.NewGray(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			dst.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	maxRadius := float64(gridSize) / 2.0

	for gy := 0; gy < h; gy += gridSize {
		for gx := 0; gx < w; gx += gridSize {
			var total float64
			var count int
			for dy := 0; dy < gridSize && gy+dy < h; dy++ {
				for dx := 0; dx < gridSize && gx+dx < w; dx++ {
					r, g, b, _ := src.At(bounds.Min.X+gx+dx, bounds.Min.Y+gy+dy).RGBA()
					lum := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
					total += lum / 65535.0
					count++
				}
			}

			darkness := 1.0 - total/float64(count)
			if darkness < 0.01 {
				continue
			}
			radius := maxRadius * math.Sqrt(darkness)

			cx := float64(gx) + maxRadius
			cy := float64(gy) + maxRadius
			r2 := radius * radius

			for dy := -int(maxRadius) - 1; dy <= int(maxRadius)+1; dy++ {
				for dx := -int(maxRadius) - 1; dx <= int(maxRadius)+1; dx++ {
					px := int(cx) + dx
					py := int(cy) + dy
					if px < 0 || px >= w || py < 0 || py >= h {
						continue
					}
					if float64(dx)*float64(dx)+float64(dy)*float64(dy) <= r2 {
						dst.SetGray(px, py, color.Gray{Y: 0})
					}
				}
			}
		}
	}

	return dst
}
