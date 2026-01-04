# M2: jj Client Implementation Plan

## Overview

Go package wrapping Jujutsu CLI commands. CWD-based design (caller manages directory).

## File Structure

```
internal/jj/
├── errors.go       # Step 1: Typed errors
├── client.go       # Step 2: Command executor
├── workspace.go    # Step 3: Workspace operations
├── log.go          # Step 4: Log parsing
├── status.go       # Step 5: Status parsing
├── diff.go         # Step 6: Diff retrieval
├── ops.go          # Step 7: Git operations
└── jj_test.go      # Integration tests
```

---

## Step 1: errors.go

Typed errors for jj operations.

```go
var (
    ErrNotJJRepo         = errors.New("not a jj repository")
    ErrWorkspaceExists   = errors.New("workspace already exists")
    ErrWorkspaceNotFound = errors.New("workspace not found")
)

type CommandError struct {
    Cmd    string
    Stderr string
    Err    error
}
```

**Detection patterns:**

| Error | Stderr contains |
|-------|-----------------|
| ErrNotJJRepo | `"There is no jj repo in"` |
| ErrWorkspaceExists | `"already exists"` |
| ErrWorkspaceNotFound | `"No such workspace"` |

---

## Step 2: client.go

Command executor foundation.

```go
type Client struct {
    jjPath string // default: "jj"
}

func NewClient(opts ...Option) *Client
func WithJJPath(path string) Option

// Internal execution
func (c *Client) run(ctx context.Context, args ...string) (string, error)
func (c *Client) parseError(subcmd, stderr string, err error) error
```

---

## Step 3: workspace.go

Workspace management.

```go
type Workspace struct {
    Name     string // "default", "agent-1"
    ChangeID string // Short change ID
    CommitID string // Short commit ID
    Summary  string // Description
}

func (c *Client) WorkspaceAdd(ctx context.Context, path, revision string) error
func (c *Client) WorkspaceForget(ctx context.Context, name string) error
func (c *Client) WorkspaceList(ctx context.Context) ([]Workspace, error)
func (c *Client) WorkspaceRoot(ctx context.Context) (string, error)
```

**Parsing `jj workspace list`:**

```
default: wpxqlmox f3c3a79d (no description set)
```

Regex: `^(\S+): (\S+) (\S+) (.*)$`

---

## Step 4: log.go

Commit log with template parsing.

```go
type Commit struct {
    ChangeID      string
    ChangeIDShort string
    Description   string
    Author        string
    Timestamp     time.Time
    IsWorkingCopy bool
}

type LogOptions struct {
    Revisions string
    Limit     int
}

func (c *Client) Log(ctx context.Context, opts *LogOptions) ([]Commit, error)
```

**Use jj template for reliable parsing:**

```
jj log --no-graph -T 'change_id.short() ++ "|" ++ description.first_line() ++ "|" ++ author.email() ++ "|" ++ author.timestamp() ++ "\n"'
```

---

## Step 5: status.go

Working copy status.

```go
type FileStatus struct {
    Status string // M, A, D, C
    Path   string
}

type Status struct {
    WorkingCopy  string
    ParentCommit string
    Changes      []FileStatus
    HasConflicts bool
}

func (c *Client) Status(ctx context.Context) (*Status, error)
```

---

## Step 6: diff.go

Raw diff output (no parsing).

```go
type DiffOptions struct {
    Revision string
}

func (c *Client) Diff(ctx context.Context, opts *DiffOptions) (string, error)
```

Command: `jj diff --color=always`

---

## Step 7: ops.go

Mutating operations.

```go
func (c *Client) Commit(ctx context.Context, message string) error
func (c *Client) Squash(ctx context.Context) error
func (c *Client) SquashInto(ctx context.Context, from, into string) error
func (c *Client) Rebase(ctx context.Context, destination string) error
func (c *Client) Describe(ctx context.Context, message string) error
func (c *Client) DescribeRevision(ctx context.Context, revision, message string) error
func (c *Client) GitPush(ctx context.Context) error
func (c *Client) GitPushBookmark(ctx context.Context, bookmark string) error
```

---

## Testing Strategy

**Single test file:** `internal/jj/jj_test.go`

**Test helper:**

```go
func setupTestRepo(t *testing.T) (repoPath string, cleanup func()) {
    tmpDir, _ := os.MkdirTemp("", "jj-test-*")
    exec.Command("git", "init").Dir = tmpDir
    exec.Command("jj", "git", "init", "--colocate").Dir = tmpDir
    return tmpDir, func() { os.RemoveAll(tmpDir) }
}

func runInDir(t *testing.T, dir string, fn func()) {
    orig, _ := os.Getwd()
    os.Chdir(dir)
    defer os.Chdir(orig)
    fn()
}
```

**Test cases:**

- errors: CommandError formatting, Unwrap, errors.Is
- client: NewClient defaults, run in non-repo returns ErrNotJJRepo
- workspace: List, Add, Add duplicate, Forget, Forget missing
- log: Parse commits, limit option
- status: Clean repo, modified files
- diff: Empty diff, with changes
- ops: Commit, Squash, Describe

---

## jj Commands Reference

| Operation | Command |
|-----------|---------|
| Add workspace | `jj workspace add <path> -r <rev>` |
| Forget workspace | `jj workspace forget <name>` |
| List workspaces | `jj workspace list` |
| Workspace root | `jj workspace root` |
| Log | `jj log --no-graph -T <template>` |
| Status | `jj status` |
| Diff | `jj diff --color=always` |
| Commit | `jj commit -m <msg>` |
| Squash | `jj squash` |
| Rebase | `jj rebase -d <dest>` |
| Describe | `jj describe -m <msg>` |
| Git push | `jj git push` |

---

## Files to Create

1. `internal/jj/errors.go`
2. `internal/jj/client.go`
3. `internal/jj/workspace.go`
4. `internal/jj/log.go`
5. `internal/jj/status.go`
6. `internal/jj/diff.go`
7. `internal/jj/ops.go`
8. `internal/jj/jj_test.go`
