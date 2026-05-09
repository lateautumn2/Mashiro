<#
.SYNOPSIS
    Mashiro Agent installer for Windows
.DESCRIPTION
    Downloads the pre-built agent binary from GitHub Releases and starts it as a background process.
.PARAMETER BaseUrl
    Mashiro server base URL (e.g. https://status.example.com)
.PARAMETER Token
    Server authentication token
.PARAMETER Version
    Agent version (default: latest)
.PARAMETER InstallDir
    Install directory (default: $env:USERPROFILE\.mashiro-agent)
.EXAMPLE
    powershell -ExecutionPolicy Bypass -NoProfile -File install.ps1 -BaseUrl "https://status.example.com" -Token "your-token"
#>

param(
    [string]$BaseUrl,
    [string]$Token,
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\.mashiro-agent"
)

if (-not $BaseUrl) { $BaseUrl = $env:MASHIRO_BASE_URL }
if (-not $Token)    { $Token    = $env:MASHIRO_AGENT_TOKEN }

if (-not $BaseUrl -or -not $Token) {
    Write-Error "Usage: install.ps1 -BaseUrl <url> -Token <token>"
    Write-Error "Or set env: MASHIRO_BASE_URL, MASHIRO_AGENT_TOKEN"
    exit 1
}

$ErrorActionPreference = "Stop"
$Repo = "lateautumn2/Mashiro"
$ReportUrl = "$($BaseUrl.TrimEnd('/'))/api/agent/report"
$ProgressPreference = "SilentlyContinue"

$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$BinaryName = "mashiro-agent_windows_$Arch.exe"

if ($Version -eq "latest") {
    Write-Host "resolving latest version..." -ForegroundColor Green
    try {
        $releaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -TimeoutSec 10
        $Version = $releaseInfo.tag_name
    } catch {
        Write-Error "failed to resolve latest version from GitHub API: $_"
        exit 1
    }
    Write-Host "resolved latest version: $Version" -ForegroundColor Green
}

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$BinaryName"

Write-Host "downloading mashiro-agent $Version for windows/$Arch..." -ForegroundColor Green
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile "$InstallDir\mashiro-agent.exe" -TimeoutSec 120
} catch {
    Write-Error "download failed: $DownloadUrl"
    Write-Error "make sure version $Version exists and has a release asset for windows/$Arch"
    exit 1
}

Write-Host "binary installed to $InstallDir\mashiro-agent.exe" -ForegroundColor Green

$StartScript = Join-Path $InstallDir "start-agent.ps1"
@"
`$env:MASHIRO_SERVER_URL = '$ReportUrl'
`$env:MASHIRO_AGENT_ID = '$Token'
& '$InstallDir\mashiro-agent.exe' *>> '$InstallDir\agent.log'
"@ | Set-Content -Path $StartScript -Encoding UTF8

Write-Host "starting Mashiro Agent..." -ForegroundColor Green
$Process = Start-Process powershell `
    -ArgumentList '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $StartScript `
    -WorkingDirectory $InstallDir `
    -WindowStyle Hidden `
    -PassThru

Start-Sleep -Seconds 2

if (Get-Process -Id $Process.Id -ErrorAction SilentlyContinue) {
    Write-Host "Mashiro Agent started. PID: $($Process.Id)" -ForegroundColor Green
    Write-Host "  log: $InstallDir\agent.log" -ForegroundColor Green
} else {
    Write-Error "agent failed to stay running, check log: $InstallDir\agent.log"
    exit 1
}
