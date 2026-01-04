package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel is the model for the confirmation dialog.
type ConfirmModel struct {
	prompt  string
	visible bool
	action  string
	data    interface{}
	width   int
	height  int
}

// NewConfirmModel creates a new confirmation dialog model.
func NewConfirmModel() ConfirmModel {
	return ConfirmModel{}
}

// Show displays the confirmation dialog.
func (m *ConfirmModel) Show(prompt, action string, data interface{}) {
	m.prompt = prompt
	m.action = action
	m.data = data
	m.visible = true
}

// Hide hides the confirmation dialog.
func (m *ConfirmModel) Hide() {
	m.visible = false
	m.prompt = ""
	m.action = ""
	m.data = nil
}

// Visible returns whether the dialog is visible.
func (m ConfirmModel) Visible() bool {
	return m.visible
}

// Init initializes the confirmation dialog.
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the confirmation dialog.
func (m ConfirmModel) Update(msg tea.Msg) (ConfirmModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			result := ConfirmResultMsg{
				Confirmed: true,
				Action:    m.action,
				Data:      m.data,
			}
			m.Hide()
			return m, func() tea.Msg { return result }

		case "n", "N", "esc":
			result := ConfirmResultMsg{
				Confirmed: false,
				Action:    m.action,
				Data:      m.data,
			}
			m.Hide()
			return m, func() tea.Msg { return result }
		}
	}

	return m, nil
}

// View renders the confirmation dialog.
func (m ConfirmModel) View() string {
	if !m.visible {
		return ""
	}

	prompt := ConfirmPromptStyle.Render(m.prompt)
	hint := lipgloss.NewStyle().Foreground(colorGray).Render("(y/n)")

	content := fmt.Sprintf("%s %s", prompt, hint)
	box := ConfirmBoxStyle.Render(content)

	return box
}

// SetSize sets the dimensions for centering the dialog.
func (m *ConfirmModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// CenteredView returns the dialog centered on screen.
func (m ConfirmModel) CenteredView() string {
	if !m.visible {
		return ""
	}

	dialog := m.View()

	// Center horizontally and vertically
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)
}
