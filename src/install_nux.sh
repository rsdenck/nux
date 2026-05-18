#!/bin/bash
# NUX - Linux CLI Manager
# One-command installer for all Linux distributions
# Usage: curl -fsSL https://denck.tech/install_nux.sh | bash
set -e

# ─────────────────────────────────────────────────────────────
#  ASCII HEADER
# ─────────────────────────────────────────────────────────────
echo -e "\033[38;5;208m"
echo " ███╗   ██╗██╗   ██╗██╗  ██╗"
echo " ████╗  ██║██║   ██║╚██╗██╔╝"
echo " ██╔██╗ ██║██║   ██║ ╚███╔╝ "
echo " ██║╚██╗██║██║   ██║ ██╔██╗ "
echo " ██║ ╚████║╚██████╔╝██╔╝ ██╗"
echo " ╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝"
echo -e "\033[0m"
echo -e "\033[38;5;208mNUX — Linux Operations Platform\033[0m"
echo "================================================"
echo " Production-Grade Linux CLI Manager"
echo " One-command installer — v0.3.0"
echo "================================================"
echo ""

# ─────────────────────────────────────────────────────────────
#  CONFIG
# ─────────────────────────────────────────────────────────────
REPO_OWNER="rsdenck"
REPO_NAME="nux"
BIN_NAME="nux"
INSTALL_DIR="/usr/local/bin"

# ─────────────────────────────────────────────────────────────
#  COLORS
# ─────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
log_err()   { echo -e "${RED}[ERROR]${NC} $1"; }
log_step()  { echo -e "${CYAN}  →${NC} $1"; }

# ─────────────────────────────────────────────────────────────
#  OS / ARCH DETECTION
# ─────────────────────────────────────────────────────────────
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux|darwin) ;;
    *)
        log_err "Unsupported OS: $OS"
        log_err "NUX supports Linux and macOS only."
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64|arm64)
        GOARCH="arm64"
        ;;
    armv7l|armv8l)
        GOARCH="armv7"
        ;;
    *)
        log_err "Unsupported architecture: $ARCH"
        log_err "Supported: amd64, arm64, armv7"
        exit 1
        ;;
esac

log_info "Detected: ${OS}/${GOARCH}"

# ─────────────────────────────────────────────────────────────
#  DISTRO DETECTION (Linux only)
# ─────────────────────────────────────────────────────────────
DISTRO=""
PKG_MANAGER=""
if [ "$OS" = "linux" ]; then
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO="$ID"
    elif command -v lsb_release &>/dev/null; then
        DISTRO=$(lsb_release -si 2>/dev/null | tr '[:upper:]' '[:lower:]')
    fi

    if command -v apt &>/dev/null; then
        PKG_MANAGER="apt"
    elif command -v dnf &>/dev/null; then
        PKG_MANAGER="dnf"
    elif command -v yum &>/dev/null; then
        PKG_MANAGER="yum"
    elif command -v pacman &>/dev/null; then
        PKG_MANAGER="pacman"
    elif command -v zypper &>/dev/null; then
        PKG_MANAGER="zypper"
    elif command -v apk &>/dev/null; then
        PKG_MANAGER="apk"
    elif command -v xbps-install &>/dev/null; then
        PKG_MANAGER="xbps"
    else
        log_warn "Could not detect package manager"
    fi

    log_info "Distribution: ${ID:-unknown}"
    log_info "Package manager: ${PKG_MANAGER:-unknown}"
fi

# ─────────────────────────────────────────────────────────────
#  CHECK DEPENDENCIES
# ─────────────────────────────────────────────────────────────
if ! command -v curl &>/dev/null; then
    log_info "curl not found, installing..."
    case "$PKG_MANAGER" in
        apt)    sudo apt update -qq && sudo apt install -y -qq curl ;;
        dnf)    sudo dnf install -y -q curl ;;
        yum)    sudo yum install -y -q curl ;;
        pacman) sudo pacman -S --noconfirm curl ;;
        zypper) sudo zypper install -y curl ;;
        apk)    sudo apk add curl ;;
        xbps)   sudo xbps-install -y curl ;;
        *)
            log_err "Please install curl manually and re-run"
            exit 1
            ;;
    esac
