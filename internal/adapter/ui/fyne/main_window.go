package fyne

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"sync"
	"time"

	fyneapp "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/tejashwikalptaru/gotune/internal/adapter/ui/fyne/rotate"
	"github.com/tejashwikalptaru/gotune/res"
)

// MainWindow is the main UI window implementing the UIView interface.
// It handles all UI rendering and user interactions.
//
// The MainWindow follows the MVP pattern:
// - It's a "dumb view" that just displays data
// - All business logic is in the Presenter
// - User interactions are forwarded to the Presenter
type MainWindow struct {
	app    fyneapp.App
	window fyneapp.Window

	// UI components
	prevButton     *widget.Button
	playButton     *widget.Button
	stopButton     *widget.Button
	nextButton     *widget.Button
	muteButton     *widget.Button
	loopButton     *widget.Button
	songInfo       *widget.Label
	currentTime    *widget.Label
	endTime        *widget.Label
	progressSlider *widget.Slider
	volumeSlider   *widget.Slider
	albumArt       *canvas.Image

	// State
	isDarkTheme      bool
	rotator          *rotate.Rotator
	stopScroll       chan struct{}
	updatingProgress bool
	progressMu       sync.Mutex

	// Playlist window (optional)
	playlistWindow *PlaylistWindow

	// Lifecycle management
	closeOnce sync.Once
	scrollWg  sync.WaitGroup // WaitGroup to wait for scroll goroutine to exit

	// Presenter (set after construction)
	presenter *Presenter
}

// NewMainWindow creates a new main window.
func NewMainWindow(app fyneapp.App) *MainWindow {
	w := &MainWindow{
		app: app,
	}

	// Create a window
	w.window = app.NewWindow(APPNAME)

	// Build UI
	w.buildUI()

	// Set window properties
	w.window.Resize(fyneapp.Size{
		Width:  WIDTH,
		Height: HEIGHT,
	})
	w.window.SetFixedSize(true)
	w.app.SetIcon(res.ResourceIconPng)

	// Detect theme
	if app.Settings().ThemeVariant() == 0 {
		w.isDarkTheme = true
	} else {
		w.isDarkTheme = false
	}

	// Initialize rotator (but don't start the goroutine yet)
	w.rotator = rotate.NewRotator(APPNAME, 15)
	w.stopScroll = make(chan struct{})

	return w
}

// SetPresenter connects the presenter to this view.
// This must be called before showing the window.
func (w *MainWindow) SetPresenter(presenter *Presenter) {
	w.presenter = presenter
	w.wirePresenterHandlers()
	w.addShortcuts()
}

// buildUI constructs the UI components.
func (w *MainWindow) buildUI() {
	// Album art display
	w.albumArt = canvas.NewImageFromResource(res.ResourceMusicPng)
	w.albumArt.FillMode = canvas.ImageFillContain

	// Control buttons
	w.prevButton = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), nil)
	w.playButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil)
	w.stopButton = widget.NewButtonWithIcon("", theme.MediaStopIcon(), nil)
	w.nextButton = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nil)
	w.muteButton = widget.NewButtonWithIcon("", theme.VolumeUpIcon(), nil)
	w.loopButton = widget.NewButtonWithIcon("", theme.MediaReplayIcon(), nil)

	// Song info label
	w.songInfo = widget.NewLabel("")
	w.songInfo.Truncation = fyneapp.TextTruncateClip
	w.songInfo.TextStyle = fyneapp.TextStyle{
		Bold:   true,
		Italic: true,
	}

	// Volume slider
	w.volumeSlider = widget.NewSlider(0, 100)
	w.volumeSlider.Orientation = widget.Horizontal
	volIcon := canvas.NewImageFromResource(theme.VolumeUpIcon())
	volumeHolder := container.NewHBox(volIcon, w.volumeSlider)

	// Button container
	buttonsHBox := container.NewHBox(
		w.prevButton, w.playButton, w.stopButton,
		w.nextButton, w.muteButton, w.loopButton,
	)
	buttonsHolder := container.NewBorder(nil, nil, buttonsHBox, volumeHolder, w.songInfo)

	// Progress slider
	w.progressSlider = widget.NewSlider(0, 100)
	w.currentTime = widget.NewLabel("00:00")
	w.endTime = widget.NewLabel("00:00")
	sliderHolder := container.NewBorder(nil, nil, w.currentTime, w.endTime, w.progressSlider)

	// Main layout
	controls := container.NewVBox(buttonsHolder, sliderHolder)
	splitContainer := container.NewBorder(nil, controls, nil, nil, w.albumArt)
	w.window.SetContent(container.NewPadded(splitContainer))

	// Menu
	w.window.SetMainMenu(fyneapp.NewMainMenu(w.createMenu()...))
}

