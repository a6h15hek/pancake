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
		&cobra.Command{Use: "install", Run: func(cmd *cobra.Command, args []string) { installTool(args) }},
		&cobra.Command{Use: "uninstall", Run: func(cmd *cobra.Command, args []string) { uninstallTool(args) }},
		&cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listTools() }},
		&cobra.Command{Use: "update", Run: func(cmd *cobra.Command, args []string) { updateTool(args) }},
		&cobra.Command{Use: "search", Run: func(cmd *cobra.Command, args []string) { searchTool(args) }},
		&cobra.Command{Use: "setup", Run: func(cmd *cobra.Command, args []string) { setupTools() }},
		&cobra.Command{Use: "info", Run: func(cmd *cobra.Command, args []string) { infoTool(args) }},
		&cobra.Command{Use: "upgrade", Run: func(cmd *cobra.Command, args []string) { upgradeTool(args) }},
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
	case "darwin", "linux":
		fmt.Println("The Pancake uses Homebrew internally.")
	default:
		fmt.Println("Unsupported platform:", platform)
		return
	}

	var response string
	fmt.Print("Do you want to proceed with the installation? (yes/no): ")
	fmt.Scanln(&response)

	if response != "yes" {
		fmt.Println("Operation aborted by the user.")
		return
	}

	switch platform {
	case "windows":
		fmt.Println("Installing Chocolatey...")
		err := utils.ExecuteCommand("Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))", "")
		if err != nil {
			fmt.Println("Error installing Chocolatey:", err)
			return
		}
	case "darwin", "linux":
		fmt.Println("Installing Homebrew...")
		err := utils.ExecuteCommand("/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"", "")
		if err != nil {
			fmt.Println("Error installing Homebrew:", err)
			return
		}
	default:
		fmt.Println("Unsupported platform:", platform)
	}
}

func ensureToolInstalled() bool {
	platform := runtime.GOOS
	switch platform {
	case "windows":
		err := utils.ExecuteCommand("choco -v", "", false)
		if err != nil {
			fmt.Println("Chocolatey is not installed. Please run 'pancake tool setup' first.")
			return false
		}
	case "darwin", "linux":
		err := utils.ExecuteCommand("brew -v", "", false)
		if err != nil {
			fmt.Println("Homebrew is not installed. Please run 'pancake tool setup' first.")
			return false
		}
	default:
		fmt.Println("Unsupported platform:", platform)
		return false
	}
	return true
}

func searchTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing search query")
		return
	}
	if !ensureToolInstalled() {
		return
	}
	platform := runtime.GOOS
	query := strings.Join(args, " ")
	var cmdStr string
	switch platform {
	case "windows":
		cmdStr = fmt.Sprintf("choco search %s", query)
	case "darwin", "linux":
		cmdStr = fmt.Sprintf("brew search %s", query)
	default:
		fmt.Println("Unsupported platform:", platform)
		return
	}
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error searching tool:", err)
	}
}

func installTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing tool name")
		return
	}
	if !ensureToolInstalled() {
		return
	}
	config := *utils.GetConfig()
	toolName := args[0]
	installPath := filepath.Join(config.Home, "tools")
	var cmdStr string
	switch runtime.GOOS {
	case "windows":
		cmdStr = fmt.Sprintf("choco install %s --cache-location %s", toolName, installPath)
	case "darwin", "linux":
		cmdStr = fmt.Sprintf("HOMEBREW_PREFIX=%s brew install %s", installPath, toolName)
	default:
		fmt.Println("Unsupported platform:", runtime.GOOS)
		return
	}
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error installing tool:", err)
	}
}

func uninstallTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing tool name")
		return
	}
	if !ensureToolInstalled() {
		return
	}
	config := *utils.GetConfig()
	toolName := args[0]
	uninstallPath := filepath.Join(config.Home, "tools")
	var cmdStr string
	switch runtime.GOOS {
	case "windows":
		cmdStr = fmt.Sprintf("choco uninstall %s --cache-location %s", toolName, uninstallPath)
	case "darwin", "linux":
		cmdStr = fmt.Sprintf("HOMEBREW_PREFIX=%s brew uninstall %s", uninstallPath, toolName)
	default:
		fmt.Println("Unsupported platform:", runtime.GOOS)
		return
	}
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error uninstalling tool:", err)
	}
}

func updateTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing tool name")
		return
	}
	if !ensureToolInstalled() {
		return
	}
	config := *utils.GetConfig()
	toolName := args[0]
	updatePath := filepath.Join(config.Home, "tools")
	var cmdStr string
	switch runtime.GOOS {
	case "windows":
		cmdStr = fmt.Sprintf("choco upgrade %s --cache-location %s", toolName, updatePath)
	case "darwin", "linux":
		cmdStr = fmt.Sprintf("HOMEBREW_PREFIX=%s brew upgrade %s", updatePath, toolName)
	default:
		fmt.Println("Unsupported platform:", runtime.GOOS)
		return
	}
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error updating tool:", err)
	}
}

func infoTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing tool name")
		return
	}
	if !ensureToolInstalled() {
		return
	}
	config := *utils.GetConfig()
	toolName := args[0]
	infoPath := filepath.Join(config.Home, "tools")
	var cmdStr string
	switch runtime.GOOS {
	case "windows":
		cmdStr = fmt.Sprintf("choco info %s", toolName)
	case "darwin", "linux":
		cmdStr = fmt.Sprintf("HOMEBREW_PREFIX=%s brew info %s", infoPath, toolName)
	default:
		fmt.Println("Unsupported platform:", runtime.GOOS)
		return
	}
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error fetching tool info:", err)
	}
}

func upgradeTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Missing tool name")
		return
	}
	if !ensureToolInstalled() {
		return
	}
	config := *utils.GetConfig()
	toolName := args[0]
	upgradePath := filepath.Join(config.Home, "tools")
	var cmdStr string
	switch runtime.GOOS {
	case "windows":
		cmdStr = fmt.Sprintf("choco upgrade %s", toolName)
	case "darwin", "linux":
		cmdStr = fmt.Sprintf("HOMEBREW_PREFIX=%s brew upgrade %s", upgradePath, toolName)
	default:
		fmt.Println("Unsupported platform:", runtime.GOOS)
		return
	}
	err := utils.ExecuteCommand(cmdStr, "")
	if err != nil {
		fmt.Println("Error upgrading tool:", err)
	}
}
