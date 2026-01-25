package visualizer

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2/canvas"
)

const (
	ledSegments      = 16  // Number of LED segments per bar
	ledGapRatio      = 0.2 // Gap as fraction of segment height
	ledPaddingTop    = 10
	ledPaddingLeft   = 10
	ledPaddingRight  = 10
	ledPaddingBottom = 10
	ledMinGap        = 2   // Minimum gap between bars
	ledCapFalloff    = 1.0 // Segments per update the cap falls
)

// LEDBars is a widget that displays audio spectrum as LED-style segmented bars.
// Based on audioMotion-analyzer's LED bars mode.
type LEDBars struct {
	BaseVisualizer

	numBars      int
	capPositions []float32 // Cap position (segment index) for each bar
	showDimLEDs  bool      // Show unlit segments dimly

	// Layout cache
	lastWidth        int
	lastHeight       int
	cachedBarWidth   int
	cachedActualGap  int
	cachedStartX     int
	cachedEffectiveW int
	cachedEffectiveH int
	cachedSegHeight  int
	cachedSegGap     int

	freq FrequencyAnalyzer
	draw DrawingUtils
}

// NewLEDBars creates a new LED bars visualizer widget.
func NewLEDBars(numBars ...int) *LEDBars {
	bars := 32
	if len(numBars) > 0 && numBars[0] > 0 {
		bars = numBars[0]
	}

	v := &LEDBars{
		numBars:      bars,
		capPositions: make([]float32, bars),
		showDimLEDs:  true,
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// Reset clears the visualizer state.
func (v *LEDBars) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	for i := range v.capPositions {
		v.capPositions[i] = 0
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the LED bars visualizer.
func (v *LEDBars) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	v.draw.FillBackground(img, color.Black)

	v.Mu.Lock()
	fftData := v.FFTData
	caps := v.capPositions
	v.Mu.Unlock()

	if len(fftData) == 0 || w == 0 || h == 0 {
		return img
	}

	// Recalculate layout only if size changed
	if v.lastWidth != w || v.lastHeight != h {
		v.recalculateLayout(w, h)
	}

	if v.cachedBarWidth == 0 || v.cachedSegHeight == 0 {
		return img
	}

	// Calculate bar heights
	maxHeight := float64(v.cachedEffectiveH)
	barHeights := v.freq.CalculateBarHeights(fftData, v.numBars, maxHeight)

	// Update cap positions
	v.Mu.Lock()
	v.updateCapPositions(barHeights, caps, maxHeight)
	v.Mu.Unlock()

	// Draw all bars
	v.drawBars(img, barHeights, caps, h, maxHeight)

	return img
}

// recalculateLayout computes and caches size-dependent layout values.
func (v *LEDBars) recalculateLayout(w, h int) {
	v.lastWidth = w
	v.lastHeight = h

	v.cachedEffectiveW = w - ledPaddingLeft - ledPaddingRight
	v.cachedEffectiveH = h - ledPaddingTop - ledPaddingBottom

	if v.cachedEffectiveW <= 0 || v.cachedEffectiveH <= 0 {
		v.cachedBarWidth = 0
		return
	}

	// Calculate segment dimensions
	totalSegHeight := v.cachedEffectiveH
	segWithGap := float64(totalSegHeight) / float64(ledSegments)
	v.cachedSegGap = max(int(segWithGap*ledGapRatio), 1)
	v.cachedSegHeight = max(int(segWithGap)-v.cachedSegGap, 2)

	// Calculate bar dimensions
	totalGapWidth := (v.numBars - 1) * ledMinGap
	availableBarWidth := v.cachedEffectiveW - totalGapWidth

	v.cachedBarWidth = max(availableBarWidth/v.numBars, 2)

	// Recalculate gap to distribute remaining space evenly
	v.cachedActualGap = ledMinGap
	if v.numBars > 1 {
		remainingSpace := v.cachedEffectiveW - (v.cachedBarWidth * v.numBars)
		v.cachedActualGap = max(remainingSpace/(v.numBars-1), ledMinGap)
	}

	// Calculate starting X position
	usedWidth := v.numBars*v.cachedBarWidth + (v.numBars-1)*v.cachedActualGap
	v.cachedStartX = ledPaddingLeft + (v.cachedEffectiveW-usedWidth)/2
}

// updateCapPositions updates cap positions with falling animation.
func (v *LEDBars) updateCapPositions(barHeights []float32, caps []float32, maxHeight float64) {
	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		// Convert height to segment position
		litSegments := float32(barHeights[i]) / float32(maxHeight) * float32(ledSegments)

		if litSegments > caps[i] {
			caps[i] = litSegments
		} else {
			caps[i] -= ledCapFalloff
			if caps[i] < 0 {
				caps[i] = 0
			}
		}
	}
}

// drawBars renders all LED bars to the image.
func (v *LEDBars) drawBars(img *image.RGBA, barHeights []float32, caps []float32, h int, maxHeight float64) {
	totalBarWidth := v.cachedBarWidth + v.cachedActualGap
	segStep := v.cachedSegHeight + v.cachedSegGap

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barX := v.cachedStartX + i*totalBarWidth

		// Calculate how many segments are lit
		litSegments := int(float64(barHeights[i]) / maxHeight * float64(ledSegments))
		capSegment := int(caps[i])

		// Draw each segment
		for seg := range ledSegments {
			segY := h - ledPaddingBottom - (seg+1)*segStep

			// Determine segment color based on position
			segRatio := float64(seg) / float64(ledSegments)
			col := v.getLEDColor(segRatio)

			switch {
			case seg < litSegments:
				// Lit segment
				v.drawSegment(img, barX, segY, col)
			case seg == capSegment && capSegment > 0:
				// Cap segment (white)
				v.drawSegment(img, barX, segY, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			case v.showDimLEDs:
				// Dim/unlit segment
				dimCol := color.RGBA{R: 30, G: 30, B: 30, A: 255}
				v.drawSegment(img, barX, segY, dimCol)
			}
		}
	}
}

// drawSegment draws a single LED segment.
func (v *LEDBars) drawSegment(img *image.RGBA, x, y int, col color.RGBA) {
	bounds := img.Bounds()

	for dy := 0; dy < v.cachedSegHeight; dy++ {
		for dx := 0; dx < v.cachedBarWidth; dx++ {
			px := x + dx
			py := y + dy

			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				img.Set(px, py, col)
			}
		}
	}
}

// getLEDColor returns the color for an LED segment based on its vertical position.
// 0-40%: Green, 40-75%: Yellow, 75-100%: Red
func (v *LEDBars) getLEDColor(ratio float64) color.RGBA {
	switch {
	case ratio < 0.4:
		// Green zone
		return color.RGBA{R: 0, G: 255, B: 0, A: 255}
	case ratio < 0.75:
		// Yellow zone - transition from green to yellow
		t := (ratio - 0.4) / 0.35
		return color.RGBA{R: uint8(255 * t), G: 255, B: 0, A: 255}
	default:
		// Red zone - transition from yellow to red
		t := (ratio - 0.75) / 0.25
		return color.RGBA{R: 255, G: uint8(255 * (1 - t)), B: 0, A: 255}
	}
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*LEDBars)(nil)
