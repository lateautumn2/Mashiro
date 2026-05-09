#!/usr/bin/env bash
set -euo pipefail

REPO="lateautumn2/Mashiro"
INSTALL_DIR="${HOME}/.mashiro-agent"
BASE_URL=""
TOKEN=""
VERSION="latest"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*" >&2; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }

usage() {
    cat <<EOF
Usage: install.sh -e <base_url> -t <token> [-v version] [-d install_dir]

  -e  Mashiro server base URL (e.g. https://status.example.com)
  -t  Server authentication token
  -v  Agent version (default: latest)
  -d  Install directory (default: ~/.mashiro-agent)
EOF
    exit 1
}

while getopts "e:t:v:d:h" opt; do
    case $opt in
        e) BASE_URL="$OPTARG" ;;
        t) TOKEN="$OPTARG" ;;
        v) VERSION="$OPTARG" ;;
        d) INSTALL_DIR="$OPTARG" ;;
        h) usage ;;
        *) usage ;;
    esac
done

if [ -z "$BASE_URL" ] || [ -z "$TOKEN" ]; then
    error "missing required arguments: -e and -t"
    usage
fi

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux)  os="linux" ;;
        Darwin) os="darwin" ;;
        *)
            if uname -s | grep -qiE "mingw|msys|cygwin"; then
                error "For Windows, use install.ps1 instead"
            else
                error "Unsupported OS: $(uname -s)"
            fi
            exit 1
            ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

download_file() {
    local url="$1" output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL --retry 3 --retry-delay 2 "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q --tries=3 --wait=2 "$url" -O "$output"
    else
        error "neither curl nor wget found"
        exit 1
    fi
}

resolve_version() {
    if [ "$VERSION" = "latest" ]; then
        local api_url="https://api.github.com/repos/${REPO}/releases/latest"
        local tag
        tag=$(download_file "$api_url" /dev/stdout 2>/dev/null | grep -o '"tag_name":[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
        if [ -z "$tag" ]; then
            error "failed to resolve latest version from GitHub API"
            exit 1
        fi
        VERSION="$tag"
        info "resolved latest version: ${VERSION}"
    fi
}

download_agent() {
    local platform="$1"
    local binary_name="mashiro-agent_${platform}"
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"

    info "downloading mashiro-agent ${VERSION} for ${platform}..."
    mkdir -p "${INSTALL_DIR}"

    download_file "$download_url" "${INSTALL_DIR}/mashiro-agent" || {
        error "download failed: ${download_url}"
        error "make sure version ${VERSION} exists and has a release asset for ${platform}"
        exit 1
    }
    chmod +x "${INSTALL_DIR}/mashiro-agent"
    info "binary installed to ${INSTALL_DIR}/mashiro-agent"
}

setup_systemd() {
    local report_url="${BASE_URL%/}/api/agent/report"
    local service_name="mashiro-agent"

    if ! command -v systemctl >/dev/null 2>&1; then
        return 1
    fi

    local service_file
    if [ "$(id -u)" -eq 0 ] || command -v sudo >/dev/null 2>&1; then
        service_file="/etc/systemd/system/${service_name}.service"
        local esc=$'\n'
        local cmd_prefix="sudo"
    else
        mkdir -p "${HOME}/.config/systemd/user"
        service_file="${HOME}/.config/systemd/user/${service_name}.service"
        local esc=$'\n'
        local cmd_prefix=""
    fi

    info "setting up systemd service: ${service_file}"

    ${cmd_prefix} tee "$service_file" > /dev/null <<EOF
[Unit]
Description=Mashiro Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Environment="MASHIRO_SERVER_URL=${report_url}"
Environment="MASHIRO_AGENT_ID=${TOKEN}"
ExecStart=${INSTALL_DIR}/mashiro-agent
Restart=always
RestartSec=10
StandardOutput=append:${INSTALL_DIR}/agent.log
StandardError=append:${INSTALL_DIR}/agent.log

[Install]
WantedBy=multi-user.target
EOF

    ${cmd_prefix} systemctl daemon-reload
    ${cmd_prefix} systemctl enable "${service_name}"
    ${cmd_prefix} systemctl restart "${service_name}"

    info "Mashiro Agent systemd service started"
    info "  status: systemctl status ${service_name}"
    info "  logs:   journalctl -u ${service_name} -f"
    return 0
}

setup_nohup() {
    local report_url="${BASE_URL%/}/api/agent/report"

    info "starting agent via nohup (systemd not available)"

    cat > "${INSTALL_DIR}/start.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
export MASHIRO_SERVER_URL="${report_url}"
export MASHIRO_AGENT_ID="${TOKEN}"
exec "${INSTALL_DIR}/mashiro-agent"
EOF
    chmod +x "${INSTALL_DIR}/start.sh"

    nohup "${INSTALL_DIR}/start.sh" > "${INSTALL_DIR}/agent.log" 2>&1 &
    local pid=$!
    sleep 1
    if kill -0 "$pid" 2>/dev/null; then
        info "Mashiro Agent started (PID: ${pid})"
        info "  log: ${INSTALL_DIR}/agent.log"
    else
        error "agent failed to stay running, check log: ${INSTALL_DIR}/agent.log"
        exit 1
    fi
}

PLATFORM=$(detect_platform)
info "detected platform: ${PLATFORM}"

resolve_version
download_agent "$PLATFORM"

if [ "$(uname -s)" = "Linux" ] && setup_systemd; then
    :
else
    setup_nohup
fi
