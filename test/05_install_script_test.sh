#!/usr/bin/env bash
# 05 — macos_linux.sh install script against a mock GitHub release
# Covers: build artifacts + checksums.txt served over local HTTP, install
# downloads, verifies checksum, installs to a writable prefix, pancake runs.
# Also: unknown flag is rejected, unsupported arch is reported clearly.

set -uo pipefail
source "$(dirname "$0")/helpers.sh"

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
