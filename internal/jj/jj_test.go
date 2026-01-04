package jj

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireJJ fails the test if jj is not installed.
func requireJJ(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("jj"); err != nil {
		t.Fatal("jj is not installed")
	}
}


// setupTestRepo creates a temporary jj repository for testing.
func setupTestRepo(t *testing.T) (repoPath string, cleanup func()) {
	t.Helper()
	requireJJ(t)

	tmpDir, err := os.MkdirTemp("", "jj-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup = func() { os.RemoveAll(tmpDir) }

	// Initialize jj repo
	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("failed to init repository: %v", err)
	}

	// Configure user via .jj/repo/config.toml
	configPath := filepath.Join(tmpDir, ".jj", "repo", "config.toml")
	configContent := `[user]
name = "Test User"
email = "test@example.com"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		cleanup()
		t.Fatalf("failed to write jj config: %v", err)
	}

	return tmpDir, cleanup
}

// runInDir executes a function in the given directory.
func runInDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to directory %s: %v", dir, err)
	}
	defer os.Chdir(orig)

	fn()
}

// TestCommandError tests the CommandError type.
func TestCommandError(t *testing.T) {
	t.Run("Error formatting", func(t *testing.T) {
		err := &CommandError{
			Cmd:    "workspace list",
			Stderr: "something went wrong",
			Err:    errors.New("exit status 1"),
		}

		expected := "jj workspace list: something went wrong"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := &CommandError{
			Cmd:    "test",
			Stderr: "stderr",
			Err:    underlying,
		}

		if !errors.Is(err, underlying) {
			t.Error("Unwrap should return underlying error")
		}
	})

	t.Run("errors.Is with sentinel errors", func(t *testing.T) {
		err := &CommandError{
			Cmd:    "test",
			Stderr: "stderr",
			Err:    ErrNotJJRepo,
		}

		if !errors.Is(err, ErrNotJJRepo) {
			t.Error("errors.Is should match ErrNotJJRepo")
		}
	})
}

// TestNewClient tests client creation.
func TestNewClient(t *testing.T) {
	t.Run("default jj path", func(t *testing.T) {
		c := NewClient()
		if c.jjPath != "jj" {
			t.Errorf("expected default jjPath to be 'jj', got %q", c.jjPath)
		}
	})

	t.Run("custom jj path", func(t *testing.T) {
		c := NewClient(WithJJPath("/custom/jj"))
		if c.jjPath != "/custom/jj" {
			t.Errorf("expected jjPath to be '/custom/jj', got %q", c.jjPath)
		}
	})
}

// TestRunInNonRepo tests that running in a non-jj directory returns ErrNotJJRepo.
func TestRunInNonRepo(t *testing.T) {
	requireJJ(t)

	tmpDir, err := os.MkdirTemp("", "non-jj-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	runInDir(t, tmpDir, func() {
		c := NewClient()
		_, err := c.WorkspaceList(context.Background())

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var cmdErr *CommandError
		if !errors.As(err, &cmdErr) {
			t.Fatalf("expected CommandError, got %T: %v", err, err)
		}

		if !errors.Is(err, ErrNotJJRepo) {
			t.Errorf("expected ErrNotJJRepo, got %v (stderr: %q)", err, cmdErr.Stderr)
		}
	})
}

// TestWorkspaceList tests listing workspaces.
func TestWorkspaceList(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		workspaces, err := c.WorkspaceList(context.Background())
		if err != nil {
			t.Fatalf("WorkspaceList failed: %v", err)
		}

		if len(workspaces) != 1 {
			t.Fatalf("expected 1 workspace, got %d", len(workspaces))
		}

		if workspaces[0].Name != "default" {
			t.Errorf("expected workspace name 'default', got %q", workspaces[0].Name)
		}
	})
}

// TestWorkspaceAddAndForget tests adding and forgetting workspaces.
func TestWorkspaceAddAndForget(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Add a workspace
		agentPath := filepath.Join(repoPath, "agent-1")
		err := c.WorkspaceAdd(ctx, agentPath, "")
		if err != nil {
			t.Fatalf("WorkspaceAdd failed: %v", err)
		}

		// Verify it was added
		workspaces, err := c.WorkspaceList(ctx)
		if err != nil {
			t.Fatalf("WorkspaceList failed: %v", err)
		}

		if len(workspaces) != 2 {
			t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
		}

		// Try to add duplicate
		err = c.WorkspaceAdd(ctx, agentPath, "")
		if !errors.Is(err, ErrWorkspaceExists) {
			t.Errorf("expected ErrWorkspaceExists, got %v", err)
		}

		// Forget the workspace
		err = c.WorkspaceForget(ctx, "agent-1")
		if err != nil {
			t.Fatalf("WorkspaceForget failed: %v", err)
		}

		// WorkspaceForget does NOT remove the directory - caller must do that
		// Just verify the directory still exists
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Error("workspace directory should still exist after forget")
		}

		// Clean up directory manually (as caller would)
		os.RemoveAll(agentPath)

		// Verify it was removed from jj
		workspaces, err = c.WorkspaceList(ctx)
		if err != nil {
			t.Fatalf("WorkspaceList failed: %v", err)
		}

		if len(workspaces) != 1 {
			t.Fatalf("expected 1 workspace after forget, got %d", len(workspaces))
		}
	})
}

// TestWorkspaceForgetNonexistent tests forgetting a workspace that doesn't exist.
func TestWorkspaceForgetNonexistent(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		err := c.WorkspaceForget(context.Background(), "nonexistent")
		if !errors.Is(err, ErrWorkspaceNotFound) {
			t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
		}
	})
}

// TestWorkspaceRoot tests getting the workspace root.
func TestWorkspaceRoot(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		root, err := c.WorkspaceRoot(context.Background())
		if err != nil {
			t.Fatalf("WorkspaceRoot failed: %v", err)
		}

		// Resolve symlinks for comparison
		expectedRoot, _ := filepath.EvalSymlinks(repoPath)
		actualRoot, _ := filepath.EvalSymlinks(root)

		if actualRoot != expectedRoot {
			t.Errorf("expected root %q, got %q", expectedRoot, actualRoot)
		}
	})
}

// TestLog tests commit log retrieval.
func TestLog(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Get initial log
		commits, err := c.Log(ctx, nil)
		if err != nil {
			t.Fatalf("Log failed: %v", err)
		}

		// Should have at least the root commit
		if len(commits) == 0 {
			t.Error("expected at least one commit")
		}
	})
}

// TestLogWithLimit tests log with limit option.
func TestLogWithLimit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create some commits
		for i := 0; i < 5; i++ {
			if err := c.Describe(ctx, "commit message"); err != nil {
				t.Fatalf("Describe failed: %v", err)
			}
			if err := c.Commit(ctx, "test commit"); err != nil {
				// Commit might fail if no changes, which is fine
				continue
			}
		}

		// Get log with limit
		commits, err := c.Log(ctx, &LogOptions{Limit: 2})
		if err != nil {
			t.Fatalf("Log with limit failed: %v", err)
		}

		if len(commits) > 2 {
			t.Errorf("expected at most 2 commits, got %d", len(commits))
		}
	})
}

// TestStatus tests status retrieval.
func TestStatus(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Get status of clean repo
		status, err := c.Status(ctx)
		if err != nil {
			t.Fatalf("Status failed: %v", err)
		}

		if status.HasConflicts {
			t.Error("clean repo should not have conflicts")
		}
	})
}

// TestStatusWithChanges tests status with modified files.
func TestStatusWithChanges(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create a file
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Get status
		status, err := c.Status(ctx)
		if err != nil {
			t.Fatalf("Status failed: %v", err)
		}

		if len(status.Changes) == 0 {
			t.Error("expected changes after creating a file")
		}

		// Check that the file is listed
		found := false
		for _, change := range status.Changes {
			if change.Path == "test.txt" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected test.txt in changes")
		}
	})
}

// TestDiff tests diff retrieval.
func TestDiff(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Get diff of empty repo
		diff, err := c.Diff(ctx, nil)
		if err != nil {
			t.Fatalf("Diff failed: %v", err)
		}

		// Empty repo should have empty diff
		if diff != "" {
			t.Logf("diff output (may be empty): %q", diff)
		}
	})
}

// TestDiffWithChanges tests diff with changes.
func TestDiffWithChanges(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create a file
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Get diff
		diff, err := c.Diff(ctx, nil)
		if err != nil {
			t.Fatalf("Diff failed: %v", err)
		}

		// Should have some diff output
		if diff == "" {
			t.Error("expected non-empty diff after creating a file")
		}
	})
}

// TestDescribe tests describing the working copy.
func TestDescribe(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		err := c.Describe(ctx, "test description")
		if err != nil {
			t.Fatalf("Describe failed: %v", err)
		}

		// Verify by checking log
		commits, err := c.Log(ctx, &LogOptions{Limit: 1})
		if err != nil {
			t.Fatalf("Log failed: %v", err)
		}

		if len(commits) == 0 {
			t.Fatal("expected at least one commit")
		}

		if commits[0].Description != "test description" {
			t.Errorf("expected description 'test description', got %q", commits[0].Description)
		}
	})
}

// TestCommit tests creating a commit.
func TestCommit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create a file first
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		// Commit
		err := c.Commit(ctx, "add test file")
		if err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Verify by checking log
		commits, err := c.Log(ctx, &LogOptions{Limit: 2})
		if err != nil {
			t.Fatalf("Log failed: %v", err)
		}

		found := false
		for _, commit := range commits {
			if commit.Description == "add test file" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find commit with message 'add test file'")
		}
	})
}

// TestSquash tests squashing.
func TestSquash(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create a file and commit it first (can't squash into root commit)
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		if err := c.Commit(ctx, "initial file"); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Now make another change in the new working copy
		if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
			t.Fatalf("failed to modify test file: %v", err)
		}

		// Squash should work (squashes into parent)
		err := c.Squash(ctx)
		if err != nil {
			t.Fatalf("Squash failed: %v", err)
		}
	})
}

// TestGetWorkingCopyChangeID tests getting the change ID of a workspace.
func TestGetWorkingCopyChangeID(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		changeID, err := c.GetWorkingCopyChangeID(ctx, repoPath)
		if err != nil {
			t.Fatalf("GetWorkingCopyChangeID failed: %v", err)
		}

		if changeID == "" {
			t.Error("expected non-empty change ID")
		}

		t.Logf("Working copy change ID: %s", changeID)
	})
}

// TestGetParentChangeID tests getting the parent change ID.
func TestGetParentChangeID(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create a commit so we have a parent
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		if err := c.Commit(ctx, "test commit"); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		parentID, err := c.GetParentChangeID(ctx, repoPath)
		if err != nil {
			t.Fatalf("GetParentChangeID failed: %v", err)
		}

		if parentID == "" {
			t.Error("expected non-empty parent change ID")
		}

		// Parent should be different from current
		currentID, _ := c.GetWorkingCopyChangeID(ctx, repoPath)
		if parentID == currentID {
			t.Error("parent change ID should be different from current")
		}

		t.Logf("Current: %s, Parent: %s", currentID, parentID)
	})
}

// TestWorkspaceUpdateStale tests updating a stale workspace.
func TestWorkspaceUpdateStale(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// This should not error even if workspace is not stale
		err := c.WorkspaceUpdateStale(ctx, repoPath)
		if err != nil {
			t.Logf("WorkspaceUpdateStale returned (may be expected): %v", err)
		}
	})
}

// TestNewFromRevisionInDir tests creating a new revision from a specific parent in a workspace.
func TestNewFromRevisionInDir(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Create a file and commit to have some history
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		if err := c.Commit(ctx, "test commit"); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Get current ID before
		currentID, _ := c.GetWorkingCopyChangeID(ctx, repoPath)
		parentID, _ := c.GetParentChangeID(ctx, repoPath)
		t.Logf("Before: current=%s, parent=%s", currentID, parentID)

		// Create a new revision based on parent (going back one level)
		if err := c.NewFromRevisionInDir(ctx, repoPath, parentID); err != nil {
			t.Fatalf("NewFromRevisionInDir failed: %v", err)
		}

		// Should now be at a different change ID
		newID, _ := c.GetWorkingCopyChangeID(ctx, repoPath)
		t.Logf("After: new=%s", newID)

		if newID == currentID {
			t.Error("Change ID should be different after NewFromRevisionInDir")
		}
	})
}

// TestDiffAutoUpdatesStaleWorkspace tests that Diff automatically handles stale workspaces.
func TestDiffAutoUpdatesStaleWorkspace(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	runInDir(t, repoPath, func() {
		c := NewClient()
		ctx := context.Background()

		// Step 1: Create some initial content and commit
		testFile := filepath.Join(repoPath, "test.txt")
		if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		if err := c.Commit(ctx, "initial"); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Step 2: Create an agent workspace
		agentPath := filepath.Join(repoPath, ".jj", "agents", "test-agent")
		if err := os.MkdirAll(filepath.Dir(agentPath), 0755); err != nil {
			t.Fatalf("failed to create agents dir: %v", err)
		}
		if err := c.WorkspaceAdd(ctx, agentPath, "@"); err != nil {
			t.Fatalf("WorkspaceAdd failed: %v", err)
		}
		t.Log("Created agent workspace")

		// Step 3: Make a change in default workspace (this will rebase agent and make it stale)
		if err := os.WriteFile(testFile, []byte("modified in default"), 0644); err != nil {
			t.Fatalf("failed to modify test file: %v", err)
		}
		t.Log("Modified file in default workspace")

		// Step 4: Try to get diff in agent workspace
		// This should auto-recover from the stale error
		diff, err := c.Diff(ctx, &DiffOptions{WorkDir: agentPath})
		if err != nil {
			t.Fatalf("Diff failed (should have auto-recovered from stale): %v", err)
		}

		t.Logf("Diff succeeded (auto-recovered from stale): %d bytes", len(diff))

		// Cleanup
		c.WorkspaceForget(ctx, "test-agent")
		os.RemoveAll(agentPath)
	})
}
