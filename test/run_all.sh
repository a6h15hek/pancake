#!/usr/bin/env bash
# Runs every *_test.sh in this directory and aggregates PASS/FAIL.
#
# Usage:
#   ./test/run_all.sh                # run everything
#   ./test/run_all.sh 01 03           # run only suites whose name contains 01 or 03
#
# Exit code is non-zero if any suite fails.

set -uo pipefail

TEST_DIR="$(cd "$(dirname "$0")" && pwd)"
FILTERS=("$@")

COLOR_GREEN=$'\033[0;32m'
COLOR_RED=$'\033[0;31m'
COLOR_BLUE=$'\033[0;34m'
COLOR_RESET=$'\033[0m'
[[ ! -t 1 ]] && { COLOR_GREEN=""; COLOR_RED=""; COLOR_BLUE=""; COLOR_RESET=""; }

TOTAL_PASS=0
TOTAL_FAIL=0
FAILED_SUITES=()

run_suite() {
    local suite="$1" output_file
    output_file="$(mktemp -t pancake_suite_XXXXXX.log)"
    printf '%s>>> %s%s\n' "$COLOR_BLUE" "$suite" "$COLOR_RESET"
    if bash "$suite" >"$output_file" 2>&1; then
        cat "$output_file"
        local pass fail
        pass=$(grep -E '^[[:space:]]*passed:' "$output_file" | awk '{print $2}' | head -n1)
        fail=$(grep -E '^[[:space:]]*failed:' "$output_file" | awk '{print $2}' | head -n1)
        pass=${pass:-0}; fail=${fail:-0}
        TOTAL_PASS=$((TOTAL_PASS + pass))
        TOTAL_FAIL=$((TOTAL_FAIL + fail))
        printf '%s[SUITE OK] %s (%s passed, %s failed)%s\n\n' "$COLOR_GREEN" "$(basename "$suite")" "$pass" "$fail" "$COLOR_RESET"
    else
        cat "$output_file"
        local fail
        fail=$(grep -E '^[[:space:]]*failed:' "$output_file" | awk '{print $2}' | head -n1)
        fail=${fail:-1}
        TOTAL_FAIL=$((TOTAL_FAIL + fail))
        FAILED_SUITES+=("$(basename "$suite")")
        printf '%s[SUITE FAIL] %s%s\n\n' "$COLOR_RED" "$(basename "$suite")" "$COLOR_RESET"
    fi
    rm -f "$output_file"
}

printf '%s=== Pancake e2e test harness ===%s\n' "$COLOR_BLUE" "$COLOR_RESET"

for suite in "$TEST_DIR"/*_test.sh; do
    [[ -f "$suite" ]] || continue
    if [[ ${#FILTERS[@]} -gt 0 ]]; then
        match=0
        for f in "${FILTERS[@]}"; do
            if [[ "$(basename "$suite")" == *"$f"* ]]; then match=1; break; fi
        done
        [[ $match -eq 0 ]] && continue
    fi
    run_suite "$suite"
done

printf '%s=== Summary ===%s\n' "$COLOR_BLUE" "$COLOR_RESET"
printf '  total passed: %s%d%s\n' "$COLOR_GREEN" "$TOTAL_PASS" "$COLOR_RESET"
printf '  total failed: %s%d%s\n' "$COLOR_RED" "$TOTAL_FAIL" "$COLOR_RESET"
if [[ ${#FAILED_SUITES[@]} -gt 0 ]]; then
    printf '  failed suites:\n'
    for s in "${FAILED_SUITES[@]}"; do printf '    - %s\n' "$s"; done
    exit 1
fi
exit 0
