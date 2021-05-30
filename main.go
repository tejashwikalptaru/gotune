package main

import (
	"github.com/tejashwikalptaru/gotune/bass"
	"github.com/tejashwikalptaru/gotune/ui"
	"log"
)

var plView *ui.PlayListView

func handleMainWindowButtonClicks(app *ui.Main, player *bass.Player) {
	app.OnPlay(func() {
		status, _ := player.Status()
		if status == bass.ChannelStatusPlaying {
			_, _ = player.Pause()
			return
		}
		_ ,_ = player.Play()
	})
	app.OnMute(func() {
		mute := !player.IsMute()
		player.Mute(mute)
	})
	app.OnStop(func() {
		player.Stop()
	})
	app.OnNext(func() {
		player.PlayNext()
	})
	app.OnPrev(func() {
		player.PlayPrevious()
	})
	app.OnLoop(func() {
		loop := !player.IsLoop()
		player.Loop(loop)
		app.SetLoopState(loop)
	})
}

func handleMainWindowCallBacks(app *ui.Main, player *bass.Player) {
	app.VolumeUpdateCallBack(func(vol float64) {
		player.SetVolume(vol)
	})
	app.SetPlayListUpdater(func(path string) {
		player.AddToQueue(path, false)
	})
	app.ProgressChanged(func(val float64) {
		player.SetChannelPosition(val)
	})
	app.SetOpenPlayListCallBack(func() {
		if plView != nil {
			plView.Show()
			return
		}
		plView = ui.NewPlayListView(app.GetApp(), player.GetPlayList(), player.GetPlaylistIndex(), func() {
			plView = nil
		}, func(path string) {
			index := player.PlayFromQueue(path)
			if index > -1 {
				plView.SetSelected(path)
			}
		})
		plView.Show()
	})
	app.SetFileOpenCallBack(func(filePath string) {
		err := player.AddToQueue(filePath, false)
		if err != nil {
			log.Fatal(err)
			return
		}
	})
	app.OnAppClose(func() {
		app.QuitApp()
	})
}

func handlePlayerCallbacks(app *ui.Main, player *bass.Player) {
	player.ChannelLoadedCallBack(func(status bass.ChannelStatus, totalTime float64, channel int64, meta bass.MusicMetaInfo, currentQueueIndex int) {
		app.SetTotalTime(totalTime)
		app.SetSongName(meta.Info.Name)
		if plView != nil {
			plView.SetSelected(meta.Path)
		}
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
	player.FileAddedCallBack(func(info bass.MusicMetaInfo) {
		if plView != nil {
			plView.FileAdded(info)
		}
		// save history
		app.SaveHistory(player.GetHistory())
	})
}

func createPlayer() *bass.Player {
	player, err := bass.New(-1, 44100, bass.InitFlag3D | bass.InitFlagSTEREO)
	if err != nil {
		log.Fatal(err)
	}
	player.SetVolume(100)
	return player
}

func createMainWindow() *ui.Main {
	app := ui.NewMainWindow()
	app.SetVolume(100)
	return app
}

func main() {
	// audio player instance
	player := createPlayer()
	defer func(player *bass.Player) {
		//player.SaveHistory()
		err := player.Free()
		if err != nil {
			log.Fatal(err)
		}
	}(player)

	// main window instance
	app := createMainWindow()
	defer func(app *ui.Main) {
		app.Free()
	}(app)

	// load history
	player.LoadHistory(app.GetHistory())

	// main window button clicks callback
	handleMainWindowButtonClicks(app, player)

	// main window callbacks
	handleMainWindowCallBacks(app, player)

	// player callbacks
	handlePlayerCallbacks(app, player)

	// add app shortcuts
	app.AddShortCuts()

	// finally run
	app.ShowAndRun()
}
