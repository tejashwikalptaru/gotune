// Package visualizer provides audio visualization widgets for the GoTune application.
package visualizer

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// BaseVisualizer provides common functionality for all visualizers.
// It is designed to be embedded in concrete visualizer implementations.
type BaseVisualizer struct {
	widget.BaseWidget

	Raster  *canvas.Raster
	FFTData []float32
	Mu      sync.RWMutex

	// Frequency state (smoothed values)
	BassAvg float64
	MidAvg  float64
	HighAvg float64

	// Animation time
	Time float64
}

// CreateRenderer implements fyne.Widget.
func (v *BaseVisualizer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.Raster)
}

// MinSize returns the minimum size of the visualizer.
func (v *BaseVisualizer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

// UpdateFFT updates the visualizer with new FFT data.
func (v *BaseVisualizer) UpdateFFT(data []float32) {
	v.Mu.Lock()
	v.FFTData = data
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// Reset clears the visualizer state.
func (v *BaseVisualizer) Reset() {
	v.Mu.Lock()
	v.FFTData = nil
	v.BassAvg = 0
	v.MidAvg = 0
	v.HighAvg = 0
	v.Time = 0
	v.Mu.Unlock()

	v.Raster.Refresh()
}

// GetFFTData safely retrieves a copy of the FFT data.
func (v *BaseVisualizer) GetFFTData() []float32 {
	v.Mu.Lock()
	defer v.Mu.Unlock()
	return v.FFTData
}
