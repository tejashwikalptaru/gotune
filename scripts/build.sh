#!/bin/bash
set -e

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
