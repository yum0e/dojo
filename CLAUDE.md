# Kekkai

Minimal CLI wrapper that launches AI coding agents (Codex, Claude) in isolated jj workspaces.

## Architecture

```
kekkai <name> [--agent=codex|claude]
  → jj workspace add ../<repo>-<name>/   (sibling directory)
  → create .git directory (scope isolation, auto-ignored by jj)
  → create .jj/kekkai-agent marker (auto-ignored by jj)
  → create git shim (blocks git)
  → exec agent (full terminal passthrough)
  → prompt cleanup on exit
```

## Key Files

| File                    | Purpose                                           |
| ----------------------- | ------------------------------------------------- |
| `src/kekkai/cli.py`     | CLI entry point, workspace setup, agent launcher  |
| `src/kekkai/jj.py`      | jj CLI wrapper + Workspace dataclass              |
| `src/kekkai/errors.py`  | Custom exception classes                          |

## Commands

- `kekkai <name>` - Create workspace and launch agent (default: codex)
- `kekkai <name> --agent=claude` - Launch with Claude instead
- `kekkai list` - List existing agent workspaces (shows agent type)

## Running

```bash
# Local development (default agent: codex)
uv run kekkai <name>
uv run kekkai <name> --agent=claude
uv run kekkai list

# After publishing to PyPI
uvx kekkai <name>
uvx kekkai <name> -a claude
uvx kekkai list
```

## Workspace Isolation Mechanisms

1. **jj workspace**: Each agent gets its own jj workspace/revision as sibling directory
2. **.git directory**: Empty directory at workspace root scopes agent (auto-ignored by jj)
3. **.jj/kekkai-agent**: Marker file with metadata including agent type (auto-ignored, inside .jj/)
4. **git shim**: Script in PATH that blocks git commands, forces jj usage
5. **PWD**: Agent runs with workspace as working directory

## Code Patterns

### Launching Agent

```python
env = os.environ.copy()
env["PATH"] = f"{shim_path}:{env.get('PATH', '')}"

subprocess.run([agent.executable], cwd=workspace_path, env=env)
```

### Cleanup

```python
shutil.rmtree(Path(workspace_path) / ".git")           # Remove .git directory
(Path(workspace_path) / AGENT_MARKER_FILE).unlink()    # Remove marker
client.workspace_forget(jj_workspace_name, cwd=root)   # Unregister from jj
shutil.rmtree(workspace_path)                          # Delete directory
```

## Testing

```bash
uv run --with pytest pytest tests/ -v
```

## When to Look Here

- Adding new CLI commands
- Modifying workspace isolation behavior
- Changing cleanup behavior
- jj integration issues
