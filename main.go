package main

import (
	"github.com/tejashwikalptaru/gotune/bass"
	"github.com/tejashwikalptaru/gotune/ui"
	"log"
)

func main() {
	player, err := bass.New(-1, 44100, bass.InitFlag3D | bass.InitFlagSTEREO)
	if err != nil {
		log.Fatal(err)
	}
	defer func(player *bass.Player) {
		err := player.Free()
		if err != nil {
			log.Fatal(err)
		}
	}(player)

	app := ui.NewMainWindow()

	app.PlayFunc(func() {
		status, _ := player.Status()
		if status == bass.ChannelStatusPlaying {
			_, _ = player.Pause()
			return
		}
		if status == bass.ChannelStatusStalled || status == bass.ChannelStatusPaused {
			_ ,_ = player.Play()
			return
		}
	})
	app.MuteFunc(func() {
		mute := !player.IsMute()
		player.Mute(mute)
		app.SetMuteState(mute)
	})
	app.StopFunc(func() {
		player.Stop()
	})
	app.SetFileOpenCallBack(func(filePath string) {
		err := player.Load(filePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		_, _ = player.Play()
	})
	player.ChannelLoadedCallBack(func(status bass.ChannelStatus, totalTime float64, channel int64, metaInfo bass.MusicMetaInfo) {
		app.SetTotalTime(totalTime)
		if metaInfo.IsMOD {
			app.SetSongName(metaInfo.ModInfo.Name)
			return
		}
		app.SetSongName(metaInfo.SongInfo.Title())
		if metaInfo.SongInfo.Picture() != nil {
			app.SetAlbumArt(metaInfo.SongInfo.Picture().Data)
		}
	})
	player.StatusCallBack(func(status bass.ChannelStatus, elapsed float64) {
		if status == bass.ChannelStatusPlaying {
			app.SetPlayState(true)
			app.SetCurrentTime(elapsed)
			return
		}
		if status == bass.ChannelStatusStopped || status == bass.ChannelStatusStalled {
			app.SetCurrentTime(0)
		}
		app.SetPlayState(false)
	})
	app.AddShortCuts()
	app.ShowAndRun()
}
