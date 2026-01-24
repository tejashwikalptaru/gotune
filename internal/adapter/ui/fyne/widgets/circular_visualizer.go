// Package widgets provides custom Fyne widgets for the GoTune application.
package widgets

import (
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	circularInnerRadiusRatio = 0.15 // Inner circle ratio of min dimension
	circularMaxBarRatio      = 0.35 // Maximum bar length ratio of min dimension
	circularCapFalloff       = 2.0  // Pixels per frame the cap falls
)

// CircularVisualizer is a widget that displays audio spectrum in a circular pattern.
// Bars radiate outward from a central circle.
type CircularVisualizer struct {
	widget.BaseWidget

	raster     *canvas.Raster
	fftData    []float32
	capHeights []float32 // Falling cap animation heights (in radial distance)
	numBars    int
	mu         sync.RWMutex
	bassAvg    float64 // Smoothed bass for center pulse
}

// NewCircularVisualizer creates a new circular visualizer with the specified number of bars.
func NewCircularVisualizer(numBars int) *CircularVisualizer {
	v := &CircularVisualizer{
		numBars:    numBars,
		capHeights: make([]float32, numBars),
		bassAvg:    0,
	}

	v.raster = canvas.NewRaster(v.draw)
	v.ExtendBaseWidget(v)

	return v
}

// CreateRenderer implements fyne.Widget.
func (v *CircularVisualizer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.raster)
}

// MinSize returns the minimum size of the visualizer.
func (v *CircularVisualizer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// UpdateFFT updates the visualizer with new FFT data.
func (v *CircularVisualizer) UpdateFFT(data []float32) {
	v.mu.Lock()
	v.fftData = data
	v.mu.Unlock()

	v.raster.Refresh()
}

// Reset clears the visualizer state.
func (v *CircularVisualizer) Reset() {
	v.mu.Lock()
	v.fftData = nil
	v.bassAvg = 0
	for i := range v.capHeights {
		v.capHeights[i] = 0
	}
	v.mu.Unlock()

	v.raster.Refresh()
}

// draw renders the circular visualizer.
func (v *CircularVisualizer) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with black background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.Black)
		}
	}

	v.mu.Lock()
	fftData := v.fftData
	caps := v.capHeights
	v.mu.Unlock()

	if len(fftData) == 0 || w == 0 || h == 0 {
		return img
	}

	// Calculate dimensions
	centerX := float64(w) / 2
	centerY := float64(h) / 2
	minDim := math.Min(float64(w), float64(h))
	innerRadius := minDim * circularInnerRadiusRatio
	maxBarLen := minDim * circularMaxBarRatio

	// Calculate bass for center pulse
	bass := v.calculateBass(fftData)
	v.mu.Lock()
	v.bassAvg = v.bassAvg*0.7 + bass*0.3
	pulseRadius := innerRadius + v.bassAvg*innerRadius*0.3
	v.mu.Unlock()

	// Calculate bar heights using logarithmic frequency distribution
	barHeights := v.calculateBarHeights(fftData, maxBarLen)

	// Update cap heights
	v.mu.Lock()
	v.updateCapHeights(barHeights, caps, maxBarLen)
	v.mu.Unlock()

	// Draw inner circle (pulsing with bass)
	v.drawFilledCircle(img, int(centerX), int(centerY), pulseRadius, color.RGBA{R: 30, G: 30, B: 40, A: 255})
	v.drawCircle(img, int(centerX), int(centerY), pulseRadius, color.RGBA{R: 100, G: 100, B: 150, A: 255})

	// Draw bars radiating from a center
	angleStep := 2 * math.Pi / float64(v.numBars)

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		angle := float64(i)*angleStep - math.Pi/2 // Start from the top

		barLen := float64(barHeights[i])
		capLen := float64(caps[i])

		// Bar start and end points
		startX := centerX + math.Cos(angle)*innerRadius
		startY := centerY + math.Sin(angle)*innerRadius
		endX := centerX + math.Cos(angle)*(innerRadius+barLen)
		endY := centerY + math.Sin(angle)*(innerRadius+barLen)

		// Get gradient color based on bar height
		heightRatio := barLen / maxBarLen
		col := v.getGradientColor(heightRatio)

		// Draw the bar
		v.drawThickLine(img, startX, startY, endX, endY, 2, col)

		// Draw cap
		if capLen > 0 {
			capStartX := centerX + math.Cos(angle)*(innerRadius+capLen)
			capStartY := centerY + math.Sin(angle)*(innerRadius+capLen)
			capEndX := centerX + math.Cos(angle)*(innerRadius+capLen+3)
			capEndY := centerY + math.Sin(angle)*(innerRadius+capLen+3)
			v.drawThickLine(img, capStartX, capStartY, capEndX, capEndY, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	return img
}

// calculateBarHeights converts FFT data to bar heights using logarithmic bin mapping.
func (v *CircularVisualizer) calculateBarHeights(fftData []float32, maxHeight float64) []float32 {
	heights := make([]float32, v.numBars)

	if len(fftData) < 2 {
		return heights
	}

	b0 := 1 // Skip DC offset

	for x := 0; x < v.numBars; x++ {
		var b1 int
		if v.numBars > 1 {
			b1 = int(math.Pow(2, float64(x)*10.0/float64(v.numBars-1)))
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

// updateCapHeights updates cap positions with falling animation.
func (v *CircularVisualizer) updateCapHeights(barHeights []float32, caps []float32, maxHeight float64) {
	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := barHeights[i]
		if barH > caps[i] {
			caps[i] = barH
		} else {
			caps[i] -= circularCapFalloff
			if caps[i] < 0 {
				caps[i] = 0
			}
		}
	}
}

// getGradientColor returns a color from the gradient based on position (0.0 to 1.0).
func (v *CircularVisualizer) getGradientColor(pos float64) color.RGBA {
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

// calculateBass returns the average bass amplitude.
func (v *CircularVisualizer) calculateBass(fftData []float32) float64 {
	if len(fftData) < 10 {
		return 0
	}

	var sum float64
	for i := 1; i < 10 && i < len(fftData); i++ {
		sum += float64(fftData[i])
	}
	return math.Sqrt(sum / 9.0)
}

// drawThickLine draws a line with specified thickness.
func (v *CircularVisualizer) drawThickLine(img *image.RGBA, x1, y1, x2, y2 float64, thickness int, col color.RGBA) {
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

// drawCircle draws a circle outline.
func (v *CircularVisualizer) drawCircle(img *image.RGBA, cx, cy int, radius float64, col color.RGBA) {
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

// drawFilledCircle draws a filled circle.
func (v *CircularVisualizer) drawFilledCircle(img *image.RGBA, cx, cy int, radius float64, col color.RGBA) {
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

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*CircularVisualizer)(nil)
