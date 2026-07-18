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

func loadConfig() bool {
	cfg, err := utils.GetConfig()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println(utils.ConfigHintEditConfig)
		return false
	}
	config = *cfg
	return true
}

func handleProjectAction(args []string, action func(string)) {
	if !loadConfig() {
		return
	}
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

func listProjects() {
	if !loadConfig() {
		return
	}
	fmt.Println("Loading projects")
	if len(config.Projects) == 0 {
		fmt.Println("No projects in pancake.yml. Run 'pancake edit config' to add one.")
		return
	}
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
		fmt.Printf("Syncing... Cloning repository for project %s\n", projectName)
		if err := utils.CloneRepository(projectPath, project.RemoteSSHURL); err != nil {
			fmt.Printf("Error syncing project %s: %v\n", projectName, err)
			return
		}
	} else {
		fmt.Printf("Syncing... Pulling changes for project %s\n", projectName)
		if err := utils.PullChanges(projectPath); err != nil {
			fmt.Printf("Error pulling changes for project %s: %v\n", projectName, err)
			return
		}
	}
	fmt.Printf("Synchronized project %s successfully.\n", projectName)
	fmt.Printf("\nTip: Run 'pancake open %s' to open the specified project in your preferred IDE.\n", projectName)
}

func syncProjects(args []string) {
	handleProjectAction(args, syncSingleProject)
}

// openProject opens a project in the configured code editor.
func openProject(args []string) {
	if !loadConfig() {
		return
	}
	fmt.Println("Loading project")
	path := config.Home
	if len(args) > 0 {
		projectName := args[0]
		if _, ok := getProject(projectName); !ok {
			return
		}
		path = filepath.Join(config.Home, projectName)
	}

	if err := utils.ExecuteCommand(config.CodeEditor, path); err != nil {
		fmt.Printf("Error opening project: %v\n", err)
		fmt.Printf("%s\n", utils.ProjectErrorAddConfig)
		fmt.Printf("%s\n", utils.ProjectErrorSync)
		return
	}
	fmt.Printf("Opened project at %s\n", path)
	if len(args) > 0 {
		fmt.Printf("\nTip: \n- Run 'pancake build %s' to build your project.\n", args[0])
		fmt.Printf("- Run 'pancake run %s' to start the project locally.", args[0])
	}
}

func pwdProject(args []string) {
	if !loadConfig() {
		return
	}
	path := config.Home
	if len(args) > 0 {
		projectName := args[0]
		if _, ok := getProject(projectName); !ok {
			return
		}
		path = filepath.Join(config.Home, projectName)
	}

	cdCommand := fmt.Sprintf("cd %s", path)

	if err := clipboard.WriteAll(cdCommand); err != nil {
		fmt.Printf("Failed to copy to clipboard: %v\n", err)
		return
	}

	fmt.Printf("Project path: %s\n", path)
	fmt.Printf("\nTip: The command 'cd %s' has been copied to your clipboard.\n", path)
	fmt.Println("Press Ctrl+V to paste and use the command.")
}

// buildSingleProject builds a single project by name.
func buildSingleProject(projectName string) {
	fmt.Printf("Building... Running build command for project %s\n", projectName)
	project, ok := getProject(projectName)
	if !ok {
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("Project path %s does not exist.\n", projectPath)
		fmt.Printf("%s\n", utils.ProjectErrorSync)
		return
	}

	if project.Build == "" {
		fmt.Println("Build command not specified in pancake.yml.")
		fmt.Printf("%s\n", utils.ProjectErrorAddCommand)
		return
	}

	if err := utils.ExecuteCommand(project.Build, projectPath); err != nil {
		fmt.Printf("Error building project %s: %v\n", projectName, err)
		return
	}
	fmt.Printf("Built project %s successfully.\n", projectName)
	fmt.Printf("\nTip: Run 'pancake run %s' to start the project locally.\n", projectName)
}

func buildProject(args []string) {
	handleProjectAction(args, buildSingleProject)
}

// runSingleProject runs a single project by name.
func runSingleProject(projectName string) {
	fmt.Printf("Running project %s\n", projectName)
	project, ok := getProject(projectName)
	if !ok {
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("Project path %s does not exist.\n", projectPath)
		fmt.Printf("%s\n", utils.ProjectErrorAddConfig)
		return
	}

	if project.Run == "" {
		fmt.Println("Run command not specified in pancake.yml.")
		fmt.Printf("%s\n", utils.ProjectErrorAddCommand)
		return
	}

	if err := utils.ExecuteCommandInNewTerminal(project.Run, projectPath, projectName, &projectPIDs); err != nil {
		fmt.Printf("Error running project %s: %v\n", projectName, err)
		return
	}
	fmt.Printf("Started project %s successfully.\n", projectName)
	if err := utils.SaveProjectPIDs(config.Home, projectPIDs); err != nil {
		fmt.Printf("Warning: could not save project PIDs: %v\n", err)
	}
}

func runProject(args []string) {
	handleProjectAction(args, runSingleProject)
}

func monitorProject() {
	if !loadConfig() {
		return
	}
	fmt.Println("Monitoring... Fetching project status")
	if err := utils.LoadProjectPIDs(config.Home, &projectPIDs); err != nil {
		fmt.Printf("Warning: could not load project PIDs: %v\n", err)
	}

	data := [][]string{
		{"Project Name", "Running", "PID", "Port", "Type"},
	}

	for projectName, project := range config.Projects {
		running := "No"
		pid := "-"
		port := project.Port
		projectType := project.Type

		if pidVal, exists := projectPIDs[projectName]; exists {
			running = "Yes"
			pid = fmt.Sprintf("%d", pidVal)
		}
		if port == "" {
			port = "-"
		}

		data = append(data, []string{projectName, running, pid, port, projectType})
	}

	utils.PrintTable(data)
}
