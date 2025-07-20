package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type GeminiConfig struct {
	APIKey      string  `yaml:"api_key"`
	Temperature float32 `yaml:"temperature"`
	URL         string  `yaml:"url"`
	Context     string  `yaml:"context"`
}

type Config struct {
	Home       string             `yaml:"home"`
	CodeEditor string             `yaml:"code_editor"`
	Tools      []string           `yaml:"tools"`
	Projects   map[string]Project `yaml:"projects"`
	Gemini     GeminiConfig       `yaml:"gemini"`
}

type Project struct {
	RemoteSSHURL string `yaml:"remote_ssh_url"`
	Type         string `yaml:"type,omitempty"`
	Port         string `yaml:"port,omitempty"`
	Run          string `yaml:"run,omitempty"`
	Build        string `yaml:"build,omitempty"`
}

func GetConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Error getting user home directory:", err)
		os.Exit(1)
	}
	configPath := filepath.Join(homeDir, "pancake.yml")

	file, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("❌ Error reading config file:", err)
		os.Exit(1)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		fmt.Println("❌ Error unmarshaling config file:", err)
		os.Exit(1)
	}

	// Replace $HOME with actual home directory
	config.Home = strings.Replace(config.Home, "$HOME", homeDir, -1)

	return &config
}

func UpdateConfig(config *Config) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Error getting user home directory:", err)
		return
	}
	configPath := filepath.Join(homeDir, "pancake.yml")

	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("❌ Error marshaling config data:", err)
		return
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		fmt.Println("❌ Error writing config file:", err)
		return
	}
	fmt.Println("✅ Config file updated successfully.")
}
