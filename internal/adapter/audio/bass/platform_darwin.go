//go:build darwin

// Package bass provides macOS-specific CGO configuration for the BASS library.
package bass

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR}/../../../../build/libs/darwin -lbass
#include "bass.h"
*/
import "C"

// Platform-specific constants for macOS
const (
	platformName = "darwin"
	libraryName  = "libbass.dylib"
)
