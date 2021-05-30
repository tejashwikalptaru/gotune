package ui

import (
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/tejashwikalptaru/gotune/res"
	"github.com/tejashwikalptaru/gotune/ui/rotate"
	"github.com/tejashwikalptaru/gotune/utils"
	"image"
	"math"
	"path/filepath"
	"time"
)

type FileOpenCallBack func(filePath string)

type Main struct {
	app    fyne.App
	window fyne.Window

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

	isDarkTheme      bool
	fileSearchStatus utils.ScanStatus
	changeListener   chan fyne.Settings

	// callbacks
	fileOpenCallBack     FileOpenCallBack
	addToPlayListFunc    func(path string)
	openPlaylistCallBack func()

	// rotator chan
	killRotate chan bool
	rotator    *rotate.Rotator
}

func NewMainWindow() *Main {
	var main Main

	a := app.NewWithID(utils.PACKAGE)
	w := a.NewWindow(utils.APPNAME)
	splitContainer := container.NewBorder(nil, main.controls(), nil, nil, main.display())
	w.SetContent(container.NewPadded(splitContainer))
	w.SetMainMenu(fyne.NewMainMenu(main.createMainMenu()...))
	main.app = a
	main.window = w
	w.Resize(fyne.Size{
		Width:  utils.WIDTH,
		Height: utils.HEIGHT,
	})
	w.SetFixedSize(true)

	main.killRotate = make(chan bool, 1)
	main.changeListener = make(chan fyne.Settings, 1)
	main.rotator = rotate.NewRotator(utils.APPNAME, 15)
	main.scrollInfoRoutine()
	return &main
}

func (main *Main) Free() {
	main.killRotate <- true
	close(main.killRotate)
	close(main.changeListener)
}

func (main *Main) createMainMenu() []*fyne.Menu {
	menus := make([]*fyne.Menu, 0)
	separator := fyne.NewMenuItemSeparator()

	openFile := fyne.NewMenuItem("Open", func() {
		main.handleOpenFile()
	})
	openFolder := fyne.NewMenuItem("Open Folder", func() {
		main.handleOpenFolder()
	})
	viewPL := fyne.NewMenuItem("View Playlist", func() {
		if main.openPlaylistCallBack != nil {
			main.openPlaylistCallBack()
		}
	})
	exitMenu := fyne.NewMenuItem("Exit", func() {
		main.window.Close()
	})

	fileMenuItems := fyne.NewMenu("File", openFile, openFolder, separator, viewPL, separator, exitMenu)
	menus = append(menus, fileMenuItems)
	return menus
}

func (main *Main) display() fyne.CanvasObject {
	main.albumArt = canvas.NewImageFromResource(res.ResourceMusicPng)
	main.albumArt.FillMode = canvas.ImageFillContain
	return main.albumArt
}
func (main *Main) controls() fyne.CanvasObject {
	main.prevButton = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), func() {})
	main.playButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {})
	main.stopButton = widget.NewButtonWithIcon("", theme.MediaStopIcon(), func() {})
	main.nextButton = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), func() {})
	main.muteButton = widget.NewButtonWithIcon("", theme.VolumeUpIcon(), func() {})
	main.loopButton = widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() {})
	main.songInfo = widget.NewLabel("")
	main.songInfo.Wrapping = fyne.TextTruncate
	main.songInfo.TextStyle = fyne.TextStyle{
		Bold:      true,
		Italic:    true,
		Monospace: false,
	}
	main.volumeSlider = widget.NewSlider(0, 100)
	main.volumeSlider.Orientation = widget.Horizontal

	volIcon := canvas.NewImageFromResource(theme.VolumeUpIcon())
	volumeHolder := container.NewHBox(volIcon, main.volumeSlider)

	buttonsHBox := container.NewHBox(main.prevButton, main.playButton, main.stopButton, main.nextButton, main.muteButton, main.loopButton)
	buttonsHolder := container.NewBorder(nil, nil, buttonsHBox, volumeHolder, main.songInfo)

	main.progressSlider = widget.NewSlider(0, 100)
	main.currentTime = widget.NewLabel("00:00")
	main.endTime = widget.NewLabel("00:00")

	sliderHolder := container.NewBorder(nil, nil, main.currentTime, main.endTime, main.progressSlider)

	return container.NewVBox(buttonsHolder, sliderHolder)
}

