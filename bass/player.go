package bass

import (
	"github.com/pkg/errors"
	"time"
)

type StatusCallBack func(status ChannelStatus, elapsed float64)
type ChannelLoadedCallBack func(status ChannelStatus, totalTime float64, channel int64, metaInfo MusicMetaInfo)

type Player struct {
	initialized         bool
	currentChannel      int64
	currentVolume       float32
	mute                bool

	killUpdateRoutine chan bool

	// callbacks
	statusCallBackFunc    StatusCallBack
	channelLoadedCallBack ChannelLoadedCallBack
}

func New(device, frequency int, flag InitFlags) (*Player, error) {
	init, err := initBass(device, frequency, flag)
	if err != nil {
		return nil, errors.Wrapf(err.Err, "Failed to initialize lib bass with error: %+v", err)
	}
	player := Player{
		initialized:       init,
		currentChannel:    0,
		killUpdateRoutine: make(chan bool),
		mute:              false,
		currentVolume:     0,
	}
	player.updateRoutine()
	return &player, nil
}

func (p *Player) Free() error {
	if p.initialized {
		p.initialized = false
		p.killUpdateRoutine <- true
		if _, err := freeBass(); err != nil {
			return errors.Wrapf(err.Err, "Failed to free lib bass with error: %+v", err)
		}
	}
	return nil
}

func (p *Player) Load(path string) *Error {
	if !p.initialized {
		return errMsg(8)
	}
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
	status, _ := p.Status()
	if status == ChannelStatusPlaying {
		return true, nil
	}
	if status == ChannelStatusStopped || status == ChannelStatusStalled {
		return channelPlay(p.currentChannel, true)
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

func (p *Player) Stop() (bool, *Error) {
	if !p.initialized {
		return false, errMsg(8)
	}
	return channelStop(p.currentChannel)
}

func (p *Player) Status() (ChannelStatus, *Error) {
	if !p.initialized {
		return ChannelStatusStopped, errMsg(8)
	}
	return channelStatus(p.currentChannel), nil
}

func (p *Player) SetVolume(vol float32) *Error {
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
					elapsed = channelBytes2Seconds(p.currentChannel, channelPosition(p.currentChannel))
				}
				if p.statusCallBackFunc != nil {
					p.statusCallBackFunc(status, elapsed)
				}
				// very important to give some rest to CPU
				time.Sleep(time.Second / 3)
			}
		}
	}()
}
