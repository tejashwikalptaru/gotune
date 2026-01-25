package visualizer

import (
	"fyne.io/fyne/v2"
)

// Type represents the type of visualizer.
type Type string

// Available visualizer types.
const (
	TypeSpectrumBars Type = "spectrum_bars"
	TypeStarfield    Type = "starfield"
	TypeCircular     Type = "circular"
	TypeTunnel       Type = "tunnel"
	TypePlasma       Type = "plasma"
	TypeFFTSpectrum  Type = "fft_spectrum"
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

// Factory creates a new visualizer of the specified type.
func Factory(visType Type, numBars int) MusicVisualizer {
	switch visType {
	case TypeSpectrumBars:
		return NewSpectrum(numBars)
	case TypeStarfield:
		return NewStarfield()
	case TypeCircular:
		return NewCircular(numBars)
	case TypeTunnel:
		return NewTunnel()
	case TypePlasma:
		return NewPlasma()
	case TypeFFTSpectrum:
		return NewFFTSpectrum()
	default:
		return NewSpectrum(numBars)
	}
}

// TypeInfo contains information about a visualizer type.
type TypeInfo struct {
	Type Type
	Name string
}

// GetTypes returns all available visualizer types with their display names.
func GetTypes() []TypeInfo {
	return []TypeInfo{
		{TypeSpectrumBars, "Spectrum Bars"},
		{TypeStarfield, "Starfield"},
		{TypeCircular, "Circular"},
		{TypeTunnel, "Tunnel"},
		{TypePlasma, "Plasma"},
		{TypeFFTSpectrum, "FFT Spectrum 3D"},
	}
}
