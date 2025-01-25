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
	Use: "tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please use one of the subcommands: install, uninstall, list, update, search, setup, info, upgrade")
	},
}

func init() {
	rootCmd.AddCommand(toolCmd)
	toolCmd.AddCommand(
		&cobra.Command{Use: "install", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "install") }},
		&cobra.Command{Use: "uninstall", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "uninstall") }},
		&cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listTools() }},
		&cobra.Command{Use: "update", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "update") }},
		&cobra.Command{Use: "search", Run: func(cmd *cobra.Command, args []string) { searchTool(args) }},
		&cobra.Command{Use: "setup", Run: func(cmd *cobra.Command, args []string) { setupTools() }},
		&cobra.Command{Use: "info", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "info") }},
		&cobra.Command{Use: "upgrade", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "upgrade") }},
	)
}

func listTools() {
	fmt.Println("🔍 Loading... Listing projects")
	config := *utils.GetConfig()
	for toolName := range config.Tools {
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
		fmt.Println("Unsupported platform:", platform)
	}
}

func searchTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing search query")
		return
	}
	if !utils.EnsureToolInstalled() {
		return
	}
	query := strings.Join(args, " ")
	cmdStr := fmt.Sprintf("%s search %s", utils.GetPackageManager(), query)
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error searching tool:", err)
	}
}

func handleToolCommand(args []string, action string) {
	if len(args) == 0 {
		fmt.Printf("Error: Missing tool name for %s\n", action)
		return
	}
	if !utils.EnsureToolInstalled() {
		return
	}
	toolName := args[0]
	cmdStr := fmt.Sprintf("%s %s %s", utils.GetPackageManager(), action, toolName)
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Printf("Error during %s of tool: %s\n", action, err)
	}
}
