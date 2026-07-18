#!/usr/bin/env bash
# Pancake installer for macOS and Linux.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash -s -- uninstall
#   curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash -s -- --version v1.2.0
#
# Flags:
#   uninstall            Remove pancake (binary + config + projects with --purge)
#   --version <tag>      Install a specific release tag (default: latest)
#   --prefix <dir>       Install prefix (default: /usr/local)
#   --purge              During uninstall, also remove ~/pancake.yml and ~/pancake/
#   --no-checksum        Skip SHA-256 checksum verification (not recommended)
#   --yes                Assume yes to prompts (non-interactive)

set -euo pipefail

REPO="github.com/a6h15hek/pancake"
BINARY_NAME="pancake"
DEFAULT_PREFIX="/usr/local"
VERSION_TAG="latest"
PURGE=0
VERIFY_CHECKSUM=1
ASSUME_YES=0
ACTION="install"

COLOR_RED=$'\033[0;31m'
COLOR_GREEN=$'\033[0;32m'
COLOR_BLUE=$'\033[0;34m'
COLOR_YELLOW=$'\033[0;33m'
COLOR_RESET=$'\033[0m'

if [[ ! -t 1 ]]; then
    COLOR_RED=""; COLOR_GREEN=""; COLOR_BLUE=""; COLOR_YELLOW=""; COLOR_RESET=""
fi

log()        { printf '%s[%s]%s %s\n' "$COLOR_BLUE" "$(date +'%H:%M:%S')" "$COLOR_RESET" "$1"; }
log_success() { printf '%s[SUCCESS]%s %s\n' "$COLOR_GREEN" "$COLOR_RESET" "$1"; }
log_warn()    { printf '%s[WARN]%s %s\n' "$COLOR_YELLOW" "$COLOR_RESET" "$1"; }
die()        { printf '%s[ERROR]%s %s\n' "$COLOR_RED" "$COLOR_RESET" "$1" >&2; exit 1; }

TEMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TEMP_DIR"; }
trap cleanup EXIT

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            uninstall)        ACTION="uninstall"; shift ;;
            --version)        VERSION_TAG="${2:-}"; shift 2 ;;
            --prefix)         DEFAULT_PREFIX="${2:-}"; shift 2 ;;
            --purge)          PURGE=1; shift ;;
            --no-checksum)    VERIFY_CHECKSUM=0; shift ;;
            --yes|-y)         ASSUME_YES=1; shift ;;
            --help|-h)
                sed -n '2,12p' "$0" 2>/dev/null || true
                exit 0 ;;
            *) die "Unknown argument: $1 (run with --help)" ;;
        esac
    done
}

confirm() {
    if [[ $ASSUME_YES -eq 1 ]]; then return 0; fi
    if [[ ! -t 0 ]]; then
        log_warn "No TTY attached; assuming 'no'. Re-run in an interactive terminal or pass --yes."
        return 1
    fi
    printf '%s[y/N]: ' "$1"
    read -r reply
    [[ "$reply" =~ ^[Yy]$ ]]
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || die "Required command '$1' not found. Install it and retry."
}

detect_platform() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"
    case "$os" in
        linux|darwin) ;;
        *) die "Unsupported OS: $os. Use windows.ps1 on Windows." ;;
    esac
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) die "Unsupported architecture: $arch (supported: amd64, arm64)." ;;
    esac
    printf '%s-%s' "$os" "$arch"
}

resolve_install_dir() {
    if [[ -w "$DEFAULT_PREFIX/bin" ]]; then
        printf '%s/bin' "$DEFAULT_PREFIX"
        return
    fi
    if [[ $ASSUME_YES -eq 0 && ! -t 0 ]]; then
        die "$DEFAULT_PREFIX/bin is not writable and there is no TTY to prompt. Re-run with sudo or pass --yes."
    fi
    if ! confirm "Need sudo to write to $DEFAULT_PREFIX/bin. Continue with sudo?"; then
        local user_bin="$HOME/.local/bin"
        log_warn "Falling back to $user_bin. Make sure it is on your PATH."
        printf '%s' "$user_bin"
        return
    fi
    printf '%s/bin' "$DEFAULT_PREFIX"
}

download_with_retry() {
    local url="$1" dest="$2" attempts=3
    for ((i=1; i<=attempts; i++)); do
        if curl -fsSL --retry 2 "$url" -o "$dest"; then return 0; fi
        log_warn "Download attempt $i failed: $url"
        sleep 2
    done
    die "Download failed after $attempts attempts: $url
Troubleshooting:
  - Check your internet connection: curl -I https://github.com
  - Check the release exists: https://github.com/a6h15hek/pancake/releases
  - If on a corporate network, set HTTPS_PROXY and retry."
}

