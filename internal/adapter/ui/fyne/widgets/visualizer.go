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

// Visualizer is a widget that displays audio spectrum bars.
// It uses logarithmic frequency distribution for better visual representation.
type Visualizer struct {
	widget.BaseWidget

	raster     *canvas.Raster
	fftData    []float32
	capHeights []float32 // Falling cap animation heights
	numBars    int
	mu         sync.RWMutex

	// Visual configuration
	capHeight  int
	capFalloff float32 // Pixels per update the cap falls
	minGap     int     // Minimum gap between bars

	// Padding configuration
	paddingTop   float32
	paddingLeft  float32
	paddingRight float32

	// Layout cache (recalculated only when size changes)
	lastWidth        int
	lastHeight       int
	cachedBarWidth   int
	cachedActualGap  int
	cachedStartX     int
	cachedEffectiveW int
	cachedEffectiveH int
	cachedPaddingL   int
	cachedPaddingR   int
	cachedPaddingT   int
}

// NewVisualizer creates a new visualizer widget with the specified number of bars.
func NewVisualizer(numBars int) *Visualizer {
	v := &Visualizer{
		numBars:      numBars,
		capHeights:   make([]float32, numBars),
		capHeight:    2,
		capFalloff:   2.0, // Pixels per frame the cap falls
		minGap:       2,   // Minimum gap between bars
		paddingTop:   10,
		paddingLeft:  10,
		paddingRight: 10,
	}

	v.raster = canvas.NewRaster(v.draw)
	v.ExtendBaseWidget(v)

	return v
}

// CreateRenderer implements fyne.Widget.
func (v *Visualizer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.raster)
}

// MinSize returns the minimum size of the visualizer.
// Returns a minimal size so the widget expands to fill available space.
func (v *Visualizer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// UpdateFFT updates the visualizer with new FFT data.
// This should be called periodically (e.g., 30fps) from the presenter.
func (v *Visualizer) UpdateFFT(data []float32) {
	v.mu.Lock()
	v.fftData = data
	v.mu.Unlock()

	// Request a redraw
	v.raster.Refresh()
}

// draw is the raster generator function that renders the visualizer.
func (v *Visualizer) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	v.fillBackground(img, w, h)

	v.mu.Lock()
	fftData := v.fftData
	caps := v.capHeights
	v.mu.Unlock()

	if len(fftData) == 0 || w == 0 || h == 0 {
		return img
	}

	// Recalculate layout only if size changed
	if v.lastWidth != w || v.lastHeight != h {
		v.recalculateLayout(w, h)
	}

	// Early return if effective area is invalid
	if v.cachedBarWidth == 0 {
		return img
	}

	// Calculate bar heights using logarithmic frequency distribution
	barHeights := v.calculateBarHeights(fftData, v.cachedEffectiveH)

	// Update cap heights with falling animation
	v.mu.Lock()
	v.updateCapHeights(barHeights, caps)
	v.mu.Unlock()

	// Draw all bars and caps
	v.drawBars(img, barHeights, caps, h)

	return img
}

// calculateBarHeights converts FFT data to bar heights using logarithmic bin mapping.
// This is based on the BASS spectrum.c example for better frequency distribution.
func (v *Visualizer) calculateBarHeights(fftData []float32, height int) []float32 {
	heights := make([]float32, v.numBars)

	if len(fftData) < 2 {
		return heights
	}

	// Skip the DC offset (first bin) like the BASS example
	// Use logarithmic bin mapping: b1 = pow(2, x * 10.0 / (numBars - 1))
	b0 := 1 // Skip DC offset

	for x := 0; x < v.numBars; x++ {
		// Calculate the upper bin index for this bar
		var b1 int
		if v.numBars > 1 {
			b1 = int(math.Pow(2, float64(x)*10.0/float64(v.numBars-1)))
		} else {
			b1 = len(fftData) - 1
		}

		// Clamp to FFT data size
		if b1 >= len(fftData) {
			b1 = len(fftData) - 1
		}
		if b1 < b0 {
			b1 = b0
		}

		// Find peak value in this bin range
		var peak float32
		for b := b0; b <= b1 && b < len(fftData); b++ {
			if fftData[b] > peak {
				peak = fftData[b]
			}
		}

		// Apply sqrt scaling to make low values more visible (from BASS example)
		// Scale: y = sqrt(peak) * 3 * height - 4
		y := float32(math.Sqrt(float64(peak))) * 3.0 * float32(height)
		if y < 0 {
			y = 0
		}
		if y > float32(height) {
			y = float32(height)
		}

		heights[x] = y

		// The next bar starts where this one ended
		b0 = b1 + 1
	}

	return heights
}

