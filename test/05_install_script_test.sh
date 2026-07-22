#!/usr/bin/env bash
# 05 — macos_linux.sh install script against a mock GitHub release
# Covers: build artifacts + checksums.txt served over local HTTP, install
# downloads, verifies checksum, installs to a writable prefix, pancake runs.
# Also: unknown flag is rejected, unsupported arch is reported clearly.

set -uo pipefail
source "$(dirname "$0")/helpers.sh"

# The installer can prompt via /dev/tty; keep the suite non-interactive.
exec </dev/null

set_suite "05 macos_linux.sh install (mock release server)"

build_pancake >/dev/null || { fail "build pancake"; exit 1; }

# Stage a release directory that mirrors GitHub's URL layout:
#   http://server/releases/latest/download/<artifact>
#   http://server/releases/latest/download/checksums.txt
STAGE_DIR="$(mktemp_dir pancake_release)"
RELEASE_DIR="$STAGE_DIR/releases/latest/download"
mkdir -p "$RELEASE_DIR"
goos="$(uname -s | tr '[:upper:]' '[:lower:]')"
goarch="$(uname -m)"
case "$goarch" in
    x86_64|amd64) goarch="amd64" ;;
    arm64|aarch64) goarch="arm64" ;;
esac
BINARY_ARTIFACT="pancake-${goos}-${goarch}"
cp "$PANCAKE_BIN" "$RELEASE_DIR/$BINARY_ARTIFACT"
( cd "$RELEASE_DIR" && shasum -a 256 "$BINARY_ARTIFACT" > checksums.txt )

# Point the installer at our mock server by rewriting the repo URL via env.
# macos_linux.sh hardcodes REPO=github.com/...; we vendor a patched copy.
INSTALL_PREFIX="$(mktemp_dir pancake_prefix)"
INSTALL_BIN_DIR="$INSTALL_PREFIX/bin"
mkdir -p "$INSTALL_BIN_DIR"

start_mock_release_server_with_root() {
    MOCK_SERVER_ROOT="$STAGE_DIR"
    local port attempts
    for attempts in 1 2 3 4 5; do
        port=$(( (RANDOM % 10000) + 20000 ))
        if python3 -m http.server "$port" --bind 127.0.0.1 --directory "$STAGE_DIR" >/tmp/pancake_mock_server.log 2>&1 &
        then
            MOCK_SERVER_PID=$!
            sleep 0.5
            if kill -0 "$MOCK_SERVER_PID" 2>/dev/null; then
                MOCK_SERVER_PORT="$port"
                return 0
            fi
        fi
    done
    echo "ERROR: could not start mock release server" >&2
    return 1
}

start_mock_release_server_with_root || { fail "start mock server"; exit 1; }

# Build a patched copy of macos_linux.sh that points at our local server.
PATCHED_SCRIPT="$(mktemp_file pancake_install).sh"
sed -E \
    -e "s#^REPO=\"github.com/a6h15hek/pancake\"#REPO=\"127.0.0.1:${MOCK_SERVER_PORT}\"#" \
    -e "s#https://\\\$\{REPO\}/releases#http://\${REPO}/releases#g" \
    "$REPO_ROOT/macos_linux.sh" > "$PATCHED_SCRIPT"
chmod +x "$PATCHED_SCRIPT"

# Sanity: the artifact is reachable at the expected URL.
assert_exit_code 0 "mock release URL is reachable" \
    curl -fsSL "http://127.0.0.1:${MOCK_SERVER_PORT}/releases/latest/download/$BINARY_ARTIFACT" -o /dev/null

# Install into our writable prefix (no sudo needed).
assert_exit_code 0 "install script succeeds against mock release" \
    bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes

assert_file_exists "binary installed to prefix" "$INSTALL_BIN_DIR/pancake"
assert_exit_code 0 "installed pancake runs version" "$INSTALL_BIN_DIR/pancake" version
assert_contains "installed pancake prints version" "Pancake" "$INSTALL_BIN_DIR/pancake" version

# Checksum mismatch is detected: tamper the artifact but LEAVE the original
# checksums.txt in place, so the stored hash no longer matches the file.
echo "tampered-content" > "$RELEASE_DIR/$BINARY_ARTIFACT"
rm -f "$INSTALL_BIN_DIR/pancake"
assert_contains "checksum mismatch -> clear error" "Checksum mismatch" \
    bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes

# Restore valid artifact (and regenerate matching checksums) for the next tests.
cp "$PANCAKE_BIN" "$RELEASE_DIR/$BINARY_ARTIFACT"
( cd "$RELEASE_DIR" && shasum -a 256 "$BINARY_ARTIFACT" > checksums.txt )

# --no-checksum skips verification and still installs.
rm -f "$INSTALL_BIN_DIR/pancake"
assert_exit_code 0 "install with --no-checksum succeeds" \
    bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes --no-checksum
assert_file_exists "binary installed without checksum" "$INSTALL_BIN_DIR/pancake"

# Unknown flag is rejected.
assert_contains "unknown flag rejected" "Unknown argument" \
    bash "$PATCHED_SCRIPT" --bogus-flag

# Flags that need a value fail clearly when the value is missing.
assert_contains "--version without value rejected" "requires a value" \
    bash "$PATCHED_SCRIPT" --version
