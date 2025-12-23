package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// ClearScreen clears the terminal screen and sets black background
func ClearScreen() {
	// Clear screen and set black background with white text
	fmt.Print("\033[H\033[2J\033[40m\033[37m")
}

// ShowBanner displays the application banner with model info
func ShowBanner(model string) {
	// ASCII art logo
	logo := LogoStyle.Render(`
    ⚔️  🎲                      🎲  ⚔️
   ╔══════════════════════════════╗
   ║      SkillsWeaver            ║
   ║   ~ Dungeon Master Agent ~   ║
   ╚══════════════════════════════╝
    ⚔️  🎲                      🎲  ⚔️
`)

	info := InfoStyle.Render(fmt.Sprintf(
		"Model: %s | Author: Nicolas Martignole | CC BY-SA 4.0",
		model,
	))

	fmt.Println()
	fmt.Println(logo)
	fmt.Println(info)
	fmt.Println()
}

// ShowAdventureInfo displays adventure details in a styled box
func ShowAdventureInfo(name, location string, gold int, party string, lastAction string) {
	// Adventure title
	title := AdventureTitleStyle.Render(name)
	fmt.Println(title)
	fmt.Println()

	// Info content
	var content string
	if location != "" {
		content += fmt.Sprintf("📍 %s\n", location)
	}
	content += fmt.Sprintf("💰 %d po\n", gold)
	if party != "" {
		content += fmt.Sprintf("👥 %s", party)
	}

	fmt.Println(InfoBoxStyle.Render(content))

	// Last action if available
	if lastAction != "" {
		fmt.Println()
		lastActionStyle := lipgloss.NewStyle().Foreground(White)
		fmt.Println(lastActionStyle.Render(fmt.Sprintf("📖 %s", lastAction)))
	}
	fmt.Println()
}

// ShowSeparator displays a visual separator
func ShowSeparator() {
	fmt.Println(SubtitleStyle.Render("─────────────────────────────────────────"))
}
