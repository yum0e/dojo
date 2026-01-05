package tui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	colorMint      = lipgloss.Color("#4ECCA3") // Modern mint green - primary accent
	colorWhite     = lipgloss.Color("#FAFAFA")
	colorOrange    = lipgloss.Color("#FF9F43") // Nice orange for Claude
	colorGray      = lipgloss.Color("#666666")
	colorDimGray   = lipgloss.Color("#555555")
	colorBorder    = lipgloss.Color("#333333")
	colorHighlight = lipgloss.Color("#2A2A2A")
	colorDark      = lipgloss.Color("#1A1A1A")

	// Chat message backgrounds - more visible
	colorUserMsgBg   = lipgloss.Color("#2A2A2A") // Visible lighter background for user
	colorClaudeMsgBg = lipgloss.Color("#1F1F1F") // Subtle dark background for Claude
	colorOrangeDim   = lipgloss.Color("#CC7A30") // Dimmer orange for context text
)

// Title styles (kept for compatibility but not used)
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorMint).
			Padding(0, 1)
)

// Pane border styles
var (
	PaneBorderFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMint)

	PaneBorderUnfocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder)
)

// Workspace list item styles
var (
	WorkspaceItemNormal = lipgloss.NewStyle().
				Foreground(colorWhite)

	WorkspaceItemSelected = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorHighlight).
				Bold(true)
)

// Agent state indicators
const (
	IndicatorDefault = "●" // Mint - default workspace
	IndicatorRunning = "◐" // Mint - agent running
	IndicatorIdle    = "○" // Gray - agent idle
)

// Spinner frames for activity indicator
var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var (
	IndicatorDefaultStyle = lipgloss.NewStyle().
				Foreground(colorMint)

	IndicatorRunningStyle = lipgloss.NewStyle().
				Foreground(colorMint)

	IndicatorIdleStyle = lipgloss.NewStyle().
				Foreground(colorDimGray)
)

// Help bar style
var HelpStyle = lipgloss.NewStyle().
	Foreground(colorGray)

// Key style for highlighting keybindings
var KeyStyle = lipgloss.NewStyle().
	Foreground(colorMint).
	Bold(true)

// Error style
var ErrorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF6B6B")).
	Bold(true)

// Empty diff style
var EmptyDiffStyle = lipgloss.NewStyle().
	Foreground(colorGray).
	Italic(true)

// Confirm dialog styles
var (
	ConfirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMint).
			Padding(1, 2).
			Background(colorDark)

	ConfirmPromptStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Bold(true)
)

// Tab bar styles
var (
	TabBarStyle = lipgloss.NewStyle().
			Background(colorDark)

	TabActiveStyle = lipgloss.NewStyle().
			Foreground(colorDark).
			Background(colorMint).
			Bold(true).
			Padding(0, 1)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorGray).
				Background(colorHighlight).
				Padding(0, 1)
)

// Chat message styles
var (
	ChatUserStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	ChatAgentStyle = lipgloss.NewStyle().
			Foreground(colorOrange).
			Bold(true)

	ChatUserMsgStyle = lipgloss.NewStyle().
				Background(colorUserMsgBg).
				Padding(0, 1)

	ChatAgentMsgStyle = lipgloss.NewStyle().
				Background(colorClaudeMsgBg).
				Padding(0, 1)

	ChatToolStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	ChatToolSuccessStyle = lipgloss.NewStyle().
				Foreground(colorMint)

	ChatToolErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B"))

	ChatModeNormalStyle = lipgloss.NewStyle().
				Foreground(colorGray)

	ChatModeInsertStyle = lipgloss.NewStyle().
				Foreground(colorMint).
				Bold(true)

	ChatProcessingStyle = lipgloss.NewStyle().
				Foreground(colorOrangeDim).
				Italic(true)

	ChatSpinnerStyle = lipgloss.NewStyle().
				Foreground(colorOrange)
)
