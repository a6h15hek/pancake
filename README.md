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

### Macos & Linux 
To install the tool on macOS or Linux, run the following command in your terminal:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh)" install
```
To uninstall the tool, use the following command:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/a6h15hek/pancake/main/macos_linux.sh)" uninstall
```

### Windows
To install the tool on Windows, open PowerShell and run the following command:
```powershell
Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/a6h15hek/pancake/main/windows.ps1')) install
```

To uninstall the tool, use the following command in PowerShell:
```powershell
Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/a6h15hek/pancake/main/windows.ps1')) uninstall
```

### Using Go
You can install the tool using `go install`:

```bash
go install github.com/a6h15hek/pancake@latest
```
Alternatively, download the pre-built binaries from the [Releases page](https://github.com/a6h15hek/pancake/releases).

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
Step 3: Save, reload: `source ~/.bashrc` \


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
default_ai: gemini # or chatgpt:

api_key: "YOUR_OPENAI_API_KEY"
temperature: 0.7
url: "https://api.openai.com/v1/chat/completions"
model: "gpt-3.5-turbo"
context: "PRINT OUTPUT IN MARKDOWN. You are a helpful assistant that translates natural language into executable shell commands..."
+gemini:

api_key: "YOUR_GEMINI_API_KEY"
temperature: 0.7
url: "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
context: "You are a helpful assistant that translates natural language into executable shell commands. Only provide the command, with no extra text or explanation."

```

### Project Structure

```bash
pancake/
├── utils/
│   ├── constants.go
│   ├── functions.go
│   ├── structure.go
├── cmd/
│   ├── root.go
│   ├── project.go
│   ├── tools.go
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

code_editor: code # Preferred code editor (code -> VS Code, idea -> IntelliJ IDE)
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

### Build Binaries
```bash
GOOS=linux GOARCH=amd64 go build -o pancake-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o pancake-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o pancake-windows-amd64.exe
```

Thank you for visiting the Pancake repository! Feel free to fork and 🌟 the repository!
