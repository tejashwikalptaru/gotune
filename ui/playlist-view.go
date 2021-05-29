package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/tejashwikalptaru/gotune/bass"
)

type PlayListView struct {
	item []bass.MusicMetaInfo
	win  fyne.Window
}

func NewPlayListView(app fyne.App, data []bass.MusicMetaInfo, currentIndex int) PlayListView {
	plView := PlayListView{item: data}
	plView.win = app.NewWindow("Playlist")
	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i].Path)
		})
	plView.win.Resize(fyne.Size{
		Width:  500,
		Height: 600,
	})
	list.OnSelected = func(id widget.ListItemID) {
		fmt.Println(id)
	}
	list.Select(currentIndex)
	plView.win.SetContent(container.NewVScroll(list))
	return plView
}
