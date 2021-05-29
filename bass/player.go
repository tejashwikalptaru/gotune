package bass

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/tejashwikalptaru/gotune/utils"
	"io/ioutil"
	"time"
)

type StatusCallBack func(status ChannelStatus, elapsed float64, mute bool)
type ChannelLoadedCallBack func(status ChannelStatus, totalTime float64, channel int64, meta MusicMetaInfo)

type Player struct {
	initialized    bool
	currentChannel int64
	currentVolume  float64
	mute           bool

	killUpdateRoutine chan bool

	// callbacks
	statusCallBackFunc    StatusCallBack
	channelLoadedCallBack ChannelLoadedCallBack

	// playlist files
	playlist             []MusicMetaInfo
	currentPlaylistIndex int

	isManualStop bool
}

func New(device, frequency int, flag InitFlags) (*Player, error) {
	init, err := initBass(device, frequency, flag)
	if err != nil {
		return nil, errors.Wrapf(err.Err, "Failed to initialize lib bass with error: %+v", err)
	}
	player := Player{
		initialized:       init,
		currentChannel:    0,
		killUpdateRoutine: make(chan bool, 1),
		mute:              false,
		currentVolume:     0,
		isManualStop:      true,
		currentPlaylistIndex: -1,
	}
	player.updateRoutine()
	return &player, nil
}

func (p *Player) Free() error {
	if p.initialized {
		p.initialized = false
		p.killUpdateRoutine <- true
		p.Stop()
		if _, err := freeBass(); err != nil {
			return errors.Wrapf(err.Err, "Failed to free lib bass with error: %+v", err)
		}
	}
	return nil
}

func (p *Player) freeChannel() {
	if !streamFree(p.currentChannel) {
		musicFree(p.currentChannel)
	}
}

func (p *Player) Load(path string) *Error {
	if !p.initialized {
		return errMsg(8)
	}
	p.Stop()

	isMOD := false
	// try to load tracker modules
	channel, err := musicLoad(path, musicPreScan|musicRamps|streamAutoFree)
	if err != nil {
		// then try to load audio files
		channel, err = streamCreateFile(path, streamAutoFree)
		if err != nil {
			// give up!
			return err
		}
	} else {
		isMOD = true
	}
	p.currentChannel = channel
	p.SetVolume(p.currentVolume)
	if p.channelLoadedCallBack != nil {
		status, _ := p.Status()
		total := channelBytes2Seconds(p.currentChannel, channelLength(p.currentChannel))
		meta := findMeta(p.currentChannel, isMOD, path)
		p.channelLoadedCallBack(status, total, p.currentChannel, meta)
	}
	return err
}

func (p *Player) Play() (bool, *Error) {
	if !p.initialized {
		return false, errMsg(8)
	}
	p.isManualStop = false
	status, _ := p.Status()
	if status == ChannelStatusPlaying {
		return true, nil
	}
	if status == ChannelStatusStopped || status == ChannelStatusStalled {
		channelPlay(p.currentChannel, true)
	}
	// it should be paused then, resume play
	return channelPlay(p.currentChannel, false)
}

func (p *Player) Pause() (bool, *Error) {
	if !p.initialized {
		return false, errMsg(8)
	}
	return channelPause(p.currentChannel)
}

func (p *Player) Stop() *Error {
	if !p.initialized {
		return errMsg(8)
	}
	p.isManualStop = true
	// graceful stop
	channelSlideAttribute(p.currentChannel, ChannelAttribFREQ, 1000, 500)
	channelSlideAttribute(p.currentChannel, ChannelAttribVOL|ChannelAttribSLIDELOG, -1, 100)
	channelStop(p.currentChannel)
	p.freeChannel()
	return nil
}

func (p *Player) Status() (ChannelStatus, *Error) {
	if !p.initialized {
		return ChannelStatusStopped, errMsg(8)
	}
	return channelStatus(p.currentChannel), nil
}

