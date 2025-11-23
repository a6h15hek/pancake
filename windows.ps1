# pancake-install.ps1

param (
    [string]$Action = "install"
)

# Configuration
$REPO = "github.com/a6h15hek/pancake"
$BINARY_NAME = "pancake"
$INSTALL_DIR = "$env:ProgramFiles\Pancake"
$BINARY_FILE_NAME = "$BINARY_NAME-windows-amd64.exe"
$DESTINATION_EXE = "$INSTALL_DIR\$BINARY_NAME.exe"
$TEMP_FILE = "$env:TEMP\$BINARY_FILE_NAME"

# --- Helper Functions ---

function Log {
    param ([string]$Message)
    Write-Host "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - $Message" -ForegroundColor Cyan
}

function Log-Error {
    param ([string]$Message)
    Write-Host "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') - ERROR: $Message" -ForegroundColor Red
}

function Assert-Admin {
    $currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
    if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        Log-Error "This script requires Administrator privileges to install to $INSTALL_DIR and update system PATH."
        Log-Error "Please run PowerShell as Administrator and try again."
        exit 1
    }
}

function Ensure-Path {
    param ([string]$PathToAdd)
    $currentPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine)
    if ($currentPath -split ';' -notcontains $PathToAdd) {
        Log "Adding '$PathToAdd' to System PATH..."
        [Environment]::SetEnvironmentVariable("Path", $currentPath + ";$PathToAdd", [EnvironmentVariableTarget]::Machine)
        Log "PATH updated. You may need to restart your terminal for changes to take effect."
    } else {
        Log "PATH is already configured correctly."
    }
}

function Kill-Process {
    $proc = Get-Process $BINARY_NAME -ErrorAction SilentlyContinue
    if ($proc) {
        Log "Stopping running instance of $BINARY_NAME..."
        Stop-Process -InputObject $proc -Force
        Start-Sleep -Seconds 1
    }
}

# --- Main Logic ---

Assert-Admin

if ($Action -eq "uninstall") {
    if (Test-Path $DESTINATION_EXE) {
        Log "Uninstalling Pancake..."
        Kill-Process
        Remove-Item -Path $DESTINATION_EXE -Force -ErrorAction Stop

        # Optional: Clean up directory if empty
        if ((Get-ChildItem $INSTALL_DIR).Count -eq 0) {
            Remove-Item -Path $INSTALL_DIR -Force
        }

        Log "Pancake uninstalled successfully."
        Log "Note: The directory '$INSTALL_DIR' was removed from PATH, but the environment variable entry might remain until manual cleanup or reboot."
    } else {
        Log "Pancake is not installed."
    }
    exit 0
}

# Install / Update Logic
try {
    # 1. Environment Checks
    Kill-Process

    if (Test-Path $DESTINATION_EXE) {
        Log "Pancake is already installed. Updating..."
    } else {
        Log "Fresh installation of Pancake detected."
    }

    # 2. Download
    Log "Downloading latest binary from $REPO..."
    $DownloadUrl = "https://$REPO/releases/latest/download/$BINARY_FILE_NAME"
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TEMP_FILE -ErrorAction Stop

    if (-not (Test-Path $TEMP_FILE)) {
        throw "Downloaded file not found at $TEMP_FILE"
    }

    # 3. Installation
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    }

    Move-Item -Path $TEMP_FILE -Destination $DESTINATION_EXE -Force -ErrorAction Stop
    Log "Binary installed to $DESTINATION_EXE"

    # 4. Post-Install Configuration
    Ensure-Path $INSTALL_DIR

    # 5. Verification
    if (Get-Command $DESTINATION_EXE -ErrorAction SilentlyContinue) {
        $Version = & $DESTINATION_EXE version
        Log "SUCCESS! Pancake installed: $Version"
        Log "Type 'pancake' in a new terminal window to get started."
    } else {
        throw "Verification failed. The binary was moved but cannot be executed."
    }

} catch {
    Log-Error $_.Exception.Message
    exit 1
}