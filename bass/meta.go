package bass

import (
	"github.com/dhowden/tag"
	"log"
	"os"
)

type ModData struct {
	Name       string
	Message    string
	Author     string
	Instrument string
}

type MusicMetaInfo struct {
	IsMOD bool
	Path  string
	ModInfo ModData
	SongInfo tag.Metadata
}

func findMeta(ch int64, isMod bool, path string) MusicMetaInfo {
	info := MusicMetaInfo{IsMOD: isMod, Path: path}

	if isMod {
		modInfo := ModData{}
		modInfo.Name = channelGetMODTags(ch, TagMusicNAME)
		modInfo.Message = channelGetMODTags(ch, TagMusicMESSAGE)
		modInfo.Author = channelGetMODTags(ch, TagMusicAUTH)
		modInfo.Instrument = channelGetMODTags(ch, TagMusicINST)
		info.ModInfo = modInfo
		return info
	}
	// get audio meta data
	currentFile, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
		return info
	}
	metadata, _ := tag.ReadFrom(currentFile)
	info.SongInfo = metadata
	err = currentFile.Close()
	if err != nil {
		log.Fatal(err)
		return info
	}
	return info
}
