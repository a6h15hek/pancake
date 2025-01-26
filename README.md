# Pancake: Developer Command Line Tool

Streamlines your project management workflow, simplifies syncing multiple projects from remote repositories, helps build and run projects on a local machine, and allows for the installation, uninstallation, and updating of development software and tools. No need to remember longer commands. Share your entire development setup with your work buddies or back up your entire development setup by sharing `$HOME/pancake.yml`.

```sh
$ pancake project list                  # View the list of all projects you are working on
$ pancake project sync                  # Sync multiple projects from the remote repository 
$ pancake project build <project_name>  # Build a project
$ pancake project run <project_name>    # Start a project on the local machine
$ pancake project monitor               # Show details about the port, process PID, and uptime

$ pancake tool install tree             # Install a tool via pancake 
$ pancake tool upgrade tree             # Update tools via pancake
```
It keeps all project files in the `$HOME/pancake` folder. Sharing this single file enables the sharing of your entire developer setup, making backups and migration from one machine to another easy.

For migration and sharing projects, just copy pancake.yml to your user home location and run:
```bash
$ pancake init
```
Everything your project needs will be installed. All the build and run configurations will already be in the configuration file, and it will sync the project from your remote repository all at once.

## Learn More
```sh
$ pancake help       # prints the version, edit config commands                    
$ pancake project    # prints all project commands
$ pancake tool       # prints all tool commands
```

Another feature in development is Pancake GPT, which allows users to write commands in natural language and convert them into actual commands. It utilizes GPT models to understand the user's natural language input, interpret it, create a corresponding command, and execute it.

```bash
$ pancake gpt <user_description_of_command>
```

# Developer Documentation

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
  - visual-studio-code
  - maven
  - node
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

Thank you for visiting the Pancake repository! Feel free to fork and 🌟 the repository!


