package visualizer

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

const (
	graphPaddingTop   = 10
	graphPaddingLeft  = 10
	graphPaddingRight = 10
	graphLineWidth    = 2
	graphFillAlpha    = 0.3
	graphPeakDecay    = 2.0
)

// Graph is a widget that displays audio spectrum as a continuous line/filled area.
// Based on audioMotion-analyzer's graph mode.
type Graph struct {
	BaseVisualizer

	numBars     int
	peakHeights []float32 // Peak hold values for each point

	freq FrequencyAnalyzer
	draw DrawingUtils
}

// NewGraph creates a new graph visualizer widget.
func NewGraph(numBars ...int) *Graph {
	bars := 64
	if len(numBars) > 0 && numBars[0] > 0 {
		bars = numBars[0]
	}

	v := &Graph{
		numBars:     bars,
		peakHeights: make([]float32, bars),
	}

	v.Raster = canvas.NewRaster(v.render)
	v.ExtendBaseWidget(v)

	return v
}

// Reset clears the visualizer state.
func (v *Graph) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	for i := range v.peakHeights {
		v.peakHeights[i] = 0
	}
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// render draws the graph visualizer.
func (v *Graph) render(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	v.draw.FillBackground(img, color.Black)

	v.Mu.Lock()
	fftData := v.FFTData
	peaks := v.peakHeights
	v.Mu.Unlock()

	if len(fftData) == 0 || w == 0 || h == 0 {
		return img
	}

	// Calculate effective drawing area
	effectiveW := w - graphPaddingLeft - graphPaddingRight
	effectiveH := h - graphPaddingTop
	baseY := h - 1

	if effectiveW <= 0 || effectiveH <= 0 {
		return img
	}

	// Calculate bar heights
	barHeights := v.freq.CalculateBarHeights(fftData, v.numBars, float64(effectiveH))

	// Calculate points
	points := make([]point, v.numBars)
	barSpacing := float64(effectiveW) / float64(v.numBars-1)
	if v.numBars == 1 {
		barSpacing = float64(effectiveW)
	}

	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		x := float64(graphPaddingLeft) + float64(i)*barSpacing
		y := float64(baseY) - float64(barHeights[i])
		points[i] = point{x, y}
	}

	// Update peak heights
	v.Mu.Lock()
	v.updatePeakHeights(barHeights, peaks)
	v.Mu.Unlock()

	// Calculate peak points
	peakPoints := make([]point, v.numBars)
	for i := 0; i < v.numBars && i < len(peaks); i++ {
		x := float64(graphPaddingLeft) + float64(i)*barSpacing
		y := float64(baseY) - float64(peaks[i])
		peakPoints[i] = point{x, y}
	}

	// Draw filled area under the curve
	v.drawFilledArea(img, points, baseY)

	// Draw main line
	lineColor := color.RGBA{R: 0, G: 255, B: 100, A: 255}
	v.drawSmoothLine(img, points, lineColor, graphLineWidth)

	// Draw peak line
	peakColor := color.RGBA{R: 255, G: 255, B: 255, A: 180}
	v.drawSmoothLine(img, peakPoints, peakColor, 1)

	return img
}

// point represents a 2D coordinate.
type point struct {
	x, y float64
}

// updatePeakHeights updates peak positions with falling animation.
func (v *Graph) updatePeakHeights(barHeights []float32, peaks []float32) {
	for i := 0; i < v.numBars && i < len(barHeights); i++ {
		barH := barHeights[i]
		if barH > peaks[i] {
			peaks[i] = barH
		} else {
			peaks[i] -= graphPeakDecay
			if peaks[i] < 0 {
				peaks[i] = 0
			}
		}
	}
}

// drawFilledArea draws a filled area under the curve with gradient and transparency.
func (v *Graph) drawFilledArea(img *image.RGBA, points []point, baseY int) {
	if len(points) < 2 {
		return
	}

	bounds := img.Bounds()

	// Draw vertical lines from each point down to baseline with gradient
	for i := 0; i < len(points)-1; i++ {
		p1 := points[i]
		p2 := points[i+1]

		// Interpolate between points
		steps := int(math.Abs(p2.x-p1.x)) + 1
		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			x := int(p1.x + (p2.x-p1.x)*t)
			topY := int(p1.y + (p2.y-p1.y)*t)

			if x < bounds.Min.X || x >= bounds.Max.X {
				continue
			}

			// Draw vertical line from top to baseline with gradient
			for y := topY; y < baseY; y++ {
				if y < bounds.Min.Y || y >= bounds.Max.Y {
					continue
				}

				// Calculate gradient color based on height
				heightRatio := float64(baseY-y) / float64(baseY-int(topY))
				col := v.draw.GetGradientColor(heightRatio)

				// Apply alpha for fill
				alpha := uint8(float64(col.A) * graphFillAlpha)
				fillCol := color.RGBA{R: col.R, G: col.G, B: col.B, A: alpha}
				img.Set(x, y, fillCol)
			}
		}
	}
}

// drawSmoothLine draws a connected line through all points.
func (v *Graph) drawSmoothLine(img *image.RGBA, points []point, col color.RGBA, thickness int) {
	if len(points) < 2 {
		return
	}

	for i := 0; i < len(points)-1; i++ {
		v.draw.DrawThickLine(img, points[i].x, points[i].y, points[i+1].x, points[i+1].y, thickness, col)
	}
}

// Verify interface implementation at compile time.
var _ MusicVisualizer = (*Graph)(nil)
