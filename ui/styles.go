package ui

import "github.com/charmbracelet/lipgloss"

// Minimal styles that follow the terminal theme.
var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	StatusStyle   = lipgloss.NewStyle().Faint(true)
	SelectedStyle = lipgloss.NewStyle().Bold(true)
	HintStyle     = lipgloss.NewStyle().Faint(true)
)
