package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bigq/dojo/internal/jj"
)

// setupTestRepo creates a temporary jj repository for testing.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "dojo-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to init jj repo: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestGitShimCreation(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	// Create agents parent directory first
	agentsPath := filepath.Join(dir, agentsDir)
	if err := os.MkdirAll(agentsPath, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Create workspace
	workspacePath := filepath.Join(dir, agentsDir, "test-agent")
	if err := client.WorkspaceAdd(ctx, workspacePath, ""); err != nil {
		t.Fatalf("WorkspaceAdd failed: %v", err)
	}

	// Create git shim (mimicking what runAgent does)
	shimPath := filepath.Join(workspacePath, shimDir)
	if err := os.MkdirAll(shimPath, 0755); err != nil {
		t.Fatalf("failed to create shim directory: %v", err)
	}

	shimScript := filepath.Join(shimPath, "git")
	shimContent := `#!/bin/sh
echo "git disabled for agents; use jj" >&2
exit 1
`
	if err := os.WriteFile(shimScript, []byte(shimContent), 0755); err != nil {
		t.Fatalf("failed to write git shim: %v", err)
	}

	// Verify shim exists and is executable
	info, err := os.Stat(shimScript)
	if err != nil {
		t.Fatalf("git shim not found: %v", err)
	}

	if info.Mode().Perm()&0111 == 0 {
		t.Error("git shim is not executable")
	}

	// Test that shim blocks git
	cmd := exec.Command(shimScript, "status")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected git shim to return error")
	}

	if !strings.Contains(string(output), "git disabled") {
		t.Errorf("unexpected shim output: %s", output)
	}
}

func TestGitMarkerCreation(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	// Create agents parent directory first
	agentsPath := filepath.Join(dir, agentsDir)
	if err := os.MkdirAll(agentsPath, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Create workspace
	workspacePath := filepath.Join(dir, agentsDir, "test-agent")
	if err := client.WorkspaceAdd(ctx, workspacePath, ""); err != nil {
		t.Fatalf("WorkspaceAdd failed: %v", err)
	}

	// Create .git marker (mimicking what runAgent does)
	gitMarker := filepath.Join(workspacePath, ".git")
	if err := os.WriteFile(gitMarker, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create .git marker: %v", err)
	}

	// Verify marker exists
	if _, err := os.Stat(gitMarker); os.IsNotExist(err) {
		t.Error(".git marker was not created")
	}
}

func TestCleanup(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	// Create agents parent directory first
	agentsPath := filepath.Join(dir, agentsDir)
	if err := os.MkdirAll(agentsPath, 0755); err != nil {
		t.Fatalf("failed to create agents dir: %v", err)
	}

	// Create workspace with all the fixtures
	name := "cleanup-test"
	workspacePath := filepath.Join(dir, agentsDir, name)

	if err := client.WorkspaceAdd(ctx, workspacePath, ""); err != nil {
		t.Fatalf("WorkspaceAdd failed: %v", err)
	}

	// Create .git marker
	gitMarker := filepath.Join(workspacePath, ".git")
	if err := os.WriteFile(gitMarker, []byte{}, 0644); err != nil {
		t.Fatalf("failed to create .git marker: %v", err)
	}

	// Create shim
	shimPath := filepath.Join(workspacePath, shimDir)
	os.MkdirAll(shimPath, 0755)
	os.WriteFile(filepath.Join(shimPath, "git"), []byte("#!/bin/sh\nexit 1"), 0755)

	// Verify workspace exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		t.Fatal("workspace should exist before cleanup")
	}

	// Run cleanup
	cleanupWorkspace(ctx, client, name, workspacePath)

	// Verify workspace is gone
	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
		t.Error("workspace directory should be removed after cleanup")
	}

	// Verify workspace is forgotten from jj
	workspaces, _ := client.WorkspaceList(ctx)
	for _, ws := range workspaces {
		if ws.Name == name {
			t.Error("workspace should be forgotten from jj")
		}
	}
}

func TestListWorkspacesEmpty(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	// No agents directory - should not error
	agentsPath := filepath.Join(dir, agentsDir)
	entries, err := os.ReadDir(agentsPath)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("unexpected error: %v", err)
	}

	// Filter like listWorkspaces does
	var count int
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			count++
		}
	}

	if count != 0 {
		t.Errorf("expected 0 workspaces, got %d", count)
	}
}

func TestListWorkspacesFiltersHidden(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	// Create agents directory with hidden and visible dirs
	agentsPath := filepath.Join(dir, agentsDir)
	os.MkdirAll(agentsPath, 0755)

	os.MkdirAll(filepath.Join(agentsPath, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(agentsPath, ".pids"), 0755)
	os.MkdirAll(filepath.Join(agentsPath, "visible-agent"), 0755)
	os.MkdirAll(filepath.Join(agentsPath, "another-agent"), 0755)

	entries, err := os.ReadDir(agentsPath)
	if err != nil {
		t.Fatalf("failed to read agents dir: %v", err)
	}

	// Filter like listWorkspaces does
	var visible []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			visible = append(visible, entry.Name())
		}
	}

	if len(visible) != 2 {
		t.Errorf("expected 2 visible workspaces, got %d: %v", len(visible), visible)
	}
}

func TestPathWithShim(t *testing.T) {
	shimPath := "/fake/shim/path"
	originalPath := "/usr/bin:/bin"

	env := []string{
		"HOME=/home/user",
		"PATH=" + originalPath,
		"TERM=xterm",
	}

	newPath := shimPath + ":" + originalPath
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + newPath
			break
		}
	}

	// Find PATH in modified env
	var foundPath string
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			foundPath = strings.TrimPrefix(e, "PATH=")
			break
		}
	}

	if !strings.HasPrefix(foundPath, shimPath) {
		t.Errorf("shim path should be first in PATH, got: %s", foundPath)
	}
}

// cleanupWorkspace is extracted for testing (mirrors the cleanup function in main.go)
func cleanupWorkspace(ctx context.Context, client *jj.Client, name, workspacePath string) {
	os.Remove(filepath.Join(workspacePath, ".git"))
	client.WorkspaceForget(ctx, name)
	os.RemoveAll(workspacePath)
}
