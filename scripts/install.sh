#!/usr/bin/env bash
# exist.io — One-line automated installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/exist.io/main/scripts/install.sh | bash
#
# Environment variables:
#   VERSION: Specific version to install (default: latest release, e.g. v0.5.0-beta)
#   INSTALL_DIR: Directory where binaries will be installed (default: /usr/local/bin or ~/.local/bin)

set -euo pipefail

REPO="RavaniRoshan/exist.io"
VERSION="${VERSION:-v0.5.0-beta}"

# Color output helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BLUE}${BOLD}==> exist.io Installer${NC}"

# Detect OS
OS_RAW="$(uname -s)"
case "$OS_RAW" in
    Linux*)     OS="linux" ;;
    Darwin*)    OS="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
    *)
        echo -e "${RED}Error: Unsupported operating system: $OS_RAW${NC}" >&2
        exit 1
        ;;
esac

# Detect Architecture
ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH_RAW${NC}" >&2
        exit 1
        ;;
esac

echo -e "    Detected Platform: ${BOLD}${OS}/${ARCH}${NC}"
echo -e "    Target Version:    ${BOLD}${VERSION}${NC}"

EXT="tar.gz"
if [ "$OS" = "windows" ]; then
    EXT="zip"
fi

ARCHIVE_NAME="exist_${VERSION}_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/SHA256SUMS"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo -e "--> Downloading ${BOLD}${ARCHIVE_NAME}${NC}..."
if ! curl -fsSL -o "$TMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Failed to download archive from $DOWNLOAD_URL${NC}" >&2
    echo -e "Please verify the release version exists at https://github.com/${REPO}/releases" >&2
    exit 1
fi

echo -e "--> Downloading checksums..."
if curl -fsSL -o "$TMP_DIR/SHA256SUMS" "$CHECKSUMS_URL" 2>/dev/null; then
    echo -e "--> Verifying SHA-256 checksum..."
    (
        cd "$TMP_DIR"
        if command -v sha256sum >/dev/null 2>&1; then
            grep "$ARCHIVE_NAME" SHA256SUMS | sha256sum -c - >/dev/null 2>&1 || {
                echo -e "${RED}Error: Checksum verification failed!${NC}" >&2
                exit 1
            }
        elif command -v shasum >/dev/null 2>&1; then
            grep "$ARCHIVE_NAME" SHA256SUMS | shasum -a 256 -c - >/dev/null 2>&1 || {
                echo -e "${RED}Error: Checksum verification failed!${NC}" >&2
                exit 1
            }
        fi
    )
    echo -e "    ${GREEN}Checksum verified successfully.${NC}"
else
    echo -e "    Notice: Checksum file not found on server; skipping verification."
fi

echo -e "--> Extracting binaries..."
if [ "$EXT" = "zip" ]; then
    unzip -q -o "$TMP_DIR/$ARCHIVE_NAME" -d "$TMP_DIR/extracted"
else
    mkdir -p "$TMP_DIR/extracted"
    tar -xzf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR/extracted"
fi

# Locate extracted binaries
STAGE_SUBDIR="$(find "$TMP_DIR/extracted" -maxdepth 2 -type f \( -name "exist" -o -name "exist.exe" \) -exec dirname {} \; | head -n 1)"
if [ -z "$STAGE_SUBDIR" ]; then
    STAGE_SUBDIR="$TMP_DIR/extracted"
fi

# Determine target install directory
if [ -n "${INSTALL_DIR:-}" ]; then
    TARGET_DIR="$INSTALL_DIR"
elif [ -w "/usr/local/bin" ]; then
    TARGET_DIR="/usr/local/bin"
elif command -v sudo >/dev/null 2>&1 && [ -t 0 ]; then
    USE_SUDO=1
    TARGET_DIR="/usr/local/bin"
else
    TARGET_DIR="$HOME/.local/bin"
fi

mkdir -p "$TARGET_DIR" 2>/dev/null || true

echo -e "--> Installing to ${BOLD}${TARGET_DIR}${NC}..."
BIN_EXT=""
if [ "$OS" = "windows" ]; then
    BIN_EXT=".exe"
fi

if [ "${USE_SUDO:-0}" = "1" ]; then
    sudo cp "$STAGE_SUBDIR/existd$BIN_EXT" "$STAGE_SUBDIR/exist$BIN_EXT" "$TARGET_DIR/"
    sudo chmod 755 "$TARGET_DIR/existd$BIN_EXT" "$TARGET_DIR/exist$BIN_EXT"
else
    cp "$STAGE_SUBDIR/existd$BIN_EXT" "$STAGE_SUBDIR/exist$BIN_EXT" "$TARGET_DIR/"
    chmod 755 "$TARGET_DIR/existd$BIN_EXT" "$TARGET_DIR/exist$BIN_EXT"
fi

# PATH check
case ":$PATH:" in
    *:"$TARGET_DIR":*) ;;
    *)
        echo -e "${BLUE}Notice: ${TARGET_DIR} is not in your PATH.${NC}"
        echo -e "Add it by running:"
        echo -e "    export PATH=\"${TARGET_DIR}:\$PATH\""
        ;;
esac

echo ""
echo -e "${GREEN}${BOLD}================================================================${NC}"
echo -e "${GREEN}${BOLD}         exist.io successfully installed!                      ${NC}"
echo -e "${GREEN}${BOLD}================================================================${NC}"
echo ""
echo -e "Installed binaries:"
echo -e "  - ${BOLD}$TARGET_DIR/existd$BIN_EXT${NC} (Forge daemon: Smart-HTTP, SSH, MCP, Web UI)"
echo -e "  - ${BOLD}$TARGET_DIR/exist$BIN_EXT${NC}  (Operator CLI & Agent integration)"
echo ""
echo -e "${BOLD}Quickstart (5-Minute Agent Run):${NC}"
echo -e "  1. Start the forge daemon:"
echo -e "     ${BLUE}$ existd${NC}"
echo ""
echo -e "  2. In another terminal, bootstrap your first repo:"
echo -e "     ${BLUE}$ exist init my-repo${NC}"
echo ""
echo -e "  3. Connect your coding agent (Claude Code, Cursor, Cline):"
echo -e "     - MCP JSON-RPC Server: ${BOLD}http://localhost:8080/mcp${NC}"
echo -e "     - Read-only Sandbox:   ${BOLD}http://localhost:8080/mcp/readonly${NC}"
echo -e "     - Web UI:              ${BOLD}http://localhost:8080/${NC}"
echo ""
echo -e "Documentation & Tutorials: ${BLUE}https://github.com/${REPO}#readme${NC}"
echo -e "Submit Beta Feedback:      ${BLUE}https://github.com/${REPO}/issues/new?template=beta_feedback.yml${NC}"
echo ""
