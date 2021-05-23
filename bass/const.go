package bass

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo darwin LDFLAGS: -L${SRCDIR}/../build -lbass
#include "bass.h"
*/
import "C"

const MusicRamps int = C.BASS_MUSIC_RAMPS
const MusicPreScan int = C.BASS_MUSIC_PRESCAN
const AttribVol int = C.BASS_ATTRIB_VOL

type Error struct {
	Err  error
	Code ErrorCode
}

type ChannelStatus int

const (
	ChannelStatusStopped ChannelStatus = 0
	ChannelStatusPlaying ChannelStatus = 1
	ChannelStatusStalled ChannelStatus = 2
	ChannelStatusPaused  ChannelStatus = 3
)

type ErrorCode int

const (
	ErrorOK       ErrorCode = 0  // all is OK
	ErrorMEM      ErrorCode = 1  // memory error
	ErrorFILEOPEN ErrorCode = 2  // can't open the file
	ErrorDRIVER   ErrorCode = 3  // can't find a free/valid driver
	ErrorBUFLOST  ErrorCode = 4  // the sample buffer was lost
	ErrorHANDLE   ErrorCode = 5  // invalid handle
	ErrorFORMAT   ErrorCode = 6  // unsupported sample format
	ErrorPOSITION ErrorCode = 7  // invalid position
	ErrorINIT     ErrorCode = 8  // BASS_Init has not been successfully called
	ErrorSTART    ErrorCode = 9  // BASS_Start has not been successfully called
	ErrorSSL      ErrorCode = 10 // SSL/HTTPS support isn't available
	ErrorALREADY  ErrorCode = 14 // already initialized/paused/whatever
	ErrorNOCHAN   ErrorCode = 18 // can't get a free channel
	ErrorILLTYPE  ErrorCode = 19 // an illegal type was specified
	ErrorILLPARAM ErrorCode = 20 // an illegal parameter was specified
	ErrorNO3D     ErrorCode = 21 // no 3D support
	ErrorNOEAX    ErrorCode = 22 // no EAX support
	ErrorDEVICE   ErrorCode = 23 // illegal device number
	ErrorNOPLAY   ErrorCode = 24 // not playing
	ErrorFREQ     ErrorCode = 25 // illegal sample rate
	ErrorNOTFILE  ErrorCode = 27 // the stream is not a file stream
	ErrorNOHW     ErrorCode = 29 // no hardware voices available
	ErrorEMPTY    ErrorCode = 31 // the MOD music has no sequence data
	ErrorNONET    ErrorCode = 32 // no internet connection could be opened
	ErrorCREATE   ErrorCode = 33 // couldn't create the file
	ErrorNOFX     ErrorCode = 34 // effects are not available
	ErrorNOTAVAIL ErrorCode = 37 // requested data is not available
	ErrorDECODE   ErrorCode = 38 // the channel is/isn't a "decoding channel"
	ErrorDX       ErrorCode = 39 // a sufficient DirectX version is not installed
	ErrorTIMEOUT  ErrorCode = 40 // connection timed out
	ErrorFILEFORM ErrorCode = 41 // unsupported file format
	ErrorSPEAKER  ErrorCode = 42 // unavailable speaker
	ErrorVERSION  ErrorCode = 43 // invalid BASS version (used by add-ons)
	ErrorCODEC    ErrorCode = 44 // codec is not available/supported
	ErrorENDED    ErrorCode = 45 // the channel/file has ended
	ErrorBUSY     ErrorCode = 46 // the device is busy
	ErrorUNKNOWN  ErrorCode = -1 // some other mystery problem
)

type InitFlags int
const (
	InitFlag8BITS      InitFlags = 1      // 8 bit
	InitFlagMONO       InitFlags = 2      // mono
	InitFlag3D         InitFlags = 4      // enable 3D functionality
	InitFlag16BITS     InitFlags = 8      // limit output to 16 bit
	InitFlagLATENCY    InitFlags = 0x100  // calculate device latency (BASS_INFO struct)
	InitFlagCPSPEAKERS InitFlags = 0x400  // detect speakers via Windows control panel
	InitFlagSPEAKERS   InitFlags = 0x800  // force enabling of speaker assignment
	InitFlagNOSPEAKER  InitFlags = 0x1000 // ignore speaker arrangement
	InitFlagDMIX       InitFlags = 0x2000 // use ALSA "dmix" plugin
	InitFlagFREQ       InitFlags = 0x4000 // set device sample rate
	InitFlagSTEREO     InitFlags = 0x8000 // limit output to stereo
)
