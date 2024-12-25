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
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pancake",
	Short: "A brief description of your application",
	Long: `Pancake Project Management Tool

Pancake is a versatile tool designed to streamline your project management workflow. It simplifies running web and server modules, monitors application status, and offers customizable project locations and override files. Best of all, you can run and open projects from anywhere!

Usage:
Use the command 'pancake [command]'. Replace '<project_name>' with the name of your project.

Commands:
- pancake project list: List all projects defined in the pancake.yml file.
- pancake project sync: Sync all projects. This clones or pulls the latest changes from the repositories.
- pancake project sync <project_name>: Sync the specified project. This clones or pulls the latest changes from the repository of the specified project.
- pancake build <project_name>: Build the specified project. This runs the build command defined in the pancake.yml file for the specified project.
- pancake run <project_name>: Run the specified project. This runs the command defined in the run variable in the pancake.yml file for the specified project.
- pancake stop <project_name>: Stop the specified project. This stops the process running the specified project.
- pancake status: Check the status of all projects. This prints the status, PID, and start time of the process for each project.
- pancake edit config: Open the pancake.yml file in the default editor.
- pancake open <project_name>: Open the specified project with the command mentioned in code_editor_command.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pancake.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


