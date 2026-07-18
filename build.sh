#!/usr/bin/env bash
# Build cross-platform pancake binaries with version metadata and checksums.
#
# Usage:
#   ./build.sh                  # builds all targets into ./build/
#   ./build.sh v1.2.0           # embeds version v1.2.0
#   ./build.sh latest           # uses git describe
#
# Outputs:
#   build/pancake-<os>-<arch>          (and .exe for windows)
#   build/checksums.txt                (SHA-256 of every binary)
#   build/readme.txt                   (copy of readme.txt)

set -euo pipefail

COLOR_RED=$'\033[0;31m'
COLOR_GREEN=$'\033[0;32m'
COLOR_BLUE=$'\033[0;34m'
COLOR_RESET=$'\033[0m'
[[ ! -t 1 ]] && { COLOR_RED=""; COLOR_GREEN=""; COLOR_BLUE=""; COLOR_RESET=""; }

log()        { printf '%s[%s]%s %s\n' "$COLOR_BLUE" "$(date +'%H:%M:%S')" "$COLOR_RESET" "$1"; }
log_success() { printf '%s[SUCCESS]%s %s\n' "$COLOR_GREEN" "$COLOR_RESET" "$1"; }
die()        { printf '%s[ERROR]%s %s\n' "$COLOR_RED" "$COLOR_RESET" "$1" >&2; exit 1; }

VERSION_ARG="${1:-}"
BUILD_DIR="./build"
README_FILE="./readme.txt"

if ! command -v go >/dev/null 2>&1; then
    die "Go is not installed. Install it from https://go.dev/dl/ and retry."
fi

resolve_version() {
    if [[ -n "$VERSION_ARG" && "$VERSION_ARG" != "latest" ]]; then
        printf '%s' "${VERSION_ARG#v}"
        return
    fi
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        local described
        described="$(git describe --tags --always 2>/dev/null || echo unknown)"
        printf '%s' "${described#v}"
        return
    fi
    printf 'dev'
}

PANCAKE_VERSION="$(resolve_version)"
LDFLAGS="-s -w -X github.com/a6h15hek/pancake/utils.Version=v${PANCAKE_VERSION}"

log "Building pancake v${PANCAKE_VERSION}"

if [[ -d "$BUILD_DIR" ]]; then
    log "Removing existing $BUILD_DIR"
    rm -rf "$BUILD_DIR"
fi
mkdir -p "$BUILD_DIR"

TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

for target in "${TARGETS[@]}"; do
    goos="${target%/*}"
    goarch="${target#*/}"
    output="${BUILD_DIR}/pancake-${goos}-${goarch}"
    if [[ "$goos" == "windows" ]]; then
        output="${output}.exe"
    fi
    log "Building ${goos}/${goarch} -> ${output}"
    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
        go build -trimpath -ldflags "$LDFLAGS" -o "$output" \
        || die "Build failed for ${goos}/${goarch}"
done

if [[ -f "$README_FILE" ]]; then
    cp "$README_FILE" "$BUILD_DIR/" || die "Could not copy $README_FILE"
else
    die "$README_FILE not found."
fi

CHECKSUMS_FILE="${BUILD_DIR}/checksums.txt"
: > "$CHECKSUMS_FILE"
if command -v sha256sum >/dev/null 2>&1; then
    ( cd "$BUILD_DIR" && sha256sum pancake-* readme.txt > checksums.txt )
elif command -v shasum >/dev/null 2>&1; then
    ( cd "$BUILD_DIR" && shasum -a 256 pancake-* readme.txt > checksums.txt )
else
    die "Neither sha256sum nor shasum available to write checksums."
fi

log_success "Build complete. Artifacts in $BUILD_DIR:"
ls -1 "$BUILD_DIR"
