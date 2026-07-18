#!/usr/bin/env bash
# 01 — pancake init / first-time setup
# Covers: fresh init, idempotent init, init --force backup, home dir creation,
# missing TTY does not crash, init refuses to overwrite without --force backup.

set -uo pipefail
source "$(dirname "$0")/helpers.sh"

set_suite "01 pancake init / first-time setup"

build_pancake >/dev/null || { fail "build pancake"; exit 1; }

# Fresh init creates pancake.yml and the home directory.
setup_mock_home
assert_exit_code 0 "init creates config and home" run_pancake init
assert_file_exists "pancake.yml created" "$MOCK_HOME/pancake.yml"
assert_file_exists "pancake home dir created" "$MOCK_HOME/pancake"
assert_file_contains "default config has home field" "$MOCK_HOME/pancake.yml" "home:"
cleanup_mock_home

# init is idempotent: re-running does not error and keeps the existing config.
setup_mock_home
run_pancake init >/dev/null 2>&1
echo "# custom-marker-$(date +%s)" >> "$MOCK_HOME/pancake.yml"
assert_exit_code 0 "init is idempotent" run_pancake init
assert_file_contains "existing config preserved on re-init" "$MOCK_HOME/pancake.yml" "custom-marker-"
cleanup_mock_home

# init --force backs up the old config before resetting.
setup_mock_home
run_pancake init >/dev/null 2>&1
echo "should-be-backed-up" >> "$MOCK_HOME/pancake.yml"
assert_exit_code 0 "init --force succeeds" run_pancake init --force
assert_file_exists "backup created" "$MOCK_HOME/pancake.yml.bak"
assert_file_contains "backup contains old content" "$MOCK_HOME/pancake.yml.bak" "should-be-backed-up"
cleanup_mock_home

# init with no TTY (piped stdin) still works for the non-interactive path.
setup_mock_home
assert_exit_code 0 "init works with piped stdin" bash -c "echo '' | $PANCAKE_BIN init"
assert_file_exists "pancake.yml created via piped stdin" "$MOCK_HOME/pancake.yml"
cleanup_mock_home

# version command works.
assert_exit_code 0 "version subcommand" run_pancake version
assert_contains "version prints Pancake" "Pancake" run_pancake version

# edit config before init gives a clear not-found message.
setup_mock_home
assert_contains "edit config before init is helpful" "pancake.yml does not exist" run_pancake edit config
cleanup_mock_home

print_summary
RESULT=$?
rm -f /tmp/pancake_test_out
exit $RESULT
