package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
	cmd := exec.Command("git", "clone", remoteURL, path)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
	} else {
		fmt.Printf("Cloned repository into %s\n", path)
	}
}

// PullChanges pulls the latest changes from a git repository.
func PullChanges(path string) {
	cmd := exec.Command("git", "-C", path, "pull")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error pulling latest changes: %v\n", err)
	} else {
		fmt.Printf("Updated repository in %s\n", path)
	}
}

// ConfirmAction prompts the user to confirm an action.
func ConfirmAction(action string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to %s for all projects? This may take some time. (yes/no): ", action)
	response, _ := reader.ReadString('\n')
	return strings.TrimSpace(response) == "yes"
}

// ExecuteCommand runs a shell command in a specified directory and prints the command and its logs.
func ExecuteCommand(cmdStr, dir string) error {
	fmt.Printf("> %s\n", cmdStr)
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = dir

	// Capture and print the command's output and error logs
	cmd.Stdout = newLoggingWriter()
	cmd.Stderr = newLoggingWriter()

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
