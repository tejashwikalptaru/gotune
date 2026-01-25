// Package widgets provides custom Fyne widgets for the GoTune application.
// This file re-exports visualizer types from the visualizer subpackage for backward compatibility.
package widgets

import (
	"github.com/tejashwikalptaru/gotune/internal/adapter/ui/fyne/widgets/visualizer"
)

// Type aliases for backward compatibility
type (
	// VisualizerType represents the type of visualizer.
	VisualizerType = visualizer.Type

	// MusicVisualizer defines the interface that all visualizers must implement.
	MusicVisualizer = visualizer.MusicVisualizer
)

// Visualizer type constants for backward compatibility.
const (
	VisualizerTypeSpectrumBars = visualizer.TypeSpectrumBars
	VisualizerTypeStarfield    = visualizer.TypeStarfield
	VisualizerTypeCircular     = visualizer.TypeCircular
	VisualizerTypeTunnel       = visualizer.TypeTunnel
	VisualizerTypePlasma       = visualizer.TypePlasma
	VisualizerTypeFFTSpectrum  = visualizer.TypeFFTSpectrum
)

// VisualizerFactory creates a new visualizer of the specified type.
// This is a wrapper around visualizer.Factory for backward compatibility.
func VisualizerFactory(visType VisualizerType, numBars int) MusicVisualizer {
	return visualizer.Factory(visType, numBars)
}

// GetVisualizerTypes returns all available visualizer types with their display names.
// This is a wrapper around visualizer.GetTypes for backward compatibility.
func GetVisualizerTypes() []struct {
	Type VisualizerType
	Name string
} {
	types := visualizer.GetTypes()
	result := make([]struct {
		Type VisualizerType
		Name string
	}, len(types))
	for i, t := range types {
		result[i] = struct {
			Type VisualizerType
			Name string
		}{t.Type, t.Name}
	}
	return result
}
