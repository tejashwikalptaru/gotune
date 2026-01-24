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
	tunnelNumRings     = 20
	tunnelMaxRadius    = 1.5 // Max radius multiplier of min dimension
	tunnelBaseSpeed    = 0.02
	tunnelWaveSegments = 60 // Number of segments per ring
)

// TunnelVisualizer is a widget that displays a tunnel/wormhole effect.
// Audio frequencies control ring expansion speed.
type TunnelVisualizer struct {
	widget.BaseWidget

	raster  *canvas.Raster
	fftData []float32
	mu      sync.RWMutex

	// Animation state
	time    float64 // Animation time
	bassAvg float64 // Smoothed bass
	midAvg  float64 // Smoothed mid
	highAvg float64 // Smoothed high
}

// NewTunnelVisualizer creates a new tunnel visualizer widget.
func NewTunnelVisualizer() *TunnelVisualizer {
	v := &TunnelVisualizer{
		time:    0,
		bassAvg: 0.1,
	}

	v.raster = canvas.NewRaster(v.draw)
	v.ExtendBaseWidget(v)

	return v
}

// CreateRenderer implements fyne.Widget.
func (v *TunnelVisualizer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.raster)
}

// MinSize returns the minimum size of the visualizer.
func (v *TunnelVisualizer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// UpdateFFT updates the visualizer with new FFT data.
func (v *TunnelVisualizer) UpdateFFT(data []float32) {
	v.mu.Lock()
	v.fftData = data
	v.mu.Unlock()

	v.raster.Refresh()
}

// Reset clears the visualizer state.
func (v *TunnelVisualizer) Reset() {
	v.mu.Lock()
	v.fftData = nil
	v.time = 0
	v.bassAvg = 0.1
	v.midAvg = 0
	v.highAvg = 0
	v.mu.Unlock()

	v.raster.Refresh()
}

// draw renders the tunnel effect.
func (v *TunnelVisualizer) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with dark background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 5, G: 5, B: 15, A: 255})
		}
	}

	if w == 0 || h == 0 {
		return img
	}

	v.mu.Lock()
	fftData := v.fftData
	v.mu.Unlock()

	// Calculate audio-reactive values
	bass := v.calculateBass(fftData)
	mid := v.calculateMid(fftData)
	high := v.calculateHigh(fftData)

	// Smooth the values
	v.mu.Lock()
	v.bassAvg = v.bassAvg*0.8 + bass*0.2
	v.midAvg = v.midAvg*0.8 + mid*0.2
	v.highAvg = v.highAvg*0.8 + high*0.2
	smoothedBass := v.bassAvg
	smoothedMid := v.midAvg
	smoothedHigh := v.highAvg
	v.mu.Unlock()

	// Update animation time (speed based on bass)
	v.mu.Lock()
	v.time += tunnelBaseSpeed + smoothedBass*0.05
	time := v.time
	v.mu.Unlock()

	centerX := float64(w) / 2
	centerY := float64(h) / 2
	minDim := math.Min(float64(w), float64(h))

	// Draw rings from back to front (largest to smallest)
	for i := tunnelNumRings - 1; i >= 0; i-- {
		// Ring position cycles based on time
		ringPhase := math.Mod(float64(i)/float64(tunnelNumRings)+time, 1.0)

		// Radius based on ring phase (perspective: far rings are smaller)
		radius := ringPhase * minDim * tunnelMaxRadius / 2

		if radius < 5 {
			continue
		}

		// Color based on depth
		depth := 1.0 - ringPhase
		intensity := depth * 0.8

		// Hue shift based on high frequencies
		hueShift := smoothedHigh * 0.3

		// Base color with hue shift
		col := v.tunnelColor(intensity, hueShift)

		// Wall distortion based on mid-frequencies
		waveAmplitude := smoothedMid * radius * 0.15

		// Draw the ring with wave distortion
		v.drawDistortedRing(img, centerX, centerY, radius, waveAmplitude, time, col, depth)
	}

	// Add lightning effect on high bass
	if smoothedBass > 0.4 {
		v.drawLightning(img, centerX, centerY, minDim, smoothedBass, time)
	}

	return img
}

