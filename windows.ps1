\# Variables
$REPO = "github.com/a6h15hek/pancake"
$BINARY_NAME = "pancake"
$INSTALL_DIR = "$env:ProgramFiles\Pancake"
$TEMP_DIR = $env:TEMP

# Logging function
function Log {
    param (
        [string]$Message
    )
    Write-Output "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - $Message"
}

# Error handling function
function ErrorExit {
    param (
        [string]$Message
    )
    Log "ERROR: $Message"
    exit 1
}

# Uninstall function
function Uninstall {
    if (Get-Command $BINARY_NAME -ErrorAction SilentlyContinue) {
        Log "Uninstalling Pancake..."
        Remove-Item -Path "$INSTALL_DIR\$BINARY_NAME.exe" -Force -ErrorAction Stop
        Log "Pancake uninstalled successfully."
    } else {
        Log "Pancake is not installed."
    }
    exit 0
}

# Check if uninstall argument is passed
if ($args[0] -eq "uninstall") {
    Uninstall
}

# Detect OS and architecture
$OS = $PSVersionTable.OS -replace "Windows NT ", ""
$ARCH = "amd64"

# Determine the binary name based on OS
if ($OS -match "Windows") {
    $BINARY_FILE = "$BINARY_NAME-windows-$ARCH.exe"
} else {
    ErrorExit "Unsupported OS: $OS"
}

# Check if pancake is already installed
if (Get-Command $BINARY_NAME -ErrorAction SilentlyContinue) {
    Log "Pancake is already installed. Checking for updates..."
    $CURRENT_VERSION = & $BINARY_NAME version | Select-String -Pattern "\d+\.\d+\.\d+" | ForEach-Object { $_.Matches.Value }
    if (-not $CURRENT_VERSION) {
        ErrorExit "Failed to get current version of Pancake."
    }
    Log "Current version: $CURRENT_VERSION"
} else {
    Log "Pancake is not installed. Proceeding with installation..."
}

# Download the latest version
Log "Downloading the latest version of Pancake..."
$DOWNLOAD_URL = "https://$REPO/releases/latest/download/$BINARY_FILE"
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "$TEMP_DIR\$BINARY_FILE" -ErrorAction Stop

if (-not (Test-Path "$TEMP_DIR\$BINARY_FILE")) {
    ErrorExit "Downloaded binary not found in $TEMP_DIR."
}
Log "Download completed."

# Install the binary
Log "Installing Pancake to $INSTALL_DIR..."
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force -ErrorAction Stop | Out-Null
}
Move-Item -Path "$TEMP_DIR\$BINARY_FILE" -Destination "$INSTALL_DIR\$BINARY_NAME.exe" -Force -ErrorAction Stop

# Verify installation
$NEW_VERSION = & "$INSTALL_DIR\$BINARY_NAME.exe" version | Select-String -Pattern "\d+\.\d+\.\d+" | ForEach-Object { $_.Matches.Value }
if (-not $NEW_VERSION) {
    ErrorExit "Failed to verify installation. Pancake may not be installed correctly."
}
Log "Pancake installed/updated successfully. Version: $NEW_VERSION"