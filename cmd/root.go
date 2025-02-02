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

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

var config utils.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pancake",
	Short: utils.Description,
	Long:  utils.LongDescription,
}

func version() {
	fmt.Println("Pancake " + utils.Version)
}

func editConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Error finding home directory:", err)
		os.Exit(1)
	}

	filePath := homeDir + "/pancake.yml"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("📌 Pancake configuration file does not exist. Do 'pancake init' for initial setup.")
		os.Exit(1)
	} else {
		fmt.Printf("✅ Opening pancake.yml file at: %s\n", filePath)
		utils.OpenTextFileInDefaultEditor(filePath)
	}
}

func initCommand() {
	fmt.Println("🔄 Setup of pancake started...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Error finding home directory:", err)
		os.Exit(1)
	}

	filePath := homeDir + "/pancake.yml"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		pancakeYAMLContent := utils.DefaultYMLContent

		newFile, err := os.Create(filePath)
		if err != nil {
			fmt.Println("❌ Error creating pancake.yml file in home directory:", err)
			os.Exit(1)
		}
		defer newFile.Close()

		_, err = newFile.Write([]byte(pancakeYAMLContent))
		if err != nil {
			fmt.Println("❌ Error writing content to new pancake.yml file:", err)
			os.Exit(1)
		}
		fmt.Println("✅ pancake.yml file created in home directory.")
		fmt.Println("✅ Setup Completed.")
	} else {
		fmt.Println("pancake.yml already exists. Using it for configuration:")
		config := *utils.GetConfig()
		fmt.Printf("%+v\n", config)
	}

	// Run the pancake tool setup command
	err = utils.ExecuteCommand("echo yes | pancake tool setup", homeDir)
	if err != nil {
		fmt.Println("❌ Error executing pancake tool setup command:", err)
		os.Exit(1)
	}
	fmt.Println("✅ Pancake tool setup command executed successfully.")
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
			Use:     "version",
			Aliases: []string{"v"},
			Run:     func(cmd *cobra.Command, args []string) { version() },
		},
		&cobra.Command{
			Use:     "edit config",
			Aliases: []string{"ec"},
			Run:     func(cmd *cobra.Command, args []string) { editConfig() },
		},
		&cobra.Command{
			Use:     "init",
			Aliases: []string{"ec"},
			Run:     func(cmd *cobra.Command, args []string) { initCommand() },
		},
	)
}
