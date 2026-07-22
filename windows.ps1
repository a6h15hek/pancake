#requires -Version 5.1
<#
.SYNOPSIS
    Install or uninstall Pancake on Windows.
.DESCRIPTION
    Installs pancake to %LOCALAPPDATA%\Programs\Pancake (no admin required) by default,
    or to $env:ProgramFiles\Pancake if run as admin. Detects amd64/arm64, verifies the
    SHA-256 checksum BEFORE installing, retries on network failures, replaces a running
    binary safely, updates PATH, and cleans up leftovers from previous installs in the
    other location (e.g. installed as admin in the past, reinstalled as user, or vice
    versa). Uninstall removes the binary from every known location it can and optionally
    purges config and projects.
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
.PARAMETER Yes
    Assume yes to prompts (non-interactive / CI).
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

    [switch]$Force,

    [switch]$Yes
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo = 'github.com/a6h15hek/pancake'
$BinaryName = 'pancake'

function Write-Log     { param([string]$Message) Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message" -ForegroundColor Cyan }
function Write-Success { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-Warn    { param([string]$Message) Write-Host "[WARN] $Message" -ForegroundColor Yellow }
function Write-Err     { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

function Test-IsAdmin {
    $current = [Security.Principal.WindowsPrincipal]::new([Security.Principal.WindowsIdentity]::GetCurrent())
    return $current.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Confirm-Action {
    param([string]$Prompt)
    if ($Yes) { return $true }
    if (-not [Environment]::UserInteractive -or [Console]::IsInputRedirected) {
        Write-Warn "No interactive console to prompt; assuming 'no'. Pass -Yes to skip prompts."
        return $false
    }
    $reply = Read-Host "$Prompt [y/N]"
    return $reply -match '^(y|yes)$'
}

function Resolve-Arch {
    # PROCESSOR_ARCHITEW6432 is set when a 32-bit host runs on a 64-bit OS.
    $arch = $env:PROCESSOR_ARCHITEW6432
    if (-not $arch) { $arch = $env:PROCESSOR_ARCHITECTURE }
    switch ($arch) {
        'ARM64' { return 'arm64' }
        'AMD64' { return 'amd64' }
        default {
            Write-Err "Unsupported architecture: $arch (supported: AMD64, ARM64)."
            exit 1
        }
    }
}

# Both locations this script ever installs to. Reinstalls must consider both:
# a machine that installed as admin in the past and reinstalls as user (or the
# reverse) would otherwise end up with two copies shadowing each other on PATH.
function Get-KnownInstallDirs {
    return @(
        (Join-Path $env:ProgramFiles 'Pancake'),
        (Join-Path $env:LOCALAPPDATA 'Programs\Pancake')
    )
}

function Resolve-InstallDir {
    if ((Test-IsAdmin) -and -not $Force) {
        return Join-Path $env:ProgramFiles 'Pancake'
    }
    return Join-Path $env:LOCALAPPDATA 'Programs\Pancake'
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
    # Make the current session work too, so `pancake` runs without a restart.
    if (($env:Path -split ';') -notcontains $PathToAdd) { $env:Path = "$env:Path;$PathToAdd" }
    Write-Log "Added '$PathToAdd' to PATH ($Scope). Restart other terminals for it to take effect."
}

function Remove-From-Path {
    param([string]$PathToRemove, [EnvironmentVariableTarget]$Scope)
    try {
        $currentPath = [Environment]::GetEnvironmentVariable('Path', $Scope)
        if (-not $currentPath) { return }
        $entries = $currentPath -split ';' | Where-Object { $_ -ne '' }
        if ($entries -notcontains $PathToRemove) { return }
        $newPath = ($entries | Where-Object { $_ -ne $PathToRemove }) -join ';'
        [Environment]::SetEnvironmentVariable('Path', $newPath, $Scope)
        Write-Log "Removed '$PathToRemove' from PATH ($Scope)."
    } catch {
        Write-Warn "Could not update PATH ($Scope): $($_.Exception.Message)"
    }
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
    [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
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
  - Check the release/tag exists: https://github.com/a6h15hek/pancake/releases
  - On corporate networks, set `$env:HTTPS_PROXY and retry."
    exit 1
}

function Confirm-Checksum {
    param([string]$BinaryPath, [string]$ChecksumsUrl)
    if ($NoChecksum) {
        Write-Warn 'Skipping checksum verification (-NoChecksum).'
        return
    }
    $tempChecksums = Join-Path $env:TEMP ("pancake-checksums-" + [guid]::NewGuid().ToString('N').Substring(0, 8) + ".txt")
    try {
        Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $tempChecksums -UseBasicParsing -ErrorAction Stop
    } catch {
        Write-Warn "Checksums file unavailable at $ChecksumsUrl; skipping verification."
        return
    }
    try {
        $binaryFile = Split-Path $BinaryPath -Leaf
        $pattern = [regex]::Escape($binaryFile)
        $expectedLine = Get-Content $tempChecksums -ErrorAction SilentlyContinue |
            Where-Object { $_ -match "\s\*?(\./)?$pattern\s*$" } | Select-Object -First 1
        if (-not $expectedLine) {
            Write-Warn "No checksum entry for $binaryFile; skipping verification."
            return
        }
        $expected = ($expectedLine.Trim() -split '\s+')[0]
        $actual = (Get-FileHash -Algorithm SHA256 -Path $BinaryPath).Hash.ToLower()
        if ($actual -ne $expected.ToLower()) {
            Write-Err "Checksum mismatch for $binaryFile
  expected: $expected
  actual:   $actual
Re-download, or report at https://github.com/a6h15hek/pancake/issues."
            exit 1
        }
        Write-Log 'Checksum verified.'
    } finally {
        Remove-Item $tempChecksums -Force -ErrorAction SilentlyContinue
    }
}

# Replace $Destination with $Source even if the destination executable is
# locked by a running process: rename the old file out of the way first,
# then delete it best-effort.
function Install-BinaryFile {
    param([string]$Source, [string]$Destination)
    $backup = "$Destination.old"
    if (Test-Path $Destination) {
        Remove-Item $backup -Force -ErrorAction SilentlyContinue
        try {
            Move-Item -Path $Destination -Destination $backup -Force -ErrorAction Stop
        } catch {
            Write-Err "Could not replace $Destination : $($_.Exception.Message)
If pancake is still running, close it (or kill it via Task Manager) and retry."
            exit 1
        }
    }
    try {
        Move-Item -Path $Source -Destination $Destination -Force -ErrorAction Stop
    } catch {
        # Roll the old binary back so the machine is not left without pancake.
        if (Test-Path $backup) { Move-Item -Path $backup -Destination $Destination -Force -ErrorAction SilentlyContinue }
        Write-Err "Could not move binary to $Destination : $($_.Exception.Message)"
        exit 1
    }
    Remove-Item $backup -Force -ErrorAction SilentlyContinue
}

# Remove leftovers from a previous install in the other location, so the old
# copy does not shadow the fresh one on PATH.
function Remove-StaleInstalls {
    param([string]$CurrentInstallDir)
    $isAdmin = Test-IsAdmin
    foreach ($dir in Get-KnownInstallDirs) {
        if ($dir -eq $CurrentInstallDir) { continue }
        $exe = Join-Path $dir "$BinaryName.exe"
        if (-not (Test-Path $exe)) { continue }
        Write-Warn "Found another pancake install at $exe (leftover from a previous install)."
        try {
            Remove-Item -Path $exe -Force -ErrorAction Stop
            Remove-Item "$exe.old" -Force -ErrorAction SilentlyContinue
            if ((Test-Path $dir) -and -not (Get-ChildItem $dir -ErrorAction SilentlyContinue)) {
                Remove-Item -Path $dir -Force -ErrorAction SilentlyContinue
            }
            Write-Log "Removed stale copy: $exe"
        } catch {
            if (-not $isAdmin -and $dir.StartsWith($env:ProgramFiles)) {
                Write-Warn "Cannot remove $exe without admin rights. Run this script once from an elevated PowerShell to clean it up."
            } else {
                Write-Warn "Could not remove $exe : $($_.Exception.Message)"
            }
        }
        Remove-From-Path -PathToRemove $dir -Scope ([EnvironmentVariableTarget]::User)
        if ($isAdmin) { Remove-From-Path -PathToRemove $dir -Scope ([EnvironmentVariableTarget]::Machine) }
    }
}

function Install-Pancake {
    $arch = Resolve-Arch
    $installDir = Resolve-InstallDir
    $destinationExe = Join-Path $installDir "$BinaryName.exe"
    $binaryFile = "$BinaryName-windows-$arch.exe"
    $tempFile = Join-Path $env:TEMP ("pancake-" + [guid]::NewGuid().ToString('N').Substring(0, 8) + ".exe")

    if ($Version -eq 'latest') {
        $downloadUrl = "https://$Repo/releases/latest/download/$binaryFile"
        $checksumsUrl = "https://$Repo/releases/latest/download/checksums.txt"
    } else {
        $tag = $Version -replace '^v', ''
        $downloadUrl = "https://$Repo/releases/download/v$tag/$binaryFile"
        $checksumsUrl = "https://$Repo/releases/download/v$tag/checksums.txt"
    }

    Write-Log "Installing pancake $Version for windows-$arch into $installDir"
    if (Test-Path $destinationExe) {
        Write-Log 'Existing installation found; updating.'
    }

    Download-With-Retry -Url $downloadUrl -Destination $tempFile
    if (-not (Test-Path $tempFile)) {
        Write-Err "Downloaded file not found at $tempFile"
        exit 1
    }

    try {
        # Verify BEFORE touching the existing install: a bad download must
        # never replace a working binary.
        Confirm-Checksum -BinaryPath $tempFile -ChecksumsUrl $checksumsUrl

        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }

        Stop-RunningPancake
        Install-BinaryFile -Source $tempFile -Destination $destinationExe
    } finally {
        Remove-Item $tempFile -Force -ErrorAction SilentlyContinue
    }

    $scope = if (Test-IsAdmin) { [EnvironmentVariableTarget]::Machine } else { [EnvironmentVariableTarget]::User }
    Ensure-Path -PathToAdd $installDir -Scope $scope

    Remove-StaleInstalls -CurrentInstallDir $installDir

    $versionOutput = & $destinationExe version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Installed pancake $versionOutput."
        Write-Log "Run 'pancake help' to get started (open a NEW terminal if the command is not found)."
    } else {
        Write-Err "Verification failed: '$destinationExe version' exited with code $LASTEXITCODE."
        exit 1
    }
}

function Uninstall-Pancake {
    $isAdmin = Test-IsAdmin
    $removedAny = $false
    $failedAny = $false
    Stop-RunningPancake

    foreach ($installDir in Get-KnownInstallDirs) {
        $destinationExe = Join-Path $installDir "$BinaryName.exe"
        if (Test-Path $destinationExe) {
            try {
                Remove-Item -Path $destinationExe -Force -ErrorAction Stop
                Write-Log "Removed $destinationExe"
                $removedAny = $true
            } catch {
                $failedAny = $true
                if (-not $isAdmin -and $installDir.StartsWith($env:ProgramFiles)) {
                    Write-Warn "Cannot remove $destinationExe without admin rights. Re-run this script from an elevated PowerShell."
                } else {
                    Write-Warn "Could not remove $destinationExe : $($_.Exception.Message)
If pancake is still running, kill it via Task Manager and retry."
                }
                continue
            }
        }
        Remove-Item (Join-Path $installDir "$BinaryName.exe.old") -Force -ErrorAction SilentlyContinue
        if ((Test-Path $installDir) -and -not (Get-ChildItem $installDir -ErrorAction SilentlyContinue)) {
            Remove-Item -Path $installDir -Force -ErrorAction SilentlyContinue
        }
        Remove-From-Path -PathToRemove $installDir -Scope ([EnvironmentVariableTarget]::User)
        if ($isAdmin) { Remove-From-Path -PathToRemove $installDir -Scope ([EnvironmentVariableTarget]::Machine) }
    }

    if ($removedAny) {
        Write-Success 'Pancake binary removed.'
    } elseif (-not $failedAny) {
        Write-Log 'Pancake binary not found. Already uninstalled.'
    }

    if ($Purge) {
        $configFile = Join-Path $env:USERPROFILE 'pancake.yml'
        $pancakeHome = Join-Path $env:USERPROFILE 'pancake'
        if (Test-Path $configFile) {
            Remove-Item $configFile -Force
            Write-Log "Removed $configFile"
        }
        if (Test-Path $pancakeHome) {
            if (Confirm-Action "Also remove pancake project directory $pancakeHome? This deletes all synced projects.") {
                Remove-Item $pancakeHome -Recurse -Force
                Write-Log "Removed $pancakeHome"
            } else {
                Write-Warn "Kept $pancakeHome."
            }
        }
    } else {
        Write-Log 'Config (~/pancake.yml) and projects (~/pancake/) were left in place. Re-run with -Purge to remove them.'
    }

    if ($failedAny) { exit 1 }
}

switch ($Action) {
    'install'   { Install-Pancake }
    'uninstall' { Uninstall-Pancake }
    default     { Write-Err "Unknown action: $Action"; exit 1 }
}
