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
	player.SetUpdateElapsedTimeFunc(app.SetCurrentTime())
	player.SetUpdateEndTimeFunc(app.SetEndTime())

	app.PlayFunc(func() {
		status, err := player.Status()
		if err != nil {
			log.Fatal(err)
		}
		if status == bass.ChannelStatusPlaying {
			_, _ = player.Pause()
			app.SetPlayState(false)
			return
		}
		if status == bass.ChannelStatusStalled || status == bass.ChannelStatusPaused {
			_ ,_ = player.Play()
			app.SetPlayState(true)
			return
		}

		err = player.MusicLoad("/Users/tejashwi/projects/personal/gotune/build/mktheme.it", bass.MusicRamps | bass.MusicPreScan | bass.MusicAutoFree)
		if err != nil {
			log.Fatal(err)
		}
		_ = player.SetVolume(5)
		_, err = player.Play()
		if err != nil {
			log.Fatal(err)
		}
		app.SetPlayState(true)
	})
	app.StopFunc(func() {
		player.Stop()
		app.SetPlayState(false)
	})
	app.ShowAndRun()
}
