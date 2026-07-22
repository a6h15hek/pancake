#!/usr/bin/env bash
# 06 — uninstall and --purge
# Covers: uninstall removes the binary, uninstall when not installed is a no-op
# message, --purge removes binary + pancake.yml + ~/pancake/, and uninstall
# without --purge keeps config and projects.

set -uo pipefail
source "$(dirname "$0")/helpers.sh"

# The installer can prompt via /dev/tty; keep the suite non-interactive.
exec </dev/null

set_suite "06 uninstall & --purge"

build_pancake >/dev/null || { fail "build pancake"; exit 1; }

# Stage a release mirroring GitHub's URL layout.
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
    return 1
}
start_mock_release_server_with_root || { fail "start mock server"; exit 1; }

PATCHED_SCRIPT="$(mktemp_file pancake_install).sh"
sed -E \
    -e "s#^REPO=\"github.com/a6h15hek/pancake\"#REPO=\"127.0.0.1:${MOCK_SERVER_PORT}\"#" \
    -e "s#https://\\\$\{REPO\}/releases#http://\${REPO}/releases#g" \
    "$REPO_ROOT/macos_linux.sh" > "$PATCHED_SCRIPT"
chmod +x "$PATCHED_SCRIPT"

INSTALL_PREFIX="$(mktemp_dir pancake_prefix)"
mkdir -p "$INSTALL_PREFIX/bin"

# Install + run init so we have config + projects on disk.
bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes >/tmp/pancake_install.log 2>&1 || { cat /tmp/pancake_install.log; fail "install step failed"; exit 1; }
setup_mock_home
# Run init with stdin piped so ConfirmAction (homebrew) returns no and init continues.
echo "" | env HOME="$MOCK_HOME" "$INSTALL_PREFIX/bin/pancake" init >/tmp/pancake_init.log 2>&1 || true
assert_file_exists "config exists before purge" "$MOCK_HOME/pancake.yml"

# Ensure the pancake home dir + a project file exist (init may have aborted home
# creation if tool setup errored before mkdir; create them explicitly).
mkdir -p "$MOCK_HOME/pancake/demo"
echo "project-data" > "$MOCK_HOME/pancake/demo/file.txt"

# Patched uninstall needs HOME set to our mock home and prefix to our install dir.
assert_exit_code 0 "uninstall --purge succeeds" \
    env HOME="$MOCK_HOME" bash "$PATCHED_SCRIPT" uninstall --prefix "$INSTALL_PREFIX" --yes --purge

assert_file_missing "binary removed by purge" "$INSTALL_PREFIX/bin/pancake"
assert_file_missing "config removed by purge" "$MOCK_HOME/pancake.yml"
assert_file_missing "projects removed by purge" "$MOCK_HOME/pancake/demo/file.txt"

# Reinstall then uninstall WITHOUT --purge keeps config and projects.
bash "$PATCHED_SCRIPT" --prefix "$INSTALL_PREFIX" --yes >/tmp/pancake_install2.log 2>&1 || { cat /tmp/pancake_install2.log; fail "reinstall step failed"; exit 1; }
setup_mock_home_2="$(mktemp_dir pancake_home2)"
echo "" | env HOME="$setup_mock_home_2" "$INSTALL_PREFIX/bin/pancake" init >/tmp/pancake_init2.log 2>&1 || true
mkdir -p "$setup_mock_home_2/pancake/demo"
echo "kept" > "$setup_mock_home_2/pancake/demo/file.txt"
assert_file_exists "config exists before non-purge uninstall" "$setup_mock_home_2/pancake.yml"

assert_exit_code 0 "uninstall without purge succeeds" \
    env HOME="$setup_mock_home_2" bash "$PATCHED_SCRIPT" uninstall --prefix "$INSTALL_PREFIX" --yes

assert_file_missing "binary removed without purge" "$INSTALL_PREFIX/bin/pancake"
assert_file_exists "config kept without purge" "$setup_mock_home_2/pancake.yml"
assert_file_exists "projects kept without purge" "$setup_mock_home_2/pancake/demo/file.txt"

# Uninstall when not installed is graceful.
rm -f "$INSTALL_PREFIX/bin/pancake"
assert_contains "uninstall when absent is graceful" "not found" \
    env HOME="$setup_mock_home_2" bash "$PATCHED_SCRIPT" uninstall --prefix "$INSTALL_PREFIX" --yes

# Reinstalled-across-locations scenario: copies exist in BOTH the prefix and
# ~/.local/bin (a past install used a different location). Uninstall must
# remove every copy, not just the current prefix.
multi_home="$(mktemp_dir pancake_home3)"
mkdir -p "$multi_home/.local/bin" "$INSTALL_PREFIX/bin"
cp "$PANCAKE_BIN" "$INSTALL_PREFIX/bin/pancake"
cp "$PANCAKE_BIN" "$multi_home/.local/bin/pancake"
assert_exit_code 0 "uninstall with copies in two locations succeeds" \
    env HOME="$multi_home" bash "$PATCHED_SCRIPT" uninstall --prefix "$INSTALL_PREFIX" --yes
assert_file_missing "prefix copy removed" "$INSTALL_PREFIX/bin/pancake"
assert_file_missing "per-user copy removed too" "$multi_home/.local/bin/pancake"
rm -rf "$multi_home"

stop_mock_release_server
rm -rf "$STAGE_DIR" "$PATCHED_SCRIPT" "$INSTALL_PREFIX" "$MOCK_HOME" "$setup_mock_home_2"
print_summary
RESULT=$?
rm -f /tmp/pancake_test_out
exit $RESULT
