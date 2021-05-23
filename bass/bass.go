package bass

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo darwin LDFLAGS: -L${SRCDIR}/../build -lbass
#include "bass.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

func initBass(device int, freq int, flags int) (bool, error) {
	if C.BASS_Init(C.int(device), C.DWORD(freq), C.DWORD(flags), nil, nil) != 0 {
		return true, nil
	} else {
		return false, errMsg(int(C.BASS_ErrorGetCode()))
	}
}

func freeBass() (bool, error) {
	if C.BASS_Free() != 0 {
		return true, nil
	} else {
		return false, errMsg(int(C.BASS_ErrorGetCode()))
	}
}

func musicLoad(file string, flags int) (int, error) {
	ch := C.BASS_MusicLoad(0, unsafe.Pointer(C.CString(file)), 0, 0, C.BASS_MUSIC_RAMPS|C.BASS_MUSIC_PRESCAN, 1)
	if ch != 0 {
		return int(ch), nil
	} else {
		return 0, errMsg(int(C.BASS_ErrorGetCode()))
	}
}

func channelPlay(ch int) (bool, error) {
	if C.BASS_ChannelPlay(C.DWORD(ch), 1) != 0 {
		return true, nil
	} else {
		return false, errMsg(int(C.BASS_ErrorGetCode()))
	}
}

func channelIsActive(ch int) (bool, error) {
	if C.BASS_ChannelIsActive(C.DWORD(ch)) != 0 {
		return true, nil
	}
	return false, errMsg(int(C.BASS_ErrorGetCode()))
}

func errMsg(c int) error {
	if c == 0 {
		return nil
	}
	codes := make(map[int]string)
	codes[1] = "memory error"
	codes[2] = "can't open the file"
	codes[3] = "can't find a free/valid driver"
	codes[4] = "the sample buffer was lost"
	codes[5] = "invalid handle"
	codes[6] = "unsupported sample format"
	codes[7] = "invalid position"
	codes[8] = "BASS_Init has not been successfully called"
	codes[9] = "BASS_Start has not been successfully called"
	codes[10] = "SSL/HTTPS support isn't available"
	codes[14] = "already initialized/paused/whatever"
	codes[18] = "can't get a free channel"
	codes[19] = "an illegal type was specified"
	codes[20] = "an illegal parameter was specified"
	codes[21] = "no 3D support"
	codes[22] = "no EAX support"
	codes[23] = "illegal device number"
	codes[24] = "not playing"
	codes[25] = "illegal sample rate"
	codes[27] = "the stream is not a file stream"
	codes[29] = "no hardware voices available"
	codes[31] = "the MOD music has no sequence data"
	codes[32] = "no internet connection could be opened"
	codes[33] = "couldn't create the file"
	codes[34] = "effects are not available"
	codes[37] = "requested data is not available"
	codes[38] = "the channel is/isn't a 'decoding channel'"
	codes[39] = "a sufficient DirectX version is not installed"
	codes[40] = "connection timedout"
	codes[41] = "unsupported file format"
	codes[42] = "unavailable speaker"
	codes[43] = "invalid BASS version (used by add-ons)"
	codes[44] = "codec is not available/supported"
	codes[45] = "the channel/file has ended"
	codes[46] = "the device is busy"
	codes[-1] = "some other mystery problem"
	return errors.New(codes[c])
}