# Pancake: Developer Command Line Tool

![GitHub release (latest by date)](https://img.shields.io/github/v/release/a6h15hek/pancake)
![GitHub](https://img.shields.io/github/license/a6h15hek/pancake)

Streamlines your project management workflow, simplifies syncing multiple projects from remote repositories, helps build and run projects on a local machine, and allows for the installation, uninstallation, and updating of development software and tools. No need to remember longer commands. Share your entire development setup with your work buddies or back up your entire development setup by sharing `$HOME/pancake.yml`.
![Pancake Project Developer Command Line Tool](https://github.com/user-attachments/assets/0d0fa2df-f997-4ba8-b65a-b2fb4337bd65)
![Pancake AI Project Developer Command Line Tool](https://github.com/user-attachments/assets/7deb4539-d14b-4c31-8e1f-976512bfe6c9)

```sh
$ pancake list                  # View the list of all projects you are working on
$ pancake sync                  # Sync multiple projects from the remote repository 
$ pancake open <project_name>   # Open a specific project in IDE mentioned in config file
$ pancake build <project_name>  # Build a project
$ pancake run <project_name>    # Start a project on the local machine
$ pancake pwd <project_name>    # Shows and copies the project path to clipboard  
$ pancake monitor               # Show details about the port, process PID, and uptime
$ pancake ai "find all files larger than 10MB in my home directory" # generate command out of natural language

$ pancake tool install tree     # Install a tool via pancake 
$ pancake tool upgrade tree     # Update tools via pancake
```
It keeps all project files in the `$HOME/pancake` folder. Sharing this single file enables the sharing of your entire developer setup, making backups and migration from one machine to another easy.

For migration and sharing projects, just copy pancake.yml to your user home location and run:
```bash
$ pancake init
```
Everything your project needs will be installed. All the build and run configurations will already be in the configuration file, and it will sync the project from your remote repository all at once.

## Installation

### macOS & Linux
To install or update the tool on macOS or Linux, run the following command in your terminal (no `sudo` needed — the script asks only if it has to):
```bash
curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash
```
To uninstall the tool (also removes config and projects):
```bash
curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh | bash -s -- uninstall --purge
```
Supported flags:
- `--version <tag>`  Install a specific release tag (default: `latest`).
- `--prefix <dir>`   Install prefix (default: `/usr/local`; falls back to `~/.local/bin`).
- `--purge`          During uninstall, also remove `~/pancake.yml` and `~/pancake/`.
- `--no-checksum`    Skip SHA-256 checksum verification (not recommended).
- `--yes`            Assume yes to prompts (non-interactive / CI).

Native binaries are provided for `amd64` and `arm64` (Apple Silicon, Linux arm64). Apple Silicon terminals running under Rosetta 2 automatically get the native `arm64` build.

How the installer handles permissions (in order):
1. If `/usr/local/bin` is writable (or you are root), it installs directly.
2. If your sudo credentials are already cached, it reuses them.
3. If a terminal is available, it asks before using sudo — even when piped through `curl | bash`.
4. Otherwise it installs to `~/.local/bin` (per-user, no privileges needed) and prints the PATH line to add.

Reinstalling and upgrading:
- Re-running the install command upgrades in place — even while pancake is running.
- If an older install exists at a different location (e.g. a past install in `~/.local/bin` and a new one in `/usr/local/bin`), the installer detects it, warns that it may shadow the new binary, and offers to remove it.
- `uninstall` removes pancake from every known location (`<prefix>/bin`, `~/.local/bin`, `~/bin`, and anything else on your PATH), and points you at `brew uninstall pancake` if a copy is managed by Homebrew.

### Windows
To install or update the tool on Windows, run the following in PowerShell (admin is optional — without admin it installs per-user to `%LOCALAPPDATA%\Programs\Pancake`):
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;
$script = (New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/a6h15hek/pancake/main/windows.ps1');
& ([scriptblock]::Create($script)) -Action install
```
To uninstall (also removes config and projects):
```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;
$script = (New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/a6h15hek/pancake/main/windows.ps1');
& ([scriptblock]::Create($script)) -Action uninstall -Purge
```
Supported flags: `-Version <tag>`, `-Purge`, `-NoChecksum`, `-Force` (use per-user location even as admin), `-Yes` (assume yes to prompts / CI).

Notes:
- Both `amd64` and `arm64` Windows are supported; the script picks the right binary automatically.
- The SHA-256 checksum is verified **before** the new binary replaces an existing install, so a bad download can never break a working setup.
- Reinstalls across privilege levels are handled: if pancake was previously installed as admin (`Program Files`) and you reinstall as a regular user (or vice versa), the leftover copy is removed (or you are told to run once elevated to clean it up) so it cannot shadow the new one.
- `uninstall` checks both install locations and cleans the PATH entries it added.

### Using Go
You can install the tool using `go install`:
```bash
go install github.com/a6h15hek/pancake@latest
```
Alternatively, download the pre-built binaries from the [Releases page](https://github.com/a6h15hek/pancake/releases). Each release ships SHA-256 checksums in `checksums.txt`.

## Learn More
```sh
$ pancake help       # prints the version, edit config commands                    
$ pancake project    # prints all project commands
$ pancake tool       # prints all tool commands
```

### Creating an Alias for Pancake
Windows (PowerShell) \
Step 1: Open PowerShell profile and Run: `notepad $PROFILE` \
Step 2: Add: `Set-Alias pc pancake` \
Step 3: Save, close and Run Command: `. $PROFILE` 

macOS and Linux \
Step 1: Open shell config file: `nano ~/.bashrc` \
Step 2: Add: `alias pc='pancake'` \
Step 3: Save, reload: `source ~/.bashrc` 


## Pancake AI

Pancake AI allows you to write commands in natural language and have them translated into executable shell commands. It utilizes AI models (supporting both Gemini and ChatGPT) to understand your input, generate the corresponding command, and offers you the choice to execute it, copy it, or ask a follow-up question.

```bash
$ pancake ai "find all files larger than 10MB in my home directory"
```

# Developer Documentation
- This will start an interactive session where the AI will provide the command. You will then have the following options: 
- Run Code: Press Ctrl+R to execute the generated command directly (for bash and python). 
- Copy Command: Press Enter to copy the command to your clipboard. +- Quit: Press Ctrl+C to exit the session. 
- Follow-up: Simply start typing a follow-up question and press Enter. 

### AI Configuration 
To use Pancake AI, you need to configure your preferred AI provider in your $HOME/pancake.yml file. Add your API key for either Gemini or ChatGPT.

```yml
default_ai: gemini # or chatgpt

chatgpt:
  api_key: "YOUR_OPENAI_API_KEY"
  temperature: 0.7
  url: "https://api.openai.com/v1/chat/completions"
  model: "gpt-3.5-turbo"
  context: "PRINT OUTPUT IN MARKDOWN. You are a helpful assistant that translates natural language into executable shell commands..."

gemini:
  api_key: "YOUR_GEMINI_API_KEY"
  temperature: 0.7
  url: "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
  context: "You are a helpful assistant that translates natural language into executable shell commands. Only provide the command, with no extra text or explanation."
```

### Project Structure

```bash
pancake/
├── utils/
│   ├── constants.go          # constants + troubleshooting messages
│   ├── functions.go          # cross-platform shell/git helpers
│   ├── structure.go          # config load/validate/update, env expansion
│   ├── gemini_client.go      # Gemini AI client
│   ├── chatgpt_client.go     # ChatGPT AI client
│   ├── functions_test.go     # unit tests
│   └── structure_test.go     # unit tests
├── cmd/
│   ├── root.go               # pancake init, version, edit config
│   ├── project.go            # list/sync/open/build/run/pwd/monitor
│   ├── tool.go               # tool install/uninstall/list/search/setup
│   └── ai.go                 # pancake ai
├── test/                     # e2e harness (mock HOME + mock release server)
├── .github/workflows/        # CI (test.yml) + release (release.yml)
├── main.go
├── go.mod
└── go.sum
```

### Running the project
```bash
go build
go install
pancake [args]
```

### Config file Structure
```yml
# Pancake Configuration File.
# Home directory for project storage
home: $HOME/pancake # For MacOS & Linux
#home: '%userprofile%/pancake' # For Windows

code_editor: code . # Preferred code editor (code -> VS Code, idea -> IntelliJ IDE)

default_ai: gemini # or chatgpt

chatgpt:
  api_key: ""
  temperature: 0.7
  url: "https://api.openai.com/v1/chat/completions"
  model: "gpt-3.5-turbo"
  context: "PRINT OUTPUT IN MARKDOWN. You are a helpful assistant that translates natural language into executable shell commands..."

gemini:
  api_key: ""
  temperature: 0.7
  url: "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
  context: "You are a helpful assistant that translates natural language into executable shell commands. Only provide the command, with no extra text or explanation."

tools:
  - tree
projects:
  spring-boot:
    remote_ssh_url: git@github.com:spring-guides/gs-spring-boot.git
    run: cd gs-spring-boot/initial && mvn spring-boot:run
    build: cd gs-spring-boot/initial && mvn clean install
  june-gpt:
    remote_ssh_url: git@github.com:suren-atoyan/react-pwa.git
    type: web
    port: "3000"
    run: npm start
    build: npm install
```

Config validation: `pancake` checks `home` is set and absolute, `default_ai` is `gemini`/`chatgpt` (or empty), project names contain no path separators, and every project has a `remote_ssh_url`. On any failure it prints an actionable message pointing you at the field to fix in `pancake.yml`.

### Build Binaries
```bash
./build.sh              # builds all 6 targets + checksums.txt into ./build/ (version from git describe)
./build.sh latest       # same as above, explicit
./build.sh v1.3.0       # embed a specific version
```
Targets: `linux/{amd64,arm64}`, `darwin/{amd64,arm64}`, `windows/{amd64,arm64}`. The version is injected via `-ldflags` and reported by `pancake version`. `checksums.txt` is what the install scripts verify.

### Releases
Releases are published **automatically** when a PR merges to `main`. The CI workflow:
1. Computes the next semver from commit messages since the last tag (`feat:` → minor, `BREAKING CHANGE` → major, everything else → patch).
2. Builds all 6 targets with that version injected via `-ldflags`.
3. Tags the merge commit and publishes a GitHub Release with binaries + `checksums.txt` + `readme.txt`.

The version badge at the top of this README and the `curl | bash` install commands always point at the latest release automatically — no manual version bumps needed in the docs. See all releases on the [Releases page](https://github.com/a6h15hek/pancake/releases).

### Testing
```bash
go test ./...           # Go unit tests (utils package)
./test/run_all.sh       # full e2e harness (builds binary, mock HOME, mock release server)
./test/run_all.sh 01 03 # run only specific suites
```
The e2e harness builds the real binary and runs it inside an isolated `mktemp` HOME so your real `~/pancake.yml` is never touched. Install-script tests spin up a local HTTP server mocking a GitHub release with real SHA-256 checksums. See [`test/README.md`](test/README.md) for details.

### Troubleshooting
- `pancake: command not found` right after installing: the install directory is not on your PATH yet — the installer prints the exact line to add (or just open a new terminal on Windows).
- An old version still runs after upgrading: a leftover copy from a previous install is earlier in your PATH. Run the installer again in an interactive terminal (it offers to remove stale copies), or run `uninstall` followed by a fresh install.
- `pancake edit config` opens `~/pancake.yml` in your default editor.
- `pancake init --force` re-creates a fresh config (backs up the old one to `pancake.yml.bak`).
- `pancake version` shows the installed version.
- If a project command fails with "pancake.yml was not found", run `pancake init`.
- If sync fails, ensure your SSH key is set up (`ssh -T git@github.com`) or switch `remote_ssh_url` to an HTTPS URL in `pancake.yml`.
- Docs: [USAGE.md](USAGE.md) and [open an issue](https://github.com/a6h15hek/pancake/issues).

Thank you for visiting the Pancake repository! Feel free to fork and 🌟 the repository!
