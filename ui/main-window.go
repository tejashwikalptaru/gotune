package ui

import (
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/tejashwikalptaru/gotune/res"
	"github.com/tejashwikalptaru/gotune/utils"
	"image"
)

type Main struct {
	app    fyne.App
	window fyne.Window

	prevButton *widget.Button
	playButton *widget.Button
	stopButton *widget.Button
	nextButton *widget.Button
	muteButton *widget.Button
	songInfo   *widget.Label

	albumArt *canvas.Image
}

func NewMainWindow() *Main {
	var window Main
	a := app.New()
	w := a.NewWindow(utils.APPNAME)
	w.Resize(fyne.Size{
		Width:  utils.WIDTH,
		Height: utils.HEIGHT,
	})
	w.SetFixedSize(true)
	splitContainer := container.NewBorder(nil, window.controls(), nil, nil, window.display())
	w.SetContent(container.NewPadded(splitContainer))
	w.CenterOnScreen()
	window.app = a
	window.window = w
	return &window
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
	main.songInfo = widget.NewLabel("Here goes the song name with the details of song in it")

	buttonsHolder := container.NewHBox(main.prevButton, main.playButton, main.stopButton, main.nextButton, main.muteButton, main.songInfo)

	mediaSlider := widget.NewSlider(0, 100)
	startTime := widget.NewLabel("00:00:00")
	endTime := widget.NewLabel("00:00:00")

	sliderHolder := container.NewGridWithColumns(3, startTime, mediaSlider, endTime)

	return container.NewVBox(buttonsHolder, sliderHolder)
}

func (main *Main) ShowAndRun() {
	main.window.ShowAndRun()
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
