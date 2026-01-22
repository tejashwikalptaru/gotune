//go:build linux

// Package bass provides Linux-specific CGO configuration for the BASS library.
package bass

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../build/libs/linux
#cgo LDFLAGS: -L${SRCDIR}/../../../../build/libs/linux -lbass -Wl,-rpath,$$ORIGIN/../libs
#include "bass.h"
*/
import "C"

// Platform-specific constants for Linux
const (
	platformName = "linux"
	libraryName  = "libbass.so"
)
