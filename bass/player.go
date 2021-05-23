package bass

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
)

type Player struct {
	initialized    bool
	currentChannel int
	quitTimeUpdate chan bool
	timeElapsed    float64

	// update callbacks
	updateElapsedTimeFunc func(text string)
	updateEndTimeFunc func(text string)
}

func New(device, frequency int, flag InitFlags) (*Player, error) {
	init, err := initBass(device, frequency, flag)
	if err != nil {
		return nil, errors.Wrapf(err.Err, "Failed to initialize lib bass with error: %+v", err)
	}
	player := Player{
		initialized:    init,
		currentChannel: 0,
		quitTimeUpdate: make(chan bool),
	}
	player.threadTimeUpdate()
	return &player, nil
}

func (p *Player) Free() error {
	if p.initialized {
		if _, err := freeBass(); err != nil {
			return errors.Wrapf(err.Err, "Failed to free lib bass with error: %+v", err)
		}
		p.initialized = false
		// kill time update routine
		p.quitTimeUpdate <- true
	}
	return nil
}

func (p *Player) MusicLoad(path string, flags int) *Error {
	if !p.initialized {
		return errMsg(8)
	}
	channel, err := musicLoad(path, flags)
	p.currentChannel = channel

	if p.updateEndTimeFunc != nil {
		posByte := channelLength(p.currentChannel)
		totalTime := channelPositionToSeconds(p.currentChannel, posByte)
		fmt.Println(totalTime)
		p.updateEndTimeFunc(fmt.Sprintf("%.2d:%.2d", int(totalTime/60), int(math.Mod(totalTime, 60))))
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
	if _, err := channelSetVolume(p.currentChannel, vol); err != nil {
		return err
	}
	return nil
}

func (p *Player) SetUpdateElapsedTimeFunc ( f func(text string)) {
	p.updateElapsedTimeFunc = f
}

func (p *Player) SetUpdateEndTimeFunc ( f func(text string)) {
	p.updateEndTimeFunc = f
}

func (p *Player) threadTimeUpdate() {
	go func() {
		for {
			select {
			case <-p.quitTimeUpdate:
				return
			default:
				status, _ := p.Status()
				if status == ChannelStatusPlaying {
					posByte := channelPosition(p.currentChannel)
					p.timeElapsed = channelPositionToSeconds(p.currentChannel, posByte)
					if p.updateElapsedTimeFunc != nil {
						p.updateElapsedTimeFunc(fmt.Sprintf("%.2d:%.2d", int(p.timeElapsed/60), int(math.Mod(p.timeElapsed, 60))))
					}
				}
				if status == ChannelStatusStopped || status == ChannelStatusStalled {
					p.timeElapsed = 0
					if p.updateElapsedTimeFunc != nil {
						p.updateElapsedTimeFunc(fmt.Sprintf("%.2d:%.2d", int(p.timeElapsed/60), int(math.Mod(p.timeElapsed, 60))))
					}
				}
			}
		}
	}()
}
