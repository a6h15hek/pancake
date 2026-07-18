#!/usr/bin/env bash
# Legacy quick smoke test. For the full e2e suite, use ./test/run_all.sh
# This script builds, installs, and runs a few basic pancake commands.

set -euo pipefail

check_status() {
    if [ $? -ne 0 ]; then
        echo "$1 failed!"
        exit 1
    fi
}

run_test() {
    echo "-----------------------------------------------------------------------"
    echo "Pancake smoke test: $2..."
    echo "Running command: $1"
    eval "$1"
    check_status "$2"
    echo "-----------------------------------------------------------------------"
}

# Ensure go-installed binaries are on PATH (needed for `pancake` after `go install`).
export PATH="${GOPATH:-$HOME/go}/bin:$PATH"

echo "Pancake smoke test: starting..."

run_test "go build" "Building"
run_test "go install" "Installing"
run_test "pancake version" "Checking version"

# Use an isolated HOME so the developer's real ~/pancake.yml is never touched.
SMOKE_HOME="$(mktemp -d "${TMPDIR:-/tmp}/pancake_smoke.XXXXXX")"
export HOME="$SMOKE_HOME"
trap 'rm -rf "$SMOKE_HOME"' EXIT

run_test "pancake init" "First-time setup"
run_test "pancake edit config" "Opening config file"
run_test "pancake project list" "Listing projects"

echo "Pancake smoke test: done."
echo "For the full e2e suite run: ./test/run_all.sh"