// getGradientColor returns a color from the gradient based on position (0.0 to 1.0).
// Gradient: Red (#f00) at bottom -> Yellow (#ff0) middle -> Green (#0f0) top
func (v *Visualizer) getGradientColor(pos float64) color.RGBA {
	if pos < 0 {
		pos = 0
	}
	if pos > 1 {
		pos = 1
	}

	var r, g uint8

	if pos < 0.5 {
		// Red to Yellow (0.0 -> 0.5)
		// R: 255, G: 0 -> 255
		r = 255
		g = uint8(pos * 2 * 255)
	} else {
		// Yellow to Green (0.5 -> 1.0)
		// R: 255 -> 0, G: 255
		r = uint8((1 - (pos-0.5)*2) * 255)
		g = 255
	}

	return color.RGBA{R: r, G: g, B: 0, A: 255}
}

// Reset clears the visualizer state.
func (v *Visualizer) Reset() {
	v.mu.Lock()
	v.fftData = nil
	for i := range v.capHeights {
		v.capHeights[i] = 0
	}
	v.mu.Unlock()

	v.raster.Refresh()
}

// recalculateLayout computes and caches size-dependent layout values.
// This should only be called when widget dimensions change.
func (v *Visualizer) recalculateLayout(w, h int) {
	v.lastWidth = w
	v.lastHeight = h

	v.cachedPaddingL = int(v.paddingLeft)
	v.cachedPaddingR = int(v.paddingRight)
	v.cachedPaddingT = int(v.paddingTop)

	v.cachedEffectiveW = w - v.cachedPaddingL - v.cachedPaddingR
	v.cachedEffectiveH = h - v.cachedPaddingT

	if v.cachedEffectiveW <= 0 || v.cachedEffectiveH <= 0 {
		v.cachedBarWidth = 0
		return
	}

	// Calculate bar dimensions dynamically based on available space
	totalGapWidth := (v.numBars - 1) * v.minGap
	availableBarWidth := v.cachedEffectiveW - totalGapWidth

	v.cachedBarWidth = max(availableBarWidth/v.numBars, 1)

	// Recalculate gap to distribute remaining space evenly
	v.cachedActualGap = v.minGap
	if v.numBars > 1 {
		remainingSpace := v.cachedEffectiveW - (v.cachedBarWidth * v.numBars)
		v.cachedActualGap = max(remainingSpace/(v.numBars-1), v.minGap)
	}

	// Calculate starting X position (with left padding)
	usedWidth := v.numBars*v.cachedBarWidth + (v.numBars-1)*v.cachedActualGap
	v.cachedStartX = v.cachedPaddingL + (v.cachedEffectiveW-usedWidth)/2
}

// fillBackground fills the image with black color.
func (v *Visualizer) fillBackground(img *image.RGBA, w, h int) {
	for y := range h {
		for x := range w {
			img.Set(x, y, color.Black)
		}
	}
}

// updateCapHeights updates cap positions with falling animation.
// Must be called with v.mu locked.
func (v *Visualizer) updateCapHeights(barHeights []float32, caps []float32) {
	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := barHeights[i]
		if barH > caps[i] {
			caps[i] = barH // Cap jumps to bar height
		} else {
			caps[i] -= v.capFalloff // Cap falls slowly
			if caps[i] < 0 {
				caps[i] = 0
			}
		}
	}
}

// drawBars renders all bars and their caps to the image.
func (v *Visualizer) drawBars(img *image.RGBA, barHeights []float32, caps []float32, h int) {
	totalBarWidth := v.cachedBarWidth + v.cachedActualGap

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := int(barHeights[i])
		barX := v.cachedStartX + i*totalBarWidth

		v.drawSingleBar(img, barX, barH, h)
		v.drawCap(img, barX, int(caps[i]), h)
	}
}

// drawSingleBar renders one bar with gradient coloring.
func (v *Visualizer) drawSingleBar(img *image.RGBA, barX, barH, h int) {
	maxX := img.Bounds().Max.X - v.cachedPaddingR

	for y := 0; y < barH && y < v.cachedEffectiveH; y++ {
		screenY := h - 1 - y
		col := v.getGradientColor(float64(y) / float64(v.cachedEffectiveH))

		for x := barX; x < barX+v.cachedBarWidth && x < maxX; x++ {
			if x >= v.cachedPaddingL {
				img.Set(x, screenY, col)
			}
		}
	}
}

// drawCap renders the falling cap for a bar.
func (v *Visualizer) drawCap(img *image.RGBA, barX, capY, h int) {
	if capY <= 0 || capY >= v.cachedEffectiveH {
		return
	}

	maxX := img.Bounds().Max.X - v.cachedPaddingR
	screenY := h - 1 - capY

	for cy := 0; cy < v.capHeight && screenY+cy < h && screenY+cy >= v.cachedPaddingT; cy++ {
		for x := barX; x < barX+v.cachedBarWidth && x < maxX; x++ {
			if x >= v.cachedPaddingL {
				img.Set(x, screenY+cy, color.White)
			}
		}
	}
}