verify_checksum() {
    local binary_path="$1" checksum_url="$2"
    if [[ $VERIFY_CHECKSUM -eq 0 ]]; then
        log_warn "Skipping checksum verification (--no-checksum)."
        return 0
    fi
    if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
        log_warn "Neither sha256sum nor shasum available; skipping checksum verification."
        return 0
    fi
    local checksum_file="$TEMP_DIR/checksums.txt"
    if ! curl -fsSL "$checksum_url" -o "$checksum_file"; then
        log_warn "Checksums file unavailable at $checksum_url; skipping verification."
        return 0
    fi
    local binary_name expected
    binary_name="$(basename "$binary_path")"
    expected="$(grep -E "[[:space:]]+${binary_name}\$|[[:space:]]+\\./${binary_name}\$" "$checksum_file" | awk '{print $1}' | head -n1)"
    if [[ -z "$expected" ]]; then
        log_warn "No checksum entry for $binary_name in checksums file; skipping verification."
        return 0
    fi
    local actual
    if command -v sha256sum >/dev/null 2>&1; then
        actual="$(sha256sum "$binary_path" | awk '{print $1}')"
    else
        actual="$(shasum -a 256 "$binary_path" | awk '{print $1}')"
    fi
    if [[ "$actual" != "$expected" ]]; then
        die "Checksum mismatch for $binary_name
  expected: $expected
  actual:   $actual
Re-download, or report at https://github.com/a6h15hek/pancake/issues."
    fi
    log "Checksum verified."
}

install_pancake() {
    require_command curl

    local platform binary_file install_dir install_path
    platform="$(detect_platform)"
    binary_file="${BINARY_NAME}-${platform}"
    install_dir="$(resolve_install_dir)"
    install_path="${install_dir}/${BINARY_NAME}"
    local need_sudo=0
    [[ ! -w "$install_dir" ]] && need_sudo=1

    local version_url download_url checksums_url
    if [[ "$VERSION_TAG" == "latest" ]]; then
        download_url="https://${REPO}/releases/latest/download/${binary_file}"
        checksums_url="https://${REPO}/releases/latest/download/checksums.txt"
    else
        version_url="${VERSION_TAG#v}"
        download_url="https://${REPO}/releases/download/v${version_url}/${binary_file}"
        checksums_url="https://${REPO}/releases/download/v${version_url}/checksums.txt"
    fi

    log "Installing pancake ${VERSION_TAG} for ${platform} into ${install_dir}"
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log "Existing installation found at $(command -v "$BINARY_NAME"); updating."
    fi

    download_with_retry "$download_url" "${TEMP_DIR}/${binary_file}"
    verify_checksum "${TEMP_DIR}/${binary_file}" "$checksums_url"

    if [[ ! -d "$install_dir" ]]; then
        if [[ $need_sudo -eq 1 ]]; then
            sudo mkdir -p "$install_dir" || die "Could not create $install_dir with sudo."
        else
            mkdir -p "$install_dir" || die "Could not create $install_dir."
        fi
    fi

    if [[ $need_sudo -eq 1 ]]; then
        sudo install -m 0755 "${TEMP_DIR}/${binary_file}" "$install_path" || die "Could not move binary to $install_path (sudo)."
    else
        install -m 0755 "${TEMP_DIR}/${binary_file}" "$install_path" || die "Could not move binary to $install_path."
    fi

    verify_install "$install_path" "$install_dir"
}

verify_install() {
    local install_path="$1" install_dir="$2"
    if [[ ":${PATH}:" != *":${install_dir}:"* ]]; then
        log_warn "${install_dir} is not in your PATH."
        printf 'Add this to your shell rc file (~/.bashrc or ~/.zshrc):\n  export PATH="%s:$PATH"\n' "$install_dir"
        printf 'Then reload: source ~/.zshrc\n'
    fi
    if "$install_path" version >/dev/null 2>&1; then
        log_success "Installed pancake $("$install_path" version 2>/dev/null || echo)"
        printf 'Run "pancake help" to get started.\n'
    else
        die "Binary at $install_path is not executable. Check permissions: ls -l $install_path"
    fi
}

uninstall_pancake() {
    local removed=0
    local prefix_bin="${DEFAULT_PREFIX}/bin"
    local candidates=("$prefix_bin/${BINARY_NAME}" "${HOME}/.local/bin/${BINARY_NAME}")
    for candidate in "${candidates[@]}"; do
        if [[ -f "$candidate" ]]; then
            log "Removing $candidate"
            if [[ -w "$(dirname "$candidate")" ]]; then
                rm -f "$candidate" || die "Could not remove $candidate."
            else
                sudo rm -f "$candidate" || die "Could not remove $candidate (sudo)."
            fi
            removed=1
        fi
    done

    if [[ $removed -eq 0 ]]; then
        log "Pancake binary not found in known locations. Already uninstalled."
    else
        log_success "Pancake binary removed."
    fi

    if [[ $PURGE -eq 1 ]]; then
        local config_file="${HOME}/pancake.yml"
        local pancake_home="${HOME}/pancake"
        if [[ -f "$config_file" ]]; then
            rm -f "$config_file" && log "Removed $config_file"
        fi
        if [[ -d "$pancake_home" ]]; then
            if confirm "Also remove pancake project directory ${pancake_home}? This deletes all synced projects."; then
                rm -rf "$pancake_home" && log "Removed $pancake_home"
            else
                log_warn "Kept $pancake_home."
            fi
        fi
    else
        log "Config (~/pancake.yml) and projects (~/pancake/) were left in place. Re-run with --purge to remove them."
    fi
}

main() {
    parse_args "$@"
    case "$ACTION" in
        install)   install_pancake ;;
        uninstall) uninstall_pancake ;;
        *) die "Unknown action: $ACTION" ;;
    esac
}

main "$@"
