# BASS Audio Engine Adapter

This package provides an adapter for the Un4seen BASS audio library, implementing the `AudioEngine` interface defined in `internal/ports/audio.go`.

## Architecture

The BASS adapter follows a clean architecture pattern with a clear separation of concerns:

```
bass/
├── constants.go         # BASS library constants and error code mappings
├── platform_darwin.go   # macOS-specific CGO configuration
├── platform_windows.go  # Windows-specific CGO configuration
├── platform_linux.go    # Linux-specific CGO configuration
├── bindings.go          # Low-level CGO wrapper functions (internal use only)
├── engine.go            # BassEngine implementing AudioEngine interface
├── metadata.go          # Metadata extraction for audio files
├── bass.h               # BASS library header file
├── engine_test.go       # Comprehensive unit tests
└── README.md            # This file
```

## Key Features

- **Thread-Safe**: All operations protected with `sync.RWMutex`
- **Cross-Platform**: Build tags enable platform-specific compilation
- **MOD Support**: Handles both MOD/tracker files and regular audio formats
- **Clean Interface**: Implements `AudioEngine` interface for dependency injection
- **Comprehensive Testing**: 20+ test cases covering all functionality

## Supported Formats

### Regular Audio Formats
- MP3, MP2, MP1
- OGG, OGA
- WAV, AIF, AIFF
- FLAC, FLA
- AAC, M4A, M4B, MP4
- WMA
- WavPack (WV)
- APE, MAC
- Musepack (MPC, MP+, MPP)
- OptimFROG (OFR, OFS)
- TTA
- ADX, AIX
- AC3
- CDA (CD Audio)

### MOD/Tracker Formats
- MOD
- XM (Extended Module)
- IT (Impulse Tracker)
- S3M (Scream Tracker 3)
- MTM (MultiTracker)
- UMX
- MO3

## Library Requirements

### macOS
- Library: `libbass.dylib`
- Location: `build/libs/darwin/` or `libs/`
- Architecture: Must match your macOS architecture (x86_64 or arm64)

### Windows
- Library: `bass.dll`
- Location: `build/libs/windows/`
- Architecture: Match your Windows architecture (x86 or x64)

### Linux
- Library: `libbass.so`
- Location: `build/libs/linux/`
- Architecture: Match your Linux architecture (x86 or x64)

## Getting the BASS Library

1. Download from https://www.un4seen.com/
2. Extract the library for your platform
3. Place in the appropriate `build/libs/{platform}/` directory

## Thread Safety

The `BassEngine` implementation is fully thread-safe:

- All state access protected with `sync.RWMutex`
- Read operations use `RLock()`/`RUnlock()`
- Write operations use `Lock()`/`Unlock()`
- Verified with `go test -race`

## Testing

### Unit Tests

Run all tests:
```bash
go test -v -race ./internal/adapter/audio/bass/
```

### Apple Silicon Note

If running on Apple Silicon (arm64) with x86_64 library:
```bash
GOARCH=amd64 go test -v ./internal/adapter/audio/bass/
```

Or download the arm64 BASS library from un4seen.com.

### Test Coverage

The test suite includes:
- Initialization and shutdown
- Track loading (valid/invalid paths)
- Playback controls (play, pause, stop)
- Volume control (0.0 to 1.0 range)
- Position and duration
- Seeking
- Metadata extraction
- Multiple simultaneous tracks
- Concurrent operations (thread safety)

## Usage Example

```go
import (
    "github.com/tejashwikalptaru/gotune/internal/adapter/audio/bass"
    "github.com/tejashwikalptaru/gotune/internal/ports"
)

func main() {
    // Create engine
    var engine ports.AudioEngine = bass.NewBassEngine()

    // Initialize
    err := engine.Initialize(-1, 44100, 0)
    if err != nil {
        panic(err)
    }
    defer engine.Shutdown()

    // Load track
    handle, err := engine.Load("/path/to/song.mp3")
    if err != nil {
        panic(err)
    }
    defer engine.Unload(handle)

    // Set volume
    engine.SetVolume(handle, 0.8)

    // Play
    err = engine.Play(handle)
    if err != nil {
        panic(err)
    }

    // Check status
    status, _ := engine.Status(handle)
    fmt.Println(status) // StatusPlaying
}
```

## Implementation Details

### Engine Lifecycle

1. **Initialize**: Sets up the BASS library with a device, frequency, and flags
2. **Load**: Creates BASS stream or music handle, stores in a tracks map
3. **Play/Pause/Stop**: Controls playback state
4. **Unload**: Frees BASS resources for a track
5. **Shutdown**: Stops and frees all tracks, shuts down BASS

### MOD File Handling

MOD files are detected by extension and loaded using `BASS_MusicLoad` instead of `BASS_StreamCreateFile`. If loading fails, the engine tries the opposite method as a fallback.

### Smooth Stop Effect

The `Stop()` method implements a smooth fade-out:
1. Slide frequency down over 500ms
2. Slide volume down (logarithmically) over 100ms
3. Stop and free the channel

### Error Handling

All BASS errors are wrapped in `domain.AudioEngineError` with:
- Operation name
- File path (if applicable)
- BASS error code
- Human-readable message

## Future Enhancements

- [ ] Support for streaming from URLs
- [ ] Equalizer/effects support
- [ ] Crossfade between tracks
- [ ] Gapless playback
- [ ] Audio visualization data
- [ ] Recording support
- [ ] CD ripping support

## License

BASS is a commercial library. See https://www.un4seen.com/ for licensing information.
This adapter code is part of the GoTune project.
