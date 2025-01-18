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
	"os/exec"
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

var config utils.Config

func listProjects() {
	config = *utils.GetConfig()
	for projectName := range config.Projects {
		fmt.Printf("- %s\n", projectName)
	}
}

func syncSingleProject(projectName string) {
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("Project %s not found in configuration.\n", projectName)
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	gitDirPath := filepath.Join(projectPath, ".git")

	projectExists := utils.CheckExists(projectPath)
	gitExists := utils.CheckExists(gitDirPath)

	if !projectExists || !gitExists {
		utils.CloneRepository(projectPath, project.RemoteSSHURL)
	} else {
		utils.PullChanges(projectPath)
	}
	fmt.Printf("Synchronized project %s successfully.\n", projectName)
}

func syncProjects(args []string) {
	config = *utils.GetConfig()
	if len(args) == 0 {
		if utils.ConfirmAction("sync") {
			for projectName := range config.Projects {
				syncSingleProject(projectName)
			}
		} else {
			fmt.Println("Sync canceled.")
		}
	} else {
		syncSingleProject(args[0])
	}
}

func openProject(args []string) {
	config = *utils.GetConfig()
	path := config.Home
	if len(args) > 0 {
		path = filepath.Join(config.Home, args[0])
	}

	cmd := exec.Command(config.CodeEditor, path)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error opening project: %v\n", err)
	} else {
		fmt.Printf("Opened project at %s\n", path)
	}
}

func buildSingleProject(projectName string) {
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("Project %s not found in configuration.\n", projectName)
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("Project path %s does not exist.\n", projectPath)
		return
	}

	if project.Build == "" {
		fmt.Println("Build command not specified in the configuration.")
		return
	}

	cmd := exec.Command("sh", "-c", project.Build)
	cmd.Dir = projectPath
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error building project %v: %v\n", projectName, err)
	} else {
		fmt.Printf("Built project %s successfully.\n", projectName)
	}
}

func buildProject(args []string) {
	config = *utils.GetConfig()
	if len(args) == 0 {
		if utils.ConfirmAction("build") {
			for projectName := range config.Projects {
				buildSingleProject(projectName)
			}
		} else {
			fmt.Println("Build canceled.")
		}
	} else {
		buildSingleProject(args[0])
	}
}

func startSingleProject(projectName string) {
	project, exists := config.Projects[projectName]
	if !exists {
		fmt.Printf("Project %s not found in configuration.\n", projectName)
		return
	}

	projectPath := filepath.Join(config.Home, projectName)
	if !utils.CheckExists(projectPath) {
		fmt.Printf("Project path %s does not exist.\n", projectPath)
		return
	}

	if project.Start == "" {
		fmt.Println("Start command not specified in the configuration.")
		return
	}

	cmd := exec.Command("sh", "-c", project.Start)
	cmd.Dir = projectPath
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error starting project %v: %v\n", projectName, err)
	} else {
		fmt.Printf("Started project %s successfully.\n", projectName)
	}
}

func startProject(args []string) {
	config = *utils.GetConfig()
	if len(args) == 0 {
		if utils.ConfirmAction("start") {
			for projectName := range config.Projects {
				startSingleProject(projectName)
			}
		} else {
			fmt.Println("Start canceled.")
		}
	} else {
		startSingleProject(args[0])
	}
}

func stopProject(args []string) {
	fmt.Println(utils.NotImplemented)
}

func monitorProject() {
	fmt.Println(utils.NotImplemented)
}

func init() {
	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(
		&cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listProjects() }},
		&cobra.Command{Use: "sync", Run: func(cmd *cobra.Command, args []string) { syncProjects(args) }},
		&cobra.Command{Use: "open", Run: func(cmd *cobra.Command, args []string) { openProject(args) }},
		&cobra.Command{Use: "build", Run: func(cmd *cobra.Command, args []string) { buildProject(args) }},
		&cobra.Command{Use: "start", Run: func(cmd *cobra.Command, args []string) { startProject(args) }},
		&cobra.Command{Use: "stop", Run: func(cmd *cobra.Command, args []string) { stopProject(args) }},
		&cobra.Command{Use: "monitor", Run: func(cmd *cobra.Command, args []string) { monitorProject() }},
	)
}
