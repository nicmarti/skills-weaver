package ui

import "fmt"

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// ShowBanner displays the application banner with model info
func ShowBanner(model string) {
	banner := TitleStyle.Render("SkillsWeaver")
	subtitle := SubtitleStyle.Render("Dungeon Master Agent")
	info := SubtitleStyle.Render(fmt.Sprintf(
		"Model: %s | Author: Nicolas Martignole | License: CC BY-SA 4.0",
		model,
	))

	fmt.Println()
	fmt.Println(banner)
	fmt.Println(subtitle)
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
		fmt.Println(SubtitleStyle.Render(fmt.Sprintf("📖 %s", lastAction)))
	}
	fmt.Println()
}

// ShowSeparator displays a visual separator
func ShowSeparator() {
	fmt.Println(SubtitleStyle.Render("─────────────────────────────────────────"))
}
