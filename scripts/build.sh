#!/bin/bash
set -e

echo "========================================"
echo "GoTune Cross-Platform Build Script"
echo "========================================"
echo ""
echo "WARNING: Cross-compilation with CGO requires platform-specific toolchains."
echo "For production builds, we recommend building natively on each platform."
echo "See BUILD.md for detailed instructions."
echo ""
echo "Continuing with cross-compilation attempt..."
echo ""

echo "Building GoTune for multiple platforms..."

# macOS (current platform)
echo "Building for macOS x64..."
GOOS=darwin GOARCH=amd64 go build -o build/gotune-darwin-amd64 ./cmd

# Windows
echo "Building for Windows x64..."
GOOS=windows GOARCH=amd64 go build -o build/gotune-windows-amd64.exe ./cmd

# Linux
echo "Building for Linux x64..."
GOOS=linux GOARCH=amd64 go build -o build/gotune-linux-amd64 ./cmd

echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o build/gotune-linux-arm64 ./cmd

echo ""
echo "Build complete!"
echo ""
ls -lh build/gotune-*
