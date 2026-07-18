package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
)

type GeminiConfig struct {
	APIKey      string  `yaml:"api_key"`
	Temperature float32 `yaml:"temperature"`
	URL         string  `yaml:"url"`
	Context     string  `yaml:"context"`
}

type ChatGPTConfig struct {
	APIKey      string  `yaml:"api_key"`
	Temperature float32 `yaml:"temperature"`
	URL         string  `yaml:"url"`
	Model       string  `yaml:"model"`
	Context     string  `yaml:"context"`
}

type Config struct {
	Home       string             `yaml:"home"`
	CodeEditor string             `yaml:"code_editor"`
	DefaultAI  string             `yaml:"default_ai"`
	Tools      []string           `yaml:"tools"`
	Projects   map[string]Project `yaml:"projects"`
	Gemini     GeminiConfig       `yaml:"gemini"`
	ChatGPT    ChatGPTConfig      `yaml:"chatgpt"`
}

type Project struct {
	RemoteSSHURL string `yaml:"remote_ssh_url"`
	Type         string `yaml:"type,omitempty"`
	Port         string `yaml:"port,omitempty"`
	Run          string `yaml:"run,omitempty"`
	Build        string `yaml:"build,omitempty"`
}

func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %w", err)
	}
	return filepath.Join(homeDir, ConfigFileName), nil
}

func ExpandHomePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	expanded := os.ExpandEnv(path)
	if runtime.GOOS == "windows" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not resolve user home directory: %w", err)
		}
		expanded = replaceCaseInsensitive(expanded, "%userprofile%", homeDir)
	}
	return expanded, nil
}

func replaceCaseInsensitive(input, target, replacement string) string {
	lower := strings.ToLower(input)
	targetLower := strings.ToLower(target)
	var builder strings.Builder
	for i := 0; i < len(input); {
		if strings.HasPrefix(lower[i:], targetLower) {
			builder.WriteString(replacement)
			i += len(targetLower)
			continue
		}
		builder.WriteByte(input[i])
		i++
	}
	return builder.String()
}

func GetConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(fmt.Sprintf(ConfigErrNotFound, configPath))
		}
		return nil, errors.New(fmt.Sprintf(ConfigErrReadFailed, configPath, configPath))
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, errors.New(fmt.Sprintf(ConfigErrParseFailed, err.Error(), configPath))
	}

	expanded, err := ExpandHomePath(config.Home)
	if err != nil {
		return nil, err
	}
	config.Home = expanded

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ValidateConfig(config *Config) error {
	var issues []string

	if strings.TrimSpace(config.Home) == "" {
		issues = append(issues, fmt.Sprintf(ConfigErrHomeEmpty))
	} else if !filepath.IsAbs(config.Home) {
		issues = append(issues, fmt.Sprintf(ConfigErrHomeRelative, config.Home))
	}

	switch config.DefaultAI {
	case "", "gemini", "chatgpt":
	default:
		issues = append(issues, fmt.Sprintf(ConfigErrDefaultAIInvalid, config.DefaultAI))
	}

	for projectName := range config.Projects {
		if strings.ContainsAny(projectName, `/\`) {
			issues = append(issues, fmt.Sprintf(ConfigErrProjectNameInvalid, projectName))
			continue
		}
		project := config.Projects[projectName]
		if strings.TrimSpace(project.RemoteSSHURL) == "" {
			issues = append(issues, fmt.Sprintf(ConfigErrProjectRemoteMissing, projectName, projectName))
		}
	}

	if len(issues) == 0 {
		return nil
	}
	return errors.New(strings.Join(append([]string{"pancake.yml has issues:"}, issues...), "\n - "))
}

func UpdateConfig(config *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not encode pancake.yml: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("could not write pancake.yml at %s: %w", configPath, err)
	}
	return nil
}
