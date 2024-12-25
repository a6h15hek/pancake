# Pancake Project Management Tool 

Pancake is a versatile tool designed to streamline your project management workflow. It simplifies running web and server modules, monitors application status, and offers customizable project locations and override files. Best of all, you can run and open projects from anywhere!

## Features:
1. Simplifies running web and server modules.
2. Monitors all running and non-running applications.
3. Customizable project locations and override files.
4. Runs and opens projects from anywhere.

```bash
pancake/
├── utils/
│   ├── constants.go
├── cmd/
│   ├── root.go
│   ├── project.go
│   ├── tools.go
│   ├── common.go
├── main.go
├── pancake.yml
├── go.mod
└── go.sum
```

## Usage:
Use the command `pancake [command]`. Replace `<project_name>` with the name of your project.

| Command | Description |
| --- | --- |
| pancake project list | List all projects defined in the pancake.yml file. |
| pancake project sync | Sync all projects. This clones or pulls the latest changes from the repositories. |
| pancake project sync <project_name> | Sync the specified project. This clones or pulls the latest changes from the repository of the specified project. |
| pancake build <project_name> | Build the specified project. This runs the build command defined in the pancake.yml file for the specified project. |
| pancake run <project_name> | Run the specified project. This runs the command defined in the run variable in the pancake.yml file for the specified project. |
| pancake stop <project_name> | Stop the specified project. This stops the process running the specified project. |
| pancake status | Check the status of all projects. This prints the status, PID, and start time of the process for each project. |
| pancake edit config | Open the pancake.yml file in the default editor. |
| pancake open <project_name> | Open the specified project with the command mentioned in code_editor_command. |

## Installation

```bash
go build
go install
pancake [args]
```

## Config file Structure
```yml
# Home directory for project storage
home: $HOME/pancake # For MacOS & Linux
#home: '%userprofile%/pancake' # For Windows

code_editor: code # Preferred code editor (code -> VS Code, idea -> IntelliJ IDE)
tools:
  visual-studio-code: 1.96.2
  maven: 3.9.9
  openjdk@21: 21.0.5
  openjdk@17: 17.0.13
  node@20: 20.18.1
  node@22: 22.12.0
projects:
  june-gpt: 
    remote_ssh_url: git@github.com:a6h15hek/june-devgpt-desktop.git
    type: web
    port: 3000
    start: npm start

  spring-helloworld:
    remote_ssh_url: git@github.com:paulczar/spring-helloworld.git
    build: mvn clean install
    start: mvn spring-boot:run

```

## 👨‍💻 Developer: Yadav, Abhishek - GitHub
Thank you for visiting Pancake repo ! If you have any questions or need further assistance, feel free to ask.


