<#
.SYNOPSIS
    Builds the Windows binary and Inno Setup installer for Missile Command.
.PARAMETER Version
    Version string to inject into the installer (e.g. "1.1.0"). Default is "1.1.0".
.PARAMETER BinaryPath
    Optional path to precompiled missile-command-windows-amd64.exe.
#>
param (
    [string]$Version = "1.1.0",
    [string]$BinaryPath = ""
)

$ErrorActionPreference = "Stop"

$RootDir = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $RootDir

$DistDir = Join-Path $RootDir "dist"
if (-not (Test-Path $DistDir)) {
    New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
}

# Determine binary path
if ([string]::IsNullOrWhiteSpace($BinaryPath)) {
    $BinaryPath = Join-Path $DistDir "missile-command-windows-amd64.exe"
}

if (-not (Test-Path $BinaryPath)) {
    Write-Host "🔨 Compiling Windows binary ($BinaryPath)..." -ForegroundColor Cyan
    go build -ldflags="-s -w" -o $BinaryPath .
}

# Locate Inno Setup Compiler (iscc.exe)
$IsccCmd = Get-Command iscc.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source -First 1

if (-not $IsccCmd) {
    $CandidatePaths = @(
        "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
        "${env:ProgramFiles}\Inno Setup 6\ISCC.exe",
        "${env:LocalAppData}\Programs\Inno Setup 6\ISCC.exe"
    )
    foreach ($Path in $CandidatePaths) {
        if (Test-Path $Path) {
            $IsccCmd = $Path
            break
        }
    }
}

if (-not $IsccCmd) {
    Write-Error "Inno Setup Compiler (ISCC.exe) not found in PATH or standard Program Files locations. Please install Inno Setup 6 via: choco install innosetup or winget install JRSoftware.InnoSetup"
    exit 1
}

Write-Host "📦 Compiling Windows Setup Installer with Inno Setup (Version: $Version)..." -ForegroundColor Cyan
$IssPath = Join-Path $RootDir "scripts\setup-windows.iss"

& $IsccCmd "/DAppVersion=$Version" "/DBinaryPath=$BinaryPath" $IssPath

$SetupExe = Join-Path $DistDir "Missile-Command-Setup-windows-amd64.exe"
if (Test-Path $SetupExe) {
    Write-Host "✅ Successfully built Windows Setup: $SetupExe" -ForegroundColor Green
} else {
    Write-Error "Installer generation completed, but $SetupExe was not found."
    exit 1
}
