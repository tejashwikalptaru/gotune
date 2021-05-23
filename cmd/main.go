package main

import (
	"fmt"
	"github.com/tejashwikalptaru/gotune/bass"
	"github.com/tejashwikalptaru/gotune/ui"
	"log"
)

func main() {
	app := ui.NewMainWindow()
	app.ShowAndRun()
	player, err := bass.New(-1, 44100, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer func(player *bass.Player) {
		err := player.Free()
		if err != nil {
			log.Fatal(err)
		}
	}(player)

	channel, err := player.MusicLoad("/Users/tejashwi/projects/personal/gotune/build/mktheme.it", bass.MusicRamps | bass.MusicPreScan)
	if err != nil {
		log.Fatal(err)
	}
	_ = player.SetVolume(channel, 5)
	_, err = player.Play(channel)
	if err != nil {
		log.Fatal(err)
	}
	for {
		active, err := player.IsChannelActive(channel)
		if err != nil {
			log.Fatal(err)
		}
		if !active {
			break
		}
	}
	fmt.Println("Done playing")
}
