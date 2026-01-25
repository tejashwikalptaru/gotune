package visualizer

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

const (
	radialInnerRadiusRatio = 0.2   // Inner radius as fraction of min(w,h)/2
	radialSpinSpeed        = 0.005 // Radians per frame
	radialCapFalloff       = 2.0   // Pixels per frame the cap falls
	radialBarThickness     = 3     // Line thickness for bars
)

// Radial is a widget that displays audio spectrum in a radial/circular pattern.
// Bars radiate outward from a central point, creating a spinning wheel effect.
// Based on audioMotion-analyzer's radial mode.
type Radial struct {
	BaseVisualizer

	numBars    int
	rotation   float64   // Current rotation offset in radians
	spinSpeed  float64   // Rotation speed (radians per frame)
	capHeights []float32 // Cap positions for each bar

	// Mirror mode: -1 = bass center, 0 = none, 1 = bass edges
	mirrorMode int

	freq FrequencyAnalyzer
	draw DrawingUtils
}

// NewRadial creates a new radial spectrum visualizer widget.
func NewRadial(numBars ...int) *Radial {
	bars := 64
	if len(numBars) > 0 && numBars[0] > 0 {
		bars = numBars[0]
	}

	v := &Radial{
		numBars:    bars,
		rotation:   0,
		spinSpeed:  radialSpinSpeed,
		capHeights: make([]float32, bars),
		mirrorMode: 0,
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// Reset clears the visualizer state.
func (v *Radial) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	v.rotation = 0
	v.BassAvg = 0
	for i := range v.capHeights {
		v.capHeights[i] = 0
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the radial visualizer.
func (v *Radial) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	v.draw.FillBackground(img, color.Black)

	v.Mu.Lock()
	fftData := v.FFTData
	caps := v.capHeights
	rotation := v.rotation
	v.Mu.Unlock()

	if len(fftData) == 0 || w == 0 || h == 0 {
		return img
	}

	// Calculate dimensions
	centerX := float64(w) / 2
	centerY := float64(h) / 2
	minDim := math.Min(float64(w), float64(h))
	maxRadius := minDim / 2
	innerRadius := maxRadius * radialInnerRadiusRatio
	maxBarLen := maxRadius - innerRadius - 10 // Leave some margin

	// Calculate bass for center pulse
	bass := v.freq.CalculateBass(fftData, 0)
	v.Mu.Lock()
	v.BassAvg = v.BassAvg*0.7 + bass*0.3
	pulseRadius := innerRadius * (1 + v.BassAvg*0.2)
	v.Mu.Unlock()

	// Calculate bar heights
	barHeights := v.freq.CalculateBarHeights(fftData, v.numBars, maxBarLen)

	// Update cap heights
	v.Mu.Lock()
	v.updateCapHeights(barHeights, caps)
	v.Mu.Unlock()

	// Draw inner circle (pulsing with bass)
	v.draw.DrawFilledCircle(img, int(centerX), int(centerY), pulseRadius, color.RGBA{R: 20, G: 20, B: 30, A: 255})
	v.draw.DrawCircle(img, int(centerX), int(centerY), pulseRadius, color.RGBA{R: 80, G: 80, B: 120, A: 255})

	// Draw bars radiating from center
	angleStep := 2 * math.Pi / float64(v.numBars)

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		// Calculate angle with rotation offset
		angle := float64(i)*angleStep + rotation - math.Pi/2 // Start from top

		barLen := float64(barHeights[i])
		capLen := float64(caps[i])

		// Bar start and end points
		startX := centerX + math.Cos(angle)*pulseRadius
		startY := centerY + math.Sin(angle)*pulseRadius
		endX := centerX + math.Cos(angle)*(pulseRadius+barLen)
		endY := centerY + math.Sin(angle)*(pulseRadius+barLen)

		// Get rainbow gradient color based on bar position around the circle
		posRatio := float64(i) / float64(v.numBars)
		col := v.getRainbowColor(posRatio)

		// Draw the bar
		v.draw.DrawThickLine(img, startX, startY, endX, endY, radialBarThickness, col)

		// Draw cap
		if capLen > 0 {
			capStartX := centerX + math.Cos(angle)*(pulseRadius+capLen)
			capStartY := centerY + math.Sin(angle)*(pulseRadius+capLen)
			capEndX := centerX + math.Cos(angle)*(pulseRadius+capLen+4)
			capEndY := centerY + math.Sin(angle)*(pulseRadius+capLen+4)
			v.draw.DrawThickLine(img, capStartX, capStartY, capEndX, capEndY, radialBarThickness, color.RGBA{R: 255, G: 255, B: 255, A: 200})
		}
	}

	// Update rotation for next frame
	v.Mu.Lock()
	v.rotation += v.spinSpeed
	if v.rotation >= 2*math.Pi {
		v.rotation -= 2 * math.Pi
	}
	v.Mu.Unlock()

	return img
}

// updateCapHeights updates cap positions with falling animation.
func (v *Radial) updateCapHeights(barHeights []float32, caps []float32) {
	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := barHeights[i]
		if barH > caps[i] {
			caps[i] = barH
		} else {
			caps[i] -= radialCapFalloff
			if caps[i] < 0 {
				caps[i] = 0
			}
		}
	}
}

// getRainbowColor returns a color from a rainbow gradient based on position (0.0 to 1.0).
func (v *Radial) getRainbowColor(pos float64) color.RGBA {
	// Use HSL with varying hue for rainbow effect
	hue := pos // 0 to 1 maps to full color wheel
	saturation := 1.0
	lightness := 0.5

	r, g, b := HSLToRGB(hue, saturation, lightness)

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Radial)(nil)
