# Pancake Test Harness

End-to-end tests for pancake run the **real binary** inside an isolated mock HOME
("mock PC"), so they never touch the developer's `~/pancake.yml` or `~/pancake/`.

## Layout

```
test/
  helpers.sh                 shared bash helpers (mock HOME, build, assertions, HTTP server)
  run_all.sh                 orchestrator; runs every *_test.sh and reports
  01_init_test.sh            pancake init / init --force / config creation / backup
  02_config_test.sh          config validation, parse errors, missing/relative home
  03_project_test.sh         list / sync / open / build / run / pwd / monitor edge cases
  04_tool_test.sh            tool list / install / uninstall / search edge cases
  05_install_script_test.sh  macos_linux.sh install + uninstall against local HTTP server
  06_uninstall_purge_test.sh uninstall --purge cleans binary + config + projects
```

## Running

```bash
# from repo root
./test/run_all.sh

# single suite
./test/01_init_test.sh
```

## What is covered

- First-time setup: `pancake init` creates `pancake.yml`, creates the home dir, and is
  idempotent. `init --force` backs up the old config.
- Config validation: missing config, unparseable YAML, empty `home`, relative `home`,
  unsupported `default_ai`, project name with `/`, project missing `remote_ssh_url`.
- Project flows: list empty / populated, sync into a non-existent dir, sync refusing to
  clobber a non-git directory, open / build / run / pwd missing-project handling,
  monitor table rendering.
- Tool flows: list empty / populated, install tracking in pancake.yml, uninstall
  removing from pancake.yml, search without a package manager.
- Install script: `macos_linux.sh` downloads from a local HTTP server (mock GitHub
  release), verifies the SHA-256 checksum, installs to a writable prefix, and
  `uninstall --purge` removes binary + config + projects.
- Cross-platform: the e2e harness runs on macOS and Linux. Windows e2e runs in CI via
  PowerShell.

## CI

`.github/workflows/test.yml` runs:
- `go vet ./...`
- `go test ./...` (Go unit tests in `utils/`)
- `./test/run_all.sh` (bash e2e harness) on ubuntu-latest and macos-latest
- `pwsh test/windows_test.ps1` on windows-latest

## Adding a test

1. Pick the matching `NN_*_test.sh` file (or create a new numbered one).
2. `source ./helpers.sh` at the top.
3. Call `setup_mock_home` to get an isolated `$MOCK_HOME`.
4. Use `assert_contains` / `assert_exit_code` / `assert_file_exists` helpers.
5. `./test/run_all.sh` picks up any `*_test.sh` automatically.
