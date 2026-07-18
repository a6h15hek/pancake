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

func OpenTextFileInDefaultEditor(filePath string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		command = exec.Command("open", filePath)
	default:
		command = exec.Command("xdg-open", filePath)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("could not open %s in default editor: %w", filePath, err)
	}
	return nil
}

func CheckExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CloneRepository(path, remoteURL string) error {
	if CheckExists(filepath.Join(path, ".git")) {
		return PullChanges(path)
	}
	if CheckExists(path) {
		return fmt.Errorf("target path %s already exists and is not a git repository; move or remove it before syncing", path)
	}
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("could not create parent directory %s: %w", parentDir, err)
	}
	if err := ExecuteCommand(fmt.Sprintf("git clone %s %s", remoteURL, path), ".", true); err != nil {
		return fmt.Errorf("git clone failed for %s: %w. Ensure your SSH key is set up (ssh -T git@github.com) or switch remote_ssh_url to an https URL in pancake.yml", remoteURL, err)
	}
	return nil
}

func PullChanges(path string) error {
	if err := ExecuteCommand("git pull", path, true); err != nil {
		return fmt.Errorf("git pull failed in %s: %w", path, err)
	}
	return nil
}

func ConfirmAction(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s", message)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	trimmed := strings.TrimSpace(strings.ToLower(response))
	if trimmed != "yes" && trimmed != "y" {
		fmt.Println("Action aborted.")
		return false
	}
	return true
}

func buildShellCommand(cmdStr string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/c", cmdStr)
	}
	return exec.Command("sh", "-c", cmdStr)
}

func ExecuteCommand(cmdStr, dir string, isLogging ...bool) error {
	logging := true
	if len(isLogging) > 0 {
		logging = isLogging[0]
	}
	if logging {
		fmt.Printf("%s > %s\n", dir, cmdStr)
	}
	command := buildShellCommand(cmdStr)
	command.Dir = dir
	if logging {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}
	return command.Run()
}

func ExecuteCommandInNewTerminal(cmdStr, dir, projectName string, projectPIDs *map[string]int) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd /d %s && %s", dir, cmdStr))
	case "darwin":
		command = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "cd %s && %s"`, dir, cmdStr))
	default:
		terminal, ok := detectLinuxTerminal()
		if !ok {
			return fmt.Errorf("no supported terminal emulator found (tried gnome-terminal, konsole, xfce4-terminal, x-terminal-emulator, xterm). Open %s and run '%s' manually", dir, cmdStr)
		}
		command = exec.Command(terminal, "--", "sh", "-c", fmt.Sprintf("cd %s && %s; exec sh", dir, cmdStr))
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("could not launch terminal for %s: %w", projectName, err)
	}
	(*projectPIDs)[projectName] = command.Process.Pid
	return nil
}

func detectLinuxTerminal() (string, bool) {
	candidates := []string{"gnome-terminal", "konsole", "xfce4-terminal", "x-terminal-emulator", "xterm"}
	for _, candidate := range candidates {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate, true
		}
	}
	return "", false
}

func PrintTable(data [][]string) {
	if len(data) == 0 {
		return
	}
	colWidths := make([]int, len(data[0]))
	for _, row := range data {
		for colIndex, col := range row {
			if len(col) > colWidths[colIndex] {
				colWidths[colIndex] = len(col)
			}
		}
	}
	for _, row := range data {
		for colIndex, col := range row {
			fmt.Printf("| %-*s ", colWidths[colIndex], col)
		}
		fmt.Println("|")
	}
	for colIndex := range data[0] {
		fmt.Printf("| %s ", strings.Repeat("-", colWidths[colIndex]))
	}
	fmt.Println("|")
}

func SaveProjectPIDs(fileLocation string, projectPIDs map[string]int) error {
	data, err := json.Marshal(projectPIDs)
	if err != nil {
		return fmt.Errorf("could not encode project pids: %w", err)
	}
	if err := os.MkdirAll(fileLocation, 0755); err != nil {
		return fmt.Errorf("could not create %s: %w", fileLocation, err)
	}
	return os.WriteFile(filepath.Join(fileLocation, "pids.json"), data, 0644)
}

func LoadProjectPIDs(fileLocation string, projectPIDs *map[string]int) error {
	data, err := os.ReadFile(filepath.Join(fileLocation, "pids.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("could not read project pids file: %w", err)
	}
	return json.Unmarshal(data, projectPIDs)
}

func SetupChocolatey() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("chocolatey is only supported on windows")
	}
	if !ConfirmAction("Do you want to install Chocolatey? (yes/no):") {
		return fmt.Errorf("chocolatey installation cancelled")
	}
	fmt.Println("Installing Chocolatey...")
	installScript := "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))"
	command := exec.Command("powershell", "-NoProfile", "-Command", installScript)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("chocolatey installation failed: %w", err)
	}
	return nil
}

func SetupHomebrew() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("homebrew is not supported on windows; pancake uses chocolatey there")
	}
	if !ConfirmAction("Do you want to install Homebrew? (yes/no):") {
		return fmt.Errorf("homebrew installation cancelled")
	}
	fmt.Println("Installing Homebrew...")
	if err := ExecuteCommand(`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`, ""); err != nil {
		return fmt.Errorf("homebrew installation failed: %w", err)
	}
	return nil
}

func EnsureToolInstalled() (string, error) {
	packageManager := GetPackageManager()
	if packageManager == "" {
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	versionFlag := "-v"
	if runtime.GOOS == "windows" {
		versionFlag = "--version"
	}
	checkCommand := buildShellCommand(fmt.Sprintf("%s %s", packageManager, versionFlag))
	checkCommand.Stdout = nil
	checkCommand.Stderr = nil
	if err := checkCommand.Run(); err != nil {
		return "", fmt.Errorf("%s is not installed or not on PATH. Run 'pancake tool setup' first, or install %s manually", packageManager, packageManager)
	}
	return packageManager, nil
}

func GetPackageManager() string {
	switch runtime.GOOS {
	case "windows":
		return "choco"
	case "darwin", "linux":
		return "brew"
	default:
		return ""
	}
}
