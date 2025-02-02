#!/bin/bash

# Variables
REPO="github.com/a6h15hek/pancake"
BINARY_NAME="pancake"
INSTALL_DIR="/usr/local/bin"

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Uninstall function
uninstall() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        log "Uninstalling Pancake..."
        sudo rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        if [[ $? -ne 0 ]]; then
            log "Failed to uninstall Pancake."
            exit 1
        fi
        log "Pancake uninstalled successfully."
    else
        log "Pancake is not installed."
    fi
    exit 0
}

# Check if uninstall argument is passed
if [[ "$1" == "uninstall" ]]; then
    uninstall
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH="amd64"

# Determine the binary name based on OS
if [[ "$OS" == "linux" ]]; then
    BINARY_FILE="${BINARY_NAME}-linux-${ARCH}"
elif [[ "$OS" == "darwin" ]]; then
    BINARY_FILE="${BINARY_NAME}-darwin-${ARCH}"
else
    log "Unsupported OS: $OS"
    exit 1
fi

# Check if pancake is already installed
if command -v "$BINARY_NAME" &> /dev/null; then
    log "Pancake is already installed. Checking for updates..."
    CURRENT_VERSION=$($BINARY_NAME version | awk '{print $2}')
    log "Current version: $CURRENT_VERSION"
else
    log "Pancake is not installed. Proceeding with installation..."
fi

# Download and install the latest version
log "Starting download of the latest version of Pancake..."
DOWNLOAD_URL="https://${REPO}/releases/latest/download/${BINARY_FILE}"
curl -sL "$DOWNLOAD_URL" -o "/tmp/${BINARY_FILE}"

if [[ $? -ne 0 ]]; then
    log "Failed to download Pancake."
    exit 1
fi
log "Download completed."

log "Starting installation of Pancake to $INSTALL_DIR..."
sudo mv "/tmp/${BINARY_FILE}" "${INSTALL_DIR}/${BINARY_NAME}"
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

if [[ $? -ne 0 ]]; then
    log "Failed to install Pancake."
    exit 1
fi
log "Installation completed."

# Verify installation
NEW_VERSION=$($BINARY_NAME version | awk '{print $2}')
log "Pancake installed/updated successfully. Version: $NEW_VERSION"