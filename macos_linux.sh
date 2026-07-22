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
#   --prefix <dir>       Install prefix (default: /usr/local; falls back to ~/.local/bin)
#   --purge              During uninstall, also remove ~/pancake.yml and ~/pancake/
#   --no-checksum        Skip SHA-256 checksum verification (not recommended)
#   --yes                Assume yes to prompts (non-interactive)
#
# Privilege model (in order):
#   1. If the target directory is writable (or you are root), no sudo is used.
#   2. If sudo credentials are already cached (e.g. you just ran sudo), they are reused.
#   3. If a terminal is available, you are asked before sudo is used.
#   4. Otherwise the install falls back to ~/.local/bin — no hard failure.

set -euo pipefail

REPO="github.com/a6h15hek/pancake"
BINARY_NAME="pancake"
PREFIX="/usr/local"
PREFIX_EXPLICIT=0
USER_BIN_DIR="${HOME}/.local/bin"
VERSION_TAG="latest"
PURGE=0
VERIFY_CHECKSUM=1
ASSUME_YES=0
ACTION="install"
SUDO=""            # set to "sudo" only when escalation is both needed and possible
INSTALL_DIR=""

COLOR_RED=$'\033[0;31m'
COLOR_GREEN=$'\033[0;32m'
COLOR_BLUE=$'\033[0;34m'
COLOR_YELLOW=$'\033[0;33m'
COLOR_RESET=$'\033[0m'

# Logs go to stderr: stdout must stay clean so `curl | bash` pipelines and
# command substitution never swallow prompts or corrupt computed paths.
if [[ ! -t 2 ]]; then
    COLOR_RED=""; COLOR_GREEN=""; COLOR_BLUE=""; COLOR_YELLOW=""; COLOR_RESET=""
fi

log()         { printf '%s[%s]%s %s\n' "$COLOR_BLUE" "$(date +'%H:%M:%S')" "$COLOR_RESET" "$1" >&2; }
log_success() { printf '%s[SUCCESS]%s %s\n' "$COLOR_GREEN" "$COLOR_RESET" "$1" >&2; }
log_warn()    { printf '%s[WARN]%s %s\n' "$COLOR_YELLOW" "$COLOR_RESET" "$1" >&2; }
die()         { printf '%s[ERROR]%s %s\n' "$COLOR_RED" "$COLOR_RESET" "$1" >&2; exit 1; }

TEMP_DIR="$(mktemp -d)" || { printf '[ERROR] Could not create a temporary directory.\n' >&2; exit 1; }
cleanup() { rm -rf "$TEMP_DIR"; }
trap cleanup EXIT

usage() {
    cat <<'EOF'
Pancake installer for macOS and Linux.

Usage:
  curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash
  curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash -s -- uninstall

Actions:
  install              Install or update pancake (default)
  uninstall            Remove pancake (add --purge to remove config + projects)

Flags:
  --version <tag>      Install a specific release tag (default: latest)
  --prefix <dir>       Install prefix (default: /usr/local; falls back to ~/.local/bin)
  --purge              During uninstall, also remove ~/pancake.yml and ~/pancake/
  --no-checksum        Skip SHA-256 checksum verification (not recommended)
  --yes, -y            Assume yes to prompts (non-interactive / CI)
  --help, -h           Show this help
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            install)          ACTION="install"; shift ;;
            uninstall)        ACTION="uninstall"; shift ;;
            --version)        [[ -n "${2:-}" ]] || die "--version requires a value, e.g. --version v1.3.0"
                              VERSION_TAG="$2"; shift 2 ;;
            --version=*)      VERSION_TAG="${1#*=}"; shift ;;
            --prefix)         [[ -n "${2:-}" ]] || die "--prefix requires a value, e.g. --prefix /opt"
                              PREFIX="${2%/}"; PREFIX_EXPLICIT=1; shift 2 ;;
            --prefix=*)       PREFIX="${1#*=}"; PREFIX="${PREFIX%/}"; PREFIX_EXPLICIT=1; shift ;;
            --purge)          PURGE=1; shift ;;
            --no-checksum)    VERIFY_CHECKSUM=0; shift ;;
            --yes|-y)         ASSUME_YES=1; shift ;;
            --help|-h)        usage; exit 0 ;;
            *)                die "Unknown argument: $1 (run with --help)" ;;
        esac
    done
    [[ -n "$PREFIX" ]] || die "--prefix cannot be empty."
}

