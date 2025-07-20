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
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.ProjectDescription)
	},
}

var projectPIDs = make(map[string]int)

func init() {
	rootCmd.AddCommand(projectCmd)

	var commandList = []*cobra.Command{
		{Use: "list", Aliases: []string{"l"}, Run: func(cmd *cobra.Command, args []string) { listProjects() }},
		{Use: "pwd", Aliases: []string{"p"}, Run: func(cmd *cobra.Command, args []string) { pwdProject(args) }},
		{Use: "sync", Aliases: []string{"s"}, Run: func(cmd *cobra.Command, args []string) { syncProjects(args) }},
		{Use: "open", Aliases: []string{"o"}, Run: func(cmd *cobra.Command, args []string) { openProject(args) }},
		{Use: "build", Aliases: []string{"b"}, Run: func(cmd *cobra.Command, args []string) { buildProject(args) }},
		{Use: "run", Aliases: []string{"r"}, Run: func(cmd *cobra.Command, args []string) { runProject(args) }},
		{Use: "monitor", Aliases: []string{"m"}, Run: func(cmd *cobra.Command, args []string) { monitorProject() }},
	}

	projectCmd.AddCommand(commandList...)
	// Add the same commands to rootCmd
	rootCmd.AddCommand(commandList...)
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
	fmt.Printf("\nTip: Run 'pancake sync <project_name>' to sync your project with the remote repository.\n")
}

// getProject retrieves a project from the configuration and handles not-found errors.
func getProject(projectName string) (*utils.Project, bool) {
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("❌ Project %s not found in configuration.\n", projectName)
		fmt.Printf("%s\n", utils.ProjectErrorAddConfig)
		return nil, false
	}
	return &project, true
}

// syncSingleProject synchronizes a single project by name.
func syncSingleProject(projectName string) {
	project, ok := getProject(projectName)
	if !ok {
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
	fmt.Printf("\nTip: Run 'pancake open %s' to open the specified project in your preferred IDE.\n", projectName)
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
		projectName := args[0]
		if _, ok := getProject(projectName); !ok {
			return
		}
		path = filepath.Join(config.Home, projectName)
	}

	err := utils.ExecuteCommand(config.CodeEditor, path)
	if err != nil {
		fmt.Printf("❌ Error opening project: %v\n", err)
		fmt.Printf("%s\n", utils.ProjectErrorAddConfig)
		fmt.Printf("%s\n", utils.ProjectErrorSync)
	} else {
		fmt.Printf("✅ Opened project at %s\n", path)
		if len(args) > 0 {
			fmt.Printf("\nTip: \n- Run 'pancake build %s' to build your project.\n", args[0])
			fmt.Printf("- Run 'pancake start %s' to start the project locally.", args[0])
		}
	}
}

func pwdProject(args []string) {
	loadConfig()
	path := config.Home
	if len(args) > 0 {
		projectName := args[0]
		if _, ok := getProject(projectName); !ok {
			return
		}
		path = filepath.Join(config.Home, projectName)
	}

	cdCommand := fmt.Sprintf("cd %s", path)

	err := clipboard.WriteAll(cdCommand)
	if err != nil {
		fmt.Printf("❌ Failed to copy to clipboard: %v\n", err)
		return
	}

	fmt.Printf("📁 Project path: %s\n", path)
	fmt.Printf("\nTip: The command 'cd %s' has been copied to your clipboard.\n", path)
	fmt.Println("Press Ctrl+V to paste and use the command.")
}

// buildSingleProject builds a single project by name.
func buildSingleProject(projectName string) {
	fmt.Printf("🔨 Building... Running build command for project %s\n", projectName)
	project, ok := getProject(projectName)
	if !ok {
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("❌ Project path %s does not exist.\n", projectPath)
		fmt.Printf("%s\n", utils.ProjectErrorSync)
		return
	}

	if project.Build == "" {
		fmt.Println("❌ Build command not specified in the configuration.")
		fmt.Printf("%s\n", utils.ProjectErrorAddCommand)
		return
	}

	err := utils.ExecuteCommand(project.Build, projectPath)
	if err != nil {
		fmt.Printf("❌ Error building project %v: %v\n", projectName, err)
	} else {
		fmt.Printf("✅ Built project %s successfully.\n", projectName)
		fmt.Printf("\nTip: Run 'pancake start %s' to start the project locally.\n", projectName)
	}
}

func buildProject(args []string) {
	handleProjectAction(args, buildSingleProject)
}

// runSingleProject runs a single project by name.
func runSingleProject(projectName string) {
	fmt.Printf("🚀 Running project %s\n", projectName)
	project, ok := getProject(projectName)
	if !ok {
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("❌ Project path %s does not exist.\n", projectPath)
		fmt.Printf("%s\n", utils.ProjectErrorAddConfig)
		return
	}

	if project.Run == "" {
		fmt.Println("❌ Run command not specified in the configuration.")
		fmt.Printf("%s\n", utils.ProjectErrorAddCommand)
		return
	}

	err := utils.ExecuteCommandInNewTerminal(project.Run, projectPath, projectName, &projectPIDs)
	if err != nil {
		fmt.Printf("❌ Error running project %v: %v\n", projectName, err)
	} else {
		fmt.Printf("✅ Started project %s successfully.\n", projectName)
		utils.SaveProjectPIDs(config.Home, projectPIDs)
	}
}

func runProject(args []string) {
	handleProjectAction(args, runSingleProject)
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
