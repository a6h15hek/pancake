package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Home       string             `yaml:"home"`
	CodeEditor string             `yaml:"code_editor"`
	Tools      map[string]string  `yaml:"tools"`
	Projects   map[string]Project `yaml:"projects"`
}

type Project struct {
	RemoteSSHURL string `yaml:"remote_ssh_url"`
	Type         string `yaml:"type,omitempty"`
	Port         string `yaml:"port,omitempty"`
	Start        string `yaml:"start,omitempty"`
	Build        string `yaml:"build,omitempty"`
}

func GetConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
	configPath := filepath.Join(homeDir, "pancake.yml")

	file, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("Error unmarshaling config file:", err)
		os.Exit(1)
	}

	// Replace $HOME with actual home directory
	config.Home = strings.Replace(config.Home, "$HOME", homeDir, -1)

	return &config
}

// func main() {
//     config := utils.GetConfig()

//     fmt.Printf("Home Directory: %s\n", config.Home)
//     fmt.Printf("Code Editor: %s\n", config.CodeEditor)
//     fmt.Println("Tools:")
//     for tool, version := range config.Tools {
//         fmt.Printf("- %s: %s\n", tool, version)
//     }
//     fmt.Println("Projects:")
//     for projectName, project := range config.Projects {
//         fmt.Printf("- %s\n", projectName)
//         fmt.Printf("  Remote SSH URL: %s\n", project.RemoteSSHURL)
//         if project.Type != "" {
//             fmt.Printf("  Type: %s\n", project.Type)
//         }
//         if project.Port != 0 {
//             fmt.Printf("  Port: %d\n", project.Port)
//         }
//         if project.Start != "" {
//             fmt.Printf("  Start: %s\n", project.Start)
//         }
//         if project.Build != "" {
//             fmt.Printf("  Build: %s\n", project.Build)
//         }
//     }
// }