assert_contains "--prefix without value rejected" "requires a value" \
    bash "$PATCHED_SCRIPT" --prefix

# Reinstall scenario: a leftover copy from a previous install elsewhere is
# detected and reported, but never deleted without an interactive yes.
STALE_HOME="$(mktemp_dir pancake_stale_home)"
mkdir -p "$STALE_HOME/.local/bin"
cp "$PANCAKE_BIN" "$STALE_HOME/.local/bin/pancake"
rm -f "$INSTALL_BIN_DIR/pancake"
assert_contains "stale copy from previous install is reported" "Found another pancake install" \
    env HOME="$STALE_HOME" bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes
assert_file_exists "stale copy is NOT auto-deleted without interactive consent" "$STALE_HOME/.local/bin/pancake"
assert_file_exists "fresh binary still installed alongside" "$INSTALL_BIN_DIR/pancake"
rm -rf "$STALE_HOME"

# Privilege edge cases only behave distinctly for non-root users (root can
# write anywhere, so the fallback path never triggers). The installer prompts
# via /dev/tty even when stdin is piped, so these no-TTY tests must run
# without a controlling terminal: detach with setsid where available, and
# skip on interactive terminals without it (CI has no TTY, so CI always runs
# them).
run_detached() {
    if command -v setsid >/dev/null 2>&1; then setsid -w "$@"; else "$@"; fi
}
HAVE_TTY=0
{ : </dev/tty; } 2>/dev/null && HAVE_TTY=1
if [[ "$HAVE_TTY" -eq 1 ]] && ! command -v setsid >/dev/null 2>&1; then
    echo "  SKIP  no-TTY fallback tests (interactive terminal and no setsid)"
elif [[ "$(id -u)" -ne 0 ]]; then
    # Simulate the `curl | bash` failure mode: default prefix not writable,
    # sudo unusable, no TTY -> installer must fall back to ~/.local/bin
    # instead of dying.
    RO_PREFIX="$(mktemp_dir pancake_ro_prefix)"
    mkdir -p "$RO_PREFIX/bin"
    chmod 555 "$RO_PREFIX/bin" "$RO_PREFIX"
    FALLBACK_HOME="$(mktemp_dir pancake_fb_home)"
    NOPRIV_SCRIPT="$(mktemp_file pancake_nopriv).sh"
    # Point the default prefix at the read-only dir and neutralize sudo.
    sed -E "s#^PREFIX=\"/usr/local\"#PREFIX=\"$RO_PREFIX\"#" "$PATCHED_SCRIPT" > "$NOPRIV_SCRIPT"
    NOPRIV_RUNNER="$(mktemp_file pancake_nopriv_run).sh"
    cat > "$NOPRIV_RUNNER" <<RUN_EOF
#!/usr/bin/env bash
sudo() { return 1; }
export -f sudo
exec bash "$NOPRIV_SCRIPT" "\$@" </dev/null
RUN_EOF
    chmod +x "$NOPRIV_RUNNER"

    assert_exit_code 0 "no-TTY + unwritable prefix falls back to ~/.local/bin" \
        run_detached env HOME="$FALLBACK_HOME" bash "$NOPRIV_RUNNER"
    assert_file_exists "binary installed to per-user fallback" "$FALLBACK_HOME/.local/bin/pancake"

    # An EXPLICIT prefix must never be silently redirected: fail loudly.
    assert_contains "explicit unwritable prefix fails with guidance" "not writable" \
        run_detached env HOME="$FALLBACK_HOME" bash "$NOPRIV_RUNNER" --prefix "$RO_PREFIX"

    chmod 755 "$RO_PREFIX" "$RO_PREFIX/bin"
    rm -rf "$RO_PREFIX" "$FALLBACK_HOME" "$NOPRIV_SCRIPT" "$NOPRIV_RUNNER"
else
    echo "  SKIP  privilege fallback tests (running as root)"
fi

# --help lists the flags (covers the help/usage path).
assert_contains "help lists --version flag" "--version" \
    bash "$PATCHED_SCRIPT" --help
assert_contains "help lists --purge flag" "--purge" \
    bash "$PATCHED_SCRIPT" --help
assert_contains "help lists uninstall action" "uninstall" \
    bash "$PATCHED_SCRIPT" --help

# Unsupported OS is reported clearly: override uname -s via a function in a
# single subshell so detect_platform sees "solaris" and rejects it.
UNSUPPORTED_OS_RUNNER="$(mktemp_file pancake_os_test).sh"
cat > "$UNSUPPORTED_OS_RUNNER" <<OS_EOF
#!/usr/bin/env bash
uname() {
    if [[ "\$1" == "-s" ]]; then
        echo "solaris"
        return 0
    fi
    /usr/bin/uname "\$@"
}
export -f uname
exec bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes
OS_EOF
chmod +x "$UNSUPPORTED_OS_RUNNER"
assert_contains "unsupported OS reported" "Unsupported OS" bash "$UNSUPPORTED_OS_RUNNER"
rm -f "$UNSUPPORTED_OS_RUNNER"

stop_mock_release_server
rm -rf "$STAGE_DIR" "$PATCHED_SCRIPT" "$INSTALL_PREFIX"
print_summary
RESULT=$?
rm -f /tmp/pancake_test_out
exit $RESULT
