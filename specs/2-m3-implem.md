# M3: Workspace List Pane - Implementation Plan

## Interview Findings (2025-01-04)

### Decisions

| Topic | Decision |
|-------|----------|
| Repo source | CWD only, error if not jj repo |
| Indicators | Full mock (●/◐/○) for visual testing |
| On selection | Show jj diff in right pane |
| Components | `internal/tui/` package |
| Layout | Adaptive left pane width |
| Tab bar | Deferred to M5 |
| Keybindings | Both vim (j/k) and arrows |
| CRUD | Add (`a`) / Delete (`d`) |
| Create base | Default workspace's working copy |
| Delete confirm | y/n prompt, block default deletion |
| Empty diff | "No changes in this workspace" message |
| Auto-refresh | On workspace focus change + manual `r` |

---

## File Structure

```
internal/tui/
├── app.go              # Root model, orchestrates child components
├── messages.go         # Custom message types
├── styles.go           # Lipgloss style definitions
├── workspace_list.go   # Left pane: workspace list component
├── diff_view.go        # Right pane: diff display component
├── confirm.go          # Y/N confirmation dialog
└── tui_test.go         # Unit tests

cmd/dojo/main.go        # Wire up tui.NewApp() with error handling
internal/jj/client.go   # Add runInDir() for workspace-specific commands
```

---

## Implementation Phases

### Phase 1: Foundation
**Files:** `styles.go`, `messages.go`

1. Create `internal/tui/styles.go`:
   - Title, pane borders (focused/unfocused)
   - Workspace item styles (normal/selected)
   - Indicators: `●` default (green), `◐` running (gold), `○` idle (gray)
   - Help bar, error, empty diff styles

2. Create `internal/tui/messages.go`:
   - `WorkspacesLoadedMsg` - workspace list fetched
   - `DiffLoadedMsg` - diff content fetched
   - `WorkspaceSelectedMsg` - selection changed
   - `ConfirmDeleteMsg` / `ConfirmResultMsg` - deletion flow

### Phase 2: Workspace List Component
**File:** `workspace_list.go`

```go
type WorkspaceItem struct {
    jj.Workspace
    State AgentState  // None, Idle, Running (mocked in M3)
}

type WorkspaceListModel struct {
    items      []WorkspaceItem
    cursor     int
    selected   int
    focused    bool
    jjClient   *jj.Client
}
```

