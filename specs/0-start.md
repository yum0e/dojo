# Dojo Spec: Minimal CLI Wrapper

## Overview

A minimal CLI that launches Claude Code directly in an isolated jj workspace. Users get the full Claude terminal experience without recreation.

## Motivation

- No need to recreate Claude's UI (syntax highlighting, markdown, tool visualization)
- Focus on core value: workspace isolation and version control orchestration
- Simple codebase (~250 LOC)

## Commands

### `dojo <name>`

Creates an isolated workspace and launches Claude interactively.

```
$ dojo feature-auth
[Creates workspace, launches Claude]
[User interacts with full Claude UI]
[On exit]
Warning: This workspace has uncommitted changes!
Keep workspace for inspection? [y/N] _
```

**Flow:**

1. Find root workspace (follows `.jj/dojo-agent` marker if run from agent)
2. Check parent directory is writable
3. Create jj workspace as sibling: `../<repo>-<name>/` (full copy including `.claude/`)
4. Create `.git` directory (scopes Claude to workspace, auto-ignored by jj)
5. Create `.jj/dojo-agent` marker file (JSON with root path, agent name, timestamp, auto-ignored)
6. Set up git shim in PATH (blocks git, forces jj)
7. Fork Claude process with full terminal passthrough
8. Wait for Claude to exit
9. Warn if uncommitted changes detected
10. Prompt: "Keep workspace for inspection? [y/N]"
11. If no: cleanup (remove markers, forget workspace, delete directory)

### `dojo list`

Shows existing agent workspaces with jj revision info.

```
$ dojo list
feature-auth: wpxqlmox f3c3a79d Add OAuth2 support
bugfix-login: rstuvwxy a1b2c3d4 Fix login redirect
```

## Workspace Isolation

### Directory Structure

```
/Users/dev/
├── myproject/                    <- Root workspace
│   ├── .claude/
│   │   └── settings.local.json   <- Permissions
│   ├── .jj/
│   │   └── [jj metadata]
│   └── [project files]
└── myproject-feature-auth/       <- Agent workspace (sibling, full copy)
    ├── .claude/                  <- Full copy (not symlink)
    ├── .git/                     <- Empty directory (scope isolation, auto-ignored)
    ├── .jj/
    │   ├── dojo-agent            <- Marker file (auto-ignored)
    │   └── .dojo-bin/
    │       └── git               <- Shim script
    └── [project files]
```

### .jj/dojo-agent Marker

JSON file identifying agent workspaces (inside `.jj/` so auto-ignored by jj):

```json
{
  "root_workspace": "/Users/dev/myproject",
  "name": "feature-auth",
  "created_at": "2025-01-05T10:30:00Z"
}
```

Used for:
- Discovering agent workspaces
- Enabling nested `dojo` calls (from agent, creates sibling to original root)
- Cleanup tracking

### .git Directory

- Empty directory at `<workspace>/.git`
- Prevents Claude from detecting parent jj repo
- Makes Claude treat workspace as standalone project root
- Auto-ignored by jj (won't appear in `jj status`)

### Git Shim

- Script at `<workspace>/.jj/.dojo-bin/git`
- Returns exit 1 with message "git disabled for agents; use jj"
- PATH prepended so shim shadows real git

## Multi-Agent Model

- User opens multiple terminals for multiple agents
- Each `dojo <name>` is independent
- No centralized orchestration
- Version control via jj directly in default workspace
- Can run `dojo` from agent workspace - creates sibling to original root

## Design Decisions

| Question       | Decision                                                        |
| -------------- | --------------------------------------------------------------- |
| TTY approach   | Fork with terminal passthrough (not exec, not PTY multiplexing) |
| Workspace path | Sibling: `../<repo>-<name>/` for visibility                     |
| Permissions    | Full copy of `.claude/` (jj workspace add copies everything)    |
| Discovery      | `.jj/dojo-agent` marker file (JSON, auto-ignored)               |
| jj status      | Clean - markers in `.jj/` or `.git/` are auto-ignored           |
| List output    | Name + jj change-id + commit-id + summary                       |
| Nesting        | Works - always creates siblings to original root                |
| Cleanup        | Prompt on exit + warn about uncommitted changes                 |
| Workspace UI   | None - CLI only                                                 |
| Diff view      | None - Claude can run jj commands                               |
| Multi-agent    | Separate terminals                                              |
| CLI commands   | `dojo <name>`, `dojo list`                                      |
| Git shim       | Keep (forces jj usage)                                          |
| Claude args    | None - user interacts directly                                  |

## Dependencies

- `os/exec` (stdlib)
- `encoding/json` (stdlib)
- `internal/jj` (workspace operations)
- No external libraries
