//go:build linux

// Package bass provides Linux-specific CGO configuration for the BASS library.
package bass

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../build/libs/linux
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../../../../build/libs/linux/x86_64 -lbass -Wl,-rpath,$$ORIGIN/../libs
#cgo linux,386 LDFLAGS: -L${SRCDIR}/../../../../build/libs/linux/x86 -lbass -Wl,-rpath,$$ORIGIN/../libs
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../../../build/libs/linux/aarch64 -lbass -Wl,-rpath,$$ORIGIN/../libs
#cgo linux,arm LDFLAGS: -L${SRCDIR}/../../../../build/libs/linux/armhf -lbass -Wl,-rpath,$$ORIGIN/../libs
#include "bass.h"
*/
import "C"

// Platform-specific constants for Linux
const (
	platformName = "linux"
	libraryName  = "libbass.so"
)