// wirePresenterHandlers connects UI events to presenter handlers.
func (w *MainWindow) wirePresenterHandlers() {
	if w.presenter == nil {
		return
	}

	// Button handlers
	w.playButton.OnTapped = func() {
		w.presenter.OnPlayClicked()
	}

	w.stopButton.OnTapped = func() {
		w.presenter.OnStopClicked()
	}

	w.nextButton.OnTapped = func() {
		w.presenter.OnNextClicked()
	}

	w.prevButton.OnTapped = func() {
		w.presenter.OnPreviousClicked()
	}

	w.muteButton.OnTapped = func() {
		w.presenter.OnMuteClicked()
	}

	w.loopButton.OnTapped = func() {
		w.presenter.OnLoopClicked()
	}

	// Volume slider
	w.volumeSlider.OnChanged = func(value float64) {
		w.presenter.OnVolumeChanged(value)
	}

	w.progressSlider.OnChanged = func(value float64) {
		w.progressMu.Lock()
		isUpdating := w.updatingProgress
		w.progressMu.Unlock()

		if !isUpdating {
			w.presenter.OnSeekRequested(value)
		}
	}
}

// createMenu creates the application menu.
func (w *MainWindow) createMenu() []*fyneapp.Menu {
	menus := make([]*fyneapp.Menu, 0)
	separator := fyneapp.NewMenuItemSeparator()

	openFile := fyneapp.NewMenuItem("Open", func() {
		w.handleOpenFile()
	})

	openFolder := fyneapp.NewMenuItem("Open Folder", func() {
		w.handleOpenFolder()
	})

	viewPlaylist := fyneapp.NewMenuItem("View Playlist", func() {
		if w.presenter != nil {
			w.presenter.OnPlaylistMenuClicked()
		}
	})

	exitMenu := fyneapp.NewMenuItem("Exit", func() {
		w.window.Close()
	})

	fileMenuItems := fyneapp.NewMenu("File", openFile, openFolder, separator, viewPlaylist, separator, exitMenu)
	menus = append(menus, fileMenuItems)

	return menus
}

// handleOpenFile handles the "Open File" menu action.
func (w *MainWindow) handleOpenFile() {
	if w.presenter == nil {
		return
	}

	// Use Fyne file dialog
	dialog := NewFileDialog(w.window, func(filePath string) {
		if err := w.presenter.OnFileOpened(filePath); err != nil {
			w.ShowNotification("Error", fmt.Sprintf("Failed to open file: %v", err))
		}
	})
	dialog.Show()
}

// handleOpenFolder handles the "Open Folder" menu action.
func (w *MainWindow) handleOpenFolder() {
	if w.presenter == nil {
		return
	}

	// Use Fyne folder dialog
	dialog := NewFolderDialog(w.window, func(folderPath string) {
		if err := w.presenter.OnFolderOpened(folderPath); err != nil {
			w.ShowNotification("Error", fmt.Sprintf("Failed to scan folder: %v", err))
		}
	})
	dialog.Show()
}

// addShortcuts adds keyboard shortcuts.
func (w *MainWindow) addShortcuts() {
	w.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyneapp.KeyUp,
		Modifier: fyneapp.KeyModifierAlt,
	}, func(shortcut fyneapp.Shortcut) {
		// Volume up
		currentVol := w.volumeSlider.Value
		newVol := currentVol + 5
		if newVol > 100 {
			newVol = 100
		}
		w.volumeSlider.SetValue(newVol)
	})

	w.window.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyneapp.KeyDown,
		Modifier: fyneapp.KeyModifierAlt,
	}, func(shortcut fyneapp.Shortcut) {
		// Volume down
		currentVol := w.volumeSlider.Value
		newVol := currentVol - 5
		if newVol < 0 {
			newVol = 0
		}
		w.volumeSlider.SetValue(newVol)
	})
}

// startScrollInfoRoutine starts the song info scrolling animation.
// This should only be called after the Fyne app is fully initialized (in ShowAndRun).
func (w *MainWindow) startScrollInfoRoutine() {
	w.scrollWg.Add(1)

	go func() {
		defer w.scrollWg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-w.stopScroll:
				return
			case <-ticker.C:
				if w.rotator != nil {
					fyneapp.Do(func() {
						w.songInfo.SetText(w.rotator.Rotate())
					})
				}
			}
		}
	}()
}

// ShowAndRun shows the window and runs the application.
// This also starts the song info scrolling animation.
func (w *MainWindow) ShowAndRun() {
	w.startScrollInfoRoutine()
	w.window.ShowAndRun()
}

// Close closes the window and stops the scrolling animation.
// It's safe to call multiple times (idempotent).
func (w *MainWindow) Close() {
	w.closeOnce.Do(func() {
		// Close playlist window if open
		w.ClosePlaylistWindow()

		// Signal the scroll goroutine to stop
		close(w.stopScroll)
	})

	// Wait for the scroll goroutine to finish (safe to call multiple times)
	w.scrollWg.Wait()

	// Close the window after goroutine cleanup
	w.window.Close()
}

// GetWindow returns the underlying Fyne window.
func (w *MainWindow) GetWindow() fyneapp.Window {
	return w.window
}

// UIView interface implementation

