package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bigq/dojo/internal/jj"
)

const (
	agentsDir = ".jj/agents"
	shimDir   = ".jj/.dojo-bin"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listWorkspaces()
	case "-h", "--help", "help":
		printUsage()
	default:
		runAgent(os.Args[1])
	}
}

func printUsage() {
	fmt.Println(`Usage: dojo <name>    Create workspace and launch Claude
       dojo list      List existing workspaces`)
}

func runAgent(name string) {
	ctx := context.Background()
	client := jj.NewClient()

	// Get repo root
	root, err := client.WorkspaceRoot(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a jj repository\n")
		os.Exit(1)
	}

	workspacePath := filepath.Join(root, agentsDir, name)
	shimPath := filepath.Join(workspacePath, shimDir)

	// 1. Create workspace via jj workspace add
	if err := client.WorkspaceAdd(ctx, workspacePath, ""); err != nil {
		// Check if it already exists
		if strings.Contains(err.Error(), "already exists") {
			fmt.Fprintf(os.Stderr, "Error: workspace '%s' already exists\n", name)
			fmt.Fprintf(os.Stderr, "Use 'dojo list' to see existing workspaces\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error creating workspace: %v\n", err)
		os.Exit(1)
	}

	// 2. Create .git marker file (scopes Claude to workspace)
	gitMarker := filepath.Join(workspacePath, ".git")
	if err := os.WriteFile(gitMarker, []byte{}, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .git marker: %v\n", err)
		cleanup(ctx, client, name, workspacePath)
		os.Exit(1)
	}

	// 3. Create git shim
	if err := os.MkdirAll(shimPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating shim directory: %v\n", err)
		cleanup(ctx, client, name, workspacePath)
		os.Exit(1)
	}

	shimScript := filepath.Join(shimPath, "git")
	shimContent := `#!/bin/sh
echo "git disabled for agents; use jj" >&2
exit 1
`
	if err := os.WriteFile(shimScript, []byte(shimContent), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating git shim: %v\n", err)
		cleanup(ctx, client, name, workspacePath)
		os.Exit(1)
	}

	// 4. Build env with shim in PATH
	env := os.Environ()
	newPath := shimPath + ":" + os.Getenv("PATH")
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + newPath
			break
		}
	}

	// 5. Fork claude with Stdin/Stdout/Stderr passthrough
	cmd := exec.Command("claude")
	cmd.Dir = workspacePath
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Claude exited with error - still prompt for cleanup
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "\nClaude exited with code %d\n", exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "\nError running claude: %v\n", err)
		}
	}

	// 6. Prompt for cleanup
	fmt.Print("\nKeep workspace for inspection? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	// 7. If no: cleanup
	if answer != "y" && answer != "yes" {
		cleanup(ctx, client, name, workspacePath)
		fmt.Printf("Workspace '%s' removed\n", name)
	} else {
		fmt.Printf("Workspace kept at: %s\n", workspacePath)
	}
}

func cleanup(ctx context.Context, client *jj.Client, name, workspacePath string) {
	// Remove .git marker first so jj can work properly
	os.Remove(filepath.Join(workspacePath, ".git"))

	// Forget workspace in jj
	if err := client.WorkspaceForget(ctx, name); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to forget workspace: %v\n", err)
	}

	// Remove directory
	if err := os.RemoveAll(workspacePath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove workspace directory: %v\n", err)
	}
}

func listWorkspaces() {
	ctx := context.Background()
	client := jj.NewClient()

	// Get repo root
	root, err := client.WorkspaceRoot(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a jj repository\n")
		os.Exit(1)
	}

	// List directories in .jj/agents/
	agentsPath := filepath.Join(root, agentsDir)
	entries, err := os.ReadDir(agentsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No agents directory yet - no workspaces
			return
		}
		fmt.Fprintf(os.Stderr, "Error listing workspaces: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			fmt.Println(entry.Name())
		}
	}
}
