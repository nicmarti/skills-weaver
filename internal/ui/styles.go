// Package ui provides terminal styling utilities using lipgloss
package ui

import "github.com/charmbracelet/lipgloss"

// Colors for dark background terminal
var (
	Purple      = lipgloss.Color("141")   // Lighter purple for visibility
	Gold        = lipgloss.Color("220")   // Gold/yellow
	LightGray   = lipgloss.Color("250")   // Light gray
	Green       = lipgloss.Color("120")   // Bright green
	Red         = lipgloss.Color("196")   // Bright red
	White       = lipgloss.Color("255")   // Pure white
	Cyan        = lipgloss.Color("51")    // Cyan for accents
	DarkGray    = lipgloss.Color("240")   // Dark gray for subtle elements
)

// Styles for different UI elements (optimized for black background)
var (
	// LogoStyle for ASCII art logo
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Gold).
			Align(lipgloss.Center)

	// InfoStyle for info lines (model, author, license)
	InfoStyle = lipgloss.NewStyle().
			Foreground(LightGray).
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
			Foreground(LightGray).
			Italic(true)

	// PromptStyle for user input prompt (bright green)
	PromptStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// DMStyle for Dungeon Master narrative text (white on black)
	DMStyle = lipgloss.NewStyle().
		Foreground(White)

	// ToolStyle for tool execution messages (lighter purple)
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
			Foreground(White).
			Padding(1, 2)

	// AdventureTitleStyle for adventure name display
	AdventureTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Gold).
				Padding(0, 1)

	// MenuItemStyle for menu items
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(White)

	// MenuSelectedStyle for selected menu items
	MenuSelectedStyle = lipgloss.NewStyle().
				Foreground(Green).
				Bold(true)
)
