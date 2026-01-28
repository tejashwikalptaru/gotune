package visualizer

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

const (
	circularInnerRadiusRatio = 0.15 // Inner circle ratio of min dimension
	circularMaxBarRatio      = 0.35 // Maximum bar length ratio of min dimension
	circularCapFalloff       = 2.0  // Pixels per frame the cap falls
)

// Circular is a widget that displays audio spectrum in a circular pattern.
// Bars radiate outward from a central circle.
type Circular struct {
	BaseVisualizer

	capHeights []float32 // Falling cap animation heights (in radial distance)
	numBars    int

	freq FrequencyAnalyzer
	draw DrawingUtils
}

// NewCircular creates a new circular visualizer with the specified number of bars.
func NewCircular(numBars int) *Circular {
	v := &Circular{
		numBars:    numBars,
		capHeights: make([]float32, numBars),
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// Reset clears the visualizer state.
func (v *Circular) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	v.BassAvg = 0
	for i := range v.capHeights {
		v.capHeights[i] = 0
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the circular visualizer.
func (v *Circular) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	v.draw.FillBackground(img, color.Black)

	v.Mu.Lock()
	fftData := v.FFTData
	caps := v.capHeights
	v.Mu.Unlock()

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
	bass := v.freq.CalculateBass(fftData, 0)
	v.Mu.Lock()
	v.BassAvg = v.BassAvg*0.7 + bass*0.3
	pulseRadius := innerRadius + v.BassAvg*innerRadius*0.3
	v.Mu.Unlock()

	// Calculate bar heights using logarithmic frequency distribution
	barHeights := v.freq.CalculateBarHeights(fftData, v.numBars, maxBarLen)

	// Update cap heights
	v.Mu.Lock()
	v.updateCapHeights(barHeights, caps)
	v.Mu.Unlock()

	// Draw inner circle (pulsing with bass)
	v.draw.DrawFilledCircle(img, int(centerX), int(centerY), pulseRadius, color.RGBA{R: 30, G: 30, B: 40, A: 255})
	v.draw.DrawCircle(img, int(centerX), int(centerY), pulseRadius, color.RGBA{R: 100, G: 100, B: 150, A: 255})

	// Draw bars radiating from center
	angleStep := 2 * math.Pi / float64(v.numBars)

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		angle := float64(i)*angleStep - math.Pi/2 // Start from top

		barLen := float64(barHeights[i])
		capLen := float64(caps[i])

		// Bar start and end points
		startX := centerX + math.Cos(angle)*innerRadius
		startY := centerY + math.Sin(angle)*innerRadius
		endX := centerX + math.Cos(angle)*(innerRadius+barLen)
		endY := centerY + math.Sin(angle)*(innerRadius+barLen)

		// Get gradient color based on bar height
		heightRatio := barLen / maxBarLen
		col := v.draw.GetGradientColor(heightRatio)

		// Draw the bar
		v.draw.DrawThickLine(img, startX, startY, endX, endY, 2, col)

		// Draw cap
		if capLen > 0 {
			capStartX := centerX + math.Cos(angle)*(innerRadius+capLen)
			capStartY := centerY + math.Sin(angle)*(innerRadius+capLen)
			capEndX := centerX + math.Cos(angle)*(innerRadius+capLen+3)
			capEndY := centerY + math.Sin(angle)*(innerRadius+capLen+3)
			v.draw.DrawThickLine(img, capStartX, capStartY, capEndX, capEndY, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	return img
}

// updateCapHeights updates cap positions with falling animation.
func (v *Circular) updateCapHeights(barHeights []float32, caps []float32) {
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

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Circular)(nil)
