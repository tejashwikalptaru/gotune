package fyne

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// FileDialog is a helper for creating file open dialogs.
type FileDialog struct {
	window   fyne.Window
	callback func(string)
}

// NewFileDialog creates a new file dialog.
func NewFileDialog(window fyne.Window, callback func(string)) *FileDialog {
	return &FileDialog{
		window:   window,
		callback: callback,
	}
}

// Show displays the file dialog.
func (d *FileDialog) Show() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()

		// Get file path
		filePath := reader.URI().Path()
		if d.callback != nil {
			d.callback(filePath)
		}
	}, d.window)
}

// FolderDialog is a helper for creating folder open dialogs.
type FolderDialog struct {
	window   fyne.Window
	callback func(string)
}

// NewFolderDialog creates a new folder dialog.
func NewFolderDialog(window fyne.Window, callback func(string)) *FolderDialog {
	return &FolderDialog{
		window:   window,
		callback: callback,
	}
}

// Show displays the folder dialog.
func (d *FolderDialog) Show() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			return
		}
		if uri == nil {
			return // User cancelled
		}

		// Get folder path
		folderPath := uri.Path()
		if d.callback != nil {
			d.callback(folderPath)
		}
	}, d.window)
}
