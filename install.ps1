$ErrorActionPreference = "Stop"

$repo = "Naveen-M99/conn"
$binaryName = "conn.exe"

# Detect architecture (64-bit vs 32-bit)
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$downloadUrl = "https://github.com/$repo/releases/latest/download/conn-windows-$arch.exe"

# Target installation directory inside user local app data (no admin prompt required)
$destDir = Join-Path $env:LOCALAPPDATA "Programs\conn"
$destPath = Join-Path $destDir $binaryName

Write-Host "==> Target directory: $destDir"
if (-not (Test-Path $destDir)) {
    New-Item -ItemType Directory -Force -Path $destDir | Out-Null
}

Write-Host "==> Downloading conn for Windows ($arch)..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $destPath -UseBasicParsing

# Register to user PATH if missing
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$destDir*") {
    Write-Host "==> Adding $destDir to user PATH environment variable..."
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$destDir", "User")
    $env:Path += ";$destDir"
}

Write-Host "==> Successfully installed conn!" -ForegroundColor Green
Write-Host "You can now run 'conn --help' (open a new PowerShell session if not recognized immediately)."