func (main *Main) ShowAndRun() {
	main.app.SetIcon(res.ResourceIconPng)
	if main.app.Settings().ThemeVariant() == 0 {
		main.isDarkTheme = true
	} else {
		main.isDarkTheme = false
	}

	go func() {
		for {
			settings := <- main.changeListener
			if settings == nil {
				return
			}
			if settings.ThemeVariant() == 0 {
				main.isDarkTheme = true
				main.loopButton.SetIcon(res.ResourceRepeatLightPng)
			} else {
				main.isDarkTheme = false
				main.loopButton.SetIcon(res.ResourceRepeatDarkPng)
			}
		}
	}()
	main.app.Settings().AddChangeListener(main.changeListener)
	main.window.ShowAndRun()
}

func (main *Main) ShowNotification(title, message string) {
	main.app.SendNotification(fyne.NewNotification(title, message))
}

func (main *Main) OnMute(f func()) {
	main.muteButton.OnTapped = f
}

func (main *Main) SetMuteState(mute bool) {
	if mute {
		main.muteButton.SetIcon(theme.VolumeMuteIcon())
	} else {
		main.muteButton.SetIcon(theme.VolumeUpIcon())
	}
}

func (main *Main) SetLoopState(loop bool) {
	var icon *fyne.StaticResource
	if main.isDarkTheme {
		icon = res.ResourceRepeatLightPng
	} else {
		icon = res.ResourceRepeatDarkPng
	}

	if loop {
		main.loopButton.SetIcon(icon)
	} else {
		main.loopButton.SetIcon(theme.MediaReplayIcon())
	}
}

func (main *Main) OnPlay(f func()) {
	main.playButton.OnTapped = f
}

func (main *Main) OnStop(f func()) {
	main.stopButton.OnTapped = f
}

func (main *Main) SetPlayState(playing bool) {
	if playing {
		main.playButton.SetIcon(theme.MediaPauseIcon())
	} else {
		main.playButton.SetIcon(theme.MediaPlayIcon())
	}
}

func (main *Main) SetAlbumArt(imgByte []byte) {
	img, _, err := image.Decode(bytes.NewReader(imgByte))
	if err != nil {
		//fmt.Print(err)
		return
	}
	main.albumArt.Resource = nil
	main.albumArt.Image = img
	canvas.Refresh(main.albumArt)
}

func (main *Main) ClearAlbumArt() {
	main.albumArt.Image = nil
	main.albumArt.Resource = res.ResourceMusicPng
	main.albumArt.Refresh()
}

func (main *Main) SetCurrentTime(timeElapsed float64) {
	format := fmt.Sprintf("%.2d:%.2d", int(timeElapsed/60), int(math.Mod(timeElapsed, 60)))
	main.progressSlider.Value = timeElapsed
	main.progressSlider.Refresh()
	main.currentTime.SetText(format)
}

func (main *Main) SetTotalTime(endTime float64) {
	format := fmt.Sprintf("%.2d:%.2d", int(endTime/60), int(math.Mod(endTime, 60)))
	main.progressSlider.Max = endTime
	main.endTime.SetText(format)
}

func (main *Main) SetSongName(name string) {
	main.rotator = rotate.NewRotator(name, 15)
}

func (main *Main) scrollInfoRoutine() {
	go func() {
		for {
			select {
			case <-main.killRotate:
				return
			default:
				time.Sleep(time.Millisecond * 400)
				main.songInfo.SetText(main.rotator.Rotate())
			}
		}
	}()
}

