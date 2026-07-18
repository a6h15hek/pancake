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
		&cobra.Command{Use: "install", Aliases: []string{"i"}, Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "install") }},
		&cobra.Command{Use: "uninstall", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "uninstall") }},
		&cobra.Command{Use: "list", Aliases: []string{"l"}, Run: func(cmd *cobra.Command, args []string) { listTools() }},
		&cobra.Command{Use: "update", Run: func(cmd *cobra.Command, args []string) { updateTools() }},
		&cobra.Command{Use: "search", Aliases: []string{"s"}, Run: func(cmd *cobra.Command, args []string) { searchTool(args) }},
		&cobra.Command{Use: "setup", Run: func(cmd *cobra.Command, args []string) { _ = setupTools() }},
		&cobra.Command{Use: "info", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "info") }},
		&cobra.Command{Use: "upgrade", Run: func(cmd *cobra.Command, args []string) { handleToolCommand(args, "upgrade") }},
	)
}

func listTools() {
	fmt.Println("Loading tools")
	cfg, err := utils.GetConfig()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println(utils.ConfigHintEditConfig)
		return
	}
	if len(cfg.Tools) == 0 {
		fmt.Println("No tools listed in pancake.yml. Run 'pancake tool install <name>' to add one.")
		return
	}
	for _, toolName := range cfg.Tools {
		fmt.Printf("- %s\n", toolName)
	}
}

func setupTools() error {
	platform := runtime.GOOS
	switch platform {
	case "windows":
		fmt.Println("Pancake uses Chocolatey internally on Windows.")
		if err := utils.SetupChocolatey(); err != nil {
			return err
		}
	case "darwin", "linux":
		fmt.Println("Pancake uses Homebrew internally on macOS/Linux.")
		if err := utils.SetupHomebrew(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	cfg, err := utils.GetConfig()
	if err != nil {
		return err
	}
	for _, toolName := range cfg.Tools {
		handleToolCommand([]string{toolName}, "install")
	}
	return nil
}

func searchTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: missing search query. Usage: pancake tool search <query>")
		return
	}
	packageManager, err := utils.EnsureToolInstalled()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	query := strings.Join(args, " ")
	if err := utils.ExecuteCommand(fmt.Sprintf("%s search %s", packageManager, query), ""); err != nil {
		fmt.Println("Error searching tool:", err)
	}
}

func handleToolCommand(args []string, action string) {
	if len(args) == 0 {
		fmt.Printf("Error: missing tool name for %s\n", action)
		return
	}
	packageManager, err := utils.EnsureToolInstalled()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	toolName := args[0]
	cfg, err := utils.GetConfig()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println(utils.ConfigHintEditConfig)
		return
	}

	switch action {
	case "install":
		for _, existing := range cfg.Tools {
			if existing == toolName {
				fmt.Printf("Tool '%s' already tracked in pancake.yml. Running upgrade instead.\n", toolName)
				action = "upgrade"
				break
			}
		}
	case "uninstall":
		found := false
		for _, existing := range cfg.Tools {
			if existing == toolName {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Tool '%s' is not tracked via pancake. Cannot uninstall.\n", toolName)
			return
		}
	}

	if err := utils.ExecuteCommand(fmt.Sprintf("%s %s %s", packageManager, action, toolName), ""); err != nil {
		fmt.Printf("Error during %s of tool '%s': %v\n", action, toolName, err)
		return
	}

	switch action {
	case "install":
		cfg.Tools = append(cfg.Tools, toolName)
		if err := utils.UpdateConfig(cfg); err != nil {
			fmt.Println("Error updating pancake.yml:", err)
		} else {
			fmt.Println("Config file updated successfully.")
		}
	case "uninstall":
		for i, existing := range cfg.Tools {
			if existing == toolName {
				cfg.Tools = append(cfg.Tools[:i], cfg.Tools[i+1:]...)
				if err := utils.UpdateConfig(cfg); err != nil {
					fmt.Println("Error updating pancake.yml:", err)
				} else {
					fmt.Println("Config file updated successfully.")
				}
				break
			}
		}
	}
}

func updateTools() {
	packageManager, err := utils.EnsureToolInstalled()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err := utils.ExecuteCommand(fmt.Sprintf("%s update", packageManager), ""); err != nil {
		fmt.Println("Error updating tools:", err)
	}
}
