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
	"github.com/tejashwikalptaru/gotune/utils"
	"image"
	"math"
	"path/filepath"
)

type Main struct {
	app    fyne.App
	window fyne.Window

	prevButton *widget.Button
	playButton *widget.Button
	stopButton *widget.Button
	nextButton *widget.Button
	muteButton *widget.Button
	loopButton *widget.Button

	songInfo    *widget.Label
	currentTime *widget.Label
	endTime     *widget.Label

	progressSlider *widget.Slider

	albumArt *canvas.Image

	fileSearchStatus utils.ScanStatus
}

func NewMainWindow() *Main {
	var main Main

	a := app.New()
	w := a.NewWindow(utils.APPNAME)
	w.Resize(fyne.Size{
		Width:  utils.WIDTH,
		Height: utils.HEIGHT,
	})
	w.SetFixedSize(true)
	splitContainer := container.NewBorder(nil, main.controls(), nil, nil, main.display())
	w.SetContent(container.NewPadded(splitContainer))
	w.SetMainMenu(fyne.NewMainMenu(main.createMainMenu()...))
	main.app = a
	main.window = w
	return &main
}

func (main *Main) createMainMenu() []*fyne.Menu {
	menus := make([]*fyne.Menu, 0)
	separator := fyne.NewMenuItemSeparator()

	openFile := fyne.NewMenuItem("Open", func() {
		_, _ = utils.OpenFile("Select a file to play")
	})
	openFolder := fyne.NewMenuItem("Open Folder", func() {
		main.handleOpenFolder()
	})
	exitMenu := fyne.NewMenuItem("Exit", func() {
		main.window.Close()
	})

	fileMenuItems := fyne.NewMenu("File", openFile, openFolder, separator, exitMenu)
	menus = append(menus, fileMenuItems)
	return menus
}

func (main *Main) display() fyne.CanvasObject {
	btn := widget.NewButton("Test", func() {})
	btn.Resize(fyne.Size{
		Width:  utils.WIDTH,
		Height: utils.HEIGHT - 50,
	})
	main.albumArt = canvas.NewImageFromResource(res.ResourceLogoPng)
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
	main.songInfo = widget.NewLabel("Here goes the song name with the details of song in it")

	buttonsHolder := container.NewHBox(main.prevButton, main.playButton, main.stopButton, main.nextButton, main.muteButton, main.loopButton, main.songInfo)

	main.progressSlider = widget.NewSlider(0, 100)
	main.currentTime = widget.NewLabel("00:00:00")
	main.endTime = widget.NewLabel("00:00:00")

	sliderHolder := container.NewGridWithColumns(4, main.currentTime, main.progressSlider, main.endTime)

	return container.NewVBox(buttonsHolder, sliderHolder)
}

func (main *Main) ShowAndRun() {
	main.window.ShowAndRun()
}

func (main *Main) ShowNotification() {
	notification := utils.Notify("Welcome", "Thank you for trying GoTune")
	main.app.SendNotification(notification)
}

func (main *Main) MuteFunc(f func()) {
	main.muteButton.OnTapped = f
}

func (main *Main) SetMuteState(mute bool) {
	if mute {
		main.muteButton.SetIcon(theme.VolumeMuteIcon())
	} else {
		main.muteButton.SetIcon(theme.VolumeUpIcon())
	}
}

func (main *Main) PlayFunc(f func()) {
	main.playButton.OnTapped = f
}

func (main *Main) StopFunc(f func()) {
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
	main.albumArt.Image = img
	main.albumArt.Refresh()
}

func (main *Main) SetCurrentTime(timeElapsed float64) {
	format := fmt.Sprintf("%.2d:%.2d", int(timeElapsed/60), int(math.Mod(timeElapsed, 60)))
	main.progressSlider.SetValue(timeElapsed)
	main.currentTime.SetText(format)
}

func (main *Main) SetTotalTime(endTime float64) {
	format := fmt.Sprintf("%.2d:%.2d", int(endTime/60), int(math.Mod(endTime, 60)))
	main.progressSlider.Max = endTime
	main.endTime.SetText(format)
}

func (main *Main) AddShortCuts() {
	ctrlP := desktop.CustomShortcut{
		KeyName:  fyne.KeyP,
		Modifier: desktop.ControlModifier,
	}
	main.window.Canvas().AddShortcut(&ctrlP, func(shortcut fyne.Shortcut) {
		utils.ShowInfo("Welcome", "Thank you for trying %s", utils.APPNAME)
	})
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
	fsw.Show()
	fsw.OnClosed(func() {
		main.fileSearchStatus = utils.ScanStopped
	})
	go func() {
		_, err := utils.ScanFolder(path, func(s string) {
			fsw.UpdateLabel(filepath.Base(s))
		}, &main.fileSearchStatus)
		if err != nil {
			fyne.LogError("failed to scan for files in folder", err)
			fsw.Close()
			main.fileSearchStatus = utils.ScanStopped
		}
		fsw.Close()
	}()
}
