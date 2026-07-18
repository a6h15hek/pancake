#!/usr/bin/env bash
# 02 — config validation & troubleshooting messages
# Covers: missing config, unparseable YAML, empty home, relative home, bad
# default_ai, unsafe project name, project missing remote_ssh_url. Every failure
# must point the developer to fix pancake.yml (not pancake itself).

set -uo pipefail
source "$(dirname "$0")/helpers.sh"

set_suite "02 config validation & troubleshooting"

build_pancake >/dev/null || { fail "build pancake"; exit 1; }

# No config at all -> tells user to run pancake init.
setup_mock_home
rm -f "$MOCK_HOME/pancake.yml"
assert_contains "no config -> mentions pancake init" "pancake init" run_pancake project list
assert_contains "no config -> mentions pancake.yml" "pancake.yml" run_pancake project list
cleanup_mock_home

# Unparseable YAML -> parse error + edit-config hint.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
  bad: : :
:invalid
YAML
assert_contains "bad yaml -> parse error" "not valid YAML" run_pancake project list
assert_contains "bad yaml -> edit config hint" "pancake edit config" run_pancake project list
cleanup_mock_home

# Empty home field.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home:
code_editor: echo
default_ai: gemini
projects: {}
YAML
assert_contains "empty home -> clear message" "'home' is empty" run_pancake project list
cleanup_mock_home

# Relative home field.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: relative/path
code_editor: echo
default_ai: gemini
projects: {}
YAML
assert_contains "relative home -> clear message" "not an absolute path" run_pancake project list
cleanup_mock_home

# Unsupported default_ai.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
default_ai: claude
projects: {}
YAML
assert_contains "bad default_ai -> mentions field" "'default_ai'" run_pancake project list
assert_contains "bad default_ai -> lists allowed" "gemini" run_pancake project list
cleanup_mock_home

# Unsafe project name (contains /).
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
default_ai: gemini
projects:
  bad/name:
    remote_ssh_url: git@github.com:org/repo.git
YAML
assert_contains "unsafe project name -> mentions path separators" "path separators" run_pancake project list
cleanup_mock_home

# Project missing remote_ssh_url.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
default_ai: gemini
projects:
  demo:
    run: echo hi
YAML
assert_contains "missing remote -> mentions remote_ssh_url" "remote_ssh_url" run_pancake project list
cleanup_mock_home

# Valid config loads fine.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
default_ai: gemini
tools: []
projects:
  demo:
    remote_ssh_url: git@github.com:org/repo.git
YAML
assert_exit_code 0 "valid config loads" run_pancake project list
assert_contains "valid config lists project" "demo" run_pancake project list
cleanup_mock_home

print_summary
RESULT=$?
rm -f /tmp/pancake_test_out
exit $RESULT
