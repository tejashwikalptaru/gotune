package visualizer

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2/canvas"
)

// Spectrum is a widget that displays audio spectrum bars.
// It uses logarithmic frequency distribution for better visual representation.
type Spectrum struct {
	BaseVisualizer

	capHeights []float32 // Falling cap animation heights
	numBars    int

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

	freq FrequencyAnalyzer
	draw DrawingUtils
}

// NewSpectrum creates a new spectrum bar visualizer widget with the specified number of bars.
func NewSpectrum(numBars int) *Spectrum {
	v := &Spectrum{
		numBars:      numBars,
		capHeights:   make([]float32, numBars),
		capHeight:    2,
		capFalloff:   2.0,
		minGap:       2,
		paddingTop:   10,
		paddingLeft:  10,
		paddingRight: 10,
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// Reset clears the visualizer state.
func (v *Spectrum) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	for i := range v.capHeights {
		v.capHeights[i] = 0
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render is the raster generator function that draws the visualizer.
func (v *Spectrum) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	v.draw.FillBackground(img, color.Black)

	v.Mu.Lock()
	fftData := v.FFTData
	caps := v.capHeights
	v.Mu.Unlock()

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
	barHeights := v.freq.CalculateBarHeights(fftData, v.numBars, float64(v.cachedEffectiveH))

	// Update cap heights with falling animation
	v.Mu.Lock()
	v.updateCapHeights(barHeights, caps)
	v.Mu.Unlock()

	// Draw all bars and caps
	v.drawBars(img, barHeights, caps, h)

	return img
}

// recalculateLayout computes and caches size-dependent layout values.
func (v *Spectrum) recalculateLayout(w, h int) {
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

// updateCapHeights updates cap positions with falling animation.
func (v *Spectrum) updateCapHeights(barHeights []float32, caps []float32) {
	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := barHeights[i]
		if barH > caps[i] {
			caps[i] = barH
		} else {
			caps[i] -= v.capFalloff
			if caps[i] < 0 {
				caps[i] = 0
			}
		}
	}
}

// drawBars renders all bars and their caps to the image.
func (v *Spectrum) drawBars(img *image.RGBA, barHeights []float32, caps []float32, h int) {
	totalBarWidth := v.cachedBarWidth + v.cachedActualGap

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := int(barHeights[i])
		barX := v.cachedStartX + i*totalBarWidth

		v.drawSingleBar(img, barX, barH, h)
		v.drawCap(img, barX, int(caps[i]), h)
	}
}

// drawSingleBar renders one bar with gradient coloring.
func (v *Spectrum) drawSingleBar(img *image.RGBA, barX, barH, h int) {
	maxX := img.Bounds().Max.X - v.cachedPaddingR

	for y := 0; y < barH && y < v.cachedEffectiveH; y++ {
		screenY := h - 1 - y
		col := v.draw.GetGradientColor(float64(y) / float64(v.cachedEffectiveH))

		for x := barX; x < barX+v.cachedBarWidth && x < maxX; x++ {
			if x >= v.cachedPaddingL {
				img.Set(x, screenY, col)
			}
		}
	}
}

// drawCap renders the falling cap for a bar.
func (v *Spectrum) drawCap(img *image.RGBA, barX, capY, h int) {
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

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Spectrum)(nil)
