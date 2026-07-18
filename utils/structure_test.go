package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	home := t.TempDir()
	configPath := filepath.Join(home, ConfigFileName)
	if err := os.WriteFile(configPath, []byte(contents), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	}
	return configPath
}

const validConfig = `home: $HOME/pancake
code_editor: echo
default_ai: gemini
tools:
  - tree
projects:
  demo:
    remote_ssh_url: git@github.com:org/repo.git
    run: echo running
    build: echo building
gemini:
  api_key: "test-key"
  url: "https://example.com"
chatgpt:
  api_key: ""
`

func TestGetConfig_NotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", t.TempDir())
	}
	_, err := GetConfig()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
	if !strings.Contains(err.Error(), "pancake.yml was not found") {
		t.Fatalf("expected not-found message, got: %v", err)
	}
}

func TestGetConfig_Valid(t *testing.T) {
	writeConfig(t, validConfig)
	cfg, err := GetConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Home == "" {
		t.Fatal("Home should not be empty")
	}
	if !strings.HasSuffix(cfg.Home, "/pancake") {
		t.Fatalf("Home should be expanded to end with /pancake, got %s", cfg.Home)
	}
	if cfg.DefaultAI != "gemini" {
		t.Fatalf("DefaultAI = %s, want gemini", cfg.DefaultAI)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(cfg.Projects))
	}
}

func TestGetConfig_ExpandsEnvVars(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	configPath := filepath.Join(home, ConfigFileName)
	contents := "home: $HOME/pancake\ncode_editor: echo\n"
	if runtime.GOOS == "windows" {
		contents = "home: '%userprofile%/pancake'\ncode_editor: echo\n"
	}
	if err := os.WriteFile(configPath, []byte(contents), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cfg, err := GetConfig()
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !filepath.IsAbs(cfg.Home) {
		t.Fatalf("Home should be absolute, got %s", cfg.Home)
	}
	if !strings.HasSuffix(filepath.ToSlash(cfg.Home), "/pancake") {
		t.Fatalf("Home should end with /pancake, got %s", cfg.Home)
	}
}

func TestGetConfig_ParseError(t *testing.T) {
	writeConfig(t, "home: $HOME/pancake\ncode_editor: echo\n  bad: : :\n:invalid")
	_, err := GetConfig()
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !strings.Contains(err.Error(), "not valid YAML") {
		t.Fatalf("expected parse-failed message, got: %v", err)
	}
}

func TestValidateConfig_HomeEmpty(t *testing.T) {
	cfg := &Config{Home: "", DefaultAI: "gemini"}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty home")
	}
	if !strings.Contains(err.Error(), "'home' is empty") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateConfig_HomeRelative(t *testing.T) {
	cfg := &Config{Home: "relative/path", DefaultAI: "gemini"}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for relative home")
	}
	if !strings.Contains(err.Error(), "not an absolute path") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateConfig_DefaultAIInvalid(t *testing.T) {
	cfg := &Config{Home: "/abs/path", DefaultAI: "claude"}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid default_ai")
	}
	if !strings.Contains(err.Error(), "'default_ai'") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateConfig_ProjectNameInvalid(t *testing.T) {
	cfg := &Config{
		Home:      "/abs/path",
		DefaultAI: "gemini",
		Projects: map[string]Project{
			"bad/name": {RemoteSSHURL: "git@github.com:org/repo.git"},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unsafe project name")
	}
	if !strings.Contains(err.Error(), "path separators") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateConfig_ProjectRemoteMissing(t *testing.T) {
	cfg := &Config{
		Home:      "/abs/path",
		DefaultAI: "gemini",
		Projects: map[string]Project{
			"demo": {},
		},
	}
	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing remote")
	}
	if !strings.Contains(err.Error(), "missing 'remote_ssh_url'") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := &Config{
		Home:      "/abs/path",
		DefaultAI: "gemini",
		Projects: map[string]Project{
			"demo": {RemoteSSHURL: "git@github.com:org/repo.git"},
		},
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateConfig_DefaultAIEmptyIsValid(t *testing.T) {
	cfg := &Config{Home: "/abs/path", DefaultAI: ""}
	if err := ValidateConfig(cfg); err != nil {
		t.Fatalf("empty default_ai should be valid (AI disabled), got: %v", err)
	}
}

func TestExpandHomePath_Empty(t *testing.T) {
	got, err := ExpandHomePath("")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

func TestUpdateConfig_RoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	cfg := &Config{
		Home:       filepath.Join(home, "pancake"),
		CodeEditor: "echo",
		DefaultAI:  "gemini",
		Tools:      []string{"tree", "jq"},
		Projects: map[string]Project{
			"demo": {RemoteSSHURL: "git@github.com:org/repo.git"},
		},
	}
	if err := UpdateConfig(cfg); err != nil {
		t.Fatalf("update: %v", err)
	}
	loaded, err := GetConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Tools) != 2 {
		t.Fatalf("tools round-trip failed: %v", loaded.Tools)
	}
	if loaded.Projects["demo"].RemoteSSHURL != "git@github.com:org/repo.git" {
		t.Fatalf("project round-trip failed: %v", loaded.Projects)
	}
}

func TestConfigPath_Absolute(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	got, err := ConfigPath()
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("config path should be absolute, got %s", got)
	}
	if !strings.HasSuffix(filepath.ToSlash(got), "/"+ConfigFileName) {
		t.Fatalf("config path should end with %s, got %s", ConfigFileName, got)
	}
}
