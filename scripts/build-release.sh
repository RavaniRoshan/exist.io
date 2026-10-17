#!/usr/bin/env bash
# exist.io — Multi-Architecture Release Packaging Script
# Builds hermetic pure-Go binaries (CGO_ENABLED=0) and packages release archives with SHA256 checksums.

set -euo pipefail

VERSION="${1:-${VERSION:-v0.5.0-beta}}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="${ROOT_DIR}/dist/release"

echo "==> Building exist.io release ${VERSION}"
echo "    Root directory:   ${ROOT_DIR}"
echo "    Output directory: ${OUTPUT_DIR}"

# 1. Build frontend production assets
echo "--> Building embedded Web UI assets..."
(
  cd "${ROOT_DIR}/web"
  npm ci --prefer-offline --no-audit 2>/dev/null || npm install
  npm run build
)

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

PLATFORMS=(
  "linux/amd64/tar.gz"
  "linux/arm64/tar.gz"
  "darwin/amd64/tar.gz"
  "darwin/arm64/tar.gz"
  "windows/amd64/zip"
)

for target in "${PLATFORMS[@]}"; do
  IFS="/" read -r GOOS GOARCH EXT <<< "${target}"
  echo "--> Compiling for ${GOOS}/${GOARCH}..."

  STAGE_DIR="$(mktemp -d)"
  BIN_EXT=""
  if [ "${GOOS}" = "windows" ]; then
    BIN_EXT=".exe"
  fi

  # Build binaries
  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "${STAGE_DIR}/existd${BIN_EXT}" \
    "${ROOT_DIR}/cmd/existd"

  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o "${STAGE_DIR}/exist${BIN_EXT}" \
    "${ROOT_DIR}/cmd/exist"

  cp "${ROOT_DIR}/README.md" "${STAGE_DIR}/"
  cp "${ROOT_DIR}/LICENSE" "${STAGE_DIR}/"

  ARCHIVE_NAME="exist_${VERSION}_${GOOS}_${GOARCH}.${EXT}"
  ARCHIVE_PATH="${OUTPUT_DIR}/${ARCHIVE_NAME}"

  echo "--> Packaging ${ARCHIVE_NAME}..."
  if [ "${EXT}" = "zip" ]; then
    (cd "${STAGE_DIR}" && zip -q -r "${ARCHIVE_PATH}" .)
  else
    tar -czf "${ARCHIVE_PATH}" -C "${STAGE_DIR}" .
  fi

  rm -rf "${STAGE_DIR}"
done

echo "--> Generating SHA256SUMS..."
(
  cd "${OUTPUT_DIR}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum exist_* > SHA256SUMS
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 exist_* > SHA256SUMS
  fi
)

echo "==> Build complete. Artifacts in ${OUTPUT_DIR}:"
ls -lh "${OUTPUT_DIR}"
