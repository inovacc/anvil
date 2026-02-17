package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("111")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))

	blurredInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("111")).
			Bold(true).
			Width(14)

	// Sidebar styles
	sidebarStyle = lipgloss.NewStyle().
			Width(sidebarWidth).
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 0, 0, 1)

	sidebarCategoryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("111")).
				Bold(true)

	sidebarItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	sidebarSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Bold(true)

	sidebarSelectedFocusedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("229")).
					Bold(true).
					Background(lipgloss.Color("62"))
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("62")).
		BorderBottom(true).
		Foreground(lipgloss.Color("111")).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("252"))
	return s
}
