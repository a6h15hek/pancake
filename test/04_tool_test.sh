#!/usr/bin/env bash
# 04 — tool commands edge cases
# Covers: list empty / populated, install without a package manager installed,
# uninstall of a tool not tracked, search without query, search without package
# manager. These tests stub brew/choco to a missing path so EnsureToolInstalled
# fails with a helpful message (no real brew/choco needed).

set -uo pipefail
source "$(dirname "$0")/helpers.sh"

set_suite "04 tool commands edge cases"

build_pancake >/dev/null || { fail "build pancake"; exit 1; }

# Stub PATH to exclude brew/choco so EnsureToolInstalled fails deterministically.
# We keep go's tmp paths out of the way; pancake binary is already built.
isolated_path() {
    local clean_dir
    clean_dir="$(mktemp_dir pancake_path)"
    # Provide minimal commands pancake's tool path needs: none for these failures.
    echo "$clean_dir"
}

write_config_with_tools() {
    setup_mock_home
    cat > "$MOCK_HOME/pancake.yml" <<YAML
home: \$HOME/pancake
code_editor: echo
default_ai: gemini
tools:
$(for t in "$@"; do echo "  - $t"; done)
projects: {}
YAML
}

# list with no tools -> helpful empty message.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
default_ai: gemini
tools: []
projects: {}
YAML
assert_contains "tool list empty -> helpful message" "No tools" run_pancake tool list
cleanup_mock_home

# list with tools -> names appear.
write_config_with_tools tree jq
assert_contains "tool list shows tree" "tree" run_pancake tool list
assert_contains "tool list shows jq" "jq" run_pancake tool list
cleanup_mock_home

# install with no tool name -> error.
write_config_with_tools
assert_contains "install without name -> error" "missing tool name" run_pancake tool install
cleanup_mock_home

# uninstall of a tool not tracked -> error.
write_config_with_tools tree
assert_contains "uninstall untracked -> error" "not tracked" run_pancake tool uninstall ghost
cleanup_mock_home

# search with no query -> error.
write_config_with_tools tree
assert_contains "search without query -> error" "missing search query" run_pancake tool search
cleanup_mock_home

# tool install / search / update without brew/choco on PATH -> helpful message.
write_config_with_tools tree
CLEAN_PATH="$(isolated_path)"
assert_contains "install without package manager -> helpful" "pancake tool setup" \
    env PATH="$CLEAN_PATH:/usr/bin:/bin" "$PANCAKE_BIN" tool install newtool
cleanup_mock_home

write_config_with_tools tree
CLEAN_PATH="$(isolated_path)"
assert_contains "search without package manager -> helpful" "pancake tool setup" \
    env PATH="$CLEAN_PATH:/usr/bin:/bin" "$PANCAKE_BIN" tool search tree
cleanup_mock_home

# Remove the temp path dirs we created.
rm -rf /tmp/pancake_path_* 2>/dev/null || true

print_summary
RESULT=$?
rm -f /tmp/pancake_test_out
exit $RESULT
