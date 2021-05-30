package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/tejashwikalptaru/gotune/utils"
)

type FileSearchWindow struct {
	win             fyne.Window
	progress        *widget.ProgressBarInfinite
	progressParsing *widget.ProgressBar
	label           *widget.Label
}

func NewFileSearchWindow(app fyne.App) *FileSearchWindow {
	fsw := FileSearchWindow{}
	fsw.win = app.NewWindow("Please wait")
	fsw.win.Resize(fyne.Size{
		Width:  utils.WIDTH - 100,
		Height: utils.HEIGHT / 4,
	})
	fsw.win.SetFixedSize(true)

	fsw.progress = widget.NewProgressBarInfinite()
	fsw.progress.Start()
	fsw.progressParsing = widget.NewProgressBar()
	fsw.progressParsing.Hide()
	fsw.label = widget.NewLabel("Processing ....")

	fsw.win.SetContent(container.NewVBox(fsw.label, fsw.progress, fsw.progressParsing))
	return &fsw
}

func (fsw *FileSearchWindow) Show() {
	fsw.win.Show()
}

func (fsw *FileSearchWindow) Close() {
	fsw.win.Close()
}

func (fsw *FileSearchWindow) UpdateLabel(label string) {
	fsw.label.SetText(label)
}

func (fsw *FileSearchWindow) OnClosed(f func()) {
	fsw.win.SetOnClosed(f)
}
