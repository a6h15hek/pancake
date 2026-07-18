#!/usr/bin/env bash
# 03 — project commands edge cases
# Covers: list empty / populated, sync into non-existent dir (mkdir), sync
# refuses to clobber a non-git dir, open / build / run / pwd for missing project,
# monitor table renders, project name with slash is rejected upstream.

set -uo pipefail
source "$(dirname "$0")/helpers.sh"
trap 'cleanup_bare_repo 2>/dev/null; cleanup_mock_home 2>/dev/null' EXIT

set_suite "03 project commands edge cases"

build_pancake >/dev/null || { fail "build pancake"; exit 1; }

write_valid_config() {
    setup_mock_home
    # Use a local bare repo as the remote so sync tests need no network.
    local bare_repo
    bare_repo="$(mktemp_dir pancake_bare)"
    git init -q --bare "$bare_repo" >/dev/null 2>&1
    # Seed an initial commit so git pull works later.
    local seed
    seed="$(mktemp_dir pancake_seed)"
    git -C "$seed" init -q >/dev/null 2>&1
    git -C "$seed" config user.email test@test
    git -C "$seed" config user.name test
    echo "hello" > "$seed/README"
    git -C "$seed" add README >/dev/null 2>&1
    git -C "$seed" commit -q -m init >/dev/null 2>&1
    git -C "$seed" remote add origin "$bare_repo" >/dev/null 2>&1
    git -C "$seed" push -q origin master:master >/dev/null 2>&1 || \
        git -C "$seed" push -q origin main:main >/dev/null 2>&1 || true
    MOCK_BARE_REPO="$bare_repo"
    cat > "$MOCK_HOME/pancake.yml" <<YAML
home: \$HOME/pancake
code_editor: echo
default_ai: gemini
tools: []
projects:
  demo:
    remote_ssh_url: $bare_repo
    run: echo running
    build: echo building
  webapp:
    remote_ssh_url: $bare_repo
    type: web
    port: "3000"
    run: echo web
    build: echo webbuild
YAML
}

cleanup_bare_repo() {
    [[ -n "${MOCK_BARE_REPO:-}" ]] && rm -rf "$MOCK_BARE_REPO"
    MOCK_BARE_REPO=""
}

# Empty projects list -> helpful empty message.
setup_mock_home
cat > "$MOCK_HOME/pancake.yml" <<'YAML'
home: $HOME/pancake
code_editor: echo
default_ai: gemini
tools: []
projects: {}
YAML
assert_contains "empty projects -> helpful message" "No projects" run_pancake project list
cleanup_mock_home

# Populated projects list -> both names appear.
write_valid_config
assert_contains "populated list shows demo" "demo" run_pancake project list
assert_contains "populated list shows webapp" "webapp" run_pancake project list
cleanup_mock_home

# sync of a missing project -> not found message (no crash).
write_valid_config
assert_contains "sync missing project -> not found" "not found" run_pancake project sync does-not-exist
cleanup_mock_home

# sync of a real project into an empty home -> clones (mkdir parent).
write_valid_config
assert_exit_code 0 "sync real project clones" run_pancake project sync demo
assert_file_exists "cloned project dir created" "$MOCK_HOME/pancake/demo/.git"
cleanup_mock_home

# sync refuses to clobber a non-git directory (user data protection).
write_valid_config
mkdir -p "$MOCK_HOME/pancake/demo"
echo "important-user-data" > "$MOCK_HOME/pancake/demo/important.txt"
assert_exit_code 0 "sync refuses to clobber (exit 0 from pancake, git fails)" run_pancake project sync demo
assert_file_exists "user data preserved during sync" "$MOCK_HOME/pancake/demo/important.txt"
cleanup_mock_home

# build of a missing project -> not found.
write_valid_config
assert_contains "build missing -> not found" "not found" run_pancake project build ghost
cleanup_mock_home

# build of a project whose directory does not exist -> sync hint.
write_valid_config
assert_contains "build without dir -> sync hint" "pancake sync" run_pancake project build demo
cleanup_mock_home

# build after sync -> succeeds.
write_valid_config
run_pancake project sync demo >/dev/null 2>&1
assert_exit_code 0 "build after sync succeeds" run_pancake project build demo
cleanup_mock_home

# pwd of a missing project -> not found.
write_valid_config
assert_contains "pwd missing -> not found" "not found" run_pancake project pwd ghost
cleanup_mock_home

# pwd of a real project -> path printed.
write_valid_config
assert_contains "pwd prints project path" "$MOCK_HOME/pancake/demo" run_pancake project pwd demo
cleanup_mock_home

# monitor renders a table with both projects.
write_valid_config
assert_contains "monitor shows demo row" "demo" run_pancake project monitor
assert_contains "monitor shows webapp row" "webapp" run_pancake project monitor
assert_contains "monitor shows port 3000" "3000" run_pancake project monitor
cleanup_mock_home

# open of a missing project -> not found (does not launch editor).
write_valid_config
assert_contains "open missing -> not found" "not found" run_pancake project open ghost
cleanup_mock_home

print_summary
RESULT=$?
rm -f /tmp/pancake_test_out
exit $RESULT
