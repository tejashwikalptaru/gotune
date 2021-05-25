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

		err := player.MusicLoad("/Users/tejashwi/projects/personal/gotune/build/mktheme.it", bass.MusicRamps | bass.MusicPreScan | bass.MusicAutoFree)
		if err != nil {
			log.Fatal(err)
			return
		}
		_, _ = player.Play()
	})
	app.MuteFunc(func() {
		mute := !player.IsMute()
		player.Mute(mute)
		app.SetMuteState(mute)
	})
	app.StopFunc(func() {
		player.Stop()
	})
	player.ChannelLoadedCallBack(func(status bass.ChannelStatus, totalTime float64, channel int) {
		app.SetTotalTime(totalTime)
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