fi

if ! command -v tar &>/dev/null; then
    log_info "tar not found, installing..."
    case "$PKG_MANAGER" in
        apt)    sudo apt update -qq && sudo apt install -y -qq tar ;;
        dnf)    sudo dnf install -y -q tar ;;
        yum)    sudo yum install -y -q tar ;;
        pacman) sudo pacman -S --noconfirm tar ;;
        zypper) sudo zypper install -y tar ;;
        apk)    sudo apk add tar ;;
        xbps)   sudo xbps-install -y tar ;;
        *)
            log_err "Please install tar manually and re-run"
            exit 1
            ;;
    esac
fi

# ─────────────────────────────────────────────────────────────
#  FETCH LATEST RELEASE
# ─────────────────────────────────────────────────────────────
log_info "Fetching latest release from GitHub..."

LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
LATEST_VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    log_err "Failed to fetch latest version from GitHub API."
    log_err "Check: https://github.com/$REPO_OWNER/$REPO_NAME/releases"
    exit 1
fi

log_step "Latest version: ${LATEST_VERSION}"

VERSION_NO_V="${LATEST_VERSION#v}"
ARTIFACT="${BIN_NAME}_${VERSION_NO_V}_${OS}_${GOARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/$ARTIFACT"
CHECKSUM_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/checksums.txt"

# ─────────────────────────────────────────────────────────────
#  DOWNLOAD
# ─────────────────────────────────────────────────────────────
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

log_step "Downloading ${ARTIFACT}..."
curl -sL -o "$TMP_DIR/$ARTIFACT" "$DOWNLOAD_URL" || {
    log_err "Download failed: $DOWNLOAD_URL"
    exit 1
}

log_step "Downloading checksums..."
curl -sL -o "$TMP_DIR/checksums.txt" "$CHECKSUM_URL" || {
    log_warn "Checksums not available, skipping verification"
}

# ─────────────────────────────────────────────────────────────
#  VERIFY CHECKSUM
# ─────────────────────────────────────────────────────────────
if [ -f "$TMP_DIR/checksums.txt" ]; then
    log_step "Verifying SHA256 checksum..."
    cd "$TMP_DIR"
    if command -v sha256sum &>/dev/null; then
        sha256sum --ignore-missing -c checksums.txt && log_info "Checksum verified ✓" || {
            log_err "Checksum verification failed!"
            exit 1
        }
    elif command -v shasum &>/dev/null; then
        shasum -a 256 --ignore-missing -c checksums.txt && log_info "Checksum verified ✓" || {
            log_err "Checksum verification failed!"
            exit 1
        }
    else
        log_warn "No checksum tool found, skipping verification"
    fi
    cd - >/dev/null
fi

# ─────────────────────────────────────────────────────────────
#  EXTRACT & INSTALL
# ─────────────────────────────────────────────────────────────
log_step "Extracting archive..."
tar -xzf "$TMP_DIR/$ARTIFACT" -C "$TMP_DIR"

log_step "Installing to ${INSTALL_DIR}/${BIN_NAME}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
else
    sudo mv "$TMP_DIR/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
fi
chmod +x "$INSTALL_DIR/$BIN_NAME"

# ─────────────────────────────────────────────────────────────
#  VERIFY
# ─────────────────────────────────────────────────────────────
if command -v $BIN_NAME &>/dev/null; then
    echo ""
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}  NUX installed successfully!${NC}"
    echo -e "${GREEN}================================================${NC}"
    echo ""
    echo "  Run 'nux --help' to get started"
    echo "  Run 'nux doctor' to check system health"
    echo "  Run 'nux onboard' for guided setup"
    echo ""
    echo "  Documentation: https://github.com/$REPO_OWNER/$REPO_NAME"
    echo ""
else
    log_err "Installation failed — binary not found in PATH"
    exit 1
fi
