# Pancake: Developer Command Line Tool to Manage Projects

Pancake is a command line tool designed to streamline your project management workflow. It simplifies and makes it easy to sync multiple projects from remote repositories, helps in building and running projects on a local machine, and allows for the installation, uninstallation, and updating of development software and tools. 

```bash
pancake list                  # View list of all projects you are working on
pancake sync                  # Sync multiple projects from the remote repository 
pancake build <project_name>  # Build a project
pancake start <project_name>  # Start a project on the local machine
pancake monitor               # Show details about the port, process PID, and uptime
pancake stop <project_name>   # Stop a project
```
It keeps all project files and tools in one place and stores all configurations in the pancake.yml file. Sharing this single file enables the sharing of your entire developer setup, making backups and migration from one machine to another easy.

For migration and sharing projects, just copy pancake.yml to your user home location and run:
```bash
pancake init
```
Everything your project needs will be installed. All the build and run configurations will already be in the configuration file, and it will sync the project from your remote repository all at once.


Another feature in development is Pancake GPT, which allows users to write commands in natural language and convert them into actual commands. It utilizes GPT models to understand the user's natural language input, interpret it, create a corresponding command, and execute it.

```bash
pancake gpt <user_description_of_command>
```

## Features:
1. Single location for all configuration and project files
2. Consistent set of commands across operating systems
3. Simplified commands for running and building projects with single commands and writing configurations only once

## Project Structure

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

## Running the project

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

Thank you for visiting the Pancake repository! Feel free to fork and 🌟 the repository!


