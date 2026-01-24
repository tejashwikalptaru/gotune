package visualizer

import (
	"image"
	"image/color"
	"math"
)

// FrequencyAnalyzer provides methods for analyzing FFT data.
type FrequencyAnalyzer struct{}

// CalculateBass returns the average bass amplitude (low frequencies).
// Uses bins 1-9 (approximately 0-200Hz).
func (FrequencyAnalyzer) CalculateBass(fftData []float32, defaultValue float64) float64 {
	if len(fftData) < 10 {
		return defaultValue
	}

	var sum float64
	for i := 1; i < 10 && i < len(fftData); i++ {
		sum += float64(fftData[i])
	}
	return math.Sqrt(sum / 9.0)
}

// CalculateMid returns the average mid-frequency amplitude.
// Uses bins 10-49 (approximately 200Hz-1kHz).
func (FrequencyAnalyzer) CalculateMid(fftData []float32, defaultValue float64) float64 {
	if len(fftData) < 50 {
		return defaultValue
	}

	var sum float64
	count := 0
	for i := 10; i < 50 && i < len(fftData); i++ {
		sum += float64(fftData[i])
		count++
	}
	if count == 0 {
		return defaultValue
	}
	return math.Sqrt(sum / float64(count))
}

// CalculateHigh returns the average high-frequency amplitude.
// Uses bins 50+ (approximately 1kHz+).
func (FrequencyAnalyzer) CalculateHigh(fftData []float32, defaultValue float64) float64 {
	if len(fftData) < 100 {
		return defaultValue
	}

	var sum float64
	count := 0
	for i := 50; i < len(fftData); i++ {
		sum += float64(fftData[i])
		count++
	}
	if count == 0 {
		return defaultValue
	}
	return math.Sqrt(sum / float64(count))
}

// CalculateBarHeights converts FFT data to bar heights using logarithmic bin mapping.
// This is based on the BASS spectrum.c example for better frequency distribution.
func (FrequencyAnalyzer) CalculateBarHeights(fftData []float32, numBars int, maxHeight float64) []float32 {
	heights := make([]float32, numBars)

	if len(fftData) < 2 {
		return heights
	}

	b0 := 1 // Skip DC offset

	for x := 0; x < numBars; x++ {
		var b1 int
		if numBars > 1 {
			b1 = int(math.Pow(2, float64(x)*10.0/float64(numBars-1)))
		} else {
			b1 = len(fftData) - 1
		}

		if b1 >= len(fftData) {
			b1 = len(fftData) - 1
		}
		if b1 < b0 {
			b1 = b0
		}

		var peak float32
		for b := b0; b <= b1 && b < len(fftData); b++ {
			if fftData[b] > peak {
				peak = fftData[b]
			}
		}

		y := float32(math.Sqrt(float64(peak))) * 3.0 * float32(maxHeight)
		if y < 0 {
			y = 0
		}
		if y > float32(maxHeight) {
			y = float32(maxHeight)
		}

		heights[x] = y
		b0 = b1 + 1
	}

	return heights
}

// DrawingUtils provides common drawing operations.
type DrawingUtils struct{}

// FillBackground fills the image with a solid color.
func (DrawingUtils) FillBackground(img *image.RGBA, col color.Color) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, col)
		}
	}
}

// DrawThickLine draws a line with the specified thickness.
func (DrawingUtils) DrawThickLine(img *image.RGBA, x1, y1, x2, y2 float64, thickness int, col color.RGBA) {
	bounds := img.Bounds()

	dx := x2 - x1
	dy := y2 - y1
	length := math.Sqrt(dx*dx + dy*dy)

	if length == 0 {
		return
	}

	// Perpendicular unit vector for thickness
	perpX := -dy / length
	perpY := dx / length

	steps := int(length) + 1

	for t := -thickness / 2; t <= thickness/2; t++ {
		offsetX := float64(t) * perpX
		offsetY := float64(t) * perpY

		for i := 0; i <= steps; i++ {
			progress := float64(i) / float64(steps)
			px := int(x1 + dx*progress + offsetX)
			py := int(y1 + dy*progress + offsetY)

			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				img.Set(px, py, col)
			}
		}
	}
}

// DrawCircle draws a circle outline.
func (DrawingUtils) DrawCircle(img *image.RGBA, cx, cy int, radius float64, col color.RGBA) {
	bounds := img.Bounds()

	steps := int(2 * math.Pi * radius)
	if steps < 36 {
		steps = 36
	}

	for i := 0; i < steps; i++ {
		angle := 2 * math.Pi * float64(i) / float64(steps)
		px := int(float64(cx) + math.Cos(angle)*radius)
		py := int(float64(cy) + math.Sin(angle)*radius)

		if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
			img.Set(px, py, col)
		}
	}
}

// DrawFilledCircle draws a filled circle.
func (DrawingUtils) DrawFilledCircle(img *image.RGBA, cx, cy int, radius float64, col color.RGBA) {
	bounds := img.Bounds()
	r := int(radius)

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				px, py := cx+dx, cy+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					img.Set(px, py, col)
				}
			}
		}
	}
}

// GetGradientColor returns a color from a red-yellow-green gradient based on position (0.0 to 1.0).
func (DrawingUtils) GetGradientColor(pos float64) color.RGBA {
	if pos < 0 {
		pos = 0
	}
	if pos > 1 {
		pos = 1
	}

	var r, g uint8

	if pos < 0.5 {
		r = 255
		g = uint8(pos * 2 * 255)
	} else {
		r = uint8((1 - (pos-0.5)*2) * 255)
		g = 255
	}

	return color.RGBA{R: r, G: g, B: 0, A: 255}
}

// HSLToRGB converts HSL to RGB (h, s, l in 0-1 range).
func HSLToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)

	return r, g, b
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 0.5 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
