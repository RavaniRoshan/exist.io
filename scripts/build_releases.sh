#!/usr/bin/env bash
# build_releases.sh — Cross-compiles pure-Go existd and exist binaries with embedded web UI.
# Targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.

set -euo pipefail

VERSION="${VERSION:-v0.5.0-beta}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RELEASE_DIR="$ROOT_DIR/dist/releases"

if ! command -v go >/dev/null 2>&1; then
    if [ -x "/home/linuxbrew/.linuxbrew/bin/go" ]; then
        export PATH="/home/linuxbrew/.linuxbrew/bin:$PATH"
    fi
fi

echo "=== Building exist.io release $VERSION ==="
echo "Target architectures: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64"

# 1. Build frontend assets for embedding
echo "--> Compiling web UI production assets..."
(cd "$ROOT_DIR/web" && npm run build >/dev/null 2>&1)

mkdir -p "$RELEASE_DIR"
rm -rf "$RELEASE_DIR"/*

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

LDFLAGS="-s -w -X 'exist.io/internal/cli.Version=$VERSION'"

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    EXT=""
    if [ "$GOOS" = "windows" ]; then
        EXT=".exe"
    fi

    ARCHIVE_NAME="exist_${VERSION}_${GOOS}_${GOARCH}"
    STAGE_DIR="$RELEASE_DIR/$ARCHIVE_NAME"
    mkdir -p "$STAGE_DIR"

    echo "--> Compiling for $GOOS/$GOARCH..."
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o "$STAGE_DIR/existd$EXT" "$ROOT_DIR/cmd/existd"
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o "$STAGE_DIR/exist$EXT" "$ROOT_DIR/cmd/exist"
    cp "$ROOT_DIR/README.md" "$ROOT_DIR/LICENSE" "$STAGE_DIR/"

    if [ "$GOOS" = "windows" ]; then
        (cd "$RELEASE_DIR" && zip -q -r "${ARCHIVE_NAME}.zip" "$ARCHIVE_NAME")
    else
        tar -C "$RELEASE_DIR" -czf "$RELEASE_DIR/${ARCHIVE_NAME}.tar.gz" "$ARCHIVE_NAME"
    fi

    rm -rf "$STAGE_DIR"
done

echo "--> Generating SHA-256 checksums..."
(
    cd "$RELEASE_DIR"
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum *.tar.gz *.zip > SHA256SUMS
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 *.tar.gz *.zip > SHA256SUMS
    fi
)

echo "=== Release build complete ==="
ls -lh "$RELEASE_DIR"
cat "$RELEASE_DIR/SHA256SUMS"
