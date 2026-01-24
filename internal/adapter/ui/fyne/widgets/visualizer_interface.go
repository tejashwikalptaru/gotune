// Package widgets provides custom Fyne widgets for the GoTune application.
package widgets

import (
	"fyne.io/fyne/v2"
)

// VisualizerType represents the type of visualizer.
type VisualizerType string

// Available visualizer types.
const (
	VisualizerTypeSpectrumBars VisualizerType = "spectrum_bars"
	VisualizerTypeStarfield    VisualizerType = "starfield"
	VisualizerTypeCircular     VisualizerType = "circular"
	VisualizerTypeTunnel       VisualizerType = "tunnel"
	VisualizerTypePlasma       VisualizerType = "plasma"
)

// MusicVisualizer defines the interface that all visualizers must implement.
// This allows the main window to switch between different visualizer types.
type MusicVisualizer interface {
	fyne.CanvasObject

	// UpdateFFT updates the visualizer with new FFT data.
	// This should be called periodically (e.g., 30fps) from the presenter.
	UpdateFFT(data []float32)

	// Reset clears the visualizer state.
	Reset()
}

// VisualizerFactory creates a new visualizer of the specified type.
func VisualizerFactory(visType VisualizerType, numBars int) MusicVisualizer {
	switch visType {
	case VisualizerTypeSpectrumBars:
		return NewSpectrumBarVisualizer(numBars)
	case VisualizerTypeStarfield:
		return NewStarfieldVisualizer()
	case VisualizerTypeCircular:
		return NewCircularVisualizer(numBars)
	case VisualizerTypeTunnel:
		return NewTunnelVisualizer()
	case VisualizerTypePlasma:
		return NewPlasmaVisualizer()
	default:
		return NewSpectrumBarVisualizer(numBars)
	}
}

// GetVisualizerTypes returns all available visualizer types with their display names.
func GetVisualizerTypes() []struct {
	Type VisualizerType
	Name string
} {
	return []struct {
		Type VisualizerType
		Name string
	}{
		{VisualizerTypeSpectrumBars, "Spectrum Bars"},
		{VisualizerTypeStarfield, "Starfield"},
		{VisualizerTypeCircular, "Circular"},
		{VisualizerTypeTunnel, "Tunnel"},
		{VisualizerTypePlasma, "Plasma"},
	}
}
