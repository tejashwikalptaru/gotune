package visualizer

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

const (
	plasmaBaseScale   = 0.02 // Base scale for plasma pattern
	plasmaBaseSpeed   = 0.03 // Base animation speed
	plasmaPaletteSize = 256  // Size of color palette
	plasmaDownscale   = 2    // Render at lower resolution for performance
)

// Plasma is a widget that displays a classic plasma effect.
// Audio frequencies modulate the plasma parameters.
type Plasma struct {
	BaseVisualizer

	// Color palette (pre-computed for performance)
	palette []color.RGBA

	freq FrequencyAnalyzer
}

// NewPlasma creates a new plasma visualizer widget.
func NewPlasma() *Plasma {
	v := &Plasma{
		palette: make([]color.RGBA, plasmaPaletteSize),
	}
	v.BassAvg = 0.5
	v.MidAvg = 0.5
	v.HighAvg = 0.5

	// Initialize default palette
	v.generatePalette(0)

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// generatePalette creates a smooth color palette with hue offset.
func (v *Plasma) generatePalette(hueOffset float64) {
	for i := 0; i < plasmaPaletteSize; i++ {
		t := float64(i) / float64(plasmaPaletteSize)

		r := math.Sin(t*math.Pi*2+hueOffset)*0.5 + 0.5
		g := math.Sin(t*math.Pi*2+hueOffset+math.Pi*2/3)*0.5 + 0.5
		b := math.Sin(t*math.Pi*2+hueOffset+math.Pi*4/3)*0.5 + 0.5

		v.palette[i] = color.RGBA{
			R: uint8(r * 255),
			G: uint8(g * 255),
			B: uint8(b * 255),
			A: 255,
		}
	}
}

// Reset clears the visualizer state.
func (v *Plasma) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	v.Time = 0
	v.BassAvg = 0.5
	v.MidAvg = 0.5
	v.HighAvg = 0.5
	v.generatePalette(0)
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the plasma effect.
func (v *Plasma) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	if w == 0 || h == 0 {
		return img
	}

	v.Mu.Lock()
	fftData := v.FFTData
	v.Mu.Unlock()

	// Calculate audio-reactive values
	bass := v.freq.CalculateBass(fftData, 0.5)
	mid := v.freq.CalculateMid(fftData, 0.5)
	high := v.freq.CalculateHigh(fftData, 0.5)

	// Smooth the values
	v.Mu.Lock()
	v.BassAvg = v.BassAvg*0.85 + bass*0.15
	v.MidAvg = v.MidAvg*0.85 + mid*0.15
	v.HighAvg = v.HighAvg*0.85 + high*0.15
	smoothedBass := v.BassAvg
	smoothedMid := v.MidAvg
	smoothedHigh := v.HighAvg
	v.Mu.Unlock()

	// Update animation time (speed based on mid frequencies)
	v.Mu.Lock()
	v.Time += plasmaBaseSpeed + smoothedMid*0.05
	time := v.Time

	// Regenerate palette based on high frequencies for color cycling
	v.generatePalette(smoothedHigh * math.Pi * 2)
	palette := v.palette
	v.Mu.Unlock()

	// Scale based on bass (zooms in/out)
	scale := plasmaBaseScale * (1.0 + smoothedBass*0.5)

	// Contrast based on overall amplitude
	contrast := 0.5 + (smoothedBass+smoothedMid+smoothedHigh)*0.3

	// Render plasma at lower resolution for performance
	renderW := w / plasmaDownscale
	renderH := h / plasmaDownscale

	for py := 0; py < renderH; py++ {
		for px := 0; px < renderW; px++ {
			x := float64(px) * scale
			y := float64(py) * scale

			value := v.plasmaValue(x, y, time, smoothedBass, smoothedMid)

			// Apply contrast
			value = (value-0.5)*contrast + 0.5
			if value < 0 {
				value = 0
			}
			if value > 1 {
				value = 1
			}

			// Map to palette
			paletteIndex := int(value * float64(plasmaPaletteSize-1))
			if paletteIndex < 0 {
				paletteIndex = 0
			}
			if paletteIndex >= plasmaPaletteSize {
				paletteIndex = plasmaPaletteSize - 1
			}

			col := palette[paletteIndex]

			// Fill the upscaled pixels
			for dy := 0; dy < plasmaDownscale; dy++ {
				for dx := 0; dx < plasmaDownscale; dx++ {
					screenX := px*plasmaDownscale + dx
					screenY := py*plasmaDownscale + dy
					if screenX < w && screenY < h {
						img.Set(screenX, screenY, col)
					}
				}
			}
		}
	}

	return img
}

// plasmaValue computes the plasma value at a given point.
func (v *Plasma) plasmaValue(x, y, time, bass, mid float64) float64 {
	// Wave 1: Horizontal waves
	v1 := math.Sin(x*10 + time*2)

	// Wave 2: Vertical waves
	v2 := math.Sin(y*10 + time*3)

	// Wave 3: Diagonal waves
	v3 := math.Sin((x+y)*7 + time*1.5)

	// Wave 4: Circular waves from center (modulated by bass)
	cx := x - 5*(1+bass*0.5)
	cy := y - 5*(1+bass*0.5)
	dist := math.Sqrt(cx*cx + cy*cy)
	v4 := math.Sin(dist*8 - time*4)

	// Wave 5: Secondary circular waves (modulated by mid)
	cx2 := x + 3*math.Sin(time)
	cy2 := y + 3*math.Cos(time*0.7)
	dist2 := math.Sqrt(cx2*cx2 + cy2*cy2)
	v5 := math.Sin(dist2*6 + time*2)

	// Wave 6: Turbulence (adds complexity)
	v6 := math.Sin(x*3+math.Sin(y*4+time)) * mid

	// Combine all waves and normalize to 0-1
	combined := (v1 + v2 + v3 + v4 + v5 + v6) / 6.0
	return combined*0.5 + 0.5
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Plasma)(nil)
