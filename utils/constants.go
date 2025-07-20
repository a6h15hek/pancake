/*
Copyright © 2024 Abhishek M. Yadav <abhishekyadav@duck.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
* Common Constants used across the project.
* This file is part of the 'Pancake' project and contains common constants
* that are utilized throughout various parts of the project. These constants
* are defined to provide a centralized repository for fixed values, ensuring
* consistency and reducing the likelihood of errors due to hardcoding values
* in multiple places.
**/

package utils

const (
	AppName         = "Pancake"
	Version         = "v1.1.1"
	Description     = "A tool to streamline project management workflow."
	LongDescription = `Example Usage:
  pancake list 
  pancake sync [PROJECT_NAME]
  pancake open [PROJECT_NAME]
  pancake build [PROJECT_NAME]
  pancake run [PROJECT_NAME]
  pancake edit config 

Troubleshooting:
  pancake edit config             or pancake p ec
  pancake version                 or pancake v

Further help:
  pancake help
  https://github.com/a6h15hek/pancake/blob/main/Usage.md
`
)

const (
	NotImplemented = "Soon to be Implemented."
)

const DefaultYMLContent = `# Pancake Configuration File.
# Home directory for project storage
home: $HOME/pancake # For MacOS & Linux
#home: '%userprofile%/pancake' # For Windows

code_editor: code . # Preferred code editor (code -> VS Code, idea -> IntelliJ IDE)

default_ai: gemini

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
`
const (
	ToolsDescription = `Usage:
  pancake tool setup                                                      or  pancake t setup
  pancake tool [install|upgrade|uninstall|list|search|info] <tool_name>   or  pancake t [i|upgrade|uninstall|l|s|info] <tool_name>
  pancake tool update                                                     or  pancake t update

Troubleshooting:
  pancake edit config             or pancake p ec
  pancake version                 or pancake v

Further Assistance:
  Search Brew Packages: https://brew.sh/
  Search Chocolatey Packages: https://community.chocolatey.org/packages
  Copy the package name and use with 'pancake tools install <package-name>'`

	ProjectDescription = `Usage:
  pancake list                                     or  pancake [project|p] l
  pancake [sync|open|build|run|pwd] <project_name> or  pancake [project|p] [s|o|b|r|p] <project_name>
  pancake monitor                                  or  pancake [project|p] m

Troubleshooting:
  pancake edit config             or pancake p ec
  pancake version                 or pancake v`

	ProjectErrorAddConfig  = `Run 'pancake edit config' to check if project exists in configuration file`
	ProjectErrorSync       = `Run 'pancake sync <project_name>' to sync the project.`
	ProjectErrorAddCommand = `Run 'pancake edit config' to add commands.`
)

const (
	Copyright = `Copyright © 2024 Abhishek M. Yadav <abhishekyadav@duck.com>
	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
	`
)
