# Dojo

A minimal CLI that launches Claude Code in isolated jj workspaces.

## Usage

```bash
# Launch Claude in a new workspace
dojo feature-auth

# List existing workspaces
dojo list
```

When you run `dojo <name>`:
1. Creates an isolated jj workspace at `.jj/agents/<name>/`
2. Launches Claude Code with full terminal experience
3. On exit, prompts whether to keep or delete the workspace

## Requirements

- Go 1.24+
- [Claude Code](https://claude.ai/code) installed and in PATH
- [Jujutsu (jj)](https://github.com/martinvonz/jj) installed and in PATH
- Must be run from inside a jj repository

## Installation

```bash
go install github.com/bigq/dojo/cmd/dojo@latest
```

Or build from source:

```bash
make build
./dojo
```

## How It Works

### Workspace Isolation

Each agent runs in its own jj workspace with:
- **Separate revision**: Changes don't affect your main workspace
- **Git shim**: Blocks `git` commands, forcing `jj` usage
- **Scoped root**: Claude sees only the workspace as project root

### Multi-Agent Workflow

Run multiple agents by opening multiple terminals:

```bash
# Terminal 1
dojo feature-auth

# Terminal 2
dojo bugfix-login

# Terminal 3
dojo refactor-api
```

### Version Control

Use jj directly in your default workspace to manage agent changes:

```bash
# See what agents have changed
jj log

# Squash an agent's changes
jj squash --from <agent-revision>

# Rebase agents on latest
jj rebase -s <agent-revision> -d @
```

## Development

```bash
make build    # Build the binary
make run      # Build and run
make clean    # Remove build artifacts
```
