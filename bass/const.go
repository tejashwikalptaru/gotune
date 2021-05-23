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
