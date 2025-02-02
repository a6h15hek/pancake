# Variables
$REPO = "github.com/a6h15hek/pancake"
$BINARY_NAME = "pancake"
$INSTALL_DIR = "$env:ProgramFiles\Pancake"

# Logging function
function Log {
    param (
        [string]$Message
    )
    Write-Output "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - $Message"
}

# Uninstall function
function Uninstall {
    if (Get-Command $BINARY_NAME -ErrorAction SilentlyContinue) {
        Log "Uninstalling Pancake..."
        Remove-Item -Path "$INSTALL_DIR\$BINARY_NAME.exe" -Force
        if (-not $?) {
            Log "Failed to uninstall Pancake."
            exit 1
        }
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
    Log "Unsupported OS: $OS"
    exit 1
}

# Check if pancake is already installed
if (Get-Command $BINARY_NAME -ErrorAction SilentlyContinue) {
    Log "Pancake is already installed. Checking for updates..."
    $CURRENT_VERSION = & $BINARY_NAME version | Select-String -Pattern "\d+\.\d+\.\d+" | ForEach-Object { $_.Matches.Value }
    Log "Current version: $CURRENT_VERSION"
} else {
    Log "Pancake is not installed. Proceeding with installation..."
}

# Download and install the latest version
Log "Starting download of the latest version of Pancake..."
$DOWNLOAD_URL = "https://$REPO/releases/latest/download/$BINARY_FILE"
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "$env:TEMP\$BINARY_FILE"

if (-not $?) {
    Log "Failed to download Pancake."
    exit 1
}
Log "Download completed."

Log "Starting installation of Pancake to $INSTALL_DIR..."
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
}
Move-Item -Path "$env:TEMP\$BINARY_FILE" -Destination "$INSTALL_DIR\$BINARY_NAME.exe" -Force
if (-not $?) {
    Log "Failed to install Pancake."
    exit 1
}
Log "Installation completed."

# Verify installation
$NEW_VERSION = & "$INSTALL_DIR\$BINARY_NAME.exe" version | Select-String -Pattern "\d+\.\d+\.\d+" | ForEach-Object { $_.Matches.Value }
Log "Pancake installed/updated successfully. Version: $NEW_VERSION"