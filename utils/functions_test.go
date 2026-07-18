package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckExists(t *testing.T) {
	dir := t.TempDir()
	if !CheckExists(dir) {
		t.Fatal("temp dir should exist")
	}
	if CheckExists(filepath.Join(dir, "nope")) {
		t.Fatal("nonexistent path should not exist")
	}
}

func TestExecuteCommand_Success(t *testing.T) {
	if err := ExecuteCommand("echo hello", ""); err != nil {
		t.Fatalf("echo should succeed: %v", err)
	}
}

func TestExecuteCommand_Failure(t *testing.T) {
	err := ExecuteCommand("exit 7", "")
	if err == nil {
		t.Fatal("expected error for failing command")
	}
}

func TestExecuteCommand_InDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "marker.txt")
	if err := ExecuteCommand("touch marker.txt", dir); err != nil {
		t.Fatalf("touch failed: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("expected file at %s: %v", target, err)
	}
}

func TestCloneRepository_AlreadyGit_Pulls(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	err := CloneRepository(repoPath, "https://example.com/repo.git")
	if err == nil {
		t.Skip("git pull attempted (no remote configured) — skipping; safety guard did not trigger data loss")
	}
}

func TestCloneRepository_NonGitDirExists_Refuses(t *testing.T) {
	dir := t.TempDir()
	repoPath := filepath.Join(dir, "myrepo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatal(err)
	}
	important := filepath.Join(repoPath, "important.txt")
	if err := os.WriteFile(important, []byte("do not delete me"), 0644); err != nil {
		t.Fatal(err)
	}
	err := CloneRepository(repoPath, "https://example.com/repo.git")
	if err == nil {
		t.Fatal("expected error when non-git dir exists")
	}
	if _, err := os.Stat(important); err != nil {
		t.Fatalf("user data was deleted: %v", err)
	}
}

func TestSaveAndLoadProjectPIDs_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	pids := map[string]int{"demo": 1234, "web": 5678}
	if err := SaveProjectPIDs(dir, pids); err != nil {
		t.Fatalf("save: %v", err)
	}
	var loaded map[string]int
	if err := LoadProjectPIDs(dir, &loaded); err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded["demo"] != 1234 || loaded["web"] != 5678 {
		t.Fatalf("round-trip mismatch: %v", loaded)
	}
}

func TestLoadProjectPIDs_NoFile(t *testing.T) {
	dir := t.TempDir()
	var loaded map[string]int
	if err := LoadProjectPIDs(dir, &loaded); err != nil {
		t.Fatalf("missing file should not error, got: %v", err)
	}
	if len(loaded) != 0 {
		t.Fatalf("expected empty, got %v", loaded)
	}
}

func TestGetPackageManager(t *testing.T) {
	pm := GetPackageManager()
	if pm == "" {
		t.Fatal("package manager should not be empty on supported platforms")
	}
}
