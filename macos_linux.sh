#!/bin/bash

# Configuration
REPO="github.com/a6h15hek/pancake"
BINARY_NAME="pancake"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR=$(mktemp -d)
CMD_INSTALL_PATH="$INSTALL_DIR/$BINARY_NAME"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# --- Helper Functions ---

log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error_exit() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# --- Checks ---

# Check dependencies
command -v curl >/dev/null 2>&1 || error_exit "curl is required but not installed."

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$OS" == "linux" ]]; then
    BINARY_FILE="${BINARY_NAME}-linux-amd64"
    if [[ "$ARCH" != "x86_64" ]]; then
        echo -e "${RED}Warning:${NC} You are running on $ARCH. The current release only supports amd64 (x86_64)."
        echo "The installation may fail or the binary may not run."
        read -p "Continue anyway? (y/n) " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && exit 1
    fi
elif [[ "$OS" == "darwin" ]]; then
    BINARY_FILE="${BINARY_NAME}-darwin-amd64"
    # Mac M1/M2/M3 (arm64) can run amd64 binaries via Rosetta 2 silently.
else
    error_exit "Unsupported OS: $OS"
fi

# --- Main Logic ---

uninstall() {
    if [ -f "$CMD_INSTALL_PATH" ]; then
        log "Uninstalling Pancake..."
        sudo rm -f "$CMD_INSTALL_PATH" || error_exit "Failed to remove binary."
        success "Pancake uninstalled successfully."
    else
        log "Pancake is not installed."
    fi
    exit 0
}

# Handle Arguments
if [[ "$1" == "uninstall" ]]; then
    uninstall
fi

# Installation / Update
if command -v "$BINARY_NAME" &> /dev/null; then
    log "Existing installation found. Checking for updates..."
else
    log "Starting fresh installation..."
fi

# Download
DOWNLOAD_URL="https://${REPO}/releases/latest/download/${BINARY_FILE}"
log "Downloading from: $DOWNLOAD_URL"
curl -sL --fail "$DOWNLOAD_URL" -o "${TEMP_DIR}/${BINARY_FILE}" || error_exit "Download failed. Check your internet connection or URL."

# Install
log "Installing binary to ${INSTALL_DIR}..."

# Create directory if missing (requires sudo)
if [ ! -d "$INSTALL_DIR" ]; then
    log "Creating $INSTALL_DIR..."
    sudo mkdir -p "$INSTALL_DIR" || error_exit "Failed to create installation directory."
fi

# Move and Chmod
sudo mv "${TEMP_DIR}/${BINARY_FILE}" "$CMD_INSTALL_PATH" || error_exit "Failed to move binary. Do you have sudo privileges?"
sudo chmod +x "$CMD_INSTALL_PATH" || error_exit "Failed to make binary executable."

# Verification
if ! command -v "$BINARY_NAME" &> /dev/null; then
    # If command -v failed, it might be because /usr/local/bin isn't in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${RED}Warning:${NC} '$INSTALL_DIR' is not in your PATH."
        echo "Please add the following to your shell config (.bashrc, .zshrc, etc.):"
        echo "  export PATH=\$PATH:$INSTALL_DIR"
    else
        error_exit "Installation verification failed. The binary exists but cannot be executed."
    fi
else
    NEW_VERSION=$($BINARY_NAME version 2>/dev/null | awk '{print $2}')
    success "Installed Pancake ${NEW_VERSION} successfully!"
    echo "Run 'pancake help' to get started."
fi