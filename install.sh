#!/bin/bash
# NUX Intelligent Installer
# Detects architecture, verifies checksum, and installs to /usr/local/bin

set -e

REPO_OWNER="rsdenck"
REPO_NAME="nux"
BIN_NAME="nux"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_err() { echo -e "${RED}[ERROR]${NC} $1"; }

# 1. Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux|darwin) ;;
    *)
        log_err "This script only supports Linux and macOS."
        exit 1
        ;;
esac

case $ARCH in
    x86_64)
        GOARCH="amd64"
        ;;
    aarch64)
        GOARCH="arm64"
        ;;
    armv7l)
        GOARCH="armv7"
        ;;
    *)
        log_err "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

log_info "Detected system: $OS/$GOARCH"

# 2. Get Latest Version
log_info "Fetching latest version..."
LATEST_RELEASE_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
LATEST_VERSION=$(curl -s $LATEST_RELEASE_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    log_err "Failed to fetch latest version."
    exit 1
fi

log_info "Latest version is $LATEST_VERSION"

# 3. Construct Download URL
# Pattern: troncli_v0.2.18_linux_amd64.tar.gz
# Note: goreleaser config uses {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}
# Version usually has 'v' prefix in tag, but goreleaser .Version might strip it or keep it depending on config.
# Standard goreleaser behavior: .Version is the tag without 'v' prefix if the tag has it? 
# Wait, goreleaser defaults: version is the git tag.
# If tag is v0.2.18, .Version is 0.2.18 usually.
# Let's check the artifact naming in .goreleaser.yaml: {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
# So it should be troncli_0.2.18_linux_amd64.tar.gz (without v prefix in version part of filename)

VERSION_NO_V="${LATEST_VERSION#v}"
ARTIFACT_NAME="${BIN_NAME}_${VERSION_NO_V}_${OS}_${GOARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/$ARTIFACT_NAME"
CHECKSUM_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_VERSION/checksums.txt"

TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# 4. Download Artifact and Checksum
log_info "Downloading $ARTIFACT_NAME..."
curl -sL -o "$TMP_DIR/$ARTIFACT_NAME" "$DOWNLOAD_URL"
curl -sL -o "$TMP_DIR/checksums.txt" "$CHECKSUM_URL"

# 5. Verify Checksum
log_info "Verifying checksum..."
cd $TMP_DIR
if sha256sum --ignore-missing -c checksums.txt; then
    log_info "Checksum verified successfully."
else
    log_err "Checksum verification failed!"
    exit 1
fi

# 6. Extract and Install
log_info "Extracting..."
tar -xzf "$ARTIFACT_NAME"

log_info "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
else
    sudo mv "$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
fi

# 7. Final Verification
if command -v $BIN_NAME >/dev/null; then
    INSTALLED_VERSION=$($BIN_NAME --version)
    log_info "Successfully installed $INSTALLED_VERSION"
else
    log_err "Installation failed."
    exit 1
fi
