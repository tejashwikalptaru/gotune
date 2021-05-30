package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/tejashwikalptaru/gotune/bass"
	"github.com/tejashwikalptaru/gotune/ui/extended"
	"strings"
)

type PlayListView struct {
	win            fyne.Window
	list           *widget.List
	searchText     *widget.Entry
	data           []bass.MusicMetaInfo
	mainCollection []bass.MusicMetaInfo
	selectedPath   string

	onClose        func()
	onItemSelected func(path string)
}

func NewPlayListView(app fyne.App, data []bass.MusicMetaInfo, currentIndex int, onClose func(), onSelected func(path string)) *PlayListView {
	plView := PlayListView{data: data, mainCollection: data, onClose: onClose, onItemSelected: onSelected}
	plView.win = app.NewWindow(fmt.Sprintf("Playlist (%d items)", len(data)))
	plView.win.SetOnClosed(plView.onClose)
	plView.list = widget.NewList(plView.getDataLength, plView.createListCell, plView.updateCell)
	plView.win.Resize(fyne.Size{
		Width:  500,
		Height: 600,
	})
	if len(data) > currentIndex && currentIndex > -1  {
		plView.selectedPath = data[currentIndex].Path
	}
	plView.searchText = widget.NewEntry()
	plView.searchText.PlaceHolder = "Search..."
	plView.searchText.OnChanged = plView.searchCollection
	plView.win.SetContent(container.NewBorder(plView.searchText, nil, nil, nil, container.NewVScroll(plView.list)))
	return &plView
}

func (plView *PlayListView) getDataLength() int {
	return len(plView.data)
}

func (plView *PlayListView) createListCell() fyne.CanvasObject {
	return extended.NewDoubleTapLabel(func(index int) {
		if plView.onItemSelected != nil {
			plView.onItemSelected(plView.data[index].Path)
		}
	})
}

func (plView *PlayListView) updateCell(i widget.ListItemID, o fyne.CanvasObject) {
	o.(*extended.DoubleTapLabel).Set(plView.data[i].Info.Name, i)
}

func (plView *PlayListView) searchCollection(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		// restore main collection
		plView.data = plView.mainCollection
		// highlight currently selected
		for _, v := range plView.data {
			if v.Path == plView.selectedPath {
				plView.SetSelected(v.Path)
			}
		}
		plView.list.Refresh()
		plView.win.SetTitle(fmt.Sprintf("Playlist (%d items)", len(plView.data)))
		return
	}
	if len(plView.data) == 0 {
		plView.win.SetTitle(fmt.Sprintf("Playlist (%d items)", len(plView.data)))
		plView.list.Refresh()
		return
	}
	temp := make([]bass.MusicMetaInfo, 0)
	for _, item := range plView.mainCollection {
		if strings.Contains(strings.ToLower(item.Path), strings.ToLower(query)) {
			temp = plView.addIfNotExists(temp, item)
		}
		if strings.Contains(strings.ToLower(item.Info.Name), strings.ToLower(query)) {
			temp = plView.addIfNotExists(temp, item)
		}
		if strings.Contains(strings.ToLower(item.Info.Author), strings.ToLower(query)) {
			temp = plView.addIfNotExists(temp, item)
		}
		if strings.Contains(strings.ToLower(item.Info.Artist), strings.ToLower(query)) {
			temp = plView.addIfNotExists(temp, item)
		}
	}
	plView.data = temp
	plView.win.SetTitle(fmt.Sprintf("Playlist (%d items)", len(plView.data)))
	plView.list.Refresh()
}

func (plView *PlayListView) addIfNotExists(array []bass.MusicMetaInfo, item bass.MusicMetaInfo) []bass.MusicMetaInfo {
	found := false
	for _, v := range array {
		if strings.EqualFold(v.Path, item.Path) {
			found = true
			break
		}
	}
	if found {
		return array
	}
	return append(array, item)
}

func (plView *PlayListView) Show() {
	plView.win.Show()
	plView.SetSelected(plView.selectedPath)
}

func (plView *PlayListView) SetSelected(path string) {
	if path == "" {
		return
	}
	for i, v := range plView.data {
		if v.Path == path {
			plView.list.Select(i)
			plView.selectedPath = path
			plView.list.Refresh()
			break
		}
	}
}

func (plView *PlayListView) FileAdded(info bass.MusicMetaInfo) {
	plView.data = append(plView.data, info)
	plView.mainCollection = append(plView.mainCollection, info)
	plView.list.Refresh()
	plView.win.SetTitle(fmt.Sprintf("Playlist (%d items)", len(plView.data)))
}