is_root() { [[ "$(id -u)" -eq 0 ]]; }

sudo_available() { command -v sudo >/dev/null 2>&1; }

# True when sudo can run without asking for a password (credentials cached or
# NOPASSWD). Covers `sudo curl ... | bash`, where sudo elevated curl but the
# script itself runs unprivileged with the credential cache still warm.
sudo_cached() { sudo -n true 2>/dev/null; }

# True when we can interact with the user even if stdin is a pipe (curl | bash
# leaves stdin connected to curl, but /dev/tty still reaches the terminal).
has_tty() {
    [[ -t 0 ]] && return 0
    { : </dev/tty; } 2>/dev/null
}

confirm() {
    if [[ $ASSUME_YES -eq 1 ]]; then return 0; fi
    if ! has_tty; then
        log_warn "No terminal available to prompt; assuming 'no'. Pass --yes to skip prompts."
        return 1
    fi
    local reply
    printf '%s [y/N]: ' "$1" >&2
    if [[ -t 0 ]]; then
        read -r reply
    else
        read -r reply </dev/tty
    fi
    [[ "$reply" =~ ^[Yy]([Ee][Ss])?$ ]]
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || die "Required command '$1' not found. Install it and retry."
}

# Writability of a directory that may not exist yet: check the nearest
# existing ancestor (mkdir -p will have to create the rest).
dir_writable() {
    local d="$1"
    while [[ ! -d "$d" ]]; do
        d="$(dirname "$d")"
    done
    [[ -w "$d" ]]
}

run_privileged() {
    if [[ -n "$SUDO" ]]; then sudo "$@"; else "$@"; fi
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
    # Apple Silicon terminals running under Rosetta 2 report x86_64;
    # install the native arm64 build instead.
    if [[ "$os" == "darwin" && "$arch" == "amd64" ]] \
        && [[ "$(sysctl -in sysctl.proc_translated 2>/dev/null || echo 0)" == "1" ]]; then
        arch="arm64"
        log "Rosetta 2 detected; selecting the native arm64 build."
    fi
    printf '%s-%s' "$os" "$arch"
}

# Decide where to install and whether sudo is needed. Sets INSTALL_DIR and
# SUDO globals (no stdout capture — prompts and logs stay visible).
resolve_install_dir() {
    local target="${PREFIX}/bin"
    SUDO=""

    if is_root || dir_writable "$target"; then
        INSTALL_DIR="$target"
        return
    fi

    if sudo_available && sudo_cached; then
        SUDO="sudo"
        INSTALL_DIR="$target"
        log "Using sudo (credentials already cached) to install into ${target}."
        return
    fi

    if sudo_available && has_tty; then
        if confirm "Administrator access is needed to write to ${target}. Use sudo?"; then
            SUDO="sudo"
            INSTALL_DIR="$target"
            return
        fi
    fi

    # Cannot (or should not) escalate. An explicitly requested prefix must not
    # be silently redirected; the default prefix falls back to a per-user dir.
    if [[ $PREFIX_EXPLICIT -eq 1 ]]; then
        die "${target} is not writable and administrator access is unavailable.
Fix one of:
  - Re-run in an interactive terminal so sudo can ask for a password.
  - Run 'sudo -v' first to cache credentials, then re-run this script.
  - Choose a writable prefix: --prefix \$HOME/.local"
    fi
    INSTALL_DIR="$USER_BIN_DIR"
    log_warn "No permission for ${target}; installing to ${INSTALL_DIR} instead (per-user)."
}

