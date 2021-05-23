package bass

import "github.com/pkg/errors"

type Player struct {
	initialized    bool
	currentChannel int
}

func New(device, frequency int, flag InitFlags) (*Player, error) {
	init, err := initBass(device, frequency, flag)
	if err != nil {
		return nil, errors.Wrapf(err.Err, "Failed to initialize lib bass with error: %+v", err)
	}
	return &Player{
		initialized: init,
		currentChannel: 0,
	}, nil
}

func (p *Player) Free() error {
	if p.initialized {
		if _, err := freeBass(); err != nil {
			return errors.Wrapf(err.Err, "Failed to free lib bass with error: %+v", err)
		}
		p.initialized = false
	}
	return nil
}

func (p *Player) MusicLoad(path string, flags int) *Error {
	if !p.initialized {
		return errMsg(8)
	}
	channel, err := musicLoad(path, flags)
	p.currentChannel = channel
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
