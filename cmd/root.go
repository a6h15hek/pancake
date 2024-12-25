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
    "os"
    constants "github.com/a6h15hek/pancake/utils"
    "github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "pancake",
    Short: constants.Description,
    Long:  constants.LongDescription,
}

func version() { fmt.Println("Pancake " + constants.Version) }
func editConfig() { fmt.Println(constants.NotImplemented) }

func Execute() {
    err := rootCmd.Execute()
    if err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
    rootCmd.AddCommand(
        &cobra.Command{
            Use: "version", 
            Run: func(cmd *cobra.Command, args []string) { version() },
        },
        &cobra.Command{
            Use: "edit-config", 
            Run: func(cmd *cobra.Command, args []string) { editConfig() },
        },
    )
}