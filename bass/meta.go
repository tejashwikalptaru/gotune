package bass

import (
	"github.com/dhowden/tag"
	"github.com/tejashwikalptaru/gotune/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type BasicMeta struct {
	Name       string `json:"name"`
	Message    string `json:"message"`
	Author     string `json:"author"`
	Instrument string `json:"instrument"`
	Album      string `json:"album"`
	Artist     string `json:"artist"`
}

type MusicMetaInfo struct {
	IsMOD          bool         `json:"isMod"`
	Path           string       `json:"path"`
	Info           BasicMeta    `json:"modInfo"`
	AdditionalMeta tag.Metadata `json:"-"`
}

func ParseFile(path string) MusicMetaInfo {
	mod := utils.IsMod(path)
	if mod {
		channel, err := musicLoad(path, streamGetData|streamAutoFree)
		if err != nil {
			return MusicMetaInfo{
				IsMOD: true,
				Path:  path,
				Info: BasicMeta{
					Name: filepath.Base(path),
				},
			}
		}
		meta := findMeta(channel, true, path)
		musicFree(channel)
		return meta
	}
	return findMeta(0, false, path)
}

func findMeta(ch int64, isMod bool, path string) MusicMetaInfo {
	meta := MusicMetaInfo{IsMOD: isMod, Path: path}

	if isMod {
		meta.Info.Name = strings.TrimSpace(channelGetMODTags(ch, TagMusicNAME))
		meta.Info.Message = strings.TrimSpace(channelGetMODTags(ch, TagMusicMESSAGE))
		meta.Info.Author = strings.TrimSpace(channelGetMODTags(ch, TagMusicAUTH))
		meta.Info.Instrument = strings.TrimSpace(channelGetMODTags(ch, TagMusicINST))
		if meta.Info.Name == "" {
			meta.Info.Name = filepath.Base(path)
		}
		return meta
	}
	// get audio meta data
	meta.Info.Name = filepath.Base(path)
	currentFile, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
		return meta
	}
	defer func(currentFile *os.File) {
		err := currentFile.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(currentFile)

	metadata, _ := tag.ReadFrom(currentFile)
	if metadata == nil {
		return meta
	}
	if strings.TrimSpace(metadata.Title()) != "" {
		meta.Info.Name = strings.TrimSpace(metadata.Title())
	}
	meta.Info.Album = strings.TrimSpace(metadata.Album())
	meta.Info.Artist = strings.TrimSpace(metadata.Artist())
	meta.Info.Message = strings.TrimSpace(metadata.Composer())
	meta.AdditionalMeta = metadata
	return meta
}
