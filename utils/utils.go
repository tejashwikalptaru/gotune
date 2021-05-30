package utils

import (
	"errors"
	"github.com/sqweek/dialog"
	"os"
	"path/filepath"
	"strings"
)

const (
	APPNAME string = "Go Tune"
	HISTORY string = "Go Tune History"
	HEIGHT float32 = 300
	WIDTH float32 = 450
	PACKAGE string = "com.tejashwi.gotune.preferences"
)

var supportedFormats = []string{
	".mp3",
	".mp2",
	".mp1",
	".ogg",
	".wav",
	".aif",
	".mo3",
	".it",
	".xm",
	".s3m",
	".mtm",
	".mod",
	".umx",
	".cda",
	".fla",
	".flac",
	".oga",
	".wma",
	".wv",
	".aac",
	".m4a",
	".m4b",
	".mp4",
	".ac3",
	".adx",
	".aix",
	".ape",
	".mac",
	".mpc",
	".mp+",
	".mpp",
	".ofr",
	".ofs",
	".tta",
}
type ScanFolderCallBack func(string)
var ScanCancelled = errors.New("scanning cancelled by user")
type ScanStatus int
const (
	ScanStopped ScanStatus = 0
	ScanRunning ScanStatus = 1
)

func ScanFolder(folderPath string, callback ScanFolderCallBack, status *ScanStatus) ([]string, error) {
	result := make([]string, 0)
	addPath := func(path string, info os.FileInfo, err error) error {
		if *status == ScanStopped {
			return ScanCancelled
		}
		if err != nil {
			return err
		}
		if !info.IsDir() && Contains(supportedFormats, filepath.Ext(path)) {
			if callback != nil {
				callback(path)
			}
			result = append(result, path)
		}
		return nil
	}
	err := filepath.Walk(folderPath, addPath)
	return result, err
}

func Contains(arr []string, input string) bool {
	for _, v := range arr {
		if strings.EqualFold(v, input) {
			return true
		}
	}
	return false
}

func ShowError(err bool, title, format string, args ...interface{}) {
	builder := dialog.Message(format, args).Title(title)
	if err {
		builder.Error()
		return
	}
	builder.Info()
}

func ShowInfo(title, format string, args ...interface{}) bool {
	return dialog.Message(format, args).Title(title).YesNo()
}

func removeDot(input []string) []string {
	removed := make([]string, 0)
	for _, val := range input {
		removed = append(removed, strings.Replace(val,".", "", -1))
	}
	return removed
}

func OpenFile(title string) (string, error) {
	return dialog.File().Title(title).Filter("Audio Files", removeDot(supportedFormats)...).Load()
}

func OpenFolder(title string) (string, error) {
	return dialog.Directory().Title(title).Browse()
}

func IsMod(path string) bool {
	var mod = []string{
		".it",
		".xm",
		".s3m",
		".mtm",
		".mod",
		".umx",
	}
	return Contains(mod, filepath.Ext(path))
}
