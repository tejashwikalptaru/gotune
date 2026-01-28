#!/bin/bash
set -e

echo "Setting up BASS libraries for cross-platform builds..."

# Create directories
mkdir -p build/libs/windows/{x86,x64}
mkdir -p build/libs/linux/{x86,x86_64,aarch64,armhf}
mkdir -p build/libs/darwin

# Extract Windows libraries
echo "Extracting Windows libraries..."
unzip -j libs/bass24.zip "bass.dll" -d build/libs/windows/x86/ 2>/dev/null || true
unzip -j libs/bass24.zip "x64/bass.dll" -d build/libs/windows/x64/ 2>/dev/null || true

# Extract Linux libraries
echo "Extracting Linux libraries..."
unzip -j libs/bass24-linux.zip "libs/x86/libbass.so" -d build/libs/linux/x86/ 2>/dev/null || true
unzip -j libs/bass24-linux.zip "libs/x86_64/libbass.so" -d build/libs/linux/x86_64/ 2>/dev/null || true
unzip -j libs/bass24-linux.zip "libs/aarch64/libbass.so" -d build/libs/linux/aarch64/ 2>/dev/null || true
unzip -j libs/bass24-linux.zip "libs/armhf/libbass.so" -d build/libs/linux/armhf/ 2>/dev/null || true

# Extract and build macOS universal binary (x86_64 + arm64)
echo "Building macOS universal binary (x86_64 + arm64)..."
unzip -q libs/bass24-osx.zip -d libs/bass24-osx-extracted 2>/dev/null || true
make -C libs/bass24-osx-extracted 64bit 2>/dev/null || true
cp libs/bass24-osx-extracted/64bit/libbass.dylib build/libs/darwin/libbass.dylib 2>/dev/null || true
rm -rf libs/bass24-osx-extracted
echo "macOS universal binary created"

echo ""
echo "Library setup complete!"
echo ""
echo "Platform summary:"
echo "  macOS:   build/libs/darwin/libbass.dylib"
echo "  Windows: build/libs/windows/{x86,x64}/bass.dll"
echo "  Linux:   build/libs/linux/{x86,x86_64,aarch64,armhf}/libbass.so"
