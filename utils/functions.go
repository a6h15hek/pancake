package utils

import (
	"os/exec"
	"runtime"
)

/* Common Functions used across the project.
* This file is part of the 'Pancake' project and contains common utility functions
* that are designed to be reusable across various parts of the project.
**/

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
