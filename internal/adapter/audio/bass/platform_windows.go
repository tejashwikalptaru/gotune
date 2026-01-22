//go:build windows

// Package bass provides Windows-specific CGO configuration for the BASS library.
package bass

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../build/libs/windows
#cgo windows,amd64 LDFLAGS: -L${SRCDIR}/../../../../build/libs/windows/x64 -lbass
#cgo windows,386 LDFLAGS: -L${SRCDIR}/../../../../build/libs/windows/x86 -lbass
#include "bass.h"
*/
import "C"

// Platform-specific constants for Windows
const (
	platformName = "windows"
	libraryName  = "bass.dll"
)
