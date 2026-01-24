package visualizer

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

const (
	tunnelNumRings     = 20
	tunnelMaxRadius    = 1.5 // Max radius multiplier of min dimension
	tunnelBaseSpeed    = 0.02
	tunnelWaveSegments = 60 // Number of segments per ring
)

// Tunnel is a widget that displays a tunnel/wormhole effect.
// Audio frequencies control ring expansion speed.
type Tunnel struct {
	BaseVisualizer

	freq FrequencyAnalyzer
}

// NewTunnel creates a new tunnel visualizer widget.
func NewTunnel() *Tunnel {
	v := &Tunnel{}
	v.BassAvg = 0.1

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// render draws the tunnel effect.
func (v *Tunnel) render(w, h int) image.Image {
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

	v.Mu.Lock()
	fftData := v.FFTData
	v.Mu.Unlock()

	// Calculate audio-reactive values
	bass := v.freq.CalculateBass(fftData, 0.1)
	mid := v.freq.CalculateMid(fftData, 0)
	high := v.freq.CalculateHigh(fftData, 0)

	// Smooth the values
	v.Mu.Lock()
	v.BassAvg = v.BassAvg*0.8 + bass*0.2
	v.MidAvg = v.MidAvg*0.8 + mid*0.2
	v.HighAvg = v.HighAvg*0.8 + high*0.2
	smoothedBass := v.BassAvg
	smoothedMid := v.MidAvg
	smoothedHigh := v.HighAvg
	v.Mu.Unlock()

	// Update animation time (speed based on bass)
	v.Mu.Lock()
	v.Time += tunnelBaseSpeed + smoothedBass*0.05
	time := v.Time
	v.Mu.Unlock()

	centerX := float64(w) / 2
	centerY := float64(h) / 2
	minDim := math.Min(float64(w), float64(h))

	// Draw rings from back to front (largest to smallest)
	for i := tunnelNumRings - 1; i >= 0; i-- {
		ringPhase := math.Mod(float64(i)/float64(tunnelNumRings)+time, 1.0)
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
func (v *Tunnel) drawDistortedRing(img *image.RGBA, cx, cy, radius, waveAmp, time float64, col color.RGBA, depth float64) {
	bounds := img.Bounds()
	thickness := int(math.Max(1, 3*depth))

	for i := 0; i < tunnelWaveSegments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(tunnelWaveSegments)
		nextAngle := 2 * math.Pi * float64(i+1) / float64(tunnelWaveSegments)

		wave := math.Sin(angle*4+time*10) * waveAmp
		nextWave := math.Sin(nextAngle*4+time*10) * waveAmp

		r := radius + wave
		nextR := radius + nextWave

		x1 := int(cx + math.Cos(angle)*r)
		y1 := int(cy + math.Sin(angle)*r)
		x2 := int(cx + math.Cos(nextAngle)*nextR)
		y2 := int(cy + math.Sin(nextAngle)*nextR)

		v.drawLineSegment(img, bounds, x1, y1, x2, y2, thickness, col)
	}
}

// drawLineSegment draws a line segment with a given thickness.
func (v *Tunnel) drawLineSegment(img *image.RGBA, bounds image.Rectangle, x1, y1, x2, y2, thickness int, col color.RGBA) {
	dx := x2 - x1
	dy := y2 - y1
	steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy)))) + 1

	for s := 0; s <= steps; s++ {
		t := float64(s) / float64(steps)
		px := int(float64(x1) + float64(dx)*t)
		py := int(float64(y1) + float64(dy)*t)

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

// drawLightning draws lightning arcs emanating from center.
func (v *Tunnel) drawLightning(img *image.RGBA, cx, cy, size, intensity, time float64) {
	bounds := img.Bounds()

	numBolts := 3
	for bolt := 0; bolt < numBolts; bolt++ {
		baseAngle := (float64(bolt)/float64(numBolts))*2*math.Pi + time*2

		x := cx
		y := cy
		segLen := 15.0

		numSegments := int(size / segLen / 3)

		for seg := 0; seg < numSegments; seg++ {
			angle := baseAngle + math.Sin(time*20+float64(seg)*0.5)*0.5

			nx := x + math.Cos(angle)*segLen
			ny := y + math.Sin(angle)*segLen

			brightness := intensity * (1.0 - float64(seg)/float64(numSegments))
			col := color.RGBA{
				R: uint8(200 * brightness),
				G: uint8(200 * brightness),
				B: uint8(255 * brightness),
				A: 255,
			}

			v.drawLineSegment(img, bounds, int(x), int(y), int(nx), int(ny), 1, col)

			x, y = nx, ny
		}
	}
}

// tunnelColor returns a color for the tunnel based on intensity and hue.
func (v *Tunnel) tunnelColor(intensity, hueShift float64) color.RGBA {
	baseHue := 0.7 + hueShift
	if baseHue > 1.0 {
		baseHue -= 1.0
	}

	r, g, b := HSLToRGB(baseHue, 0.8, intensity*0.5)

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Tunnel)(nil)
