# Audio Visualizer Implementation Guide

This document provides a step-by-step guide on how to implement a real-time audio visualizer in the GoTune application, inspired by the provided source code. The visualizer will display spectrum bars that react to the currently playing audio, using an efficient `canvas.Raster` approach.

## 1. Overview

The implementation can be broken down into these main parts:

1.  **Getting Audio Data:** We'll use the BASS audio library to get the Fast Fourier Transform (FFT) data of the audio stream.
2.  **Exposing Data to the UI:** We'll modify the service and port layers to make the FFT data available to the UI.
3.  **Creating a Custom Fyne Widget:** We'll build a new Fyne widget using `canvas.Raster` to render the visualizer bars.
4.  **Integrating the Widget:** We'll replace the existing album art with our new visualizer widget.
5.  **Updating the UI:** We'll update the presenter to periodically fetch the FFT data and redraw the visualizer.

## 2. Getting FFT Data from BASS

The BASS library provides the `BASS_ChannelGetData` function, which can be used to get the FFT data.

### 2.1. Update `internal/adapter/audio/bass/bindings.go`

First, we need to add a Go binding for the `BASS_ChannelGetData` function.

```go
// internal/adapter/audio/bass/bindings.go

// ... existing code ...

// bassChannelGetData gets data from a channel.
func bassChannelGetData(handle int64, buffer unsafe.Pointer, length int) int {
	return int(C.BASS_ChannelGetData(C.DWORD(handle), buffer, C.DWORD(length)))
}

// ... existing code ...
```

### 2.2. Update `internal/adapter/audio/bass/engine.go`

Now, let's add a method to the `Engine` to get the FFT data. We also need to define the FFT constants in `constants.go`.

#### `internal/adapter/audio/bass/constants.go`

```go
// internal/adapter/audio/bass/constants.go

// ... existing code ...

// BASS_ChannelGetData flags
const (
	dataFFT256  = C.BASS_DATA_FFT256
	dataFFT512  = C.BASS_DATA_FFT512
	dataFFT1024 = C.BASS_DATA_FFT1024
	dataFFT2048 = C.BASS_DATA_FFT2048
	dataFFT4096 = C.BASS_DATA_FFT4096
	dataFFT8192 = C.BASS_DATA_FFT8192
)
```

#### `internal/adapter/audio/bass/engine.go`

```go
// internal/adapter/audio/bass/engine.go
import (
	"unsafe"
	// ... other imports
)

// ... existing code ...

// GetFFTData gets the FFT data for a channel.
func (e *Engine) GetFFTData(handle domain.TrackHandle, fftSize int) ([]float32, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return nil, domain.ErrNotInitialized
	}

	track, exists := e.tracks[handle]
	if !exists {
		return nil, domain.ErrInvalidTrackHandle
	}

	var flag int
	switch fftSize {
	case 256:
		flag = dataFFT256
	case 512:
		flag = dataFFT512
	case 1024:
		flag = dataFFT1024
	case 2048:
		flag = dataFFT2048
	case 4096:
		flag = dataFFT4096
	case 8192:
		flag = dataFFT8192
	default:
		return nil, domain.ErrInvalidFFTSize // You'll need to define this error
	}

	// BASS returns half the FFT size
	buffer := make([]float32, fftSize/2)
	ret := bassChannelGetData(track.handle, unsafe.Pointer(&buffer[0]), flag)
	if ret == -1 {
		return nil, createBassError("get_fft_data", "", C.BASS_ErrorGetCode())
	}

	return buffer, nil
}
```

You will also need to add `ErrInvalidFFTSize` to `internal/domain/errors.go`.

```go
// internal/domain/errors.go
var (
	// ...
	ErrInvalidFFTSize     = errors.New("invalid FFT size")
)
```

## 3. Exposing FFT Data to the UI

### 3.1. Update `internal/ports/audio.go`

Add `GetFFTData` to the `AudioEngine` interface.

