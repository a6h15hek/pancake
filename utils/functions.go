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

// ExecuteCommand runs a shell command in a specified directory.
func ExecuteCommand(cmdStr, dir string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Dir = dir
	return cmd.Run()
}
