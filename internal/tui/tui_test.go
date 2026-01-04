package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/bigq/dojo/internal/jj"
)

func TestWorkspaceListMinWidth(t *testing.T) {
	client := jj.NewClient()
	m := NewWorkspaceListModel(client)

	// With no items, MinWidth should be 0 + 4 = 4
	if got := m.MinWidth(); got != 4 {
		t.Errorf("MinWidth() with no items = %d, want 4", got)
	}

	// Simulate loading workspaces
	m, _ = m.Update(WorkspacesLoadedMsg{
		Workspaces: []jj.Workspace{
			{Name: "default"},
			{Name: "agent-1"},
			{Name: "agent-test-long-name"},
		},
	})

	// MinWidth should be longest name (20) + 4 = 24
	expected := len("agent-test-long-name") + 4
	if got := m.MinWidth(); got != expected {
		t.Errorf("MinWidth() = %d, want %d", got, expected)
	}
}

func TestWorkspaceNavigation(t *testing.T) {
	client := jj.NewClient()
	m := NewWorkspaceListModel(client)
	m.SetFocused(true)

	// Load some workspaces
	m, _ = m.Update(WorkspacesLoadedMsg{
		Workspaces: []jj.Workspace{
			{Name: "default"},
			{Name: "agent-1"},
			{Name: "agent-2"},
		},
	})

	// Initial cursor should be 0
	if m.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.cursor)
	}

	// Test j (down)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Errorf("cursor after 'j' = %d, want 1", m.cursor)
	}

	// Test down arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("cursor after down = %d, want 2", m.cursor)
	}

	// Test k (up)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Errorf("cursor after 'k' = %d, want 1", m.cursor)
	}

	// Test up arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("cursor after up = %d, want 0", m.cursor)
	}

	// Test boundary - can't go above 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor != 0 {
		t.Errorf("cursor after up at 0 = %d, want 0", m.cursor)
	}

	// Test boundary - can't go beyond last item
	m.cursor = 2
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor != 2 {
		t.Errorf("cursor after down at end = %d, want 2", m.cursor)
	}
}

func TestConfirmDialog(t *testing.T) {
	m := NewConfirmModel()

	// Initially not visible
	if m.Visible() {
		t.Error("confirm dialog should not be visible initially")
	}

	// Show dialog
	m.Show("Delete workspace 'test'?", "delete", "test")
	if !m.Visible() {
		t.Error("confirm dialog should be visible after Show()")
	}

	// Test 'n' cancels
	m2 := m
	m2, cmd := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m2.Visible() {
		t.Error("confirm dialog should be hidden after 'n'")
	}
	if cmd == nil {
		t.Error("expected command after 'n'")
	} else {
		msg := cmd()
		result, ok := msg.(ConfirmResultMsg)
		if !ok {
			t.Error("expected ConfirmResultMsg")
		}
		if result.Confirmed {
			t.Error("expected Confirmed = false after 'n'")
		}
	}

	// Show again and test 'y' confirms
	m.Show("Delete workspace 'test'?", "delete", "test")
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if m.Visible() {
		t.Error("confirm dialog should be hidden after 'y'")
	}
	if cmd == nil {
		t.Error("expected command after 'y'")
	} else {
		msg := cmd()
		result, ok := msg.(ConfirmResultMsg)
		if !ok {
			t.Error("expected ConfirmResultMsg")
		}
		if !result.Confirmed {
			t.Error("expected Confirmed = true after 'y'")
		}
		if result.Action != "delete" {
			t.Errorf("expected Action = 'delete', got %q", result.Action)
		}
		if result.Data != "test" {
			t.Errorf("expected Data = 'test', got %v", result.Data)
		}
	}
}

func TestDefaultWorkspaceCannotBeDeleted(t *testing.T) {
	client := jj.NewClient()
	m := NewWorkspaceListModel(client)
	m.SetFocused(true)

	// Load workspaces with default first
	m, _ = m.Update(WorkspacesLoadedMsg{
		Workspaces: []jj.Workspace{
			{Name: "default"},
			{Name: "agent-1"},
		},
	})

	// Cursor is on default (index 0)
	if m.cursor != 0 {
		t.Fatalf("cursor should be 0, got %d", m.cursor)
	}

	// Press 'd' - should NOT trigger confirm since it's default workspace
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(ConfirmDeleteMsg); ok {
			t.Error("pressing 'd' on default workspace should not trigger delete confirm")
		}
	}

	// Move to agent-1 and press 'd' - should trigger confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("cursor should be 1, got %d", m.cursor)
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Error("pressing 'd' on agent workspace should trigger command")
	} else {
		msg := cmd()
		confirmMsg, ok := msg.(ConfirmDeleteMsg)
		if !ok {
			t.Error("expected ConfirmDeleteMsg")
		}
		if confirmMsg.WorkspaceName != "agent-1" {
			t.Errorf("expected workspace name 'agent-1', got %q", confirmMsg.WorkspaceName)
		}
	}
}

func TestMockAgentStates(t *testing.T) {
	workspaces := []jj.Workspace{
		{Name: "default"},
		{Name: "agent-1"},
		{Name: "agent-2"},
		{Name: "test-workspace"},
	}

	items := mockAgentStates(workspaces)

	// Default should have AgentStateNone
	if items[0].State != AgentStateNone {
		t.Errorf("default workspace state = %d, want AgentStateNone", items[0].State)
	}

	// agent-1 (index 1, odd) should be idle
	if items[1].State != AgentStateIdle {
		t.Errorf("agent-1 state = %d, want AgentStateIdle", items[1].State)
	}

	// agent-2 (index 2, even) should be running
	if items[2].State != AgentStateRunning {
		t.Errorf("agent-2 state = %d, want AgentStateRunning", items[2].State)
	}

	// Non-agent workspace should have AgentStateNone
	if items[3].State != AgentStateNone {
		t.Errorf("test-workspace state = %d, want AgentStateNone", items[3].State)
	}
}

func TestDiffViewEmptyState(t *testing.T) {
	client := jj.NewClient()
	m := NewDiffViewModel(client)
	m.SetSize(80, 24)

	// Simulate empty diff loaded
	m, _ = m.Update(DiffLoadedMsg{
		Workspace: "default",
		Content:   "",
	})

	view := m.View()
	if view != EmptyDiffStyle.Render("No changes in this workspace") {
		t.Errorf("expected empty diff message, got %q", view)
	}
}
