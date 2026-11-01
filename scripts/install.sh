#!/usr/bin/env bash
# exist.io — One-line automated installer
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/exist.io/main/scripts/install.sh | bash
#
# Environment variables:
#   EXIST_VERSION: Specific release tag to install (default: newest GitHub release)
#   INSTALL_DIR: Directory where binaries will be installed (default: /usr/local/bin or ~/.local/bin)
#   EXIST_INSTALL_INSECURE=1: Bypass checksum verification when no hashing tool exists (warned)

set -euo pipefail

REPO="RavaniRoshan/exist.io"
VERSION="${EXIST_VERSION:-}"

# Color output helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BLUE}${BOLD}==> exist.io Installer${NC}"

while [ $# -gt 0 ]; do
    case "$1" in
        --version)
            if [ $# -lt 2 ]; then
                echo -e "${RED}Error: --version requires a release tag${NC}" >&2
                exit 1
            fi
            VERSION="$2"
            shift 2
            ;;
        --version=*)
            VERSION="${1#*=}"
            shift
            ;;
        -h|--help)
            echo "Usage: install.sh [--version <tag>]"
            exit 0
            ;;
        *)
            echo -e "${RED}Error: unknown argument: $1${NC}" >&2
            exit 1
            ;;
    esac
done

resolve_latest() {
    local json tag
    local -a auth=()
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
    fi
    json="$(curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors -H 'Accept: application/vnd.github+json' "${auth[@]}" \
        "https://api.github.com/repos/${REPO}/releases?per_page=30")" || return 1
    tag="$(printf '%s' "$json" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
    [ -n "$tag" ] || return 1
    printf '%s' "$tag"
}

if [ -z "$VERSION" ]; then
    if ! VERSION="$(resolve_latest)"; then
        echo -e "${RED}Error: could not resolve the latest release from the GitHub API.${NC}" >&2
        echo -e "Set EXIST_VERSION=<tag> or pass --version <tag> to install a specific release." >&2
        exit 1
    fi
fi

case "$VERSION" in
    v[0-9]*) ;;
    *)
        echo -e "${RED}Error: invalid release tag: $VERSION${NC}" >&2
        exit 1
        ;;
esac

echo -e "    Resolved Version:  ${BOLD}${VERSION}${NC}"

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

EXT="tar.gz"
if [ "$OS" = "windows" ]; then
    EXT="zip"
fi

ARCHIVE_NAME="exist_${VERSION}_${OS}_${ARCH}.${EXT}"
RELEASE_BASE_URL="${RELEASE_BASE_URL:-https://github.com/${REPO}/releases/download/${VERSION}}"
DOWNLOAD_URL="${RELEASE_BASE_URL}/${ARCHIVE_NAME}"
CHECKSUMS_URL="${RELEASE_BASE_URL}/SHA256SUMS"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if ! curl -fsSI -L --retry 5 --retry-delay 2 --retry-all-errors "$DOWNLOAD_URL" >/dev/null 2>&1; then
    echo -e "${RED}Error: release asset not found: $DOWNLOAD_URL${NC}" >&2
    exit 1
fi

echo -e "--> Downloading ${BOLD}${ARCHIVE_NAME}${NC}..."
if ! curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors -o "$TMP_DIR/$ARCHIVE_NAME" "$DOWNLOAD_URL"; then
    echo -e "${RED}Error: Failed to download archive from $DOWNLOAD_URL${NC}" >&2
    echo -e "Please verify the release version exists at https://github.com/${REPO}/releases" >&2
    exit 1
fi

echo -e "--> Downloading checksums..."
if ! curl -fsSL --retry 5 --retry-delay 2 --retry-all-errors -o "$TMP_DIR/SHA256SUMS" "$CHECKSUMS_URL"; then
    echo -e "${RED}Error: could not retrieve the SHA-256 manifest from $CHECKSUMS_URL${NC}" >&2
    echo -e "Refusing to install an unverified binary." >&2
    exit 1
fi

if ! grep -Eq "[[:space:]]${ARCHIVE_NAME}$" "$TMP_DIR/SHA256SUMS"; then
    echo -e "${RED}Error: SHA256SUMS contains no entry for ${ARCHIVE_NAME}.${NC}" >&2
    echo -e "Refusing to install an unverified binary." >&2
    exit 1
fi

HASH_TOOL=""
if command -v sha256sum >/dev/null 2>&1; then
    HASH_TOOL="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
    HASH_TOOL="shasum -a 256"
fi

if [ -z "$HASH_TOOL" ]; then
    if [ "${EXIST_INSTALL_INSECURE:-0}" = "1" ]; then
        echo -e "${RED}${BOLD}WARNING: EXIST_INSTALL_INSECURE=1 — checksum verification bypassed.${NC}" >&2
        echo -e "${RED}The downloaded binary is UNVERIFIED. Do not do this in production.${NC}" >&2
    else
        echo -e "${RED}Error: no SHA-256 tool available (sha256sum or shasum required).${NC}" >&2
        echo -e "Refusing to install an unverified binary. Set EXIST_INSTALL_INSECURE=1 to override." >&2
        exit 1
    fi
else
    echo -e "--> Verifying SHA-256 checksum..."
    if ! (cd "$TMP_DIR" && grep -E "[[:space:]]${ARCHIVE_NAME}$" SHA256SUMS | $HASH_TOOL -c - >/dev/null 2>&1); then
        echo -e "${RED}Error: Checksum verification failed!${NC}" >&2
        echo -e "Refusing to install a binary that does not match the manifest." >&2
        exit 1
    fi
    echo -e "    ${GREEN}Checksum verified successfully.${NC}"
fi

echo -e "--> Extracting binaries..."
if [ "$EXT" = "zip" ]; then
    mkdir -p "$TMP_DIR/extracted"
    if command -v unzip >/dev/null 2>&1; then
        unzip -q -o "$TMP_DIR/$ARCHIVE_NAME" -d "$TMP_DIR/extracted"
    else
        tar -xf "$TMP_DIR/$ARCHIVE_NAME" -C "$TMP_DIR/extracted"
    fi
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
echo -e "  1. Bootstrap your forge administrator and starter repository:"
echo -e "     ${BLUE}$ existd bootstrap --email admin@corp.exist.io --namespace acme --data-dir ./data${NC}"
echo -e "     Store the API key printed once by bootstrap."
echo ""
echo -e "  2. Start the forge daemon:"
echo -e "     ${BLUE}$ existd${NC}"
echo ""
echo -e "  3. Connect your coding agent (Claude Code, Cursor, Cline) using the MCP snippet printed by bootstrap."
echo -e "     Replace <your-admin-token> with the printed API key:"
echo -e "     - MCP JSON-RPC Server: ${BOLD}http://localhost:8080/mcp${NC}"
echo -e "     - Read-only Sandbox:   ${BOLD}http://localhost:8080/mcp/readonly${NC}"
echo -e "     - Web UI:              ${BOLD}http://localhost:8080/${NC}"
echo ""
echo -e "Documentation & Tutorials: ${BLUE}https://ravaniroshan.github.io/exist.io/?ref=installer${NC}"
echo -e "Submit Beta Feedback:      ${BLUE}https://github.com/${REPO}/issues/new?template=beta_feedback.yml&ref=installer${NC}"
echo ""
