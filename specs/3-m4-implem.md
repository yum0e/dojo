# M4: Agent Spawning - Implementation Spec

## Overview

Implement `internal/agent/` package to spawn and manage Claude Code subprocesses. The manager is TUI-agnostic, communicating via Go channels.

## Interview Decisions (2026-01-04)

| Topic | Decision |
|-------|----------|
| CLI invocation | `claude --output-format stream-json` |
| Workspace naming | User-named on creation |
| Base revision | New empty revision on top of @ (or @ if empty) |
| Lifecycle | Keep workspace on agent exit, user deletes manually |
| Crash handling | Auto-restart with same context |
| M4 scope | Spawn + read output only (M5 handles input) |
| Concurrency | Configurable agent limit |
| Notification | Channel-based (`Events() <-chan Event`) |
| Init prompt | Empty - user sends first message in M5 |
| Events parsed | `assistant`, `tool_use`, `result`, `error` |
| Tool tracking | Yes, expose tool name + status |
| WS creation fail | Return error, don't spawn agent |
| Testing | Mock process + fake JSON stream |
| Persistence | None in M4 (deferred to M7) |
| Signal handling | Forward SIGINT/SIGTERM to agents, wait for exit |
| Shutdown timeout | 30 seconds before SIGKILL |
| Orphan handling | PID file tracking, offer kill/attach on startup |
| CLI flags | Just `--output-format stream-json` for MVP |

---

## Files to Create

### `internal/agent/types.go`

```go
package agent

type State int

const (
    StateIdle State = iota
    StateRunning
    StateStopped
    StateError
)

type EventType int

const (
    EventOutput EventType = iota  // Text output from agent
    EventToolUse                   // Agent started using a tool
    EventToolResult                // Tool execution completed
    EventError                     // Error occurred
    EventStateChange               // Agent state changed
)

type Event struct {
    AgentName string
    Type      EventType
    Data      any  // Type-specific payload
}

type OutputData struct {
    Text string
}

type ToolUseData struct {
    ToolName string
    Input    string
}

type ToolResultData struct {
    ToolName string
    Output   string
    Success  bool
}

type ErrorData struct {
    Message string
    Err     error
}

type StateChangeData struct {
    OldState State
    NewState State
}
```

### `internal/agent/process.go`

```go
package agent

// Process represents a single Claude Code subprocess
type Process struct {
    Name      string
    WorkDir   string
    State     State
    cmd       *exec.Cmd
    stdin     io.WriteCloser
    cancel    context.CancelFunc
    events    chan<- Event
    pid       int
}

// Start spawns the claude process
func (p *Process) Start(ctx context.Context) error

// Stop gracefully stops the process (SIGTERM, then SIGKILL after timeout)
func (p *Process) Stop(timeout time.Duration) error

// SendInput writes to stdin (used by M5)
func (p *Process) SendInput(input string) error

// readOutput goroutine that parses stream-json and emits events
func (p *Process) readOutput(stdout io.Reader)
```

### `internal/agent/protocol.go`

```go
package agent

// Parse Claude Code --output-format stream-json events
// See: https://docs.anthropic.com/en/docs/claude-code

type StreamEvent struct {
    Type string `json:"type"`
    // Fields vary by type
}

type AssistantEvent struct {
    Type    string `json:"type"`    // "assistant"
    Message struct {
        Content []ContentBlock `json:"content"`
    } `json:"message"`
}

type ContentBlock struct {
    Type string `json:"type"` // "text" or "tool_use"
    Text string `json:"text,omitempty"`
    Name string `json:"name,omitempty"` // tool name
}

type ResultEvent struct {
    Type   string `json:"type"` // "result"
    Result string `json:"result"`
}

// ParseEvent parses a single JSON line from stream
func ParseEvent(line []byte) (Event, error)
```

### `internal/agent/manager.go`

```go
package agent

type ManagerConfig struct {
    MaxAgents       int           // Configurable limit
    ShutdownTimeout time.Duration // 30s default
    PIDDir          string        // For orphan tracking
}

type Manager struct {
    config    ManagerConfig
    processes map[string]*Process
    events    chan Event
    mu        sync.RWMutex
    jjClient  *jj.Client
}

func NewManager(cfg ManagerConfig, jjClient *jj.Client) *Manager

// Events returns read-only channel for consumers
func (m *Manager) Events() <-chan Event

// SpawnAgent creates workspace and starts claude process
// 1. Validate agent limit not exceeded
// 2. Create new revision on top of @ (jj new)
// 3. Create workspace at that revision (jj workspace add)
// 4. Spawn claude --output-format stream-json in workspace dir
// 5. Start output reader goroutine
func (m *Manager) SpawnAgent(ctx context.Context, name string) error

// StopAgent gracefully stops an agent (keeps workspace)
func (m *Manager) StopAgent(name string) error

// RestartAgent stops and restarts with same context
func (m *Manager) RestartAgent(ctx context.Context, name string) error

// GetState returns current state of an agent
func (m *Manager) GetState(name string) State

// ListAgents returns all agent names and states
func (m *Manager) ListAgents() map[string]State

// Shutdown stops all agents gracefully
func (m *Manager) Shutdown(ctx context.Context) error

// DetectOrphans checks PID file for orphaned processes
func (m *Manager) DetectOrphans() ([]OrphanInfo, error)

// KillOrphan terminates an orphaned process
func (m *Manager) KillOrphan(pid int) error
```

