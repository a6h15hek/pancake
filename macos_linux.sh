#!/bin/bash

# Variables
REPO="github.com/a6h15hek/pancake"
BINARY_NAME="pancake"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR="/tmp"

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Error handling function
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Uninstall function
uninstall() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        log "Uninstalling Pancake..."
        sudo rm -f "${INSTALL_DIR}/${BINARY_NAME}" || error_exit "Failed to remove ${BINARY_NAME} from ${INSTALL_DIR}."
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
    error_exit "Unsupported OS: $OS"
fi

# Check if pancake is already installed
if command -v "$BINARY_NAME" &> /dev/null; then
    log "Pancake is already installed. Checking for updates..."
    CURRENT_VERSION=$($BINARY_NAME version 2>/dev/null | awk '{print $2}')
    if [[ -z "$CURRENT_VERSION" ]]; then
        error_exit "Failed to get current version of Pancake."
    fi
    log "Current version: $CURRENT_VERSION"
else
    log "Pancake is not installed. Proceeding with installation..."
fi

# Download the latest version
log "Downloading the latest version of Pancake..."
DOWNLOAD_URL="https://${REPO}/releases/latest/download/${BINARY_FILE}"
curl -sL "$DOWNLOAD_URL" -o "${TEMP_DIR}/${BINARY_FILE}" || error_exit "Failed to download Pancake."

if [[ ! -f "${TEMP_DIR}/${BINARY_FILE}" ]]; then
    error_exit "Downloaded binary not found in ${TEMP_DIR}."
fi
log "Download completed."

# Install the binary
log "Installing Pancake to ${INSTALL_DIR}..."
sudo mv "${TEMP_DIR}/${BINARY_FILE}" "${INSTALL_DIR}/${BINARY_NAME}" || error_exit "Failed to move binary to ${INSTALL_DIR}."
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}" || error_exit "Failed to set executable permissions."

# Verify installation
NEW_VERSION=$($BINARY_NAME version 2>/dev/null | awk '{print $2}')
if [[ -z "$NEW_VERSION" ]]; then
    error_exit "Failed to verify installation. Pancake may not be installed correctly."
fi
log "Pancake installed/updated successfully. Version: $NEW_VERSION"