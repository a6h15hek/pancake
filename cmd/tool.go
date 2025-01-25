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
	"runtime"
	"strings"

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

var toolCmd = &cobra.Command{
	Use:     "tool",
	Aliases: []string{"t"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(utils.ToolsDescription)
	},
}

func init() {
	rootCmd.AddCommand(toolCmd)
	toolCmd.AddCommand(
		&cobra.Command{Use: "install", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "install") }},
		&cobra.Command{Use: "uninstall", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "uninstall") }},
		&cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listTools() }},
		&cobra.Command{Use: "update", Run: func(cmd *cobra.Command, args []string) { updateTools() }},
		&cobra.Command{Use: "search", Run: func(cmd *cobra.Command, args []string) { searchTool(args) }},
		&cobra.Command{Use: "setup", Run: func(cmd *cobra.Command, args []string) { setupTools() }},
		&cobra.Command{Use: "info", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "info") }},
		&cobra.Command{Use: "upgrade", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "upgrade") }},
	)
}

func listTools() {
	fmt.Println("🔍 Loading... Listing tools")
	config := *utils.GetConfig()
	for _, toolName := range config.Tools {
		fmt.Printf("- %s\n", toolName)
	}
}

func setupTools() {
	platform := runtime.GOOS
	switch platform {
	case "windows":
		fmt.Println("The Pancake uses Chocolatey internally.")
		utils.SetupChocolatey()
	case "darwin", "linux":
		fmt.Println("The Pancake uses Homebrew internally.")
		utils.SetupHomebrew()
	default:
		fmt.Println("❌ Unsupported platform:", platform)
	}
}

func searchTool(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Error: Missing search query")
		return
	}
	if !utils.EnsureToolInstalled() {
		return
	}
	query := strings.Join(args, " ")
	cmdStr := fmt.Sprintf("%s search %s", utils.GetPackageManager(), query)
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("❌ Error searching tool:", err)
	}
}

func handleToolCommand(args []string, action string) {
	if len(args) == 0 {
		fmt.Printf("❌ Error: Missing tool name for %s\n", action)
		return
	}
	if !utils.EnsureToolInstalled() {
		return
	}
	toolName := args[0]
	config := *utils.GetConfig()

	switch action {
	case "install":
		for _, t := range config.Tools {
			if t == toolName {
				fmt.Printf("🔄 Tool '%s' already installed. Updating instead.\n", toolName)
				action = "upgrade"
				break
			}
		}
	case "uninstall":
		found := false
		for _, t := range config.Tools {
			if t == toolName {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("❌ Tool '%s' is not installed via pancake. Cannot uninstall.\n", toolName)
			return
		}
	}

	cmdStr := fmt.Sprintf("%s %s %s", utils.GetPackageManager(), action, toolName)
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Printf("❌ Error during %s of tool: %s\n", action, err)
		return
	}

	if action == "install" {
		config.Tools = append(config.Tools, toolName)
		utils.UpdateConfig(&config)
	} else if action == "uninstall" {
		for i, t := range config.Tools {
			if t == toolName {
				config.Tools = append(config.Tools[:i], config.Tools[i+1:]...)
				utils.UpdateConfig(&config)
				break
			}
		}
	}
}

func updateTools() {
	if !utils.EnsureToolInstalled() {
		return
	}
	cmdStr := fmt.Sprintf("%s update", utils.GetPackageManager())
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("❌ Error updating tools:", err)
	}
}
