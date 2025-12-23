// Package ui provides terminal styling utilities using lipgloss
package ui

import "github.com/charmbracelet/lipgloss"

// Colors used throughout the application
var (
	Purple = lipgloss.Color("99")
	Gold   = lipgloss.Color("220")
	Gray   = lipgloss.Color("245")
	Green  = lipgloss.Color("82")
	Red    = lipgloss.Color("196")
)

// Styles for different UI elements
var (
	// TitleStyle for main titles with double border
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Gold).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(Purple).
			Padding(0, 2)

	// SubtitleStyle for subtitles and info lines
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	// PromptStyle for user input prompt
	PromptStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// DMStyle for Dungeon Master narrative text
	DMStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	// ToolStyle for tool execution messages
	ToolStyle = lipgloss.NewStyle().
			Foreground(Purple).
			Faint(true)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	// InfoBoxStyle for adventure info boxes with rounded border
	InfoBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(1, 2)

	// AdventureTitleStyle for adventure name display
	AdventureTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Gold).
				Padding(0, 1)

	// MenuItemStyle for menu items
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// MenuSelectedStyle for selected menu items
	MenuSelectedStyle = lipgloss.NewStyle().
				Foreground(Green).
				Bold(true)
)