```go
// internal/ports/audio.go

type AudioEngine interface {
	// ... existing methods ...
	GetFFTData(handle domain.TrackHandle, fftSize int) ([]float32, error)
}
```

### 3.2. Update `internal/service/player_service.go`

Add a method to the `PlayerService` to get the FFT data.

```go
// internal/service/player_service.go

// ... existing code ...

// GetFFTData returns the FFT data for the current track.
func (s *PlayerService) GetFFTData(fftSize int) ([]float32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentTrack == nil {
		return nil, domain.ErrNoTrackLoaded
	}

	return s.audioEngine.GetFFTData(s.currentTrack.Handle, fftSize)
}
```

## 4. Creating the Visualizer Widget with `canvas.Raster`

Create a new file `internal/adapter/ui/fyne/widgets/visualizer.go`. This widget will use `canvas.Raster` for efficient rendering.

```go
// internal/adapter/ui/fyne/widgets/visualizer.go
package widgets

import (
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Visualizer is a custom widget to display an audio visualizer.
type Visualizer struct {
	canvas.Raster
	data          []float32
	capYPositions []float32
	numBars       int
}

// NewVisualizer creates a new visualizer widget.
func NewVisualizer() *Visualizer {
	v := &Visualizer{
		numBars: 100,
	}
	v.Generator = v.generate
	v.ExtendBaseWidget(v)
	return v
}

// UpdateData updates the FFT data for the visualizer.
func (v *Visualizer) UpdateData(data []float32) {
	v.data = data
	v.Refresh()
}

func (v *Visualizer) generate(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if v.data == nil {
		return img
	}

	// Black background
	draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)

	if len(v.capYPositions) < v.numBars {
		v.capYPositions = make([]float32, v.numBars)
	}

	const gap = 2
	meterWidth := (w - (v.numBars-1)*gap) / v.numBars

	// Create a gradient
	gradient := func(val float32) color.Color {
		// From red to green
		red := uint8(255 * val)
		green := uint8(255 * (1 - val))
		return color.RGBA{R: red, G: green, B: 0, A: 255}
	}

	step := len(v.data) / v.numBars
	for i := 0; i < v.numBars; i++ {
		value := v.data[i*step]
		if value > 1.0 {
			value = 1.0
		}

		barHeight := int(value * float32(h))
		capHeight := 2

		// Cap
		if value > v.capYPositions[i] {
			v.capYPositions[i] = value
		} else {
			v.capYPositions[i] -= 0.01 // Fall speed
			if v.capYPositions[i] < 0 {
				v.capYPositions[i] = 0
			}
		}

		// Draw the bar
		barX := i * (meterWidth + gap)
		for y := h - barHeight; y < h; y++ {
			for x := barX; x < barX+meterWidth; x++ {
				img.Set(x, y, gradient(float32(h-y)/float32(h)))
			}
		}

		// Draw the cap
		capY := float32(h) - (v.capYPositions[i] * float32(h))
		for y := int(capY); y < int(capY)+capHeight; y++ {
			for x := barX; x < barX+meterWidth; x++ {
				img.Set(x, y, color.White)
			}
		}
	}

	return img
}

// MinSize returns the minimum size of the widget.
func (v *Visualizer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}
```

## 5. Integrating the Visualizer Widget

Now, let's replace the `albumArt` with our new `Visualizer` in `main_window.go`.

### 5.1. Update `internal/adapter/ui/fyne/main_window.go`