func (p *Player) SetVolume(vol float64) *Error {
	if !p.initialized {
		return errMsg(8)
	}
	p.currentVolume = vol
	if _, err := channelSetVolume(p.currentChannel, vol); err != nil {
		return err
	}
	return nil
}

func (p *Player) Mute(mute bool) *Error {
	if !p.initialized {
		return errMsg(8)
	}
	if mute {
		p.mute = true
		temp := p.currentVolume
		err := p.SetVolume(0)
		p.currentVolume = temp
		return err
	}
	p.mute = false
	return p.SetVolume(p.currentVolume)
}

func (p *Player) IsMute() bool {
	return p.mute
}

func (p *Player) StatusCallBack(f StatusCallBack) {
	p.statusCallBackFunc = f
}

func (p *Player) ChannelLoadedCallBack(f ChannelLoadedCallBack) {
	p.channelLoadedCallBack = f
}

func (p *Player) updateRoutine() {
	go func() {
		var elapsed float64
		for {
			select {
			case <-p.killUpdateRoutine:
				close(p.killUpdateRoutine)
				return
			default:
				status, _ := p.Status()
				if status == ChannelStatusPlaying {
					elapsed = channelBytes2Seconds(p.currentChannel, channelGetPosition(p.currentChannel))
				}
				if p.statusCallBackFunc != nil {
					p.statusCallBackFunc(status, elapsed, p.IsMute())
				}
				if status == ChannelStatusStopped && !p.isManualStop && len(p.playlist) > 0 && p.currentPlaylistIndex < len(p.playlist) {
					p.currentPlaylistIndex++
					p.Stop()
					p.Load(p.playlist[p.currentPlaylistIndex].Path)
					p.Play()
				}
				// very important to give some rest to CPU
				time.Sleep(time.Second / 3)
			}
		}
	}()
}

func (p *Player) PlayNext() {
	if !p.initialized {
		return
	}
	if len(p.playlist) == 0 {
		return
	}
	if p.currentPlaylistIndex < len(p.playlist) {
		p.currentPlaylistIndex++
		p.Stop()
		p.Load(p.playlist[p.currentPlaylistIndex].Path)
		p.Play()
	}
}

func (p *Player) PlayPrevious() {
	if !p.initialized {
		return
	}
	if len(p.playlist) == 0 {
		return
	}
	if p.currentPlaylistIndex > 0 {
		p.currentPlaylistIndex--
		p.Stop()
		p.Load(p.playlist[p.currentPlaylistIndex].Path)
		p.Play()
	}
}

func (p *Player) SetChannelPosition(val float64) {
	if !p.initialized {
		return
	}
	bytes := channelSeconds2Bytes(p.currentChannel, val)
	channelSetPosition(p.currentChannel, bytes)
}

func (p *Player) AddPlayListFile(path string) {
	p.playlist = append(p.playlist, ParseFile(path))
}

func (p *Player) GetPlayList() []MusicMetaInfo {
	return p.playlist
}

func (p *Player) SavePlayList() {
	if len(p.playlist) == 0 {
		return
	}
	jsonByte, err := json.Marshal(p.playlist)
	if err != nil {
		utils.ShowError(true, "Failed", err.Error())
		return
	}
	err = ioutil.WriteFile("/Users/tejashwi/projects/personal/gotune/playlist.gtp", jsonByte, 0644)
	if err != nil {
		utils.ShowError(true, "Failed", err.Error())
	}
}

func (p *Player) OpenPlayList() {
	file, err := ioutil.ReadFile("/Users/tejashwi/projects/personal/gotune/playlist.gtp")
	if err != nil {
		return
	}
	playlist := make([]MusicMetaInfo, 0)
	err = json.Unmarshal(file, &playlist)
	if err != nil {
		return
	}
	if len(playlist) > 0 {
		p.playlist = playlist
		p.currentPlaylistIndex = -1 //reset
	}
}
