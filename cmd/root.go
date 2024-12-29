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
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pancake",
	Short: utils.Description,
	Long:  utils.LongDescription,
}

func version() { fmt.Println("Pancake " + utils.Version) }

func editConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		return
	}

	filePath := homeDir + "/pancake.yml"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("pancake.yml file does not exist. Do you want to create it? (yes/no)")

		input, _ := reader.ReadString('\n')
		if input == "yes\n" {
			rootFilePath := "pancake.yml" // Directly refer to pancake.yml in the root directory
			rootFile, err := os.Open(rootFilePath)
			if err != nil {
				fmt.Println("Error opening root pancake.yml file:", err)
				return
			}
			defer rootFile.Close()

			newFile, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating pancake.yml file in home directory:", err)
				return
			}
			defer newFile.Close()

			_, err = io.Copy(newFile, rootFile)
			if err != nil {
				fmt.Println("Error copying content to new pancake.yml file:", err)
				return
			}
			fmt.Println("pancake.yml file created in home directory.")
			fmt.Printf("Opening pancake.yml file at: %s\n", filePath)
			utils.OpenTextFileInDefaultEditor(filePath)
		} else {
			fmt.Println("Invalid Input. Command aborted.")
		}
	} else {
		fmt.Printf("Opening pancake.yml file at: %s\n", filePath)
		utils.OpenTextFileInDefaultEditor(filePath)
	}
}

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
