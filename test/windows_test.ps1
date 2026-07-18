#requires -Version 5.1
<#
.SYNOPSIS
    Windows e2e tests for pancake.
.DESCRIPTION
    Builds pancake, runs it inside an isolated mock USERPROFILE, and exercises
    init, config validation, and project list edge cases. Mirrors the bash
    harness in test/01..04. Install/uninstall of windows.ps1 itself is tested
    separately via a mock HTTP server in CI.
#>
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$repoRoot = Resolve-Path "$PSScriptRoot/.."

$script:PassCount = 0
$script:FailCount = 0

function Write-Pass { param([string]$Label) $script:PassCount++; Write-Host "  PASS  $Label" -ForegroundColor Green }
function Write-Fail { param([string]$Label, [string]$Expected, [string]$Actual)
    $script:FailCount++
    Write-Host "  FAIL  $Label" -ForegroundColor Red
    if ($Expected) { Write-Host "        expected: $Expected" -ForegroundColor Yellow }
    if ($Actual)   { Write-Host "        actual:   $Actual" -ForegroundColor Yellow }
}

function Invoke-Pancake {
    param([string[]]$PancakeArgs, [string]$MockHome)
    $env:USERPROFILE = $MockHome
    $env:HOME = $MockHome
    & $script:PancakeBin @PancakeArgs 2>&1
}

function Assert-ExitCode {
    param([int]$Expected, [string]$Label, [string[]]$PancakeArgs, [string]$MockHome)
    $output = Invoke-Pancake -PancakeArgs $PancakeArgs -MockHome $MockHome
    $code = $LASTEXITCODE
    if ($code -eq $Expected) {
        Write-Pass "$Label (exit $code)"
    } else {
        Write-Fail $Label "exit $Expected" "exit $code`n$output"
    }
    return $output
}

function Assert-Contains {
    param([string]$Label, [string]$Needle, [string[]]$PancakeArgs, [string]$MockHome)
    $output = Invoke-Pancake -PancakeArgs $PancakeArgs -MockHome $MockHome
    if ($output -match [regex]::Escape($Needle)) {
        Write-Pass $Label
    } else {
        Write-Fail $Label "output contains '$Needle'" "$output"
    }
}

function Get-TempDir {
    if ($env:TEMP) { return $env:TEMP }
    if ($env:TMPDIR) { return $env:TMPDIR }
    if ($IsWindows -or $PSVersionTable.Platform -eq 'Win32NT') { return "$env:LOCALAPPDATA\Temp" }
    return '/tmp'
}

function New-MockHome {
    $dir = Join-Path (Get-TempDir) ("pancake_home_" + [guid]::NewGuid().ToString('N').Substring(0,8))
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    return $dir
}

Write-Host "=== Windows pancake e2e tests ===" -ForegroundColor Cyan

# Build pancake for the current platform.
Write-Host "Building pancake..." -ForegroundColor Cyan
$script:PancakeBin = Join-Path (Get-TempDir) "pancake-test.exe"
& go build -o $script:PancakeBin . 2>&1
if ($LASTEXITCODE -ne 0) { Write-Fail "build pancake" "exit 0" "exit $LASTEXITCODE"; exit 1 }
Write-Pass "build pancake"

# 1. version subcommand works.
Assert-Contains "version prints Pancake" "Pancake" -PancakeArgs @('version') -MockHome (New-MockHome)

# 2. init creates config and home.
$home1 = New-MockHome
Assert-ExitCode 0 "init creates config" -PancakeArgs @('init') -MockHome $home1 | Out-Null
if (Test-Path (Join-Path $home1 'pancake.yml')) { Write-Pass "pancake.yml created" }
else { Write-Fail "pancake.yml created" "$home1\pancake.yml" "missing" }
if (Test-Path (Join-Path $home1 'pancake')) { Write-Pass "pancake home dir created" }
else { Write-Fail "pancake home dir created" "$home1\pancake" "missing" }

# 3. edit config before init -> helpful not-found message.
$home2 = New-MockHome
Assert-Contains "edit config before init is helpful" "pancake.yml does not exist" -PancakeArgs @('edit','config') -MockHome $home2

# 4. No config -> mentions pancake init.
$home3 = New-MockHome
Assert-Contains "no config -> mentions pancake init" "pancake init" -PancakeArgs @('project','list') -MockHome $home3

# 5. Empty home field -> clear message.
$home4 = New-MockHome
@"
home:
code_editor: echo
default_ai: gemini
projects: {}
"@ | Set-Content -Path (Join-Path $home4 'pancake.yml') -Encoding UTF8
Assert-Contains "empty home -> clear message" "'home' is empty" -PancakeArgs @('project','list') -MockHome $home4

# 6. Unsupported default_ai -> mentions field.
$home5 = New-MockHome
@"
home: `$HOME/pancake
code_editor: echo
default_ai: claude
projects: {}
"@ | Set-Content -Path (Join-Path $home5 'pancake.yml') -Encoding UTF8
Assert-Contains "bad default_ai -> mentions field" "'default_ai'" -PancakeArgs @('project','list') -MockHome $home5

# 7. Valid config loads and lists project.
$home6 = New-MockHome
@"
home: `$HOME/pancake
code_editor: echo
default_ai: gemini
tools: []
projects:
  demo:
    remote_ssh_url: git@github.com:org/repo.git
"@ | Set-Content -Path (Join-Path $home6 'pancake.yml') -Encoding UTF8
Assert-Contains "valid config lists project" "demo" -PancakeArgs @('project','list') -MockHome $home6

# 8. Empty projects -> helpful message.
$home7 = New-MockHome
@"
home: `$HOME/pancake
code_editor: echo
default_ai: gemini
tools: []
projects: {}
"@ | Set-Content -Path (Join-Path $home7 'pancake.yml') -Encoding UTF8
Assert-Contains "empty projects -> helpful message" "No projects" -PancakeArgs @('project','list') -MockHome $home7

# 9. Project missing remote -> mentions remote_ssh_url.
$home8 = New-MockHome
@"
home: `$HOME/pancake
code_editor: echo
default_ai: gemini
projects:
  demo:
    run: echo hi
"@ | Set-Content -Path (Join-Path $home8 'pancake.yml') -Encoding UTF8
Assert-Contains "missing remote -> mentions remote_ssh_url" "remote_ssh_url" -PancakeArgs @('project','list') -MockHome $home8

# Summary.
Write-Host "--- Windows e2e summary ---" -ForegroundColor Cyan
$total = $script:PassCount + $script:FailCount
Write-Host "  passed: $script:PassCount / $total"
Write-Host "  failed: $script:FailCount / $total"
if ($script:FailCount -gt 0) { exit 1 }
exit 0
