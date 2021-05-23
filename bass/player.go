package bass

import "github.com/pkg/errors"

type Player struct {
	initialized bool
}

func New (device, frequency, flag int) (*Player, error) {
	init, err := initBass(device, frequency, flag)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to initialize lib bass with error: %+v", err)
	}
	return &Player{
		initialized: init,
	}, nil
}

func (p *Player) Free () error {
	if p.initialized {
		if _, err := freeBass(); err != nil {
			return errors.Wrapf(err, "Failed to free lib bass with error: %+v", err)
		}
		p.initialized = false
	}
	return nil
}

func (p *Player) MusicLoad (path string, flags int) (int, error) {
	if !p.initialized {
		return 0, errors.New("Player is not initialized")
	}
	return musicLoad(path, flags)
}

func (p *Player) Play (channel int) (bool, error) {
	if !p.initialized {
		return false, errors.New("Player is not initialized")
	}
	return channelPlay(channel)
}

func (p *Player) IsChannelActive (channel int) (bool, error) {
	if !p.initialized {
		return false, errors.New("Player is not initialized")
	}
	return channelIsActive(channel)
}