# internal/tui

Bubbletea TUI components for the Dojo application.

## Purpose

This package contains all UI components using the Elm architecture (Model-Update-View). It handles user input, rendering, and coordinates child components.

## Key Files

| File                | Purpose                                                 |
| ------------------- | ------------------------------------------------------- |
| `app.go`            | Root model, orchestrates layout and child components    |
| `workspace_list.go` | Left pane: lists workspaces with state indicators       |
| `right_pane.go`     | Right pane: tabbed container (Chat/Diff)                |
| `chat_view.go`      | Chat tab: vim-style input, streaming output, tool states|
| `diff_view.go`      | Diff tab: displays jj diff output with scrolling        |
| `confirm.go`        | Modal Y/N confirmation dialog                           |
| `styles.go`         | Lipgloss style definitions (colors, borders, tabs)      |
| `messages.go`       | Custom tea.Msg types for component communication        |

## Architecture

```
AppModel
├── WorkspaceListModel  (left pane, focused by default)
├── RightPaneModel      (right pane, tabbed)
│   ├── ChatViewModel   (Chat tab - agent interaction)
│   └── DiffViewModel   (Diff tab - jj diff output)
├── ConfirmModel        (overlay dialog)
└── agentManager        (eager init in NewApp, manages agent processes)
```

## Theme & Styling

- Primary accent color: mint green (`#4ECCA3`)
- Keybindings in help bar are highlighted in mint green
- No title bar (clean, minimal UI)
- Focused panes have mint green borders

## Chat View Features

- Vim-style input: Normal mode (j/k scroll, i insert) and Insert mode (Enter submit, Shift+Enter newline)
- Smart scroll: auto-scrolls only when at bottom
- Activity spinner: shows animated spinner with current activity (e.g., "ls -la", "reading file.go", "task: explore codebase")
- Cancel: press Esc in normal mode to cancel current operation
- Auto-spawn: agents spawned automatically when entering Chat tab

## Workspace Management

- Press `a` to create a new workspace
- Input field appears pre-filled with next available `agent-N` name
- Edit name or keep default, then press Enter to create
- Press Esc to cancel workspace creation
- Valid characters: a-z, A-Z, 0-9, hyphen, underscore

## Tab Navigation

- `Shift+Tab` or `Ctrl+Tab`: cycle between Chat/Diff tabs
- Tab bar always visible for consistent layout
- Default workspace shows "No chat" message in Chat tab (centered box with mint border)
- Tab preference remembered per workspace (non-default only)

## App Initialization Sequence

The correct initialization order in `main()` is critical:

```go
app, _ := tui.NewApp()           // 1. Create app (creates agentManager eagerly)
p := tea.NewProgram(app, ...)    // 2. Create tea.Program
app.SetProgram(p)                // 3. Set program reference
app.StartEventListener()         // 4. Start event listener goroutine
p.Run()                          // 5. Run the TUI
app.Shutdown()                   // 6. Cleanup on exit
```

### Why This Order Matters

- **agentManager is created eagerly in `NewApp()`** - not lazily, because the event listener needs a stable reference
- **`StartEventListener()` MUST be called from `main()`** - NOT from `Update()`. Bubbletea's Update method operates on model copies, so goroutines started from Update hold stale pointers to copied structs
- **`SetProgram()` before `StartEventListener()`** - the listener needs the program reference to send events

## Important Notes

- Agent workspaces are created at `.jj/agents/<name>/` (not repo root)
- `DeleteWorkspace()` removes both jj workspace and the directory
- Agent events flow: `manager.Events()` → `StartEventListener` goroutine → `tea.Program.Send()`
- Error messages from agent are shown in chat (stderr captured, SendInput errors displayed)
- Retry (`r` key) falls back to `StartAgent` if process not in manager map (handles app restart scenarios)

## Key Messages

| Message               | Purpose                                    |
| --------------------- | ------------------------------------------ |
| `AgentEventMsg`       | Wraps agent.Event from manager             |
| `SpawnAgentMsg`       | Request to spawn agent for workspace       |
| `SpawnAgentResultMsg` | Result of spawn attempt                    |
| `ChatInputMsg`        | Send user input to agent                   |
| `RestartAgentMsg`     | Restart crashed agent                      |
| `CancelAgentMsg`      | Cancel current agent operation (Esc key)   |
| `StatusFlashClearMsg` | Clear temporary error flash                |
| `SpinnerTickMsg`      | Advances spinner animation frame           |

## When to Look Here

- UI bugs or styling issues
- Keybinding changes
- Chat/agent interaction issues
- Tab switching behavior
- Layout/responsive design issues
