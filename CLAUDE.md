# Dojo

Minimal CLI wrapper that launches Claude Code in isolated jj workspaces.

## Architecture

```
dojo <name>
  → jj workspace add ../<repo>-<name>/   (sibling directory)
  → create .git directory (scope isolation, auto-ignored by jj)
  → create .jj/dojo-agent marker (auto-ignored by jj)
  → create git shim (blocks git)
  → exec claude (full terminal passthrough)
  → prompt cleanup on exit
```

## Key Files

| File                       | Purpose                                           |
| -------------------------- | ------------------------------------------------- |
| `cmd/dojo/main.go`         | CLI entry point, workspace setup, claude launcher |
| `internal/jj/client.go`    | jj CLI wrapper                                    |
| `internal/jj/workspace.go` | Workspace operations (add, forget, list)          |
| `internal/jj/errors.go`    | Error types                                       |

## Commands

- `dojo <name>` - Create workspace and launch Claude interactively
- `dojo list` - List existing agent workspaces

## Workspace Isolation Mechanisms

1. **jj workspace**: Each agent gets its own jj workspace/revision as sibling directory
2. **.git directory**: Empty directory at workspace root scopes Claude (auto-ignored by jj)
3. **.jj/dojo-agent**: Marker file with metadata (auto-ignored, inside .jj/)
4. **git shim**: Script in PATH that blocks git commands, forces jj usage
5. **PWD**: Claude runs with workspace as working directory

## Code Patterns

### Launching Claude

```go
cmd := exec.Command("claude")
cmd.Dir = workspacePath
cmd.Env = envWithShimInPath
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
cmd.Run()
```

### Cleanup

```go
os.RemoveAll(filepath.Join(workspacePath, ".git"))  // Remove .git directory
os.Remove(filepath.Join(workspacePath, ".jj/dojo-agent"))  // Remove marker
client.WorkspaceForget(ctx, name)                   // Unregister from jj
os.RemoveAll(workspacePath)                         // Delete directory
```

## When to Look Here

- Adding new CLI commands
- Modifying workspace isolation behavior
- Changing cleanup behavior
- jj integration issues