### `internal/agent/pidfile.go`

```go
package agent

// PID file management for orphan detection

func WritePIDFile(dir, agentName string, pid int) error
func ReadPIDFile(dir, agentName string) (int, error)
func RemovePIDFile(dir, agentName string) error
func ListPIDFiles(dir string) ([]string, error)
func IsProcessRunning(pid int) bool
```

### `internal/agent/runner.go` (for testing)

```go
package agent

// ProcessRunner interface for dependency injection
type ProcessRunner interface {
    Start(ctx context.Context, workDir string, args ...string) (Process, error)
}

// RealRunner uses exec.CommandContext
type RealRunner struct{}

// MockRunner reads from predefined JSON streams
type MockRunner struct {
    Events []string // JSON lines to emit
    Delay  time.Duration
}
```

---

## Implementation Steps

### Step 1: Types and Protocol
1. Create `internal/agent/types.go` with state and event definitions
2. Create `internal/agent/protocol.go` with stream-json parsing
3. Write unit tests with sample JSON payloads

### Step 2: Process Management
1. Create `internal/agent/process.go` with Process struct
2. Implement Start() using exec.CommandContext
3. Implement readOutput() goroutine with JSON line parsing
4. Implement Stop() with SIGTERM + timeout + SIGKILL
5. Write tests using MockRunner

### Step 3: PID File Tracking
1. Create `internal/agent/pidfile.go`
2. Implement write/read/remove/list operations
3. Implement IsProcessRunning() using os.FindProcess + signal 0

### Step 4: Manager
1. Create `internal/agent/manager.go`
2. Implement NewManager with config
3. Implement SpawnAgent():
   - Check agent limit
   - Call jjClient.New() to create revision
   - Call jjClient.WorkspaceAdd() with name
   - Get workspace root path
   - Spawn Process in that directory
   - Write PID file
4. Implement StopAgent(), RestartAgent()
5. Implement Shutdown() for graceful exit
6. Implement DetectOrphans(), KillOrphan()

### Step 5: Testing
1. Create `internal/agent/mock_runner.go`
2. Create test fixtures with sample stream-json
3. Write unit tests for protocol parsing
4. Write integration tests for manager (with mock runner)

### Step 6: TUI Integration
1. Add new message types to `internal/tui/messages.go`:
   - `AgentEventMsg` wrapping agent.Event
2. Create tea.Cmd that reads from manager.Events() channel
3. Update WorkspaceListModel to use real agent states
4. Wire manager into AppModel

---

## jj Operations Needed

The manager needs these jj client operations:

```go
// Create new empty revision on top of current
jjClient.New(ctx) error

// Create workspace at specific revision
jjClient.WorkspaceAdd(ctx, name string, atRevision string) error

// Get workspace root directory
jjClient.WorkspaceRoot(ctx, name string) (string, error)
```

Note: `jj new` may need to be added to `internal/jj/ops.go` if not present.

---

## Stream-JSON Format Reference

Claude Code `--output-format stream-json` emits newline-delimited JSON:

```json
{"type":"system","message":"..."}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{...}}]}}
{"type":"result","result":"..."}
{"type":"error","error":"..."}
```

Events to parse for M4:
- `assistant` with `content[].type == "text"` → OutputData
- `assistant` with `content[].type == "tool_use"` → ToolUseData
- `result` → ToolResultData
- `error` → ErrorData

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Agent limit exceeded | Return `ErrMaxAgentsReached` |
| Workspace creation fails | Return error, don't spawn |
| Process fails to start | Return error, cleanup workspace |
| Process crashes | Emit EventError, auto-restart |
| Restart fails 3 times | Emit EventError, mark StateStopped |
| Shutdown timeout | SIGKILL after 30s |

---

## Testing Strategy

1. **Unit tests** (`*_test.go`):
   - Protocol parsing with fixture JSON
   - PID file operations
   - State machine transitions

2. **Integration tests** (MockRunner):
   - Full spawn/stop lifecycle
   - Event emission verification
   - Restart behavior
   - Orphan detection

3. **Manual testing**:
   - Real Claude Code in temp repo
   - Signal handling (Ctrl+C)
   - Crash recovery
