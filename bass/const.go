package bass

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo darwin LDFLAGS: -L${SRCDIR}/../build -lbass
#include "bass.h"
*/
import "C"

const MusicRamps int = C.BASS_MUSIC_RAMPS
const MusicPreScan int = C.BASS_MUSIC_PRESCAN
const MusicAutoFree int = C.BASS_STREAM_AUTOFREE
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
	ErrorOK       ErrorCode = C.BASS_OK             // all is OK
	ErrorMEM      ErrorCode = C.BASS_ERROR_MEM      // memory error
	ErrorFILEOPEN ErrorCode = C.BASS_ERROR_FILEOPEN // can't open the file
	ErrorDRIVER   ErrorCode = C.BASS_ERROR_DRIVER   // can't find a free/valid driver
	ErrorBUFLOST  ErrorCode = C.BASS_ERROR_BUFLOST  // the sample buffer was lost
	ErrorHANDLE   ErrorCode = C.BASS_ERROR_HANDLE   // invalid handle
	ErrorFORMAT   ErrorCode = C.BASS_ERROR_FORMAT   // unsupported sample format
	ErrorPOSITION ErrorCode = C.BASS_ERROR_POSITION // invalid position
	ErrorINIT     ErrorCode = C.BASS_ERROR_INIT     // BASS_Init has not been successfully called
	ErrorSTART    ErrorCode = C.BASS_ERROR_START    // BASS_Start has not been successfully called
	ErrorSSL      ErrorCode = C.BASS_ERROR_SSL      // SSL/HTTPS support isn't available
	ErrorALREADY  ErrorCode = C.BASS_ERROR_ALREADY  // already initialized/paused/whatever
	ErrorNOCHAN   ErrorCode = C.BASS_ERROR_NOCHAN   // can't get a free channel
	ErrorILLTYPE  ErrorCode = C.BASS_ERROR_ILLTYPE  // an illegal type was specified
	ErrorILLPARAM ErrorCode = C.BASS_ERROR_ILLPARAM // an illegal parameter was specified
	ErrorNO3D     ErrorCode = C.BASS_ERROR_NO3D     // no 3D support
	ErrorNOEAX    ErrorCode = C.BASS_ERROR_NOEAX    // no EAX support
	ErrorDEVICE   ErrorCode = C.BASS_ERROR_DEVICE   // illegal device number
	ErrorNOPLAY   ErrorCode = C.BASS_ERROR_NOPLAY   // not playing
	ErrorFREQ     ErrorCode = C.BASS_ERROR_FREQ     // illegal sample rate
	ErrorNOTFILE  ErrorCode = C.BASS_ERROR_NOTFILE  // the stream is not a file stream
	ErrorNOHW     ErrorCode = C.BASS_ERROR_NOHW     // no hardware voices available
	ErrorEMPTY    ErrorCode = C.BASS_ERROR_EMPTY    // the MOD music has no sequence data
	ErrorNONET    ErrorCode = C.BASS_ERROR_NONET    // no internet connection could be opened
	ErrorCREATE   ErrorCode = C.BASS_ERROR_CREATE   // couldn't create the file
	ErrorNOFX     ErrorCode = C.BASS_ERROR_NOFX     // effects are not available
	ErrorNOTAVAIL ErrorCode = C.BASS_ERROR_NOTAVAIL // requested data is not available
	ErrorDECODE   ErrorCode = C.BASS_ERROR_DECODE   // the channel is/isn't a "decoding channel"
	ErrorDX       ErrorCode = C.BASS_ERROR_DX       // a sufficient DirectX version is not installed
	ErrorTIMEOUT  ErrorCode = C.BASS_ERROR_TIMEOUT  // connection timed out
	ErrorFILEFORM ErrorCode = C.BASS_ERROR_FILEFORM // unsupported file format
	ErrorSPEAKER  ErrorCode = C.BASS_ERROR_SPEAKER  // unavailable speaker
	ErrorVERSION  ErrorCode = C.BASS_ERROR_VERSION  // invalid BASS version (used by add-ons)
	ErrorCODEC    ErrorCode = C.BASS_ERROR_CODEC    // codec is not available/supported
	ErrorENDED    ErrorCode = C.BASS_ERROR_ENDED    // the channel/file has ended
	ErrorBUSY     ErrorCode = C.BASS_ERROR_BUSY     // the device is busy
	ErrorUNKNOWN  ErrorCode = C.BASS_ERROR_UNKNOWN  // some other mystery problem
)

type InitFlags int

const (
	InitFlag8BITS      InitFlags = C.BASS_DEVICE_8BITS      // 8 bit
	InitFlagMONO       InitFlags = C.BASS_DEVICE_MONO       // mono
	InitFlag3D         InitFlags = C.BASS_DEVICE_3D         // enable 3D functionality
	InitFlag16BITS     InitFlags = C.BASS_DEVICE_16BITS     // limit output to 16 bit
	InitFlagLATENCY    InitFlags = C.BASS_DEVICE_LATENCY    // calculate device latency (BASS_INFO struct)
	InitFlagCPSPEAKERS InitFlags = C.BASS_DEVICE_CPSPEAKERS // detect speakers via Windows control panel
	InitFlagSPEAKERS   InitFlags = C.BASS_DEVICE_SPEAKERS   // force enabling of speaker assignment
	InitFlagNOSPEAKER  InitFlags = C.BASS_DEVICE_NOSPEAKER  // ignore speaker arrangement
	InitFlagDMIX       InitFlags = C.BASS_DEVICE_DMIX       // use ALSA "dmix" plugin
	InitFlagFREQ       InitFlags = C.BASS_DEVICE_FREQ       // set device sample rate
	InitFlagSTEREO     InitFlags = C.BASS_DEVICE_STEREO     // limit output to stereo
)
