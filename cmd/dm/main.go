package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"dungeons/internal/adventure"
	"dungeons/internal/agent"
	"dungeons/internal/ui"
)

const (
	dataDir       = "data"
	adventuresDir = "data/adventures"
)

func main() {
	// Clear screen at startup
	ui.ClearScreen()

	// Initial title (will be replaced by banner after adventure selection)
	fmt.Println(ui.SubtitleStyle.Render("SkillsWeaver - Sélection d'aventure"))

	// Check API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY environment variable not set")
		fmt.Fprintln(os.Stderr, "Please set it in your .envrc file or export it")
		os.Exit(1)
	}

	// List adventures
	adventures, err := listAdventures()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing adventures: %v\n", err)
		os.Exit(1)
	}

	if len(adventures) == 0 {
		fmt.Println(ui.ErrorStyle.Render("No adventures found in " + adventuresDir))
		fmt.Println(ui.MenuItemStyle.Render("Create an adventure first using: ./sw-adventure create \"<name>\" \"<description>\""))
		os.Exit(1)
	}

	// Show menu
	selectedAdventure := showAdventureMenu(adventures)
	if selectedAdventure == "" {
		fmt.Println(ui.SubtitleStyle.Render("No adventure selected. Exiting."))
		return
	}

	// Clear screen and show banner
	ui.ClearScreen()
	ui.ShowBanner("Claude Haiku 4.5")

	// Load adventure context
	fmt.Println(ui.SubtitleStyle.Render(fmt.Sprintf("Chargement de l'aventure '%s'...\n", selectedAdventure)))
	adventureCtx, err := agent.LoadAdventureContext(adventuresDir, selectedAdventure)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading adventure: %v\n", err)
		os.Exit(1)
	}

	// Create output handler
	terminalOutput := NewTerminalOutput()

	// Create agent (tools are registered automatically in New)
	dmAgent, err := agent.New(apiKey, adventureCtx, terminalOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// Display welcome
	displayWelcome(adventureCtx)

	// Start REPL
	fmt.Println(ui.SubtitleStyle.Render("Tapez 'exit' ou 'quit' pour quitter.\n"))
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(ui.PromptStyle.Render("> "))
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println(ui.MenuItemStyle.Render("\nAu revoir, aventuriers !"))
			break
		}

		// Process user message
		fmt.Println()
		if err := dmAgent.ProcessUserMessage(input); err != nil {
			fmt.Fprintf(os.Stderr, "\nErreur: %v\n", err)
		}
		fmt.Println()
	}
}

// listAdventures lists all available adventures.
func listAdventures() ([]adventure.Adventure, error) {
	entries, err := os.ReadDir(adventuresDir)
	if err != nil {
		return nil, err
	}

	adventures := []adventure.Adventure{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		adv, err := adventure.LoadByName(adventuresDir, entry.Name())
		if err != nil {
			continue
		}

		adventures = append(adventures, *adv)
	}

	return adventures, nil
}

// showAdventureMenu displays the adventure selection menu.
func showAdventureMenu(adventures []adventure.Adventure) string {
	fmt.Println(ui.MenuItemStyle.Render("Aventures disponibles:\n"))

	for i, adv := range adventures {
		timeSince := time.Since(adv.LastPlayed)
		timeStr := formatTimeSince(timeSince)

		fmt.Println(ui.MenuItemStyle.Render(fmt.Sprintf("%d. %s", i+1, adv.Name)))
		fmt.Println(ui.SubtitleStyle.Render(fmt.Sprintf("   Dernière session: %s", timeStr)))
		fmt.Println(ui.SubtitleStyle.Render(fmt.Sprintf("   Sessions: %d | Statut: %s\n", adv.SessionCount, adv.Status)))
	}

	fmt.Println(ui.MenuItemStyle.Render("0. Quitter"))
	fmt.Print(ui.MenuItemStyle.Render(fmt.Sprintf("\nChoisissez une aventure (1-%d): ", len(adventures))))

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return ""
	}

	choice := strings.TrimSpace(scanner.Text())
	if choice == "0" {
		return ""
	}

	// Parse choice
	var idx int
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(adventures) {
		fmt.Println(ui.ErrorStyle.Render("Choix invalide"))
		return ""
	}

	return adventures[idx-1].Name
}

// displayWelcome displays the welcome message with adventure context.
func displayWelcome(ctx *agent.AdventureContext) {
	// Build party string
	var partyNames []string
	for _, charName := range ctx.Party.Characters {
		for _, char := range ctx.Characters {
			if char.Name == charName {
				partyNames = append(partyNames, fmt.Sprintf("%s (%s %s)", char.Name, char.Race, char.Class))
				break
			}
		}
	}
	partyStr := strings.Join(partyNames, ", ")

	// Get last action
	var lastAction string
	if len(ctx.RecentJournal) > 0 {
		lastEntry := ctx.RecentJournal[len(ctx.RecentJournal)-1]
		lastAction = lastEntry.Content
	}

	// Display using UI package
	ui.ShowAdventureInfo(
		ctx.Adventure.Name,
		ctx.State.CurrentLocation,
		ctx.Inventory.Gold,
		partyStr,
		lastAction,
	)
}

// formatTimeSince formats a duration as a human-readable string.
func formatTimeSince(d time.Duration) string {
	if d < time.Hour {
		return "il y a moins d'une heure"
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "il y a 1 heure"
		}
		return fmt.Sprintf("il y a %d heures", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "il y a 1 jour"
	}
	if days < 7 {
		return fmt.Sprintf("il y a %d jours", days)
	}
	weeks := days / 7
	if weeks == 1 {
		return "il y a 1 semaine"
	}
	if weeks < 4 {
		return fmt.Sprintf("il y a %d semaines", weeks)
	}
	months := days / 30
	if months == 1 {
		return "il y a 1 mois"
	}
	return fmt.Sprintf("il y a %d mois", months)
}

// TerminalOutput implements the OutputHandler interface for terminal display.
type TerminalOutput struct{}

// NewTerminalOutput creates a new terminal output handler.
func NewTerminalOutput() *TerminalOutput {
	return &TerminalOutput{}
}

// OnTextChunk displays a text chunk immediately (streaming).
func (to *TerminalOutput) OnTextChunk(text string) {
	fmt.Print(text)
}

// OnToolStart displays when a tool starts executing.
func (to *TerminalOutput) OnToolStart(toolName, toolID string) {
	msg := fmt.Sprintf("\n[🎲 %s...]\n", toolName)
	fmt.Print(ui.ToolStyle.Render(msg))
}

// OnToolComplete displays when a tool completes.
func (to *TerminalOutput) OnToolComplete(toolName string, result interface{}) {
	var msg string
	// Extract display message if available
	if m, ok := result.(map[string]interface{}); ok {
		if display, ok := m["display"].(string); ok {
			msg = fmt.Sprintf("[✓ %s]", display)
		} else {
			msg = fmt.Sprintf("[✓ %s complete]", toolName)
		}
	} else {
		msg = fmt.Sprintf("[✓ %s complete]", toolName)
	}
	fmt.Print(ui.ToolStyle.Render(msg))
	fmt.Println() // Ensure newline after tool result
}

// OnError displays an error.
func (to *TerminalOutput) OnError(err error) {
	msg := fmt.Sprintf("\n⚠️  Erreur: %v\n", err)
	fmt.Fprint(os.Stderr, ui.ErrorStyle.Render(msg))
}

// OnComplete is called when the agent finishes processing.
func (to *TerminalOutput) OnComplete() {
	// Nothing to do for terminal output
}
