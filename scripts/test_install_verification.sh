#!/usr/bin/env bash
# Exercises the installer's checksum enforcement against a local release mirror.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALLER="$ROOT_DIR/scripts/install.sh"
VERSION="v9.9.9"
ARCHIVE="exist_${VERSION}_linux_amd64.tar.gz"

WORK="$(mktemp -d)"
SERVER_PID=""
cleanup() {
    if [ -n "$SERVER_PID" ]; then kill "$SERVER_PID" 2>/dev/null || true; fi
    rm -rf "$WORK"
}
trap cleanup EXIT

fail() {
    echo "FAIL: $*" >&2
    echo "--- stdout ---" >&2
    cat "$WORK/stdout.log" >&2 || true
    echo "--- stderr ---" >&2
    cat "$WORK/stderr.log" >&2 || true
    exit 1
}

# Fixture release.
mkdir -p "$WORK/release" "$WORK/stage"
printf '#!/bin/sh\necho existd\n' > "$WORK/stage/existd"
printf '#!/bin/sh\necho exist\n' > "$WORK/stage/exist"
chmod +x "$WORK/stage/existd" "$WORK/stage/exist"
tar -czf "$WORK/release/$ARCHIVE" -C "$WORK/stage" .
(cd "$WORK/release" && sha256sum "$ARCHIVE" > SHA256SUMS)

# Local mirror.
PORT=$(( (RANDOM % 2000) + 28000 ))
python3 -m http.server "$PORT" --directory "$WORK/release" >/dev/null 2>&1 &
SERVER_PID=$!
READY=false
for _ in $(seq 1 50); do
    if curl -fs "http://127.0.0.1:$PORT/" >/dev/null 2>&1; then READY=true; break; fi
    sleep 0.1
done
[ "$READY" = true ] || fail "local release mirror did not start"

run_installer() {
    local out="$1" shim="${2:-}" insecure="${3:-0}"
    rm -rf "$out"
    mkdir -p "$out"
    local path="$PATH"
    if [ -n "$shim" ]; then path="$shim"; fi
    env PATH="$path" \
        RELEASE_BASE_URL="http://127.0.0.1:$PORT" \
        VERSION="$VERSION" \
        INSTALL_DIR="$out" \
        EXIST_INSTALL_INSECURE="$insecure" \
        /bin/bash "$INSTALLER" >"$WORK/stdout.log" 2>"$WORK/stderr.log"
}

# Case 1: correct manifest.
run_installer "$WORK/out-ok" && true
[ -f "$WORK/out-ok/existd" ] || fail "case 1: existd was not installed"
[ -f "$WORK/out-ok/exist" ] || fail "case 1: exist was not installed"
echo "PASS: case 1 (manifest present and correct)"

# Case 2: manifest present but mismatched.
echo "0000000000000000000000000000000000000000000000000000000000000000  $ARCHIVE" > "$WORK/release/SHA256SUMS"
if run_installer "$WORK/out-mismatch"; then fail "case 2: installer accepted a mismatched checksum"; fi
[ -z "$(ls -A "$WORK/out-mismatch")" ] || fail "case 2: installer wrote binaries despite a mismatch"
echo "PASS: case 2 (manifest mismatched, install refused)"

# Case 3: manifest absent.
(cd "$WORK/release" && sha256sum "$ARCHIVE" > SHA256SUMS)
mv "$WORK/release/SHA256SUMS" "$WORK/release/SHA256SUMS.off"
if run_installer "$WORK/out-missing"; then fail "case 3: installer accepted a missing manifest"; fi
[ -z "$(ls -A "$WORK/out-missing")" ] || fail "case 3: installer wrote binaries without a manifest"
mv "$WORK/release/SHA256SUMS.off" "$WORK/release/SHA256SUMS"
echo "PASS: case 3 (manifest absent, install refused)"

# Case 4: no hashing tool available.
SHIM="$WORK/shim"
mkdir -p "$SHIM"
for tool in curl mktemp uname find dirname head grep mkdir cp chmod tar gzip rm cat sed tr; do
    source_path="$(command -v "$tool" || true)"
    if [ -n "$source_path" ]; then ln -s "$source_path" "$SHIM/$tool"; fi
done
if run_installer "$WORK/out-notool" "$SHIM" 0; then fail "case 4: installer accepted a missing hashing tool"; fi
grep -q "no SHA-256 tool" "$WORK/stderr.log" || fail "case 4: missing tool error was not reported"
[ -z "$(ls -A "$WORK/out-notool")" ] || fail "case 4: installer wrote binaries without a hashing tool"
echo "PASS: case 4 (no hashing tool, install refused)"

# Case 5: explicit warned override.
run_installer "$WORK/out-insecure" "$SHIM" 1 || fail "case 5: override did not proceed"
grep -q "WARNING: EXIST_INSTALL_INSECURE=1" "$WORK/stderr.log" || fail "case 5: override warning missing"
[ -f "$WORK/out-insecure/existd" ] || fail "case 5: override did not install"
echo "PASS: case 5 (explicit override warns and installs)"

echo "All installer verification cases passed."
