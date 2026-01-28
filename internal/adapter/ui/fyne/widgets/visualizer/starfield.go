package visualizer

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"fyne.io/fyne/v2/canvas"
)

const (
	starfieldNumStars      = 200
	starfieldMaxZ          = 1000.0
	starfieldMinZ          = 1.0
	starfieldBaseSpeed     = 5.0
	starfieldMaxTrailLen   = 20.0
	starfieldSpawnDistance = 800.0
)

// Star represents a single star in the starfield.
type Star struct {
	x, y, z    float64 // 3D position
	prevX      float64 // Previous screen X for trail
	prevY      float64 // Previous screen Y for trail
	brightness float64 // Base brightness
}

// Starfield is a widget that displays a warp-speed starfield effect.
// Star speed is controlled by bass frequencies from the audio.
type Starfield struct {
	BaseVisualizer

	stars []Star

	freq FrequencyAnalyzer
}

// NewStarfield creates a new starfield visualizer widget.
func NewStarfield() *Starfield {
	v := &Starfield{
		stars: make([]Star, starfieldNumStars),
	}
	v.BassAvg = 0.1

	// Initialize stars with random positions
	for i := range v.stars {
		v.initStar(&v.stars[i], true)
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// initStar initializes a star with a random position.
// nolint:gosec // G404 - weak random is fine for visual effects
func (v *Starfield) initStar(s *Star, randomZ bool) {
	spread := 1500.0
	s.x = (rand.Float64() - 0.5) * spread
	s.y = (rand.Float64() - 0.5) * spread

	if randomZ {
		s.z = rand.Float64()*starfieldMaxZ + starfieldMinZ
	} else {
		s.z = starfieldSpawnDistance + rand.Float64()*200
	}

	s.brightness = 0.5 + rand.Float64()*0.5
	s.prevX = 0
	s.prevY = 0
}

// Reset clears the visualizer state.
func (v *Starfield) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	v.BassAvg = 0.1
	v.MidAvg = 0
	for i := range v.stars {
		v.initStar(&v.stars[i], true)
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the starfield.
func (v *Starfield) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with black background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.Black)
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

	// Smooth the values
	v.Mu.Lock()
	v.BassAvg = v.BassAvg*0.7 + bass*0.3
	v.MidAvg = v.MidAvg*0.7 + mid*0.3
	smoothedBass := v.BassAvg
	smoothedMid := v.MidAvg
	v.Mu.Unlock()

	// Speed based on bass
	speed := starfieldBaseSpeed + smoothedBass*30.0

	// Trail length based on speed
	trailLen := math.Min(speed*0.5, starfieldMaxTrailLen)

	// Center of screen
	centerX := float64(w) / 2
	centerY := float64(h) / 2

	v.Mu.Lock()
	defer v.Mu.Unlock()

	for i := range v.stars {
		s := &v.stars[i]

		// Save previous position for trail
		if s.z > starfieldMinZ {
			s.prevX = s.x / s.z * 300
			s.prevY = s.y / s.z * 300
		}

		// Move star toward viewer
		s.z -= speed

		// Respawn if the star passed the viewer
		if s.z < starfieldMinZ {
			v.initStar(s, false)
			continue
		}

		// Project 3D to 2D
		scale := 300.0 / s.z
		screenX := s.x*scale + centerX
		screenY := s.y*scale + centerY

		// Skip if outside the screen
		if screenX < 0 || screenX >= float64(w) || screenY < 0 || screenY >= float64(h) {
			v.initStar(s, false)
			continue
		}

		// Calculate brightness based on depth
		depthFactor := 1.0 - (s.z / starfieldMaxZ)
		brightness := s.brightness * depthFactor
		brightness = math.Min(brightness+smoothedMid*0.3, 1.0)

		// Star color - white near, blue far
		var col color.RGBA
		switch {
		case s.z < starfieldMaxZ*0.3:
			col = color.RGBA{
				R: uint8(255 * brightness),
				G: uint8(255 * brightness),
				B: uint8(200 * brightness),
				A: 255,
			}
		case s.z < starfieldMaxZ*0.6:
			col = color.RGBA{
				R: uint8(255 * brightness),
				G: uint8(255 * brightness),
				B: uint8(255 * brightness),
				A: 255,
			}
		default:
			col = color.RGBA{
				R: uint8(150 * brightness),
				G: uint8(180 * brightness),
				B: uint8(255 * brightness),
				A: 255,
			}
		}

		// Draw star trail if moving fast enough
		if trailLen > 1 && s.prevX != 0 && s.prevY != 0 {
			prevScreenX := s.prevX + centerX
			prevScreenY := s.prevY + centerY
			v.drawLine(img, int(prevScreenX), int(prevScreenY), int(screenX), int(screenY), col, brightness)
		}

		// Draw star point (larger when closer)
		starSize := int(math.Max(1, 3*depthFactor))
		v.drawStar(img, int(screenX), int(screenY), starSize, col)
	}

	return img
}

// drawStar draws a star at the given position with size.
func (v *Starfield) drawStar(img *image.RGBA, x, y, size int, col color.RGBA) {
	bounds := img.Bounds()
	for dy := -size / 2; dy <= size/2; dy++ {
		for dx := -size / 2; dx <= size/2; dx++ {
			px, py := x+dx, y+dy
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				img.Set(px, py, col)
			}
		}
	}
}

// drawLine draws a line between two points with a fading effect.
func (v *Starfield) drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.RGBA, brightness float64) {
	bounds := img.Bounds()

	dx := x2 - x1
	dy := y2 - y1
	steps := int(math.Max(math.Abs(float64(dx)), math.Abs(float64(dy))))

	if steps == 0 {
		return
	}

	xInc := float64(dx) / float64(steps)
	yInc := float64(dy) / float64(steps)

	x := float64(x1)
	y := float64(y1)

	for i := 0; i <= steps; i++ {
		px, py := int(x), int(y)
		if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
			fade := float64(i) / float64(steps) * brightness
			fadedCol := color.RGBA{
				R: uint8(float64(col.R) * fade),
				G: uint8(float64(col.G) * fade),
				B: uint8(float64(col.B) * fade),
				A: 255,
			}
			img.Set(px, py, fadedCol)
		}
		x += xInc
		y += yInc
	}
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Starfield)(nil)
