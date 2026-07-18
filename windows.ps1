#requires -Version 5.1
<#
.SYNOPSIS
    Install or uninstall Pancake on Windows.
.DESCRIPTION
    Installs pancake to %LOCALAPPDATA%\Programs\Pancake (no admin required) by default,
    or to $env:ProgramFiles if run as admin. Verifies SHA-256 checksum, retries on
    network failures, and updates the user PATH. Uninstall removes the binary and
    optionally purges config and projects.
.PARAMETER Action
    'install' (default) or 'uninstall'.
.PARAMETER Version
    Release tag to install (default: latest).
.PARAMETER Purge
    During uninstall, also remove ~/pancake.yml and ~/pancake/.
.PARAMETER NoChecksum
    Skip SHA-256 checksum verification.
.PARAMETER Force
    Use the per-user install location even when run as admin.
.EXAMPLE
    & .\windows.ps1 -Action install
    & .\windows.ps1 -Action uninstall -Purge
#>
[CmdletBinding()]
param(
    [ValidateSet('install', 'uninstall')]
    [string]$Action = 'install',

    [string]$Version = 'latest',

    [switch]$Purge,

    [switch]$NoChecksum,

    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo = 'github.com/a6h15hek/pancake'
$BinaryName = 'pancake'

function Write-Log     { param([string]$Message) Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message" -ForegroundColor Cyan }
function Write-Success { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-Warn    { param([string]$Message) Write-Host "[WARN] $Message" -ForegroundColor Yellow }
function Write-Err     { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

function Resolve-InstallDir {
    $current = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    $isAdmin = $current.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if ($isAdmin -and -not $Force) {
        return Join-Path $env:ProgramFiles 'Pancake'
    }
    return Join-Path $env:LOCALAPPDATA 'Programs\Pancake'
}

function Resolve-PathScope {
    $current = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    if ($current.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        return [EnvironmentVariableTarget]::Machine
    }
    return [EnvironmentVariableTarget]::User
}

function Ensure-Path {
    param([string]$PathToAdd, [EnvironmentVariableTarget]$Scope)
    $currentPath = [Environment]::GetEnvironmentVariable('Path', $Scope)
    if (-not $currentPath) { $currentPath = '' }
    $entries = $currentPath -split ';' | Where-Object { $_ -ne '' }
    if ($entries -contains $PathToAdd) {
        Write-Log 'PATH already configured.'
        return
    }
    $newPath = if ($currentPath) { "$currentPath;$PathToAdd" } else { $PathToAdd }
    [Environment]::SetEnvironmentVariable('Path', $newPath, $Scope)
    Write-Log "Added '$PathToAdd' to PATH ($Scope). Restart your terminal for it to take effect."
}

function Remove-From-Path {
    param([string]$PathToRemove, [EnvironmentVariableTarget]$Scope)
    $currentPath = [Environment]::GetEnvironmentVariable('Path', $Scope)
    if (-not $currentPath) { return }
    $entries = $currentPath -split ';' | Where-Object { $_ -ne '' -and $_ -ne $PathToRemove }
    $newPath = $entries -join ';'
    [Environment]::SetEnvironmentVariable('Path', $newPath, $Scope)
    Write-Log "Removed '$PathToRemove' from PATH ($Scope)."
}

function Stop-RunningPancake {
    $proc = Get-Process -Name $BinaryName -ErrorAction SilentlyContinue
    if ($proc) {
        Write-Log "Stopping running instance(s) of $BinaryName..."
        $proc | Stop-Process -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    }
}

function Download-With-Retry {
    param([string]$Url, [string]$Destination, [int]$Attempts = 3)
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    for ($i = 1; $i -le $Attempts; $i++) {
        try {
            Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing -ErrorAction Stop
            return
        } catch {
            Write-Warn "Download attempt $i failed: $Url -- $($_.Exception.Message)"
            Start-Sleep -Seconds 2
        }
    }
    Write-Err "Download failed after $Attempts attempts: $Url
Troubleshooting:
  - Check your internet connection.
  - Check the release exists: https://github.com/a6h15hek/pancake/releases
  - On corporate networks, set \$env:HTTPS_PROXY and retry."
    exit 1
}

function Confirm-Checksum {
    param([string]$BinaryPath, [string]$ChecksumsUrl)
    if ($NoChecksum) {
        Write-Warn 'Skipping checksum verification (-NoChecksum).'
        return
    }
    $tempChecksums = Join-Path $env:TEMP 'pancake-checksums.txt'
    try {
        Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $tempChecksums -UseBasicParsing -ErrorAction Stop
    } catch {
        Write-Warn "Checksums file unavailable at $ChecksumsUrl; skipping verification."
        return
    }
    $binaryFile = Split-Path $BinaryPath -Leaf
    $expectedLine = Get-Content $tempChecksums -ErrorAction SilentlyContinue | Where-Object { $_ -match "\s\./?$binaryFile\s*$" -or $_ -match "^\s*$binaryFile\s*$" } | Select-Object -First 1
    if (-not $expectedLine) {
        Write-Warn "No checksum entry for $binaryFile; skipping verification."
        return
    }
    $expected = ($expectedLine -split '\s+')[0]
    $actual = (Get-FileHash -Algorithm SHA256 -Path $BinaryPath).Hash.ToLower()
    if ($actual -ne $expected.ToLower()) {
        Write-Err "Checksum mismatch for $binaryFile
  expected: $expected
  actual:   $actual
Re-download, or report at https://github.com/a6h15hek/pancake/issues."
        exit 1
    }
    Write-Log 'Checksum verified.'
}

function Install-Pancake {
    $installDir = Resolve-InstallDir
    $destinationExe = Join-Path $installDir "$BinaryName.exe"
    $binaryFile = "$BinaryName-windows-amd64.exe"
    $tempFile = Join-Path $env:TEMP $binaryFile

    if ($Version -eq 'latest') {
        $downloadUrl = "https://$Repo/releases/latest/download/$binaryFile"
        $checksumsUrl = "https://$Repo/releases/latest/download/checksums.txt"
    } else {
        $tag = $Version -replace '^v', ''
        $downloadUrl = "https://$Repo/releases/download/v$tag/$binaryFile"
        $checksumsUrl = "https://$Repo/releases/download/v$tag/checksums.txt"
    }

    Write-Log "Installing pancake $Version for windows-amd64 into $installDir"
    Stop-RunningPancake
    if (Test-Path $destinationExe) {
        Write-Log 'Existing installation found; updating.'
    }

    Download-With-Retry -Url $downloadUrl -Destination $tempFile
    if (-not (Test-Path $tempFile)) {
        Write-Err "Downloaded file not found at $tempFile"
        exit 1
    }

    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    try {
        Move-Item -Path $tempFile -Destination $destinationExe -Force -ErrorAction Stop
    } catch {
        Write-Err "Could not move binary to $destinationExe : $($_.Exception.Message)"
        exit 1
    }

    Confirm-Checksum -BinaryPath $destinationExe -ChecksumsUrl $checksumsUrl

    $scope = Resolve-PathScope
    Ensure-Path -PathToAdd $installDir -Scope $scope

    if (& $destinationExe version) {
        Write-Success "Installed pancake $(& $destinationExe version 2>$null)."
        Write-Log "Open a NEW terminal and run 'pancake help' to get started."
    } else {
        Write-Err "Verification failed: $destinationExe is not executable."
        exit 1
    }
}

function Uninstall-Pancake {
    $installDir = Resolve-InstallDir
    $destinationExe = Join-Path $installDir "$BinaryName.exe"
    Stop-RunningPancake

    if (Test-Path $destinationExe) {
        try {
            Remove-Item -Path $destinationExe -Force -ErrorAction Stop
            Write-Success 'Pancake binary removed.'
        } catch {
            Write-Err "Could not remove $destinationExe : $($_.Exception.Message)
If pancake is still running, kill it via Task Manager and retry."
            exit 1
        }
    } else {
        Write-Log 'Pancake binary not found. Already uninstalled.'
    }

    if ((Test-Path $installDir) -and -not (Get-ChildItem $installDir -ErrorAction SilentlyContinue)) {
        Remove-Item -Path $installDir -Force -ErrorAction SilentlyContinue
    }

    $scope = Resolve-PathScope
    Remove-From-Path -PathToRemove $installDir -Scope $scope

    if ($Purge) {
        $configFile = Join-Path $env:USERPROFILE 'pancake.yml'
        $pancakeHome = Join-Path $env:USERPROFILE 'pancake'
        if (Test-Path $configFile) {
            Remove-Item $configFile -Force
            Write-Log "Removed $configFile"
        }
        if (Test-Path $pancakeHome) {
            Remove-Item $pancakeHome -Recurse -Force
            Write-Log "Removed $pancakeHome"
        }
    } else {
        Write-Log 'Config (~/pancake.yml) and projects (~/pancake/) were left in place. Re-run with -Purge to remove them.'
    }
}

switch ($Action) {
    'install'   { Install-Pancake }
    'uninstall' { Uninstall-Pancake }
    default     { Write-Err "Unknown action: $Action"; exit 1 }
}
