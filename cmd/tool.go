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
    "github.com/a6h15hek/pancake/utils"
    "github.com/spf13/cobra"
)

// toolCmd represents the tool command
var toolCmd = &cobra.Command{
    Use:   "tool",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(utils.NotImplemented)
    },
}

func installTool(args []string) {
    fmt.Println(utils.NotImplemented)
}

func uninstallTool(args []string) {
    fmt.Println(utils.NotImplemented)
}

func listTools() {
    fmt.Println(utils.NotImplemented)
}

func updateTool(args []string) {
    fmt.Println(utils.NotImplemented)
}

func init() {
    rootCmd.AddCommand(toolCmd)

    toolCmd.AddCommand(
        &cobra.Command{Use: "install", Run: func(cmd *cobra.Command, args []string) { installTool(args) }},
        &cobra.Command{Use: "uninstall", Run: func(cmd *cobra.Command, args []string) { uninstallTool(args) }},
        &cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listTools() }},
        &cobra.Command{Use: "update", Run: func(cmd *cobra.Command, args []string) { updateTool(args) }},
    )
}