// drawDistortedRing draws a ring with wave distortion.
func (v *TunnelVisualizer) drawDistortedRing(img *image.RGBA, cx, cy, radius, waveAmp, time float64, col color.RGBA, depth float64) {
	bounds := img.Bounds()

	// Line thickness based on depth (closer = thicker)
	thickness := int(math.Max(1, 3*depth))

	for i := 0; i < tunnelWaveSegments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(tunnelWaveSegments)
		nextAngle := 2 * math.Pi * float64(i+1) / float64(tunnelWaveSegments)

		// Wave distortion
		wave := math.Sin(angle*4+time*10) * waveAmp
		nextWave := math.Sin(nextAngle*4+time*10) * waveAmp

		r := radius + wave
		nextR := radius + nextWave

		x1 := int(cx + math.Cos(angle)*r)
		y1 := int(cy + math.Sin(angle)*r)
		x2 := int(cx + math.Cos(nextAngle)*nextR)
		y2 := int(cy + math.Sin(nextAngle)*nextR)

		// Draw line segment
		v.drawLineSegment(img, bounds, x1, y1, x2, y2, thickness, col)
	}
}

// drawLineSegment draws a line segment with a given thickness.
func (v *TunnelVisualizer) drawLineSegment(img *image.RGBA, bounds image.Rectangle, x1, y1, x2, y2, thickness int, col color.RGBA) {
	dx := x2 - x1
	dy := y2 - y1
	steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy)))) + 1

	for s := 0; s <= steps; s++ {
		t := float64(s) / float64(steps)
		px := int(float64(x1) + float64(dx)*t)
		py := int(float64(y1) + float64(dy)*t)

		// Draw with thickness
		for ty := -thickness / 2; ty <= thickness/2; ty++ {
			for tx := -thickness / 2; tx <= thickness/2; tx++ {
				ppx, ppy := px+tx, py+ty
				if ppx >= bounds.Min.X && ppx < bounds.Max.X && ppy >= bounds.Min.Y && ppy < bounds.Max.Y {
					img.Set(ppx, ppy, col)
				}
			}
		}
	}
}

// drawLightning draws lightning arcs emanating from a center.
func (v *TunnelVisualizer) drawLightning(img *image.RGBA, cx, cy, size, intensity, time float64) {
	bounds := img.Bounds()

	numBolts := 3
	for bolt := 0; bolt < numBolts; bolt++ {
		// Random-ish angle based on time
		baseAngle := (float64(bolt)/float64(numBolts))*2*math.Pi + time*2

		x := cx
		y := cy
		segLen := 15.0

		numSegments := int(size / segLen / 3)

		for seg := 0; seg < numSegments; seg++ {
			// Add jitter to angle
			angle := baseAngle + math.Sin(time*20+float64(seg)*0.5)*0.5

			nx := x + math.Cos(angle)*segLen
			ny := y + math.Sin(angle)*segLen

			// Lightning color (bright white/blue)
			brightness := intensity * (1.0 - float64(seg)/float64(numSegments))
			col := color.RGBA{
				R: uint8(200 * brightness),
				G: uint8(200 * brightness),
				B: uint8(255 * brightness),
				A: 255,
			}

			// Draw segment
			v.drawLineSegment(img, bounds, int(x), int(y), int(nx), int(ny), 1, col)

			x, y = nx, ny
		}
	}
}

// tunnelColor returns a color for the tunnel based on intensity and hue.
func (v *TunnelVisualizer) tunnelColor(intensity, hueShift float64) color.RGBA {
	// Base purple/blue color with hue shift
	baseHue := 0.7 + hueShift // Purple-ish
	if baseHue > 1.0 {
		baseHue -= 1.0
	}

	r, g, b := hslToRGB(baseHue, 0.8, intensity*0.5)

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// hslToRGB converts HSL to RGB (h, s, l in 0-1 range).
func hslToRGB(h, s, l float64) (r, g, b float64) {
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

// calculateBass returns the average bass amplitude.
func (v *TunnelVisualizer) calculateBass(fftData []float32) float64 {
	if len(fftData) < 10 {
		return 0.1
	}

	var sum float64
	for i := 1; i < 10 && i < len(fftData); i++ {
		sum += float64(fftData[i])
	}
	return math.Sqrt(sum / 9.0)
}

// calculateMid returns the average mid amplitude.
func (v *TunnelVisualizer) calculateMid(fftData []float32) float64 {
	if len(fftData) < 50 {
		return 0
	}

	var sum float64
	count := 0
	for i := 10; i < 50 && i < len(fftData); i++ {
		sum += float64(fftData[i])
		count++
	}
	if count == 0 {
		return 0
	}
	return math.Sqrt(sum / float64(count))
}

// calculateHigh returns the average high frequency amplitude.
func (v *TunnelVisualizer) calculateHigh(fftData []float32) float64 {
	if len(fftData) < 100 {
		return 0
	}

	var sum float64
	count := 0
	for i := 50; i < len(fftData); i++ {
		sum += float64(fftData[i])
		count++
	}
	if count == 0 {
		return 0
	}
	return math.Sqrt(sum / float64(count))
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*TunnelVisualizer)(nil)
