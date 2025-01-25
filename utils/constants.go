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
	Version         = "v1.0.0"
	Description     = "A tool to streamline project management workflow."
	LongDescription = `Manage multiple projects' lifecycle and workflow.
> pancake edit-config #Add your project's Git SSH links
> pancake sync #Sync your projects 
> pancake list #List all projects
> pancake project open <project-name> #Open project in default IDE
And do this from any location`
)

const (
	NotImplemented = "Soon to be Implemented."
)

const DefaultYMLContent = `# Pancake Configuration File.
# Home directory for project storage
home: $HOME/pancake # For MacOS & Linux
#home: '%userprofile%/pancake' # For Windows

code_editor: code # Preferred code editor (code -> VS Code, idea -> IntelliJ IDE)
tools:
  - visual-studio-code
  - tree
projects:
  june-gpt: 
    remote_ssh_url: git@github.com:a6h15hek/june-devgpt-desktop.git
    type: web
    port: 3000
    build: npm install
    run: npm start

  spring-helloworld:
    remote_ssh_url: git@github.com:paulczar/spring-helloworld.git
    build: mvn clean install
    run: mvn spring-boot:run
`
const (
	ToolsDescription = `Usage:
  pancake tool search <TEXT|/REGEX/>
  pancake tool info [FORMULA|CASK...]
  pancake tool install <FORMULA|CASK...>
  pancake tool update
  pancake tool upgrade [FORMULA|CASK...]
  pancake tool uninstall <FORMULA|CASK...>
  pancake tool list [FORMULA|CASK...]

  pancake t search <TEXT|/REGEX/>
  pancake t info [FORMULA|CASK...]
  pancake t install <FORMULA|CASK...>
  pancake t update
  pancake t upgrade [FORMULA|CASK...]
  pancake t uninstall <FORMULA|CASK...>
  pancake t list [FORMULA|CASK...]

Troubleshooting:
  pancake edit-config
  pancake version

Further Assistance:
  Search Brew Packages: https://brew.sh/
  Search Chocolatey Packages: https://community.chocolatey.org/packages
  Copy the package name and use with 'pancake tools install <package-name>'`

	ProjectDescription = `Usage:
  pancake project list
  pancake project sync [args...]
  pancake project open [args...]
  pancake project build [args...]
  pancake project start [args...]
  pancake project monitor

  pancake p list
  pancake p sync [args...]
  pancake p open [args...]
  pancake p build [args...]
  pancake p start
  pancake p monitor

Troubleshooting:
  pancake project edit-config
  pancake project version`
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
