#!/bin/bash

# Create dist directory if it doesn't exist
mkdir -p dist

# Build for Linux (amd64)
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o dist/s3-deploy-linux-amd64 main.go

# Build for Linux (arm64)
echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -o dist/s3-deploy-linux-arm64 main.go

# Build for Windows (amd64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o dist/s3-deploy-windows-amd64.exe main.go

# Build for macOS (amd64)
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o dist/s3-deploy-darwin-amd64 main.go

# Build for macOS (arm64)
echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o dist/s3-deploy-darwin-arm64 main.go

echo "Build complete. Binaries are in dist/"
