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
	"sync"
	"time"

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

	// conversationHistory holds the entire chat session.
	var conversationHistory []string
	// Add a system prompt to guide the AI's behavior.
	systemPrompt := `You are a helpful command-line assistant. Generate only the command and nothing else. If the request is not for a command, respond with helpful text.`
	conversationHistory = append(conversationHistory, "system: "+systemPrompt)

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

		// Add the latest user input to the history.
		conversationHistory = append(conversationHistory, "user: "+userInput)
		// Join the entire history to provide full context to the AI.
		fullPrompt := strings.Join(conversationHistory, "\n\n")

		response, err := getAIResponse(client, fullPrompt)
		if err != nil {
			log.Printf("❌ Error getting AI response: %v", err)
			// Remove the failed user prompt from history before trying again.
			conversationHistory = conversationHistory[:len(conversationHistory)-1]
			userInput = getUserFollowUp()
			continue
		}

		// Add the AI's response to the history for the next turn.
		conversationHistory = append(conversationHistory, "model: "+response)

		out, err := renderer.Render(response)
		if err != nil {
			log.Printf("❌ Error rendering markdown: %v", err)
			fmt.Println(response)
		} else {
			fmt.Print(out)
		}

		lang, code := extractCodeBlock(response)

		userInput = handleUserAction(code, lang)
		if userInput == "quit" {
			break
		}
	}
}

// getAIResponse sends a prompt to the AI and shows a loading animation.
func getAIResponse(client *utils.Client, prompt string) (string, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan bool)

	go func() {
		defer wg.Done()
		dots := ""
		for {
			select {
			case <-done:
				fmt.Print("\r" + strings.Repeat(" ", 80) + "\r") // Clear the line
				return
			default:
				if len(dots) >= 15 {
					dots = ""
				}
				dots += "."
				fmt.Printf("\rThinking%-15s", dots)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	response, err := client.GenerateContent(prompt)
	close(done)
	wg.Wait()
	return response, err
}

// extractCodeBlock finds and extracts the language and content of the first code block.
func extractCodeBlock(response string) (lang, code string) {
	re := regexp.MustCompile("(?s)```(.*?)\n(.*?)\n```")
	matches := re.FindStringSubmatch(response)

	if len(matches) >= 3 {
		language := strings.TrimSpace(matches[1])
		if language == "" || language == "sh" {
			language = "bash"
		}
		return language, strings.TrimSpace(matches[2])
	}
	return "text", strings.TrimSpace(response)
}

// handleUserAction presents options to the user and waits for their choice.
func handleUserAction(code, lang string) string {
	if err := keyboard.Open(); err != nil {
		log.Fatalf("❌ Could not open keyboard: %v", err)
	}
	defer keyboard.Close() // Ensure keyboard is always closed on exit

	fmt.Println(strings.Repeat("-", 70))
	if lang == "bash" || lang == "python" {
		fmt.Print("[Ctrl+R] Run | [Enter] Copy | [Ctrl+C] Quit | Type a follow-up > ")
	} else {
		fmt.Print("[Enter] Copy | [Ctrl+C] Quit | Type a follow-up > ")
	}

	const clearCurrentLine = "\r\x1b[K"
	const clearTwoLines = "\r\x1b[K\x1b[1A\x1b[K"

	var followUpInput []rune
	isFollowUpMode := false

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			log.Printf("❌ Error reading keypress: %v", err)
			return "quit"
		}

		switch {
		case key == keyboard.KeyCtrlC:
			fmt.Println("\nQuitting.")
			return "quit"

		// Handle backspace: only when in follow-up mode.
		case (key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2) && isFollowUpMode:
			if len(followUpInput) > 0 {
				followUpInput = followUpInput[:len(followUpInput)-1]
				fmt.Printf("%s> %s", clearCurrentLine, string(followUpInput))
			}

		// ADDED: Handle spacebar explicitly when in follow-up mode.
		case key == keyboard.KeySpace && isFollowUpMode:
			followUpInput = append(followUpInput, ' ')
			fmt.Printf(" ")

		// Handle Enter key.
		case key == keyboard.KeyEnter:
			if isFollowUpMode {
				// If typing a follow-up, Enter submits it.
				fmt.Println()
				return strings.TrimSpace(string(followUpInput))
			}
			// Otherwise, Enter copies the code.
			fmt.Print(clearTwoLines)
			if err := clipboard.WriteAll(code); err != nil {
				log.Printf("❌ Failed to copy to clipboard: %v", err)
			} else {
				fmt.Println("✅ Copied to clipboard!")
			}
			return "quit"

		// Handle Run command: only if not in follow-up mode.
		case key == keyboard.KeyCtrlR && !isFollowUpMode && (lang == "bash" || lang == "python"):
			fmt.Print(clearTwoLines)
			executeCommand(code, lang)
			return "quit"

		// Handle any other printable character.
		case char != 0:
			if !isFollowUpMode {
				isFollowUpMode = true
				// Clear the menu prompt and start the input prompt.
				fmt.Printf("%s> ", clearCurrentLine)
			}
			followUpInput = append(followUpInput, char)
			fmt.Printf("%c", char)
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
