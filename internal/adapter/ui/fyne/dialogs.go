package fyne

import (
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// FileDialog is a helper for creating file open dialogs.
type FileDialog struct {
	window   fyne.Window
	callback func(string)
	logger   *slog.Logger
}

// NewFileDialog creates a new file dialog.
func NewFileDialog(window fyne.Window, callback func(string), logger *slog.Logger) *FileDialog {
	return &FileDialog{
		window:   window,
		callback: callback,
		logger:   logger,
	}
}

// Show displays the file dialog.
func (d *FileDialog) Show() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			d.logger.Error("file dialog error", slog.Any("error", err))
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
	logger   *slog.Logger
}

// NewFolderDialog creates a new folder dialog.
func NewFolderDialog(window fyne.Window, callback func(string), logger *slog.Logger) *FolderDialog {
	return &FolderDialog{
		window:   window,
		callback: callback,
		logger:   logger,
	}
}

// Show displays the folder dialog.
func (d *FolderDialog) Show() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			d.logger.Error("folder dialog error", slog.Any("error", err))
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
