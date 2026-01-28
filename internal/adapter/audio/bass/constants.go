// Package bass provides a BASS audio library adapter implementing the AudioEngine interface.
// This package wraps the Un4seen BASS library (https://www.un4seen.com) for audio playback.
package bass

/*
#include "bass.h"
*/
import "C"

// InitFlags represents BASS library initialization flags.
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

// Stream and music loading flags
const (
	musicRamps       = C.BASS_MUSIC_RAMPS
	musicPreScan     = C.BASS_MUSIC_PRESCAN
	streamAutoFree   = C.BASS_STREAM_AUTOFREE
	streamDecodeOnly = C.BASS_STREAM_DECODE
	posReset         = C.BASS_MUSIC_POSRESET
	posResetEx       = C.BASS_MUSIC_POSRESETEX
)

// Tag represents metadata tag types for BASS_ChannelGetTags.
type Tag int

const (
	TagID3          Tag = C.BASS_TAG_ID3           // ID3v1 tags : TAG_ID3 structure
	TagID3V2        Tag = C.BASS_TAG_ID3V2         // ID3v2 tags : variable length block
	TagOGG          Tag = C.BASS_TAG_OGG           // OGG comments: series of null-terminated UTF-8 strings
	TagHTTP         Tag = C.BASS_TAG_HTTP          // HTTP headers: series of null-terminated ANSI strings
	TagICY          Tag = C.BASS_TAG_ICY           // ICY headers: a series of null-terminated ANSI strings
	TagMETA         Tag = C.BASS_TAG_META          // ICY metadata : ANSI string
	TagAPE          Tag = C.BASS_TAG_APE           // APE tags: series of null-terminated UTF-8 strings
	TagMP4          Tag = C.BASS_TAG_MP4           // MP4/iTunes metadata: series of null-terminated UTF-8 strings
	TagWMA          Tag = C.BASS_TAG_WMA           // WMA tags: series of null-terminated UTF-8 strings
	TagVENDOR       Tag = C.BASS_TAG_VENDOR        // OGG encoder : UTF-8 string
	TagLYRICS3      Tag = C.BASS_TAG_LYRICS3       // Lyric3v2 tag : ASCII string
	TagCaCODEC      Tag = C.BASS_TAG_CA_CODEC      // CoreAudio codec info : TAG_CA_CODEC structure
	TagMF           Tag = C.BASS_TAG_MF            // Media Foundation tags: a series of null-terminated UTF-8 strings
	TagWaveFORMAT   Tag = C.BASS_TAG_WAVEFORMAT    // WAVE format : WAVEFORMATEEX structure
	TagRiffINFO     Tag = C.BASS_TAG_RIFF_INFO     // RIFF "INFO" tags: series of null-terminated ANSI strings
	TagRiffBEXT     Tag = C.BASS_TAG_RIFF_BEXT     // RIFF/BWF "bext" tags : TAG_BEXT structure
	TagRiffCART     Tag = C.BASS_TAG_RIFF_CART     // RIFF/BWF "cart" tags : TAG_CART structure
	TagRiffDISP     Tag = C.BASS_TAG_RIFF_DISP     // RIFF "DISP" text tag : ANSI string
	TagApeBINARY    Tag = C.BASS_TAG_APE_BINARY    // + index #, binary APE tag: TAG_APE_BINARY structure
	TagMusicNAME    Tag = C.BASS_TAG_MUSIC_NAME    // MOD music name: ANSI string
	TagMusicMESSAGE Tag = C.BASS_TAG_MUSIC_MESSAGE // MOD message: ANSI string
	TagMusicORDERS  Tag = C.BASS_TAG_MUSIC_ORDERS  // MOD order list: BYTE array of pattern numbers
	TagMusicAUTH    Tag = C.BASS_TAG_MUSIC_AUTH    // MOD author: UTF-8 string
	TagMusicINST    Tag = C.BASS_TAG_MUSIC_INST    // + instrument #, MOD instrument name : ANSI string
	TagMusicSAMPLE  Tag = C.BASS_TAG_MUSIC_SAMPLE  // + sample #, MOD sample name : ANSI string
)

// ChannelAttributes represents channel attribute types.
type ChannelAttributes int