```go
// internal/adapter/ui/fyne/main_window.go

// ... imports ...
"github.com/tejashwikalptaru/gotune/internal/adapter/ui/fyne/widgets"

// ...
type MainWindow struct {
	// ...
	albumArt   *canvas.Image
	visualizer *widgets.Visualizer // Add this
	// ...
}

// ...

func (w *MainWindow) buildUI() {
	// Album art display
	w.albumArt = canvas.NewImageFromResource(res.ResourceMusicPng)
	w.albumArt.FillMode = canvas.ImageFillContain

	// Visualizer
	w.visualizer = widgets.NewVisualizer()
	w.visualizer.Hide() // Hide by default

	// ...

	// Main layout
	controls := container.NewVBox(buttonsHolder, sliderHolder)
	// Create a stack to overlay album art and visualizer
	artStack := container.NewStack(w.albumArt, w.visualizer)
	splitContainer := container.NewBorder(nil, controls, nil, nil, artStack)
	w.window.SetContent(container.NewPadded(splitContainer))

	// ...
}

// ...

// SetAlbumArt updates the album artwork.
func (w *MainWindow) SetAlbumArt(imageData []byte) {
	fyneapp.Do(func() {
		w.visualizer.Hide() // Hide visualizer when album art is shown
		w.albumArt.Show()

		img, _, err := image.Decode(bytes.NewReader(imageData))
		if err != nil {
			// If decode fails, use default
			w.ClearAlbumArt()
			return
		}

		w.albumArt.Image = img
		w.albumArt.Refresh()
	})
}

// ClearAlbumArt resets the album artwork to the default.
func (w *MainWindow) ClearAlbumArt() {
	fyneapp.Do(func() {
		w.visualizer.Hide()
		w.albumArt.Show()
		w.albumArt.Resource = res.ResourceMusicPng
		w.albumArt.Image = nil
		w.albumArt.Refresh()
	})
}

// UpdateVisualizer updates the visualizer with new data.
func (w *MainWindow) UpdateVisualizer(data []float32) {
	fyneapp.Do(func() {
		if !w.visualizer.Visible() {
			w.albumArt.Hide()
			w.visualizer.Show()
		}
		w.visualizer.UpdateData(data)
	})
}
```

## 6. Updating the Presenter

The presenter is responsible for fetching the FFT data and telling the view to update.

### 6.1. Update `internal/ports/ui.go`

Add `UpdateVisualizer` to the `UIView` interface.

```go
// internal/ports/ui.go

type UIView interface {
	// ... existing methods ...
	UpdateVisualizer(data []float32)
}
```

### 6.2. Update `internal/adapter/ui/fyne/presenter.go`

We'll add a goroutine to the presenter to periodically fetch the FFT data.

```go
// internal/adapter/ui/fyne/presenter.go

// ... imports ...

type Presenter struct {
	// ...
	stopVisualizer chan struct{}
	wg             sync.WaitGroup
	// ...
}

func NewPresenter(...) *Presenter {
	p := &Presenter{
		// ...
		stopVisualizer: make(chan struct{}),
	}
	// ...
	return p
}

func (p *Presenter) Start() {
	// ...
	p.startVisualizerUpdates()
}

func (p *Presenter) Shutdown() {
	// ...
	close(p.stopVisualizer)
	p.wg.Wait()
}


func (p *Presenter) startVisualizerUpdates() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(50 * time.Millisecond) // ~20 FPS
		defer ticker.Stop()

		for {
			select {
			case <-p.stopVisualizer:
				return
			case <-ticker.C:
				if p.player.IsPlaying() {
					data, err := p.player.GetFFTData(2048) // Higher resolution for more detail
					if err == nil && data != nil {
						p.view.UpdateVisualizer(data)
					}
				}
			}
		}
	}()
}
```

You'll need to call `Start()` when the application starts and `Shutdown()` when it closes. This can be done in `main.go`.

In `main.go`, you would modify it like this:

```go
// main.go

func main() {
	// ...
	presenter := fyne_adapter.NewPresenter(playerService, playlistService, preferenceService)
	ui.SetPresenter(presenter)

	presenter.Start()
	defer presenter.Shutdown()

	ui.ShowAndRun()
}

```

## 7. Conclusion

With these changes, the GoTune application will now display a real-time audio visualizer when a track is playing. The visualizer will replace the album art and will update approximately 20 times per second, providing a smooth and responsive animation. You can adjust the `fftSize` and the ticker duration to fine-tune the visualizer's appearance and performance.
