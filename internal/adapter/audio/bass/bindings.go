// Package bass provides low-level CGO bindings to the BASS audio library.
// These functions are internal and should not be used directly - use the Engine instead.
package bass

/*
#include <stdlib.h>
#include "bass.h"
*/
import "C"
import (
	"time"
	"unsafe"

	"github.com/tejashwikalptaru/gotune/internal/domain"
)

// bassInit initializes the BASS library.
func bassInit(device int, freq int, flags int) error {
	if C.BASS_Init(C.int(device), C.DWORD(freq), C.DWORD(flags), nil, nil) == 0 {
		return createBassError("initialize", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassFree releases the BASS library resources.
func bassFree() error {
	if C.BASS_Free() == 0 {
		return createBassError("free", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassMusicLoad loads a MOD music file.
func bassMusicLoad(filePath string, flags int) (int64, error) {
	cPath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cPath))

	handle := C.BASS_MusicLoad(0, unsafe.Pointer(cPath), 0, 0, C.DWORD(flags), 1)
	if handle == 0 {
		return 0, createBassError("load_music", filePath, C.BASS_ErrorGetCode())
	}
	return int64(handle), nil
}

// bassMusicFree frees a MOD music handle.
func bassMusicFree(handle int64) bool {
	return C.BASS_MusicFree(C.DWORD(handle)) != 0
}

// bassStreamCreateFile loads a stream from a file.
func bassStreamCreateFile(filePath string, flags int) (int64, error) {
	cPath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cPath))

	handle := C.BASS_StreamCreateFile(0, unsafe.Pointer(cPath), 0, 0, C.DWORD(flags))
	if handle == 0 {
		return 0, createBassError("load_stream", filePath, C.BASS_ErrorGetCode())
	}
	return int64(handle), nil
}

// bassStreamFree frees a stream handle.
func bassStreamFree(handle int64) bool {
	return C.BASS_StreamFree(C.DWORD(handle)) != 0
}

// bassChannelPlay starts or resumes playback.
func bassChannelPlay(handle int64, restart bool) error {
	restartVal := C.int(0)
	if restart {
		restartVal = 1
	}

	if C.BASS_ChannelPlay(C.DWORD(handle), restartVal) == 0 {
		return createBassError("play", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassChannelPause pauses playback.
func bassChannelPause(handle int64) error {
	if C.BASS_ChannelPause(C.DWORD(handle)) == 0 {
		return createBassError("pause", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassChannelStop stops playback.
func bassChannelStop(handle int64) error {
	if C.BASS_ChannelStop(C.DWORD(handle)) == 0 {
		return createBassError("stop", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassChannelIsActive returns the playback status.
func bassChannelIsActive(handle int64) domain.PlaybackStatus {
	status := C.BASS_ChannelIsActive(C.DWORD(handle))

	switch status {
	case C.BASS_ACTIVE_STOPPED:
		return domain.StatusStopped
	case C.BASS_ACTIVE_PLAYING:
		return domain.StatusPlaying
	case C.BASS_ACTIVE_PAUSED:
		return domain.StatusPaused
	case C.BASS_ACTIVE_STALLED:
		return domain.StatusStalled
	default:
		return domain.StatusStopped
	}
}

// bassChannelGetLength returns the channel length in bytes.
func bassChannelGetLength(handle int64) uint64 {
	return uint64(C.BASS_ChannelGetLength(C.DWORD(handle), C.BASS_POS_BYTE))
}

// bassChannelGetPosition returns the current position in bytes.
func bassChannelGetPosition(handle int64) uint64 {
	return uint64(C.BASS_ChannelGetPosition(C.DWORD(handle), C.BASS_POS_BYTE))
}

// bassChannelSetPosition sets the position in bytes.
func bassChannelSetPosition(handle int64, pos uint64) error {
	if C.BASS_ChannelSetPosition(C.DWORD(handle), C.QWORD(pos), C.BASS_POS_BYTE) == 0 {
		return createBassError("seek", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassChannelBytes2Seconds converts bytes to seconds.
func bassChannelBytes2Seconds(handle int64, pos uint64) time.Duration {
	seconds := C.BASS_ChannelBytes2Seconds(C.DWORD(handle), C.QWORD(pos))
	return time.Duration(float64(seconds) * float64(time.Second))
}

// bassChannelSeconds2Bytes converts seconds to bytes.
func bassChannelSeconds2Bytes(handle int64, duration time.Duration) uint64 {
	seconds := duration.Seconds()
	bytes := C.BASS_ChannelSeconds2Bytes(C.DWORD(handle), C.double(seconds))
	return uint64(bytes)
}

// bassChannelSetAttribute sets a channel attribute.
func bassChannelSetAttribute(handle int64, attrib ChannelAttributes, value float32) error {
	if C.BASS_ChannelSetAttribute(C.DWORD(handle), C.DWORD(attrib), C.float(value)) == 0 {
		return createBassError("set_attribute", "", C.BASS_ErrorGetCode())
	}
	return nil
}

// bassChannelGetAttribute gets a channel attribute.
func bassChannelGetAttribute(handle int64, attrib ChannelAttributes) (float32, error) {
	var value C.float
	if C.BASS_ChannelGetAttribute(C.DWORD(handle), C.DWORD(attrib), &value) == 0 {
		return 0, createBassError("get_attribute", "", C.BASS_ErrorGetCode())
	}
	return float32(value), nil
}

// bassChannelSlideAttribute slides a channel attribute.
func bassChannelSlideAttribute(handle int64, attrib ChannelAttributes, value float32, timeMs int) bool {
	return C.BASS_ChannelSlideAttribute(C.DWORD(handle), C.DWORD(attrib), C.float(value), C.DWORD(timeMs)) != 0
}

// bassChannelGetTags gets channel tags (for MOD files).
func bassChannelGetTags(handle int64, tag Tag) string {
	tags := C.BASS_ChannelGetTags(C.DWORD(handle), C.DWORD(tag))
	if tags == nil {
		return ""
	}
	return C.GoString((*C.char)(tags))
}

// createBassError creates an AudioEngineError from a BASS error code.
func createBassError(op, path string, code C.int) error {
	errorCode := ErrorCode(code)
	message := errorCodeToMessage(errorCode)

	return domain.NewAudioEngineError(op, path, int(code), message, nil)
}
