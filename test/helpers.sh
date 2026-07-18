#!/usr/bin/env bash
# Shared helpers for the pancake e2e test harness.
#
# Each test sources this file. It provides:
#   - An isolated MOCK_HOME so the real ~/pancake.yml is never touched.
#   - A built pancake binary at $PANCAKE_BIN.
#   - A local HTTP server mocking GitHub releases (for install-script tests).
#   - Assertion helpers that record PASS/FAIL and let run_all.sh collect results.

set -uo pipefail

TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$TEST_DIR/.." && pwd)"

mktemp_dir() {
    local tmp="${TMPDIR:-/tmp}"
    tmp="${tmp%/}"
    mktemp -d "${tmp}/${1:-pancake}.XXXXXX"
}

mktemp_file() {
    local tmp="${TMPDIR:-/tmp}"
    tmp="${tmp%/}"
    mktemp "${tmp}/${1:-pancake}.XXXXXX"
}

COLOR_GREEN=$'\033[0;32m'
COLOR_RED=$'\033[0;31m'
COLOR_YELLOW=$'\033[0;33m'
COLOR_BLUE=$'\033[0;34m'
COLOR_RESET=$'\033[0m'
[[ ! -t 1 ]] && { COLOR_GREEN=""; COLOR_RED=""; COLOR_YELLOW=""; COLOR_BLUE=""; COLOR_RESET=""; }

PASS_COUNT=0
FAIL_COUNT=0
CURRENT_SUITE=""

set_suite() {
    CURRENT_SUITE="$1"
    printf '\n%s=== %s ===%s\n' "$COLOR_BLUE" "$CURRENT_SUITE" "$COLOR_RESET"
}

pass() {
    PASS_COUNT=$((PASS_COUNT + 1))
    printf '  %sPASS%s  %s\n' "$COLOR_GREEN" "$COLOR_RESET" "$1"
}

fail() {
    FAIL_COUNT=$((FAIL_COUNT + 1))
    printf '  %sFAIL%s  %s\n' "$COLOR_RED" "$COLOR_RESET" "$1"
    if [[ -n "${2:-}" ]]; then
        printf '        %sexpected:%s %s\n' "$COLOR_YELLOW" "$COLOR_RESET" "$2"
        printf '        %sactual:%s   %s\n' "$COLOR_YELLOW" "$COLOR_RESET" "$3"
    fi
}

assert_exit_code() {
    local expected="$1" label="$2" actual
    shift 2
    "$@" >/tmp/pancake_test_out 2>&1
    actual=$?
    if [[ "$actual" -eq "$expected" ]]; then
        pass "$label (exit $actual)"
    else
        fail "$label" "exit $expected" "exit $actual"
        sed 's/^/        | /' /tmp/pancake_test_out >&2
    fi
}

assert_contains() {
    local label="$1" needle="$2"
    shift 2
    "$@" >/tmp/pancake_test_out 2>&1
    if grep -qF -- "$needle" /tmp/pancake_test_out; then
        pass "$label"
    else
        fail "$label" "output contains '$needle'" "$(cat /tmp/pancake_test_out)"
    fi
}

assert_file_exists() {
    local label="$1" path="$2"
    if [[ -e "$path" ]]; then
        pass "$label ($path exists)"
    else
        fail "$label" "file/dir at $path" "missing"
    fi
}

assert_file_missing() {
    local label="$1" path="$2"
    if [[ ! -e "$path" ]]; then
        pass "$label ($path absent)"
    else
        fail "$label" "no file at $path" "still present"
    fi
}

assert_file_contains() {
    local label="$1" path="$2" needle="$3"
    if [[ -f "$path" ]] && grep -qF -- "$needle" "$path"; then
        pass "$label"
    else
        fail "$label" "$path contains '$needle'" "missing or no match"
    fi
}

setup_mock_home() {
    MOCK_HOME="$(mktemp_dir pancake_home)"
    export HOME="$MOCK_HOME"
    export USERPROFILE="$MOCK_HOME"
    export XDG_CONFIG_HOME="$MOCK_HOME/.config"
    mkdir -p "$MOCK_HOME"
}

build_pancake() {
    local build_out
    build_out="$(mktemp_dir pancake_build)"
    local goos goarch output
    goos="$(uname -s | tr '[:upper:]' '[:lower:]')"
    goarch="$(uname -m)"
    case "$goarch" in
        x86_64|amd64) goarch="amd64" ;;
        arm64|aarch64) goarch="arm64" ;;
    esac
    output="${build_out}/pancake-${goos}-${goarch}"
    ( cd "$REPO_ROOT" && GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
        go build -trimpath -o "$output" . ) || return 1
    PANCAKE_BIN="$output"
    PANCAKE_BUILD_DIR="$build_out"
    printf '%s\n' "$output"
}

run_pancake() {
    "$PANCAKE_BIN" "$@"
}

start_mock_release_server() {
    # Serves $PANCAKE_BUILD_DIR over HTTP, mimicking a GitHub release directory.
    MOCK_SERVER_ROOT="$PANCAKE_BUILD_DIR"
    MOCK_SERVER_PORT=""
    local port attempts=0
    for attempts in 1 2 3 4 5; do
        port=$(( (RANDOM % 10000) + 20000 ))
        if python3 -m http.server "$port" --bind 127.0.0.1 --directory "$MOCK_SERVER_ROOT" >/tmp/pancake_mock_server.log 2>&1 &
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

stop_mock_release_server() {
    if [[ -n "${MOCK_SERVER_PID:-}" ]] && kill -0 "$MOCK_SERVER_PID" 2>/dev/null; then
        kill "$MOCK_SERVER_PID" 2>/dev/null || true
        wait "$MOCK_SERVER_PID" 2>/dev/null || true
    fi
}

cleanup_mock_home() {
    if [[ -n "${MOCK_HOME:-}" ]] && [[ -d "$MOCK_HOME" ]]; then
        rm -rf "$MOCK_HOME"
    fi
    stop_mock_release_server
}

cleanup_build() {
    if [[ -n "${PANCAKE_BUILD_DIR:-}" ]] && [[ -d "$PANCAKE_BUILD_DIR" ]]; then
        rm -rf "$PANCAKE_BUILD_DIR"
    fi
    stop_mock_release_server
}

print_summary() {
    local total=$((PASS_COUNT + FAIL_COUNT))
    printf '\n%s--- %s summary ---%s\n' "$COLOR_BLUE" "$CURRENT_SUITE" "$COLOR_RESET"
    printf '  passed: %s%d%s / %d\n' "$COLOR_GREEN" "$PASS_COUNT" "$COLOR_RESET" "$total"
    printf '  failed: %s%d%s / %d\n' "$COLOR_RED" "$FAIL_COUNT" "$COLOR_RESET" "$total"
    if [[ "$FAIL_COUNT" -gt 0 ]]; then return 1; fi
    return 0
}
