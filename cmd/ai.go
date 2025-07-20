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
	"strings"

	"github.com/a6h15hek/pancake/utils"
	"github.com/spf13/cobra"
)

/*
- This will have implementation of "pancake ai <user_description_of_command>".
It will utilize the AI models to understand the user's natural language input
in <user_description_of_command>, interpret it, create a corresponding command,
and execute it.
*/

var aiCmd = &cobra.Command{
	Use:   "ai [description]",
	Short: "Executes a command from a natural language description.",
	Long: `This command utilizes AI models to understand a natural language input,
interpret it, create a corresponding command, and execute it.`,
	Run: func(cmd *cobra.Command, args []string) {
		aiCommand(args)
	},
}

func aiCommand(args []string) {
	loadConfig()
	userInput := strings.Join(args, " ")
	if strings.TrimSpace(userInput) == "" {
		fmt.Println("❌ Error: Input string is empty. Please provide a description for the AI command.")
		return
	}

	fmt.Println("🧠 Thinking...")

	client, err := utils.NewAIClient(config.Gemini)
	if err != nil {
		fmt.Println(err)
		return
	}

	generatedCommand, err := client.GenerateContent(userInput)
	if err != nil {
		fmt.Println(err)
		return
	}

	generatedCommand = strings.TrimSpace(generatedCommand)
	fmt.Printf("🤖 Generated command: %s\n", generatedCommand)

	if utils.ConfirmAction("Do you want to execute this command? (yes/no)") {
		fmt.Println("🚀 Executing...")
		err := utils.ExecuteCommand(generatedCommand, ".")
		if err != nil {
			fmt.Printf("❌ Error executing command: %v\n", err)
		}
	}
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
