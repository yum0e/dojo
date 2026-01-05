# Dojo Spec: Minimal CLI Wrapper

## Overview

A minimal CLI that launches Claude Code directly in an isolated jj workspace. Users get the full Claude terminal experience without recreation.

## Motivation

- No need to recreate Claude's UI (syntax highlighting, markdown, tool visualization)
- Focus on core value: workspace isolation and version control orchestration
- Simple codebase (~150 LOC)

## Commands

### `dojo <name>`

Creates an isolated workspace and launches Claude interactively.

```
$ dojo feature-auth
[Creates workspace, launches Claude]
[User interacts with full Claude UI]
[On exit]
Keep workspace for inspection? [y/N] _
```

**Flow:**

1. Create jj workspace at `.jj/agents/<name>/`
2. Create `.git` marker file (scopes Claude to workspace)
3. Set up git shim in PATH (blocks git, forces jj)
4. Fork Claude process with full terminal passthrough
5. Wait for Claude to exit
6. Prompt: "Keep workspace for inspection? [y/N]"
7. If no: `jj workspace forget <name>` + remove directory

### `dojo list`

Shows existing agent workspaces.

```
$ dojo list
feature-auth
bugfix-login
refactor-api
```

## Workspace Isolation

### Directory Structure

```
repository/
├── .jj/
│   ├── agents/
│   │   ├── <name>/           ← Agent workspace
│   │   │   ├── .git          ← Marker file (scope isolation)
│   │   │   ├── .jj/
│   │   │   │   └── .dojo-bin/
│   │   │   │       └── git   ← Shim script
│   │   │   └── [project files]
│   │   └── ...
│   └── [jj metadata]
└── [default workspace]
```

### .git Marker

- Empty file at `<workspace>/.git`
- Prevents Claude from detecting parent jj repo
- Makes Claude treat workspace as standalone project root

### Git Shim

- Script at `<workspace>/.jj/.dojo-bin/git`
- Returns exit 1 with message "git disabled for agents; use jj"
- PATH prepended so shim shadows real git

## Multi-Agent Model

- User opens multiple terminals for multiple agents
- Each `dojo <name>` is independent
- No centralized orchestration
- Version control via jj directly in default workspace

## Design Decisions

| Question       | Decision                                                        |
| -------------- | --------------------------------------------------------------- |
| TTY approach   | Fork with terminal passthrough (not exec, not PTY multiplexing) |
| Workspace UI   | None - CLI only                                                 |
| Diff view      | None - Claude can run jj commands                               |
| Multi-agent    | Separate terminals                                              |
| CLI commands   | `dojo <name>`, `dojo list`                                      |
| Git shim       | Keep (forces jj usage)                                          |
| Claude args    | None - user interacts directly                                  |
| Crash cleanup  | Prompt user "Keep workspace?"                                   |
| Workspace path | `.jj/agents/<name>/` (invisible to user)                        |

## Dependencies

- `os/exec` (stdlib)
- `internal/jj` (workspace operations)
- No external libraries
