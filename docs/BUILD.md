# Building GoTune

GoTune is a cross-platform music player written in Go with CGO dependencies for the BASS audio library.

## Prerequisites

### All Platforms
- Go 1.25 or later
- BASS audio library (included in `libs/` directory)
- Make (optional, but recommended)

### Platform-Specific Requirements

#### macOS
- Xcode Command Line Tools (`xcode-select --install`)
- No additional setup required for native builds

#### Linux
- GCC (`sudo apt-get install build-essential` on Ubuntu/Debian)
- Development libraries: `libasound2-dev` (ALSA)
  ```bash
  sudo apt-get install libasound2-dev
  ```

#### Windows
- GCC via MinGW-w64 or TDM-GCC
- Make sure `gcc` is in your PATH
- Make tool (can use Git Bash or install separately)

## Setup

### 1. Extract BASS Libraries

Run the setup script to extract platform-specific BASS libraries:

```bash
./scripts/setup-libs.sh
```

This creates the following directory structure:
```
build/libs/
├── darwin/
│   └── libbass.dylib          (universal binary: x86_64 + arm64)
├── windows/
│   ├── x86/bass.dll
│   └── x64/bass.dll
└── linux/
    ├── x86/libbass.so
    ├── x86_64/libbass.so
    ├── aarch64/libbass.so
    └── armhf/libbass.so
```

## Building

### Native Builds (Recommended)

Build for your current platform:

```bash
# Using Make
make build

# Or directly with go build
go build -o build/gotune ./cmd
```

The binary will be created at `build/gotune` (or `build/gotune.exe` on Windows).

### Running the Application

```bash
# Build and run
make run

# Or run directly
make execute
```

### Cross-Platform Builds

**IMPORTANT:** Cross-compilation with CGO requires platform-specific toolchains and is complex. For production builds, we recommend building natively on each platform.

#### Building for Windows from macOS/Linux

Requires MinGW-w64:

```bash
# Install MinGW-w64 (macOS)
brew install mingw-w64

# Install MinGW-w64 (Ubuntu/Debian)
sudo apt-get install mingw-w64

# Build for Windows amd64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
  go build -o build/gotune-windows-amd64.exe ./cmd
```

#### Building for Linux from macOS

Requires cross-compiler:

```bash
# Install cross-compiler (macOS with Homebrew)
brew install FiloSottile/musl-cross/musl-cross

# Build for Linux amd64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc \
  go build -o build/gotune-linux-amd64 ./cmd
```

#### Building for macOS from Linux

Not officially supported due to macOS SDK licensing requirements and toolchain complexity.

### Cross-Platform Build Recommendations

For production builds, we recommend:
- Build macOS binaries on macOS (natively)
- Build Linux binaries on Linux (natively or in Docker)
- Build Windows binaries on Windows (natively or in Docker)

This approach avoids cross-compilation complexity and ensures maximum compatibility with system libraries.

## Running Tests

Tests require the BASS library to be available at runtime.

### macOS

```bash
# Using Make (recommended)
make test

# Or manually
DYLD_LIBRARY_PATH=$(PWD)/build/libs/darwin go test ./internal/... -v
```

### Linux

```bash
# Using Make (recommended - auto-detects architecture)
make test

# Or manually (replace x86_64 with your architecture)
LD_LIBRARY_PATH=$(PWD)/build/libs/linux/x86_64 go test ./internal/... -v

# Detect your architecture
uname -m  # outputs: x86_64, aarch64, armhf, etc.
```

### Windows (PowerShell)

```powershell
# Add DLL directory to PATH
$env:PATH = "$PWD\build\libs\windows\x64;$env:PATH"
go test ./internal/... -v
```

### Windows (Git Bash)

```bash
PATH=$(pwd)/build/libs/windows/x64:$PATH go test ./internal/... -v
```

### Race Detection

Run tests with race condition detection:

```bash
make test-race
```

## Platform-Specific Notes

### macOS

The BASS library (`libbass.dylib`) is a universal binary supporting both Intel (x86_64) and Apple Silicon (arm64) architectures.

**CGO Configuration** (in `internal/adapter/audio/bass/platform_darwin.go`):
```go
#cgo LDFLAGS: -L${SRCDIR}/../../../../build/libs/darwin -lbass
```

**Library Path at Runtime:**
- Development: Set `DYLD_LIBRARY_PATH` to `build/libs/darwin`
- Production: Package with Fyne bundling (see `make package`)

### Linux

The BASS library is architecture-specific. The build system automatically selects the correct version based on `GOARCH`.

**CGO Configuration** (in `internal/adapter/audio/bass/platform_linux.go`):
```go
#cgo LDFLAGS: -L${SRCDIR}/../../../../build/libs/linux -lbass -Wl,-rpath,$$ORIGIN/../libs
```

**RPATH Note:** The `-rpath,$$ORIGIN/../libs` allows the binary to find `libbass.so` relative to the executable location at runtime.

**Supported Architectures:**
- `x86` - 32-bit Intel/AMD
- `x86_64` (amd64) - 64-bit Intel/AMD
- `aarch64` (arm64) - 64-bit ARM
- `armhf` (arm) - 32-bit ARM with hardware floating point

### Windows

The BASS library is architecture-specific (x86 vs x64).

**CGO Configuration** (in `internal/adapter/audio/bass/platform_windows.go`):
```go
#cgo LDFLAGS: -L${SRCDIR}/../../../../build/libs/windows/${GOARCH} -lbass
```