func (main *Main) AddShortCuts() {
	volUp := desktop.CustomShortcut{
		KeyName:  fyne.KeyUp,
		Modifier: desktop.AltModifier,
	}
	main.window.Canvas().AddShortcut(&volUp, func(shortcut fyne.Shortcut) {
		main.volumeSlider.SetValue(main.volumeSlider.Value + 1)
	})
	volDown := desktop.CustomShortcut{
		KeyName:  fyne.KeyDown,
		Modifier: desktop.AltModifier,
	}
	main.window.Canvas().AddShortcut(&volDown, func(shortcut fyne.Shortcut) {
		main.volumeSlider.SetValue(main.volumeSlider.Value - 1)
	})
}

func (main *Main) SetPlayListUpdater(f func(p string)) {
	main.addToPlayListFunc = f
}

func (main *Main) handleOpenFolder() {
	if main.fileSearchStatus == utils.ScanRunning {
		utils.ShowError(true, "Please wait", "Scanning is already running, please wait for that to finish%s", "")
		return
	}
	main.fileSearchStatus = utils.ScanRunning
	path, err := utils.OpenFolder("Select folder to search for files...")
	if err != nil {
		fyne.LogError("failed to open folder browser", err)
		main.fileSearchStatus = utils.ScanStopped
		return
	}
	fsw := NewFileSearchWindow(main.app)
	fsw.OnClosed(func() {
		main.fileSearchStatus = utils.ScanStopped
	})
	fsw.Show()
	go func() {
		files, err := utils.ScanFolder(path, func(s string) {
			fsw.UpdateLabel(filepath.Base(s))
		}, &main.fileSearchStatus)
		if err != nil {
			fyne.LogError("failed to scan for files in folder", err)
			fsw.Close()
		}
		if main.addToPlayListFunc != nil {
			fsw.progress.Hide()
			fsw.progressParsing.Min = 0
			fsw.progressParsing.Max = float64(len(files))
			fsw.progressParsing.Show()
			for i, f := range files {
				fsw.UpdateLabel(fmt.Sprintf("Found: %d items, processing: %d", len(files), i+1))
				fsw.progressParsing.SetValue(float64(i + 1))
				main.addToPlayListFunc(f)
			}
		}
		fsw.Close()
	}()
}

func (main *Main) SetOpenPlayListCallBack(f func()) {
	main.openPlaylistCallBack = f
}

func (main *Main) SetFileOpenCallBack(f FileOpenCallBack) {
	main.fileOpenCallBack = f
}

func (main *Main) handleOpenFile() {
	path, err := utils.OpenFile("Select a file to play")
	if err != nil {
		fyne.LogError("failed to select file", err)
		return
	}
	if main.fileOpenCallBack == nil {
		return
	}
	main.fileOpenCallBack(path)
}

func (main *Main) SetVolume(vol float64) {
	main.volumeSlider.Value = vol
}

func (main *Main) VolumeUpdateCallBack(f func(float64)) {
	main.volumeSlider.OnChanged = f
}

func (main *Main) OnNext(f func()) {
	main.nextButton.OnTapped = f
}

func (main *Main) OnPrev(f func()) {
	main.prevButton.OnTapped = f
}

func (main *Main) ProgressChanged(f func(val float64)) {
	main.progressSlider.OnChanged = f
}

func (main *Main) GetApp() fyne.App {
	return main.app
}

func (main *Main) OnAppClose(f func()) {
	main.window.SetOnClosed(f)
}

func (main *Main) QuitApp() {
	main.app.Quit()
}

func (main *Main) OnLoop(f func()) {
	main.loopButton.OnTapped = f
}

func (main *Main) SaveHistory(data string) {
	main.app.Preferences().SetString(utils.HISTORY, data)
}

func (main *Main) GetHistory() string {
	return main.app.Preferences().String(utils.HISTORY)
}
