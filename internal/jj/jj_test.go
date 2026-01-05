package jj_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bigq/dojo/internal/jj"
)

// setupTestRepo creates a temporary jj repository for testing.
// Returns the repo path and a cleanup function.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	dir, err := os.MkdirTemp("", "jj-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Initialize jj repo
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

func TestWorkspaceRoot(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Change to repo directory for this test
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	root, err := client.WorkspaceRoot(ctx)
	if err != nil {
		t.Fatalf("WorkspaceRoot failed: %v", err)
	}

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	expectedDir, _ := filepath.EvalSymlinks(dir)
	actualRoot, _ := filepath.EvalSymlinks(root)

	if actualRoot != expectedDir {
		t.Errorf("WorkspaceRoot = %q, want %q", root, dir)
	}
}

func TestWorkspaceAdd(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	workspacePath := filepath.Join(dir, "test-workspace")

	// Add workspace
	err := client.WorkspaceAdd(ctx, workspacePath, "")
	if err != nil {
		t.Fatalf("WorkspaceAdd failed: %v", err)
	}

	// Verify workspace directory exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		t.Error("workspace directory was not created")
	}

	// Adding same workspace again should fail
	err = client.WorkspaceAdd(ctx, workspacePath, "")
	if err == nil {
		t.Error("expected error when adding duplicate workspace")
	}
}

func TestWorkspaceList(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	// Initially should have just "default" workspace
	workspaces, err := client.WorkspaceList(ctx)
	if err != nil {
		t.Fatalf("WorkspaceList failed: %v", err)
	}

	if len(workspaces) != 1 {
		t.Errorf("expected 1 workspace, got %d", len(workspaces))
	}

	if workspaces[0].Name != "default" {
		t.Errorf("expected workspace name 'default', got %q", workspaces[0].Name)
	}

	// Add a workspace
	workspacePath := filepath.Join(dir, "agent-1")
	if err := client.WorkspaceAdd(ctx, workspacePath, ""); err != nil {
		t.Fatalf("WorkspaceAdd failed: %v", err)
	}

	// Should now have 2 workspaces
	workspaces, err = client.WorkspaceList(ctx)
	if err != nil {
		t.Fatalf("WorkspaceList failed: %v", err)
	}

	if len(workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(workspaces))
	}

	// Find agent-1
	found := false
	for _, ws := range workspaces {
		if ws.Name == "agent-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("workspace 'agent-1' not found in list")
	}
}

func TestWorkspaceForget(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	// Add a workspace
	workspacePath := filepath.Join(dir, "to-forget")
	if err := client.WorkspaceAdd(ctx, workspacePath, ""); err != nil {
		t.Fatalf("WorkspaceAdd failed: %v", err)
	}

	// Forget it
	if err := client.WorkspaceForget(ctx, "to-forget"); err != nil {
		t.Fatalf("WorkspaceForget failed: %v", err)
	}

	// Should only have default workspace now
	workspaces, err := client.WorkspaceList(ctx)
	if err != nil {
		t.Fatalf("WorkspaceList failed: %v", err)
	}

	for _, ws := range workspaces {
		if ws.Name == "to-forget" {
			t.Error("workspace 'to-forget' still exists after forget")
		}
	}

	// Note: WorkspaceForget does NOT delete the directory
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		t.Error("workspace directory should still exist after forget")
	}
}

func TestNotJJRepo(t *testing.T) {
	// Create a temp directory that is NOT a jj repo
	dir, err := os.MkdirTemp("", "not-jj-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	client := jj.NewClient()
	ctx := context.Background()

	_, err = client.WorkspaceRoot(ctx)
	if err == nil {
		t.Error("expected error for non-jj directory")
	}
}
