package visualizer

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

// FFT Spectrum 3D visualizer constants.
const (
	fftSpectrumNumBars       = 32 // Number of frequency bars
	fftSpectrumTrailCount    = 12 // Number of trail snapshots
	fftSpectrumPerspectiveX  = 5  // X offset per trail row (pixels)
	fftSpectrumPerspectiveY  = 4  // Y offset per trail row (pixels)
	fftSpectrumScanlineGap   = 3  // Scanline every N rows
	fftSpectrumScanlineAlpha = 40 // Alpha value for scanline darkening
)

// FFTSpectrum is a widget that displays a 3D-style FFT spectrum visualization.
// It renders frequency bars with a trailing effect that creates depth perception
// through simulated perspective projection.
type FFTSpectrum struct {
	BaseVisualizer

	numBars      int
	trailCount   int
	trailHistory [][]float32 // Ring buffer of historical bar heights
	trailHead    int         // Current insertion index
	loudness     float64     // Smoothed loudness for bloom effect

	// Layout cache
	lastWidth    int
	lastHeight   int
	cachedStartX int
	cachedStartY int
	cachedBarW   int
	cachedBarGap int
	cachedMaxH   int

	freq FrequencyAnalyzer
	draw DrawingUtils
}

// NewFFTSpectrum creates a new FFT spectrum 3D visualizer widget.
func NewFFTSpectrum() *FFTSpectrum {
	v := &FFTSpectrum{
		numBars:    fftSpectrumNumBars,
		trailCount: fftSpectrumTrailCount,
		loudness:   0.0,
	}

	// Initialize trail history ring buffer
	v.trailHistory = make([][]float32, v.trailCount)
	for i := range v.trailHistory {
		v.trailHistory[i] = make([]float32, v.numBars)
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// Reset clears the visualizer state.
func (v *FFTSpectrum) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	v.loudness = 0
	v.trailHead = 0
	v.Time = 0

	// Clear trail history
	for i := range v.trailHistory {
		for j := range v.trailHistory[i] {
			v.trailHistory[i][j] = 0
		}
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the FFT spectrum with 3D trail effect.
func (v *FFTSpectrum) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Dark background with slight blue tint
	bgColor := color.RGBA{R: 8, G: 10, B: 18, A: 255}
	v.draw.FillBackground(img, bgColor)

	if w == 0 || h == 0 {
		return img
	}

	v.Mu.Lock()
	fftData := v.FFTData
	v.Mu.Unlock()

	// Recalculate layout if size changed
	if v.lastWidth != w || v.lastHeight != h {
		v.recalculateLayout(w, h)
	}

	// Calculate bar heights from FFT data
	barHeights := v.freq.CalculateBarHeights(fftData, v.numBars, float64(v.cachedMaxH))

	// Calculate loudness (RMS of all bins) for bloom effect
	loudness := v.calculateLoudness(fftData)

	v.Mu.Lock()
	// Smooth loudness
	v.loudness = v.loudness*0.8 + loudness*0.2
	smoothedLoudness := v.loudness

	// Push current bar heights to trail ring buffer
	copy(v.trailHistory[v.trailHead], barHeights)
	v.trailHead = (v.trailHead + 1) % v.trailCount

	// Copy trail history for rendering
	trailCopy := make([][]float32, v.trailCount)
	for i := range v.trailHistory {
		trailCopy[i] = make([]float32, v.numBars)
		copy(trailCopy[i], v.trailHistory[i])
	}
	currentHead := v.trailHead
	v.Mu.Unlock()

	// Draw trails from back (oldest) to front (newest)
	v.drawTrails(img, trailCopy, currentHead, smoothedLoudness)

	// Apply scanlines overlay
	v.drawScanlines(img)

	return img
}

// recalculateLayout computes layout values based on current size.
func (v *FFTSpectrum) recalculateLayout(w, h int) {
	v.lastWidth = w
	v.lastHeight = h

	// Account for perspective offset space
	perspectiveOffsetX := fftSpectrumPerspectiveX * (v.trailCount - 1)
	perspectiveOffsetY := fftSpectrumPerspectiveY * (v.trailCount - 1)

	// Calculate effective area
	padding := 20
	effectiveW := w - padding*2 - perspectiveOffsetX
	effectiveH := h - padding*2 - perspectiveOffsetY

	if effectiveW <= 0 || effectiveH <= 0 {
		v.cachedBarW = 0
		return
	}

	// Calculate bar dimensions
	minGap := 2
	totalGaps := (v.numBars - 1) * minGap
	availableWidth := effectiveW - totalGaps

	v.cachedBarW = max(availableWidth/v.numBars, 2)

	// Recalculate gap to center bars
	v.cachedBarGap = minGap
	if v.numBars > 1 {
		usedWidth := v.numBars*v.cachedBarW + (v.numBars-1)*minGap
		extraSpace := effectiveW - usedWidth
		if extraSpace > 0 {
			v.cachedBarGap = minGap + extraSpace/(v.numBars-1)
		}
	}

	// Starting position (bottom-right, with room for perspective offset to top-left)
	v.cachedStartX = padding + perspectiveOffsetX
	v.cachedStartY = h - padding
	v.cachedMaxH = effectiveH
}

// calculateLoudness computes overall loudness from FFT data.
func (v *FFTSpectrum) calculateLoudness(fftData []float32) float64 {
	if len(fftData) == 0 {
		return 0
	}

	var sum float64
	for _, val := range fftData {
		sum += float64(val * val)
	}
	return math.Sqrt(sum / float64(len(fftData)))
}

// drawTrails renders all trail rows from back to front.
func (v *FFTSpectrum) drawTrails(img *image.RGBA, trails [][]float32, head int, loudness float64) {
	if v.cachedBarW == 0 {
		return
	}

	// Draw from oldest (back) to newest (front)
	for age := v.trailCount - 1; age >= 0; age-- {
		// Calculate trail index in ring buffer
		trailIdx := (head - 1 - age + v.trailCount) % v.trailCount

		// Age factor: 0 = oldest, 1 = newest
		ageFactor := 1.0 - float64(age)/float64(v.trailCount-1)

		// Perspective offset: older trails are further top-left
		offsetX := fftSpectrumPerspectiveX * age
		offsetY := fftSpectrumPerspectiveY * age

		// Depth scale: older trails appear smaller
		depthScale := 0.7 + 0.3*ageFactor

		v.drawTrailRow(img, trails[trailIdx], ageFactor, depthScale, offsetX, offsetY, loudness)
	}
}

// drawTrailRow renders a single trail row of bars.
func (v *FFTSpectrum) drawTrailRow(img *image.RGBA, heights []float32, ageFactor, depthScale float64, offsetX, offsetY int, loudness float64) {
	bounds := img.Bounds()

	// Bloom boost based on loudness
	bloomBoost := 1.0 + loudness*0.4

	for i := range v.numBars {
		if i >= len(heights) {
			break
		}
		barH := int(float64(heights[i]) * depthScale)
		if barH <= 0 {
			continue
		}

		// Frequency ratio for color (0 = bass, 1 = treble)
		freqRatio := float64(i) / float64(v.numBars-1)

		// Calculate bar position (adjusted for perspective)
		barX := v.cachedStartX + i*(v.cachedBarW+v.cachedBarGap) - offsetX
		barY := v.cachedStartY - offsetY

		// Calculate scaled bar width
		scaledBarW := max(int(float64(v.cachedBarW)*depthScale), 1)

		// Calculate color based on age and frequency
		col := v.calculateBarColor(ageFactor, freqRatio, bloomBoost)

		// Draw the bar
		for dy := range barH {
			screenY := barY - dy
			if screenY < bounds.Min.Y || screenY >= bounds.Max.Y {
				continue
			}

			for dx := range scaledBarW {
				screenX := barX + dx
				if screenX >= bounds.Min.X && screenX < bounds.Max.X {
					// Alpha blend with existing pixel
					v.setPixelWithAlpha(img, screenX, screenY, col)
				}
			}
		}
	}
}

// calculateBarColor computes the color for a bar based on age and frequency.
func (v *FFTSpectrum) calculateBarColor(ageFactor, freqRatio, bloomBoost float64) color.RGBA {
	// Red tracks trail age (newer = brighter red)
	r := clampFloat(255 * ageFactor * bloomBoost)

	// Green inversely tracks frequency (bass = green, treble = less green)
	g := clampFloat(255 * (1.0 - freqRatio) * ageFactor * 0.7)

	// Blue is constant low for depth
	b := clampFloat(40 * ageFactor)

	// Alpha decays with age (exponential decay for trail effect)
	a := clampFloat(255 * math.Pow(ageFactor, 1.3))

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
}

// setPixelWithAlpha blends a color onto the image with alpha compositing.
func (v *FFTSpectrum) setPixelWithAlpha(img *image.RGBA, x, y int, col color.RGBA) {
	if col.A == 0 {
		return
	}

	existing := img.RGBAAt(x, y)

	// Simple alpha blending
	alpha := float64(col.A) / 255.0
	invAlpha := 1.0 - alpha

	newR := uint8(float64(col.R)*alpha + float64(existing.R)*invAlpha)
	newG := uint8(float64(col.G)*alpha + float64(existing.G)*invAlpha)
	newB := uint8(float64(col.B)*alpha + float64(existing.B)*invAlpha)

	img.SetRGBA(x, y, color.RGBA{R: newR, G: newG, B: newB, A: 255})
}

// drawScanlines overlays semi-transparent horizontal lines for CRT aesthetic.
func (v *FFTSpectrum) drawScanlines(img *image.RGBA) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Scanline darkening factor
	darkenFactor := 1.0 - float64(fftSpectrumScanlineAlpha)/255.0*0.5

	for y := 0; y < h; y += fftSpectrumScanlineGap {
		for x := range w {
			existing := img.RGBAAt(x, y)

			// Darken the pixel slightly
			newR := uint8(float64(existing.R) * darkenFactor)
			newG := uint8(float64(existing.G) * darkenFactor)
			newB := uint8(float64(existing.B) * darkenFactor)

			img.SetRGBA(x, y, color.RGBA{R: newR, G: newG, B: newB, A: 255})
		}
	}
}

// clampFloat clamps a float64 value to a range.
func clampFloat(val float64) float64 {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return val
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*FFTSpectrum)(nil)
