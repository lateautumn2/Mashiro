package controllers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"backend/models"

	"github.com/gin-gonic/gin"
)

type DeployCommandSet struct {
	Linux   string `json:"linux"`
	Windows string `json:"windows"`
	MacOS   string `json:"macos"`
}

func buildDeployCommands(c *gin.Context, server models.Server) DeployCommandSet {
	baseURL := getRequestBaseURL(c)
	token := server.AuthToken
	baseURL = strings.TrimRight(baseURL, "/")

	linuxCurl := fmt.Sprintf(`bash <(curl -fsSL "https://raw.githubusercontent.com/lateautumn2/Mashiro/main/install.sh") -e "%s" -t "%s"`, baseURL, token)
	windowsPS := fmt.Sprintf(`powershell -ExecutionPolicy Bypass -NoProfile -Command "$env:MASHIRO_BASE_URL='%s'; $env:MASHIRO_AGENT_TOKEN='%s'; iex (irm 'https://raw.githubusercontent.com/lateautumn2/Mashiro/main/install.ps1')"`, baseURL, token)

	return DeployCommandSet{
		Linux:   linuxCurl,
		Windows: windowsPS,
		MacOS:   linuxCurl,
	}
}

func renderInstallShellScript(baseURL, token string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail

BASE_URL=%q
TOKEN=%q
INSTALL_DIR="${HOME}/.mashiro-agent"
PACKAGE_URL="${BASE_URL}/api/agent/package.zip"
REPORT_URL="${BASE_URL}/api/agent/report"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd unzip
require_cmd go

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

mkdir -p "${INSTALL_DIR}"
curl -fsSL "${PACKAGE_URL}" -o "${TMP_DIR}/agent.zip"
rm -rf "${INSTALL_DIR}/src"
mkdir -p "${INSTALL_DIR}/src"
unzip -oq "${TMP_DIR}/agent.zip" -d "${INSTALL_DIR}/src"

cd "${INSTALL_DIR}/src"
go mod tidy
go build -o "${INSTALL_DIR}/mashiro-agent" .

cat > "${INSTALL_DIR}/start.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
MASHIRO_SERVER_URL=%q MASHIRO_AGENT_ID=%q exec "${INSTALL_DIR}/mashiro-agent"
EOF

chmod +x "${INSTALL_DIR}/start.sh" "${INSTALL_DIR}/mashiro-agent"
nohup "${INSTALL_DIR}/start.sh" > "${INSTALL_DIR}/agent.log" 2>&1 &
echo "Mashiro Agent started. Log: ${INSTALL_DIR}/agent.log"
`, baseURL, token, reportURLForScript(baseURL), token)
}

func renderInstallPowerShellScript(baseURL, token string) string {
	reportURL := reportURLForScript(baseURL)
	return fmt.Sprintf(`$ErrorActionPreference = 'Stop'

$BaseUrl = %q
$Token = %q
$InstallDir = Join-Path $env:USERPROFILE ".mashiro-agent"
$PackageUrl = "$BaseUrl/api/agent/package.zip"
$ReportUrl = "$BaseUrl/api/agent/report"
$ZipPath = Join-Path $env:TEMP "mashiro-agent.zip"
$LogPath = Join-Path $InstallDir "agent.log"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest -Uri $PackageUrl -OutFile $ZipPath

$SourceDir = Join-Path $InstallDir "src"
if (Test-Path $SourceDir) {
  Remove-Item -Recurse -Force $SourceDir
}
Expand-Archive -Path $ZipPath -DestinationPath $SourceDir -Force

Push-Location $SourceDir
go mod tidy
go build -o (Join-Path $InstallDir "mashiro-agent.exe") .
Pop-Location

$StartScript = Join-Path $InstallDir "start-agent.ps1"
$AgentExePath = Join-Path $InstallDir "mashiro-agent.exe"
@'
$env:MASHIRO_SERVER_URL = %q
$env:MASHIRO_AGENT_ID = %q
'@ | Set-Content -Path $StartScript -Encoding UTF8

Add-Content -Path $StartScript -Value ('& "' + $AgentExePath + '" *>> "' + $LogPath + '"')

$Process = Start-Process powershell -ArgumentList '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $StartScript -WorkingDirectory $InstallDir -WindowStyle Hidden -PassThru
Start-Sleep -Seconds 2
if (Get-Process -Id $Process.Id -ErrorAction SilentlyContinue) {
  Write-Output "Mashiro Agent started. PID: $($Process.Id) Log: $LogPath"
} else {
  throw "Mashiro Agent failed to stay running. Check log: $LogPath"
}
`, baseURL, token, reportURL, token)
}

func reportURLForScript(baseURL string) string {
	return fmt.Sprintf("%s/api/agent/report", baseURL)
}

func agentSourceDir() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to locate controller path")
	}

	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	return filepath.Join(rootDir, "agent"), nil
}

func buildAgentPackageArchive() ([]byte, error) {
	sourceDir, err := agentSourceDir()
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	err = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)
		if strings.Contains(relativePath, "/.") || strings.HasPrefix(filepath.Base(relativePath), ".") {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = relativePath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func AgentInstallShell(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.String(http.StatusBadRequest, "missing token")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, renderInstallShellScript(getRequestBaseURL(c), token))
}

func AgentInstallPowerShell(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.String(http.StatusBadRequest, "missing token")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, renderInstallPowerShellScript(getRequestBaseURL(c), token))
}

func AgentPackageArchive(c *gin.Context) {
	archiveBytes, err := buildAgentPackageArchive()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="mashiro-agent.zip"`)
	c.Data(http.StatusOK, "application/zip", archiveBytes)
}