const (
	ChannelAttribFREQ           ChannelAttributes = C.BASS_ATTRIB_FREQ
	ChannelAttribVOL            ChannelAttributes = C.BASS_ATTRIB_VOL
	ChannelAttribPAN            ChannelAttributes = C.BASS_ATTRIB_PAN
	ChannelAttribEAXMIX         ChannelAttributes = C.BASS_ATTRIB_EAXMIX
	ChannelAttribNOBUFFER       ChannelAttributes = C.BASS_ATTRIB_NOBUFFER
	ChannelAttribVBR            ChannelAttributes = C.BASS_ATTRIB_VBR
	ChannelAttribCPU            ChannelAttributes = C.BASS_ATTRIB_CPU
	ChannelAttribSRC            ChannelAttributes = C.BASS_ATTRIB_SRC
	ChannelAttribNetResume      ChannelAttributes = C.BASS_ATTRIB_NET_RESUME
	ChannelAttribSCANINFO       ChannelAttributes = C.BASS_ATTRIB_SCANINFO
	ChannelAttribNORAMP         ChannelAttributes = C.BASS_ATTRIB_NORAMP
	ChannelAttribBITRATE        ChannelAttributes = C.BASS_ATTRIB_BITRATE
	ChannelAttribBUFFER         ChannelAttributes = C.BASS_ATTRIB_BUFFER
	ChannelAttribGRANULE        ChannelAttributes = C.BASS_ATTRIB_GRANULE
	ChannelAttribMusicAmplify   ChannelAttributes = C.BASS_ATTRIB_MUSIC_AMPLIFY
	ChannelAttribMusicPANSEP    ChannelAttributes = C.BASS_ATTRIB_MUSIC_PANSEP
	ChannelAttribMusicPSCALER   ChannelAttributes = C.BASS_ATTRIB_MUSIC_PSCALER
	ChannelAttribMusicBPM       ChannelAttributes = C.BASS_ATTRIB_MUSIC_BPM
	ChannelAttribMusicSPEED     ChannelAttributes = C.BASS_ATTRIB_MUSIC_SPEED
	ChannelAttribMusicVOLGLOBAL ChannelAttributes = C.BASS_ATTRIB_MUSIC_VOL_GLOBAL
	ChannelAttribMusicACTIVE    ChannelAttributes = C.BASS_ATTRIB_MUSIC_ACTIVE
	ChannelAttribMusicVOLCHAN   ChannelAttributes = C.BASS_ATTRIB_MUSIC_VOL_CHAN // + channel #
	ChannelAttribMusicVOLINST   ChannelAttributes = C.BASS_ATTRIB_MUSIC_VOL_INST // + instrument #
	ChannelAttribSLIDELOG       ChannelAttributes = C.BASS_SLIDE_LOG             // BASS_ChannelSlideAttribute flags
)

// BASS_ChannelGetData flags
const (
	dataFFT256  = C.BASS_DATA_FFT256
	dataFFT512  = C.BASS_DATA_FFT512
	dataFFT1024 = C.BASS_DATA_FFT1024
	dataFFT2048 = C.BASS_DATA_FFT2048
	dataFFT4096 = C.BASS_DATA_FFT4096
	dataFFT8192 = C.BASS_DATA_FFT8192
)

// ErrorCode represents BASS library error codes.
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

// errorCodeToMessage maps BASS error codes to human-readable messages.
func errorCodeToMessage(code ErrorCode) string {
	messages := map[ErrorCode]string{
		ErrorOK:       "all is OK",
		ErrorMEM:      "memory error",
		ErrorFILEOPEN: "can't open the file",
		ErrorDRIVER:   "can't find a free/valid driver",
		ErrorBUFLOST:  "the sample buffer was lost",
		ErrorHANDLE:   "invalid handle",
		ErrorFORMAT:   "unsupported sample format",
		ErrorPOSITION: "invalid position",
		ErrorINIT:     "BASS_Init has not been successfully called",
		ErrorSTART:    "BASS_Start has not been successfully called",
		ErrorSSL:      "SSL/HTTPS support isn't available",
		ErrorALREADY:  "already initialized/paused/whatever",
		ErrorNOCHAN:   "can't get a free channel",
		ErrorILLTYPE:  "an illegal type was specified",
		ErrorILLPARAM: "an illegal parameter was specified",
		ErrorNO3D:     "no 3D support",
		ErrorNOEAX:    "no EAX support",
		ErrorDEVICE:   "illegal device number",
		ErrorNOPLAY:   "not playing",
		ErrorFREQ:     "illegal sample rate",
		ErrorNOTFILE:  "the stream is not a file stream",
		ErrorNOHW:     "no hardware voices available",
		ErrorEMPTY:    "the MOD music has no sequence data",
		ErrorNONET:    "no internet connection could be opened",
		ErrorCREATE:   "couldn't create the file",
		ErrorNOFX:     "effects are not available",
		ErrorNOTAVAIL: "requested data is not available",
		ErrorDECODE:   "the channel is/isn't a 'decoding channel'",
		ErrorDX:       "a sufficient DirectX version is not installed",
		ErrorTIMEOUT:  "connection timed out",
		ErrorFILEFORM: "unsupported file format",
		ErrorSPEAKER:  "unavailable speaker",
		ErrorVERSION:  "invalid BASS version (used by add-ons)",
		ErrorCODEC:    "codec is not available/supported",
		ErrorENDED:    "the channel/file has ended",
		ErrorBUSY:     "the device is busy",
		ErrorUNKNOWN:  "some other mystery problem",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "unknown error"
}
