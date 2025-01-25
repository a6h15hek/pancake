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
	"fmt"
	"path/filepath"

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		listProjects()
	},
}

var projectPIDs = make(map[string]int)

func init() {
	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(
		&cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listProjects() }},
		&cobra.Command{Use: "sync", Run: func(cmd *cobra.Command, args []string) { syncProjects(args) }},
		&cobra.Command{Use: "open", Run: func(cmd *cobra.Command, args []string) { openProject(args) }},
		&cobra.Command{Use: "build", Run: func(cmd *cobra.Command, args []string) { buildProject(args) }},
		&cobra.Command{Use: "start", Run: func(cmd *cobra.Command, args []string) { startProject(args) }},
		&cobra.Command{Use: "monitor", Run: func(cmd *cobra.Command, args []string) { monitorProject() }},
	)
}

func loadConfig() {
	config = *utils.GetConfig()
}

func handleProjectAction(args []string, action func(string)) {
	loadConfig()
	if len(args) == 0 {
		if utils.ConfirmAction("Are you sure you want to run for all projects? This may take some time. (yes/no)") {
			for projectName := range config.Projects {
				action(projectName)
			}
		}
	} else {
		action(args[0])
	}
}

// listProjects prints all projects listed in the configuration.
func listProjects() {
	loadConfig()
	fmt.Println("🔍 Loading... Listing projects")
	for projectName := range config.Projects {
		fmt.Printf("- %s\n", projectName)
	}
}

// syncSingleProject synchronizes a single project by name.
func syncSingleProject(projectName string) {
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("❌ Project %s not found in configuration.\n", projectName)
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	gitDirPath := filepath.Join(projectPath, ".git")

	projectExists := utils.CheckExists(projectPath)
	gitExists := utils.CheckExists(gitDirPath)

	if !projectExists || !gitExists {
		fmt.Printf("🔄 Syncing... Cloning repository for project %s\n", projectName)
		utils.CloneRepository(projectPath, project.RemoteSSHURL)
	} else {
		fmt.Printf("🔄 Syncing... Pulling changes for project %s\n", projectName)
		utils.PullChanges(projectPath)
	}
	fmt.Printf("✅ Synchronized project %s successfully.\n", projectName)
}

func syncProjects(args []string) {
	handleProjectAction(args, syncSingleProject)
}

// openProject opens a project in the configured code editor.
func openProject(args []string) {
	loadConfig()
	fmt.Println("🔍 Loading... Opening project")
	path := config.Home
	if len(args) > 0 {
		path = filepath.Join(config.Home, args[0])
	}

	err := utils.ExecuteCommand(config.CodeEditor, path)
	if err != nil {
		fmt.Printf("❌ Error opening project: %v\n", err)
	} else {
		fmt.Printf("✅ Opened project at %s\n", path)
	}
}

// buildSingleProject builds a single project by name.
func buildSingleProject(projectName string) {
	fmt.Printf("🔨 Building... Running build command for project %s\n", projectName)
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("❌ Project %s not found in configuration.\n", projectName)
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("❌ Project path %s does not exist.\n", projectPath)
		return
	}

	if project.Build == "" {
		fmt.Println("❌ Build command not specified in the configuration.")
		return
	}

	err := utils.ExecuteCommand(project.Build, projectPath)
	if err != nil {
		fmt.Printf("❌ Error building project %v: %v\n", projectName, err)
	} else {
		fmt.Printf("✅ Built project %s successfully.\n", projectName)
	}
}

func buildProject(args []string) {
	handleProjectAction(args, buildSingleProject)
}

// startSingleProject starts a single project by name.
func startSingleProject(projectName string) {
	fmt.Printf("🚀 Starting... Running start command for project %s\n", projectName)
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("❌ Project %s not found in configuration.\n", projectName)
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("❌ Project path %s does not exist.\n", projectPath)
		return
	}

	if project.Start == "" {
		fmt.Println("❌ Start command not specified in the configuration.")
		return
	}

	err := utils.ExecuteCommandInNewTerminal(project.Start, projectPath, projectName, &projectPIDs)
	if err != nil {
		fmt.Printf("❌ Error starting project %v: %v\n", projectName, err)
	} else {
		fmt.Printf("✅ Started project %s successfully.\n", projectName)
		utils.SaveProjectPIDs(config.Home, projectPIDs)
	}
}

func startProject(args []string) {
	handleProjectAction(args, startSingleProject)
}

// monitorProject prints a table with information about all projects.
func monitorProject() {
	loadConfig()
	fmt.Println("🔍 Monitoring... Fetching project status")
	utils.LoadProjectPIDs(config.Home, &projectPIDs)

	data := [][]string{
		{"Project Name", "Running", "PID", "Port", "Type"},
	}

	for projectName, project := range config.Projects {
		running := "No"
		pid := "-"
		port := "-"
		projectType := project.Type

		if pidVal, exists := projectPIDs[projectName]; exists {
			running = "Yes"
			pid = fmt.Sprintf("%d", pidVal)
			if project.Port != "" {
				port = project.Port
			}
		}

		data = append(data, []string{projectName, running, pid, port, projectType})
	}

	utils.PrintTable(data)
}

// stopProject stops a project by name.
// func stopProject(args []string) {
// 	fmt.Println("🛑 Stopping... Running stop command")
// 	config = *utils.GetConfig()

// 	if len(args) == 0 {
// 		if utils.ConfirmAction("Are you sure you want to stop for all projects? This may take some time. (yes/no)") {
// 			for projectName, pid := range projectPIDs {
// 				if err := utils.KillProjectByPID(pid); err == nil {
// 					fmt.Printf("✅ Stopped project %s successfully.\n", projectName)
// 					delete(projectPIDs, projectName)
// 					saveProjectPIDs()
// 				} else {
// 					fmt.Printf("❌ Error stopping project %s: %v\n", projectName, err)
// 				}
// 			}
// 		} else {
// 			fmt.Println("Stop canceled.")
// 		}
// 	} else {
// 		projectName := args[0]
// 		if pid, exists := projectPIDs[projectName]; exists {
// 			if err := utils.KillProjectByPID(pid); err == nil {
// 				fmt.Printf("✅ Stopped project %s successfully.\n", projectName)
// 				delete(projectPIDs, projectName)
// 				saveProjectPIDs()
// 			} else {
// 				fmt.Printf("❌ Error stopping project %s: %v\n", projectName, err)
// 			}
// 		} else {
// 			fmt.Printf("❌ No running process found for project %s.\n", projectName)
// 		}
// 	}
// }