- `Init()` → trigger `loadWorkspaces()` command
- `Update()`:
  - `j`/`down` → cursor++
  - `k`/`up` → cursor--
  - `Enter` → select, emit `WorkspaceSelectedMsg`
  - `a` → add workspace (from default's working copy)
  - `d` → trigger delete confirm (block if default)
- `View()` → render list with indicators
- `MinWidth()` → longest name + 4 (indicator + padding)

### Phase 3: Diff View Component
**File:** `diff_view.go`

```go
type DiffViewModel struct {
    content    string
    workspace  string
    scrollY    int
    focused    bool
    jjClient   *jj.Client
}
```

- `loadDiff(workspaceName)` → fetch diff using `jj.Diff()` with `WorkDir` option
- `Update()` → handle scroll keys
- `View()`:
  - Loading: "Loading..."
  - Empty: "No changes in this workspace"
  - Content: raw diff with ANSI colors

**Requires:** Add `runInDir()` to `internal/jj/client.go` for non-default workspaces.

### Phase 4: Confirmation Dialog
**File:** `confirm.go`

```go
type ConfirmModel struct {
    prompt   string
    visible  bool
    action   string
    data     interface{}
}
```

- `Show(prompt, action, data)` → display modal
- `Update()`: `y` → confirm, `n`/`esc` → cancel
- `View()` → "Delete workspace X? (y/n)"

### Phase 5: App Integration
**File:** `app.go`

```go
type AppModel struct {
    workspaceList WorkspaceListModel
    diffView      DiffViewModel
    confirm       ConfirmModel
    jjClient      *jj.Client
    focusedPane   FocusedPane  // WorkspaceList | DiffView
    width, height int
}
```

- `NewApp()`:
  - Create jj.Client
  - Validate jj repo via `WorkspaceRoot()`, exit if `ErrNotJJRepo`
  - Initialize child components

- `Init()` → trigger workspace loading

- `Update()`:
  - `tea.WindowSizeMsg` → recalculate layout
  - `q`/`ctrl+c` → quit
  - `tab` → toggle focus
  - `r` → refresh diff
  - Route to confirm dialog if visible
  - Route to focused child component
  - Handle `WorkspaceSelectedMsg` → trigger diff reload

- `View()`:
  - Title bar: "DOJO"
  - Two-pane layout via `lipgloss.JoinHorizontal()`
  - Help bar: "j/k: navigate | Enter: select | a: add | d: delete | r: refresh | q: quit"
  - Overlay confirm dialog when visible

- Workspace add flow:
  1. Generate name: `agent-1`, `agent-2`, etc.
  2. Get default workspace path
  3. Call `jjClient.WorkspaceAdd(ctx, path, defaultWorkingCopy)`
  4. Refresh workspace list

- Workspace delete flow:
  1. Block if name == "default"
  2. Show confirm dialog
  3. On confirm: `jjClient.WorkspaceForget(ctx, name)`
  4. Refresh workspace list

### Phase 6: Entry Point Update
**File:** `cmd/dojo/main.go`

```go
func main() {
    app, err := tui.NewApp()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    p := tea.NewProgram(app, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        os.Exit(1)
    }
}
```

### Phase 7: Testing
**File:** `internal/tui/tui_test.go`

- `TestWorkspaceListMinWidth` - adaptive width calculation
- `TestWorkspaceNavigation` - j/k/arrows move cursor
- `TestConfirmDialog` - y/n flow
- `TestDefaultWorkspaceCannotBeDeleted` - block delete on default

---

## Layout Calculation

```
┌─────────────────────────────────────────────────────────────┐
│ DOJO                                                        │ <- Title (1 line)
├─────────────────┬───────────────────────────────────────────┤
│ ● default       │                                           │
│ ◐ agent-1       │  (diff content with ANSI colors)          │
│ ○ agent-2       │                                           │
│                 │  or "No changes in this workspace"        │
│                 │                                           │
├─────────────────┴───────────────────────────────────────────┤
│ j/k: navigate | Enter: select | a: add | d: delete | q: quit│ <- Help (1 line)
└─────────────────────────────────────────────────────────────┘

Left width  = max(workspace_name_lengths) + 4
Right width = terminal_width - left_width - 3 (borders)
Height      = terminal_height - 4 (title + help + borders)
```

---

## jj Client Enhancement

Add to `internal/jj/client.go`:

```go
func (c *Client) runInDir(ctx context.Context, dir string, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, c.jjPath, args...)
    cmd.Dir = dir
    // ... same error handling as run()
}
```

Update `internal/jj/diff.go`:

```go
type DiffOptions struct {
    Revision string
    WorkDir  string  // NEW: directory to run diff in
}
```

---

## Mock Agent States (M3 Only)

Without M4 agents, mock states for visual testing:

```go
func mockAgentStates(workspaces []jj.Workspace) []WorkspaceItem {
    items := make([]WorkspaceItem, len(workspaces))
    for i, ws := range workspaces {
        items[i] = WorkspaceItem{Workspace: ws, State: AgentStateNone}
        if strings.HasPrefix(ws.Name, "agent-") {
            // Alternate between running/idle for visual testing
            if i%2 == 0 {
                items[i].State = AgentStateRunning
            } else {
                items[i].State = AgentStateIdle
            }
        }
    }
    return items
}
```

---

## Key Keybindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Enter` | Select workspace, load diff |
| `a` | Add new workspace |
| `d` | Delete workspace (with confirm) |
| `r` | Refresh diff |
| `Tab` | Toggle focus between panes |
| `q` / `Ctrl+C` | Quit |

---

## Success Criteria

After M3:
- [ ] Run `go run ./cmd/dojo` in a jj repo
- [ ] See two-pane layout with workspace list
- [ ] Navigate with j/k/arrows
- [ ] Select workspace shows its diff
- [ ] Press `a` creates new agent workspace
- [ ] Press `d` deletes agent workspace (with confirm)
- [ ] Press `r` refreshes diff
- [ ] "Not a jj repo" error when run outside jj repo
