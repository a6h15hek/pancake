#!/bin/bash

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

# Error handling function
handle_error() {
    log "ERROR: $1"
    log "Build failed. Exiting."
    exit 1
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    handle_error "Go command not found. Please install Go and try again."
fi

# Create build directory
BUILD_DIR="./build"
log "Starting Pancake build..."
log "Creating build directory: $BUILD_DIR"
mkdir -p "$BUILD_DIR" || handle_error "Failed to create build directory."

# Build for Linux
log "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o "$BUILD_DIR/pancake-linux-amd64" || handle_error "Failed to build for Linux."

# Build for macOS
log "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o "$BUILD_DIR/pancake-darwin-amd64" || handle_error "Failed to build for macOS."

# Build for Windows
log "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$BUILD_DIR/pancake-windows-amd64.exe" || handle_error "Failed to build for Windows."

# Copy readme.txt to build directory
README_FILE="./readme.txt"
log "Copying $README_FILE to $BUILD_DIR..."
if [ -f "$README_FILE" ]; then
    cp "$README_FILE" "$BUILD_DIR/" || handle_error "Failed to copy $README_FILE."
else
    handle_error "$README_FILE not found. Please ensure it exists in the adjacent folder."
fi

# Log success
log "Build completed successfully. Binaries and readme.txt are in $BUILD_DIR."