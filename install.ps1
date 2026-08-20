# Copyright 2026 Dafa. MIT License.
#
# Install script for svault, the local encrypted secret vault, on Windows.
#
#   irm https://raw.githubusercontent.com/dafagareth/svault/main/install.ps1 | iex
#
# Environment overrides:
#   $env:SVAULT_VERSION   install a specific version (default: latest release)
#   $env:SVAULT_BIN_DIR   install directory (default: %LOCALAPPDATA%\svault)

$ErrorActionPreference = 'Stop'
$repo = 'dafagareth/svault'

# Only amd64 binaries are published for Windows.
$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -ne 'AMD64') {
    Write-Error "unsupported architecture: $arch (only AMD64 is published for Windows)"
}
$asset = 'svault-windows-amd64.exe'

# Resolve the version to install.
$version = $env:SVAULT_VERSION
if (-not $version) {
    Write-Host 'Resolving latest release...'
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $version = $release.tag_name
    if (-not $version) { Write-Error 'could not resolve latest version' }
}
Write-Host "Installing svault $version (windows/amd64)"

$base = "https://github.com/$repo/releases/download/$version"
$tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP ("svault-" + [guid]::NewGuid()))
try {
    # Download the binary and checksums.
    Write-Host "Downloading $asset..."
    $exePath = Join-Path $tmp $asset
    Invoke-WebRequest -Uri "$base/$asset" -OutFile $exePath
    $sumPath = Join-Path $tmp 'checksums.txt'
    Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $sumPath

    # Verify the checksum.
    $expected = (Select-String -Path $sumPath -Pattern ([regex]::Escape($asset)) |
        Select-Object -First 1).Line -split '\s+' | Select-Object -First 1
    if ($expected) {
        $actual = (Get-FileHash -Path $exePath -Algorithm SHA256).Hash.ToLower()
        if ($actual -ne $expected.ToLower()) {
            Write-Error "checksum mismatch: expected $expected, got $actual"
        }
        Write-Host 'Checksum verified.'
    } else {
        Write-Host "Warning: no checksum found for $asset, skipping verification."
    }

    # Pick an install directory.
    $binDir = $env:SVAULT_BIN_DIR
    if (-not $binDir) { $binDir = Join-Path $env:LOCALAPPDATA 'svault' }
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null

    $dest = Join-Path $binDir 'svault.exe'
    Move-Item -Force -Path $exePath -Destination $dest
    Write-Host "Installed to $dest"

    # Add to the user PATH if missing.
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -notlike "*$binDir*") {
        [Environment]::SetEnvironmentVariable('Path', "$userPath;$binDir", 'User')
        Write-Host "Added $binDir to your user PATH. Restart your terminal to use 'svault'."
    }
}
finally {
    Remove-Item -Recurse -Force $tmp
}

Write-Host "Run 'svault init' to get started."
