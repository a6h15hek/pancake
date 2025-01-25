package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

/* Common Functions used across the project.
* This file is part of the 'Pancake' project and contains common utility functions
* that are designed to be reusable across various parts of the project.
**/

// OpenTextFileInDefaultEditor opens a text file in the default editor based on the OS.
func OpenTextFileInDefaultEditor(filePath string) {
	switch os := runtime.GOOS; os {
	case "windows":
		exec.Command("notepad.exe", filePath).Start()
	case "darwin":
		exec.Command("open", filePath).Start() // MacOS
	default: // Assume Linux/Unix
		exec.Command("xdg-open", filePath).Start()
	}
}

// CheckExists checks if a given path exists.
func CheckExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CloneRepository clones a git repository to a specified path.
func CloneRepository(path, remoteURL string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Error removing existing folder: %v\n", err)
		return
	}
	cmdStr := fmt.Sprintf("git clone %s %s", remoteURL, path)
	err = ExecuteCommand(cmdStr, ".", true)
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
	} else {
		fmt.Printf("Cloned repository into %s\n", path)
	}
}

// PullChanges pulls the latest changes from a git repository.
func PullChanges(path string) {
	cmdStr := "git pull"
	err := ExecuteCommand(cmdStr, path, true)
	if err != nil {
		fmt.Printf("Error pulling latest changes: %v\n", err)
	} else {
		fmt.Printf("Updated repository in %s\n", path)
	}
}

// ConfirmAction prompts the user to confirm an action.
func ConfirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", message)
	response, _ := reader.ReadString('\n')
	confirm := strings.TrimSpace(response) == "yes" || strings.TrimSpace(response) == "y"
	if !confirm {
		fmt.Println("Action aborted.")
		return false
	}
	return true
}

// ExecuteCommand runs a shell command in a specified directory and prints the command and its logs.
func ExecuteCommand(cmdStr, dir string, isLogging ...bool) error {
	logging := true
	if len(isLogging) > 0 {
		logging = isLogging[0]
	}

	if logging {
		fmt.Printf("> %s\n", cmdStr)
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = dir

	if logging {
		// Capture and print the command's output and error logs
		cmd.Stdout = newLoggingWriter()
		cmd.Stderr = newLoggingWriter()
	}

	return cmd.Run()
}

// NewLoggingWriter creates a writer that prints the logs.
func newLoggingWriter() *loggingWriter {
	return &loggingWriter{}
}

type loggingWriter struct{}

func (lw *loggingWriter) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))
	return len(p), nil
}

// ExecuteCommandInNewTerminal runs a shell command in a specified directory in a new terminal window/tab and prints the command and its logs.
func ExecuteCommandInNewTerminal(cmdStr, dir, projectName string, projectPIDs *map[string]int) error {
	var command *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd /d %s && %s", dir, cmdStr))
	case "darwin":
		command = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "cd %s && %s"`, dir, cmdStr))
	default:
		command = exec.Command("gnome-terminal", "--", "sh", "-c", fmt.Sprintf("cd %s && %s", dir, cmdStr))
	}

	err := command.Start()
	if err != nil {
		return err
	}

	(*projectPIDs)[projectName] = command.Process.Pid
	return nil
}

// printTable prints a table with the given data.
func PrintTable(data [][]string) {
	// Find the maximum width for each column
	colWidths := make([]int, len(data[0]))
	for _, row := range data {
		for colIndex, col := range row {
			if len(col) > colWidths[colIndex] {
				colWidths[colIndex] = len(col)
			}
		}
	}

	// Print the table with separators and proper alignment
	for _, row := range data {
		for colIndex, col := range row {
			fmt.Printf("| %-*s ", colWidths[colIndex], col)
		}
		fmt.Println("|")
	}

	// Print the table border
	for colIndex := range data[0] {
		fmt.Printf("| %s ", strings.Repeat("-", colWidths[colIndex]))
	}
	fmt.Println("|")
}

// SaveProjectPIDs saves the project PIDs to a specified file.
func SaveProjectPIDs(fileLocation string, projectPIDs map[string]int) {
	data, err := json.Marshal(projectPIDs)
	if err != nil {
		fmt.Printf("❌ Error saving project PIDs: %v\n", err)
		return
	}

	err = os.WriteFile(filepath.Join(fileLocation, "pids.json"), data, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing project PIDs file: %v\n", err)
	}
}

// LoadProjectPIDs loads the project PIDs from a specified file.
func LoadProjectPIDs(fileLocation string, projectPIDs *map[string]int) {
	data, err := os.ReadFile(filepath.Join(fileLocation, "pids.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("❌ Error reading project PIDs file: %v\n", err)
		}
		return
	}

	err = json.Unmarshal(data, projectPIDs)
	if err != nil {
		fmt.Printf("❌ Error unmarshaling project PIDs: %v\n", err)
	}
}

func SetupChocolatey() bool {
	if ConfirmAction("Do you want to proceed with the installation? (yes/no):") {
		fmt.Println("Installing Chocolatey...")
		err := ExecuteCommand("Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))", "")
		if err != nil {
			fmt.Println("Error installing Chocolatey:", err)
		} else {
			return true
		}
	}
	return false
}

func SetupHomebrew() bool {
	if ConfirmAction("Do you want to proceed with the installation? (yes/no):") {
		fmt.Println("Installing Homebrew...")
		err := ExecuteCommand("/bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"", "")
		if err != nil {
			fmt.Println("Error installing Homebrew:", err)
		} else {
			return true
		}
	}
	return false
}

func EnsureToolInstalled() bool {
	platform := runtime.GOOS
	var cmdStr string
	switch platform {
	case "windows":
		cmdStr = "choco -v"
	case "darwin", "linux":
		cmdStr = "brew -v"
	default:
		fmt.Println("Unsupported platform:", platform)
		return false
	}
	err := ExecuteCommand(cmdStr, "", false)
	if err != nil {
		fmt.Printf("%s is not installed. Please run 'pancake tool setup' first.\n", GetPackageManager())
		return false
	}
	return true
}

func GetPackageManager() string {
	switch runtime.GOOS {
	case "windows":
		return "choco"
	case "darwin", "linux":
		return "brew"
	default:
		fmt.Println("Unsupported platform:", runtime.GOOS)
		return ""
	}
}
