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
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/a6h15hek/pancake/utils"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/glamour"
	"github.com/eiannone/keyboard"
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

// aiCommand starts an interactive AI session.
func aiCommand(args []string) {
	// Ensure keyboard is closed on exit
	defer func() {
		_ = keyboard.Close()
	}()

	loadConfig()

	client, err := utils.NewGeminiClient(config.Gemini)
	if err != nil {
		log.Fatalf("❌ Failed to create AI client: %v", err)
	}

	userInput := strings.Join(args, " ")
	if strings.TrimSpace(userInput) == "" {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		userInput, err = reader.ReadString('\n')
		if err != nil {
			log.Fatalf("❌ Failed to read user input: %v", err)
		}
		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			return
		}
	}

	// Create a glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create markdown renderer: %v", err)
	}

	// Conversation loop
	for {
		if strings.ToLower(userInput) == "exit" || strings.ToLower(userInput) == "quit" {
			break
		}

		response, err := client.GenerateContent(userInput)
		if err != nil {
			log.Printf("❌ Error getting AI response: %v", err)
			userInput = getUserFollowUp()
			continue
		}

		// Render the markdown response
		out, err := renderer.Render(response)
		if err != nil {
			log.Printf("❌ Error rendering markdown: %v", err)
			// Print the raw response as a fallback
			fmt.Println(response)
		} else {
			fmt.Print(out)
		}

		lang, code := extractCodeBlock(response)

		// handleUserAction will return the next prompt or "quit" to exit.
		userInput = handleUserAction(code, lang)
		if userInput == "quit" {
			break
		}
	}
}

// extractCodeBlock finds and extracts the language and content of the first code block.
func extractCodeBlock(response string) (lang, code string) {
	// Regex to find ```lang\ncode\n```. It's more flexible and captures any language.
	re := regexp.MustCompile("(?s)```(.*?)\n(.*?)\n```")
	matches := re.FindStringSubmatch(response)

	if len(matches) >= 3 {
		language := strings.TrimSpace(matches[1])
		if language == "" || language == "sh" {
			language = "bash" // Default to bash for shell scripts
		}
		return language, strings.TrimSpace(matches[2])
	}

	// If no code block is found, return the whole response as plain text.
	return "text", strings.TrimSpace(response)
}

// handleUserAction presents options to the user and waits for their choice.
func handleUserAction(code, lang string) string {
	if err := keyboard.Open(); err != nil {
		// Use log.Fatalf to print the error and exit with a status code of 1.
		log.Fatalf("❌ Could not open keyboard: %v", err)
	}

	fmt.Println(strings.Repeat("-", 70))
	// Show different prompts based on whether the content is executable code or text
	if lang == "bash" || lang == "python" {
		fmt.Print("[Ctrl+R] Run Command | [Enter] Copy | [Ctrl+C] Quit | Type a follow-up > ")
	} else {
		fmt.Print("[Enter] Copy | [Ctrl+C] Quit | Type a follow-up > ")
	}

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Printf("❌ Error reading keypress: %v", err)
			return "quit" // Exit on error
		}

		// On Ctrl+C, exit gracefully.
		if key == keyboard.KeyCtrlC {
			return "quit"
		}

		// Handle actions for executable code
		if lang == "bash" || lang == "python" {
			if key == keyboard.KeyCtrlR {
				fmt.Println()
				executeCommand(code, lang)
				return "quit" // Exit after executing the command.
			}
		}

		// Handle copy action for both code and plain text.
		if key == keyboard.KeyEnter {
			clipboard.WriteAll(code)
			fmt.Println("\nCopied to clipboard!")
			return "quit" // Exit after copying.
		}

		// If a printable character is typed, start a follow-up prompt
		if char != 0 {
			// Clear the prompt line and start the new user input
			fmt.Printf("\r%s\r", strings.Repeat(" ", 80))
			fmt.Printf("> %c", char)
			reader := bufio.NewReader(os.Stdin)
			followUp, _ := reader.ReadString('\n')
			// Close the keyboard here so the main loop can potentially reopen it
			_ = keyboard.Close()
			return string(char) + strings.TrimSpace(followUp)
		}
	}
}

// executeCommand runs the provided code string using the appropriate interpreter.
func executeCommand(code, lang string) {
	var cmd *exec.Cmd
	switch lang {
	case "python":
		cmd = exec.Command("python", "-c", code)
	case "bash":
		cmd = exec.Command("bash", "-c", code)
	default:
		fmt.Printf("Unsupported language for execution: %s\n", lang)
		return
	}

	// Stream the command's output to the application's stdout and stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n❌ Error during execution: %v\n", err)
	}
}

// getUserFollowUp prompts the user for their next input.
func getUserFollowUp() string {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("> ")
	reader := bufio.NewReader(os.Stdin)
	userInput, _ := reader.ReadString('\n')
	return strings.TrimSpace(userInput)
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