**DLL Loading:**
- Windows automatically searches the executable directory for DLLs
- For development, add the DLL directory to `PATH`
- For production, place `bass.dll` next to the executable

## Troubleshooting

### "library not found" or "undefined symbol" errors

#### macOS

**Symptoms:**
```
dyld: Library not loaded: libbass.dylib
```

**Solutions:**
1. Ensure `build/libs/darwin/libbass.dylib` exists:
   ```bash
   ls -l build/libs/darwin/libbass.dylib
   ```
2. Run setup script if missing:
   ```bash
   ./scripts/setup-libs.sh
   ```
3. Set runtime library path:
   ```bash
   export DYLD_LIBRARY_PATH=$(PWD)/build/libs/darwin
   ```
4. Check binary's library dependencies:
   ```bash
   otool -L build/gotune
   ```

#### Linux

**Symptoms:**
```
error while loading shared libraries: libbass.so: cannot open shared object file
```

**Solutions:**
1. Ensure `build/libs/linux/<arch>/libbass.so` exists:
   ```bash
   ls -l build/libs/linux/$(uname -m)/libbass.so
   ```
2. Run setup script if missing:
   ```bash
   ./scripts/setup-libs.sh
   ```
3. Set runtime library path:
   ```bash
   export LD_LIBRARY_PATH=$(PWD)/build/libs/linux/$(uname -m)
   ```
4. Check binary's library dependencies:
   ```bash
   ldd build/gotune
   ```

#### Windows

**Symptoms:**
```
The program can't start because bass.dll is missing
```

**Solutions:**
1. Ensure `build/libs/windows/x64/bass.dll` exists:
   ```cmd
   dir build\libs\windows\x64\bass.dll
   ```
2. Run setup script if missing:
   ```bash
   ./scripts/setup-libs.sh
   ```
3. Add DLL directory to PATH:
   ```powershell
   $env:PATH = "$PWD\build\libs\windows\x64;$env:PATH"
   ```
4. Check binary's DLL dependencies:
   ```cmd
   dumpbin /dependents build\gotune.exe
   ```

### CGO Compilation Errors

**Symptoms:**
```
# command-line-arguments
cgo: C compiler "gcc" not found
```

**Solutions:**
1. Verify GCC is installed:
   ```bash
   gcc --version
   ```
2. Install GCC if missing (see Prerequisites above)
3. Verify CGO is enabled:
   ```bash
   go env CGO_ENABLED  # should output: 1
   ```
4. Check BASS header file exists:
   ```bash
   ls internal/adapter/audio/bass/bass.h
   ```

### Cross-Compilation Failures

Cross-compilation with CGO is inherently complex and may fail for various reasons:

**Common Issues:**
- Missing cross-compiler toolchain
- Incorrect compiler configuration
- Missing target platform libraries
- Linker errors due to architecture mismatches

**Recommended Solutions:**
1. **Use native builds** on the target platform
2. **Use CI/CD** with platform-specific runners (GitHub Actions, GitLab CI)
3. **Use Docker** with appropriate toolchains for Linux builds
4. **Avoid cross-compilation** for production releases

## Makefile Targets

### Build Targets

```bash
make build        # Build native binary to build/gotune
make execute      # Run the binary
make run          # Build and run (combined)
make clean        # Remove build artifacts
make package      # Create Fyne package bundle with libraries
```

### Testing Targets

```bash
make test         # Run all tests (with proper library paths)
make test-race    # Run tests with race detection
```

### Code Quality Targets

```bash
make lint         # Run golangci-lint
make lint-fix     # Auto-fix linting issues
make lint-install # Install golangci-lint v2.6.2
make deadcode     # Find unreachable/dead code
make ci-local     # Simulate CI workflow (lint + deadcode + build)
```

## CI/CD Considerations

For automated builds across platforms, consider:

### GitHub Actions (Recommended)

Use matrix builds with platform-specific runners:

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Setup BASS libraries
        run: ./scripts/setup-libs.sh
        shell: bash

      - name: Build
        run: make build

      - name: Test
        run: make test
```

### Docker Builds

For Linux builds in containerized environments:

```dockerfile
FROM golang:1.25-alpine

RUN apk add --no-cache gcc musl-dev make unzip

WORKDIR /app
COPY .. .

RUN ./scripts/setup-libs.sh
RUN make build
```

### Multi-Platform Considerations

- **Native compilation** is most reliable for CGO projects
- Use platform-specific build agents/runners
- Package BASS libraries with the executable
- Test on target platforms before releasing

## BASS Library License

GoTune uses the Un4seen BASS audio library for cross-platform audio playback.

**License:**
- BASS is **free for non-commercial use**
- For commercial use, a license must be purchased from [https://www.un4seen.com/](https://www.un4seen.com/)

**Distribution:**
- The BASS library binaries are included in the `libs/` directory as ZIP archives
- These are extracted during the setup process
- The binaries are not tracked in git history to reduce repository size

**Attribution:**
When distributing GoTune, ensure proper attribution to Un4seen for the BASS library as per their license terms.

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Fyne GUI Framework](https://fyne.io/)
- [BASS Audio Library](https://www.un4seen.com/)
- [CGO Documentation](https://pkg.go.dev/cmd/cgo)

For development workflow and architecture information, see [DEVELOPMENT.md](DEVELOPMENT.md).