// SetPlayState updates the play/pause button state.
func (w *MainWindow) SetPlayState(playing bool) {
	fyneapp.Do(func() {
		if playing {
			w.playButton.SetIcon(theme.MediaPauseIcon())
		} else {
			w.playButton.SetIcon(theme.MediaPlayIcon())
		}
		w.playButton.Refresh()
	})
}

// SetMuteState updates the mute button state.
func (w *MainWindow) SetMuteState(muted bool) {
	fyneapp.Do(func() {
		if muted {
			w.muteButton.SetIcon(theme.VolumeMuteIcon())
		} else {
			w.muteButton.SetIcon(theme.VolumeUpIcon())
		}
		w.muteButton.Refresh()
	})
}

// SetLoopState updates the loop button state.
func (w *MainWindow) SetLoopState(enabled bool) {
	fyneapp.Do(func() {
		var icon *fyneapp.StaticResource
		if w.isDarkTheme {
			icon = res.ResourceRepeatLightPng
		} else {
			icon = res.ResourceRepeatDarkPng
		}

		if enabled {
			w.loopButton.SetIcon(icon)
		} else {
			w.loopButton.SetIcon(theme.MediaReplayIcon())
		}
		w.loopButton.Refresh()
	})
}

// SetVolume updates the volume slider.
func (w *MainWindow) SetVolume(volume float64) {
	fyneapp.Do(func() {
		// Convert from 0.0-1.0 to 0-100
		w.volumeSlider.Value = volume * 100.0
		w.volumeSlider.Refresh()
	})
}

// SetTrackInfo updates the displayed track information.
func (w *MainWindow) SetTrackInfo(title, artist, album string) {
	fyneapp.Do(func() {
		// Format: "Artist - Title"
		var text string
		switch {
		case artist != "" && title != "":
			text = fmt.Sprintf("%s - %s", artist, title)
		case title != "":
			text = title
		default:
			text = "No track loaded"
		}

		// Update rotator for scrolling text
		w.rotator = rotate.NewRotator(text, 15)
		// Set initial label text
		w.songInfo.SetText(text)
	})
}

// SetAlbumArt updates the album artwork.
func (w *MainWindow) SetAlbumArt(imageData []byte) {
	fyneapp.Do(func() {
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
		w.albumArt.Resource = res.ResourceMusicPng
		w.albumArt.Image = nil
		w.albumArt.Refresh()
	})
}

// SetCurrentTime updates the current playback time display.
func (w *MainWindow) SetCurrentTime(seconds float64) {
	fyneapp.Do(func() {
		format := fmt.Sprintf("%.2d:%.2d", int(seconds/60), int(math.Mod(seconds, 60)))
		w.currentTime.SetText(format)
	})
}

// SetTotalTime updates the total track duration display.
func (w *MainWindow) SetTotalTime(seconds float64) {
	fyneapp.Do(func() {
		format := fmt.Sprintf("%.2d:%.2d", int(seconds/60), int(math.Mod(seconds, 60)))
		w.progressSlider.Max = seconds
		w.endTime.SetText(format)
	})
}

// SetProgress updates the progress slider position.
func (w *MainWindow) SetProgress(position, duration float64) {
	fyneapp.Do(func() {
		if duration > 0 {
			w.progressMu.Lock()
			w.updatingProgress = true
			w.progressSlider.Value = position
			w.progressSlider.Refresh()
			w.updatingProgress = false
			w.progressMu.Unlock()
		}
	})
}

// UpdatePlaylistSelection updates the selected track in the playlist view.
func (w *MainWindow) UpdatePlaylistSelection(index int) {
	fyneapp.Do(func() {
		if w.playlistWindow != nil && w.playlistWindow.IsVisible() {
			w.playlistWindow.SetSelected(index)
		}
	})
}

// ShowPlaylistWindow displays the playlist window.
func (w *MainWindow) ShowPlaylistWindow() {
	fyneapp.Do(func() {
		if w.playlistWindow == nil {
			w.playlistWindow = NewPlaylistWindow(
				w.app,
				w.presenter,
				w.presenter.EventBus,
			)
			// Set callback to clear reference when window is closed
			w.playlistWindow.SetOnWindowClosed(func() {
				fyneapp.Do(func() {
					w.playlistWindow = nil
				})
			})
		}
		w.playlistWindow.Show()
	})
}

// ClosePlaylistWindow closes the playlist window if it's open.
func (w *MainWindow) ClosePlaylistWindow() {
	fyneapp.Do(func() {
		if w.playlistWindow != nil {
			w.playlistWindow.Close()
		}
	})
}

// IsPlaylistWindowOpen returns whether the playlist window is currently open.
func (w *MainWindow) IsPlaylistWindowOpen() bool {
	return w.playlistWindow != nil && w.playlistWindow.IsVisible()
}

// ShowNotification displays a system notification.
func (w *MainWindow) ShowNotification(title, message string) {
	fyneapp.Do(func() {
		w.app.SendNotification(fyneapp.NewNotification(title, message))
	})
}

// Verify UIView implementation
var _ UIView = (*MainWindow)(nil)
