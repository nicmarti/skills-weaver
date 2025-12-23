// Package ui provides terminal styling utilities using lipgloss
package ui

import "github.com/charmbracelet/lipgloss"

// Adaptive colors that work on both light and dark backgrounds
var (
	Purple = lipgloss.AdaptiveColor{Light: "99", Dark: "141"}
	Gold   = lipgloss.AdaptiveColor{Light: "136", Dark: "220"}
	Gray   = lipgloss.AdaptiveColor{Light: "240", Dark: "250"}
	Green  = lipgloss.AdaptiveColor{Light: "28", Dark: "120"}
	Red    = lipgloss.AdaptiveColor{Light: "160", Dark: "196"}
	Text   = lipgloss.AdaptiveColor{Light: "235", Dark: "255"}
	Cyan   = lipgloss.AdaptiveColor{Light: "37", Dark: "51"}
)

// Styles for different UI elements (work on both light and dark backgrounds)
var (
	// LogoStyle for ASCII art logo
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Gold).
			Align(lipgloss.Center)

	// InfoStyle for info lines (model, author, license)
	InfoStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true).
			Align(lipgloss.Center)

	// TitleStyle for main titles with double border
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Gold).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(Purple).
			Padding(0, 2)

	// SubtitleStyle for subtitles
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	// PromptStyle for user input prompt
	PromptStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// DMStyle for Dungeon Master narrative text
	DMStyle = lipgloss.NewStyle().
		Foreground(Text)

	// ToolStyle for tool execution messages
	ToolStyle = lipgloss.NewStyle().
			Foreground(Purple)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	// InfoBoxStyle for adventure info boxes with rounded border
	InfoBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Cyan).
			Foreground(Text).
			Padding(1, 2)

	// AdventureTitleStyle for adventure name display
	AdventureTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Gold).
				Padding(0, 1)

	// MenuItemStyle for menu items
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(Text)

	// MenuSelectedStyle for selected menu items
	MenuSelectedStyle = lipgloss.NewStyle().
				Foreground(Green).
				Bold(true)
)
