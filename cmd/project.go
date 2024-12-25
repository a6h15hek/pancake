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
    constants "github.com/a6h15hek/pancake/utils"
    "github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
    Use:   "project",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(constants.NotImplemented)
    },
}

func addProject(name string) {
    fmt.Println(constants.NotImplemented)
}

func removeProject(name string) {
    fmt.Println(constants.NotImplemented)
}

func listProjects() {
    fmt.Println(constants.NotImplemented)
}

func syncProjects(name string) {
    fmt.Println(constants.NotImplemented)
}

func openProject(name string) {
    fmt.Println(constants.NotImplemented)
}

func buildProject(name string) {
    fmt.Println(constants.NotImplemented)
}

func startProject(name string) {
    fmt.Println(constants.NotImplemented)
}

func stopProject(name string) {
    fmt.Println(constants.NotImplemented)
}

func monitorProject() {
    fmt.Println(constants.NotImplemented)
}

func init() {
    rootCmd.AddCommand(projectCmd)

    projectCmd.AddCommand(
        &cobra.Command{Use: "add", Run: func(cmd *cobra.Command, args []string) { addProject(args[0]) }},
        &cobra.Command{Use: "remove", Run: func(cmd *cobra.Command, args []string) { removeProject(args[0]) }},
        &cobra.Command{Use: "list", Run: func(cmd *cobra.Command, args []string) { listProjects() }},
        &cobra.Command{Use: "sync", Run: func(cmd *cobra.Command, args []string) { syncProjects(args[0]) }},
        &cobra.Command{Use: "open", Run: func(cmd *cobra.Command, args []string) { openProject(args[0]) }},
        &cobra.Command{Use: "build", Run: func(cmd *cobra.Command, args []string) { buildProject(args[0]) }},
        &cobra.Command{Use: "start", Run: func(cmd *cobra.Command, args []string) { startProject(args[0]) }},
        &cobra.Command{Use: "stop", Run: func(cmd *cobra.Command, args []string) { stopProject(args[0]) }},
        &cobra.Command{Use: "monitor", Run: func(cmd *cobra.Command, args []string) { monitorProject() }},
    )
}
