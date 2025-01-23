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
	LongDescription = `It's a command line tool to manage the project lifecycle and workflow. 
    It helps in syncing projects from a remote repository, building and running projects, 
    and installing tools needed for development. It keeps everything in one place and stores all these 
    configurations in the pancake.yml file. Sharing this single file enables sharing 
    your entire developer setup.`
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
    build: npm install
    start: npm start

  spring-helloworld:
    remote_ssh_url: git@github.com:paulczar/spring-helloworld.git
    build: mvn clean install
    start: mvn spring-boot:run
`

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