download_with_retry() {
    local url="$1" dest="$2" attempts=3
    for ((i=1; i<=attempts; i++)); do
        if curl -fsSL --retry 2 --connect-timeout 15 "$url" -o "$dest"; then return 0; fi
        log_warn "Download attempt $i failed: $url"
        sleep 2
    done
    die "Download failed after $attempts attempts: $url
Troubleshooting:
  - Check your internet connection: curl -I https://github.com
  - Check the release/tag exists: https://github.com/a6h15hek/pancake/releases
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
    if ! curl -fsSL --connect-timeout 15 "$checksum_url" -o "$checksum_file"; then
        log_warn "Checksums file unavailable at $checksum_url; skipping verification."
        return 0
    fi
    local binary_name expected
    binary_name="$(basename "$binary_path")"
    expected="$(grep -E "[[:space:]]+\\*?(\\./)?${binary_name}\$" "$checksum_file" | awk '{print $1}' | head -n1)"
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

# All places this script ever installs to, for the current prefix. Combined
# with a PATH scan, this cleans up leftovers from earlier installs that used
# a different location (the "installed in the past, reinstalled elsewhere"
# scenario).
known_install_paths() {
    printf '%s\n' \
        "${PREFIX}/bin/${BINARY_NAME}" \
        "${HOME}/.local/bin/${BINARY_NAME}" \
        "${HOME}/bin/${BINARY_NAME}" | awk '!seen[$0]++'
}

is_homebrew_managed() {
    local p="$1" resolved="$1"
    if [[ -L "$p" ]]; then
        resolved="$(readlink "$p" 2>/dev/null || printf '%s' "$p")"
    fi
    [[ "$p" == *[Cc]ellar* || "$p" == *homebrew* || "$resolved" == *[Cc]ellar* || "$resolved" == *homebrew* ]]
}

remove_binary_at() {
    local f="$1" d
    d="$(dirname "$f")"
    if [[ -w "$d" ]] || is_root; then
        rm -f "$f" && return 0
    fi
    if sudo_available && { sudo_cached || has_tty; }; then
        sudo rm -f "$f" && return 0
    fi
    return 1
}

# Like confirm, but never auto-approved by --yes: deleting files this run did
# not create requires an explicit interactive yes.
confirm_interactive() {
    if [[ $ASSUME_YES -eq 1 ]] || ! has_tty; then return 1; fi
    confirm "$1"
}

# The classic reinstall bug: an old copy from a previous install lives earlier
# in PATH and keeps shadowing the fresh binary. Find every other copy, remove
# the ones this script owns (with consent), and explain the rest.
handle_stale_copies() {
    local install_path="$1" p
    local -a others=()
    while IFS= read -r p; do
        [[ -n "$p" && "$p" != "$install_path" && -e "$p" ]] && others+=("$p")
    done < <({ type -ap "$BINARY_NAME" 2>/dev/null || true; known_install_paths; } | awk '!seen[$0]++')

    [[ ${#others[@]} -eq 0 ]] && return 0

    for p in "${others[@]}"; do
        if is_homebrew_managed "$p"; then
            log_warn "Another pancake is managed by Homebrew: $p"
            log_warn "If you want only this install, run: brew uninstall pancake"
            continue
        fi
        log_warn "Found another pancake install at ${p} (leftover from a previous install)."
        if confirm_interactive "Remove the old copy at ${p}?"; then
            if remove_binary_at "$p"; then
                log "Removed stale copy: $p"
            else
                log_warn "Could not remove $p (permission denied). Remove it manually: sudo rm $p"
            fi
        else
            log_warn "Kept ${p}. It may shadow the new install if it comes first in PATH. Remove it with: rm $p (or run uninstall)."
        fi
    done

    # After cleanup, make sure `pancake` on PATH is the binary we just installed.
    local active
    active="$(command -v "$BINARY_NAME" 2>/dev/null || true)"
    if [[ -n "$active" && "$active" != "$install_path" && -e "$active" ]]; then
        log_warn "'${BINARY_NAME}' currently resolves to ${active}, not ${install_path}."
        log_warn "Open a new terminal, or remove the other copy, so the new version is used."
    fi
}

path_hint() {
    local dir="$1"
    case "${SHELL:-}" in
        */zsh)
            printf 'Add it to your PATH:\n  echo '\''export PATH="%s:$PATH"'\'' >> ~/.zshrc && source ~/.zshrc\n' "$dir" >&2 ;;
        */fish)
            printf 'Add it to your PATH:\n  fish_add_path %s\n' "$dir" >&2 ;;
        *)
            printf 'Add it to your PATH:\n  echo '\''export PATH="%s:$PATH"'\'' >> ~/.bashrc && source ~/.bashrc\n' "$dir" >&2 ;;
    esac
}

install_pancake() {
    require_command curl

    local platform binary_file install_path
    platform="$(detect_platform)"
    binary_file="${BINARY_NAME}-${platform}"
    resolve_install_dir
    install_path="${INSTALL_DIR}/${BINARY_NAME}"

    local download_url checksums_url version
    if [[ "$VERSION_TAG" == "latest" ]]; then
        download_url="https://${REPO}/releases/latest/download/${binary_file}"
        checksums_url="https://${REPO}/releases/latest/download/checksums.txt"
    else
        version="${VERSION_TAG#v}"
        download_url="https://${REPO}/releases/download/v${version}/${binary_file}"
        checksums_url="https://${REPO}/releases/download/v${version}/checksums.txt"
    fi

    log "Installing pancake ${VERSION_TAG} for ${platform} into ${INSTALL_DIR}"
    if [[ -f "$install_path" ]]; then
        log "Existing installation found at ${install_path}; updating in place."
    fi

    download_with_retry "$download_url" "${TEMP_DIR}/${binary_file}"
    verify_checksum "${TEMP_DIR}/${binary_file}" "$checksums_url"

    if [[ ! -d "$INSTALL_DIR" ]]; then
        run_privileged mkdir -p "$INSTALL_DIR" || die "Could not create $INSTALL_DIR."
    fi

    # install(1) replaces the destination atomically enough for our purposes
    # and unlinks first, so upgrading while pancake is running does not fail
    # with "text file busy" on Linux.
    run_privileged install -m 0755 "${TEMP_DIR}/${binary_file}" "$install_path" \
        || die "Could not install binary to $install_path."

    verify_install "$install_path" "$INSTALL_DIR"
    handle_stale_copies "$install_path"
}

verify_install() {
    local install_path="$1" install_dir="$2"
    if ! "$install_path" version >/dev/null 2>&1; then
        die "Binary at $install_path did not run. Check permissions: ls -l $install_path"
    fi
    log_success "Installed pancake $("$install_path" version 2>/dev/null || echo)"
    if [[ ":${PATH}:" != *":${install_dir}:"* ]]; then
        log_warn "${install_dir} is not in your PATH."
        path_hint "$install_dir"
    fi
    printf 'Run "pancake help" to get started.\n' >&2
}

uninstall_pancake() {
    local removed=0 skipped=0 p
    local -a targets=()
    while IFS= read -r p; do
        [[ -n "$p" && -f "$p" ]] && targets+=("$p")
    done < <({ known_install_paths; type -ap "$BINARY_NAME" 2>/dev/null || true; } | awk '!seen[$0]++')

    # Guard the expansion: bash 3.2 (macOS default) treats an empty array as
    # unbound under set -u.
    [[ ${#targets[@]} -gt 0 ]] || targets=("")
    for p in "${targets[@]}"; do
        [[ -n "$p" ]] || continue
        if is_homebrew_managed "$p"; then
            log_warn "Skipping ${p}: it is managed by Homebrew. Remove it with: brew uninstall pancake"
            skipped=1
            continue
        fi
        log "Removing $p"
        if remove_binary_at "$p"; then
            removed=1
        else
            log_warn "Could not remove ${p} (permission denied)."
            log_warn "Re-run with cached sudo credentials: sudo -v, then re-run uninstall."
            skipped=1
        fi
    done

    if [[ $removed -eq 0 && $skipped -eq 0 ]]; then
        log "Pancake binary not found in known locations. Already uninstalled."
    elif [[ $removed -eq 1 ]]; then
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
