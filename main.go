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
		player.SavePlayList()
		err := player.Free()
		if err != nil {
			log.Fatal(err)
		}
	}(player)

	app := ui.NewMainWindow()
	defer func(m *ui.Main) {
		m.Free()
	}(app)

	player.OpenPlayList()
	player.SetVolume(100)
	app.SetVolume(100)
	app.VolumeUpdateCallBack(func(vol float64) {
		player.SetVolume(vol)
	})

	app.SetPlayListUpdater(player.AddPlayListFile)
	app.ProgressChanged(func(val float64) {
		player.SetChannelPosition(val)
	})

	app.PlayFunc(func() {
		status, _ := player.Status()
		if status == bass.ChannelStatusPlaying {
			_, _ = player.Pause()
			return
		}
		_ ,_ = player.Play()
	})
	app.MuteFunc(func() {
		mute := !player.IsMute()
		player.Mute(mute)
	})
	app.StopFunc(func() {
		player.Stop()
	})
	app.OnNextClick(func() {
		player.PlayNext()
	})
	app.OnPrevClick(func() {
		player.PlayPrevious()
	})
	app.SetFileOpenCallBack(func(filePath string) {
		err := player.Load(filePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		_, _ = player.Play()
	})
	player.ChannelLoadedCallBack(func(status bass.ChannelStatus, totalTime float64, channel int64, meta bass.MusicMetaInfo) {
		app.SetTotalTime(totalTime)
		app.SetSongName(meta.Info.Name)
		if meta.IsMOD {
			return
		}
		if meta.AdditionalMeta != nil && meta.AdditionalMeta.Picture() != nil {
			app.SetAlbumArt(meta.AdditionalMeta.Picture().Data)
		} else {
			app.ClearAlbumArt()
		}
	})
	player.StatusCallBack(func(status bass.ChannelStatus, elapsed float64, mute bool) {
		app.SetMuteState(mute)
		if status == bass.ChannelStatusPlaying {
			app.SetPlayState(true)
			app.SetCurrentTime(elapsed)
			return
		}
		app.SetPlayState(false)
	})
	app.AddShortCuts()
	app.ShowAndRun()
}
