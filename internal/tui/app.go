package tui

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bigq/dojo/internal/jj"
)

// FocusedPane indicates which pane is focused.
type FocusedPane int

const (
	FocusWorkspaceList FocusedPane = iota
	FocusDiffView
)

// AppModel is the root model for the TUI application.
type AppModel struct {
	workspaceList WorkspaceListModel
	diffView      DiffViewModel
	confirm       ConfirmModel
	jjClient      *jj.Client
	focusedPane   FocusedPane
	width         int
	height        int
	err           error
}

// NewApp creates a new TUI application.
func NewApp() (*AppModel, error) {
	client := jj.NewClient()

	// Validate we're in a jj repo
	ctx := context.Background()
	_, err := client.WorkspaceRoot(ctx)
	if err != nil {
		if errors.Is(err, jj.ErrNotJJRepo) {
			return nil, fmt.Errorf("not a jj repository (or any parent up to mount point /)")
		}
		return nil, fmt.Errorf("failed to detect jj repository: %w", err)
	}

	app := &AppModel{
		workspaceList: NewWorkspaceListModel(client),
		diffView:      NewDiffViewModel(client),
		confirm:       NewConfirmModel(),
		jjClient:      client,
		focusedPane:   FocusWorkspaceList,
	}

	// Set initial focus
	app.workspaceList.SetFocused(true)
	app.diffView.SetFocused(false)

	return app, nil
}

// Init initializes the application.
func (m AppModel) Init() tea.Cmd {
	return m.workspaceList.Init()
}

// Update handles messages for the application.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateLayout()
		return m, nil

	case tea.KeyMsg:
		// Handle confirm dialog first if visible
		if m.confirm.Visible() {
			var cmd tea.Cmd
			m.confirm, cmd = m.confirm.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.toggleFocus()
			return m, nil
		case "r":
			return m, func() tea.Msg { return RefreshDiffMsg{} }
		}

	case ConfirmDeleteMsg:
		m.confirm.Show(
			fmt.Sprintf("Delete workspace '%s'?", msg.WorkspaceName),
			"delete",
			msg.WorkspaceName,
		)
		return m, nil

	case ConfirmResultMsg:
		if msg.Confirmed && msg.Action == "delete" {
			if name, ok := msg.Data.(string); ok {
				return m, m.workspaceList.DeleteWorkspace(name)
			}
		}
		return m, nil
	}

	// Route to child components
	var cmd tea.Cmd

	// Update workspace list
	m.workspaceList, cmd = m.workspaceList.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update diff view
	m.diffView, cmd = m.diffView.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// toggleFocus switches focus between panes.
func (m *AppModel) toggleFocus() {
	if m.focusedPane == FocusWorkspaceList {
		m.focusedPane = FocusDiffView
		m.workspaceList.SetFocused(false)
		m.diffView.SetFocused(true)
	} else {
		m.focusedPane = FocusWorkspaceList
		m.workspaceList.SetFocused(true)
		m.diffView.SetFocused(false)
	}
}

// recalculateLayout recalculates the layout based on terminal size.
func (m *AppModel) recalculateLayout() {
	// Calculate left pane width (adaptive based on workspace names)
	leftWidth := m.workspaceList.MinWidth()
	if leftWidth < 15 {
		leftWidth = 15 // Minimum width
	}
	if leftWidth > m.width/3 {
		leftWidth = m.width / 3 // Max 1/3 of screen
	}

	// Right pane gets remaining width
	rightWidth := m.width - leftWidth - 3 // 3 for borders/gap

	// Height for content (minus title and help bar)
	contentHeight := m.height - 4 // title (1) + help (1) + borders (2)

	m.workspaceList.SetSize(leftWidth, contentHeight)
	m.diffView.SetSize(rightWidth, contentHeight)
	m.confirm.SetSize(m.width, m.height)
}

// View renders the application.
func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Title bar
	title := TitleStyle.Render("DOJO")
	titleBar := lipgloss.NewStyle().Width(m.width).Render(title)

	// Calculate pane dimensions
	leftWidth := m.workspaceList.MinWidth()
	if leftWidth < 15 {
		leftWidth = 15
	}
	if leftWidth > m.width/3 {
		leftWidth = m.width / 3
	}
	rightWidth := m.width - leftWidth - 1 // 1 for gap

	contentHeight := m.height - 4

	// Left pane (workspace list)
	leftBorder := m.workspaceList.borderStyle().
		Width(leftWidth - 2).
		Height(contentHeight)
	leftPane := leftBorder.Render(m.workspaceList.View())

	// Right pane (diff view)
	rightBorder := m.diffView.borderStyle().
		Width(rightWidth - 2).
		Height(contentHeight)
	rightPane := rightBorder.Render(m.diffView.View())

	// Join panes horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Help bar
	helpText := "j/k: navigate | Enter: select | a: add | d: delete | r: refresh | Tab: switch pane | q: quit"
	helpBar := HelpStyle.Width(m.width).Render(helpText)

	// Combine all
	view := lipgloss.JoinVertical(lipgloss.Left, titleBar, content, helpBar)

	// Overlay confirm dialog if visible
	if m.confirm.Visible() {
		// Create overlay
		overlay := m.confirm.CenteredView()
		return overlay
	}

	return view
}
