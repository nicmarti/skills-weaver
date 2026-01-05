// Package main implements the sw-validate CLI for validating game data.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dungeons/internal/data"
)

// Monster represents a monster entry for validation.
type Monster struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TreasureType string `json:"treasure_type"`
}

// MonstersFile represents the monsters.json structure.
type MonstersFile struct {
	Monsters []Monster `json:"monsters"`
}

// GenderNames represents names for a specific gender.
type GenderNames struct {
	First []string `json:"first"`
	Last  []string `json:"last"`
}

// RaceNames represents names for a specific race.
type RaceNames struct {
	Male   GenderNames `json:"male"`
	Female GenderNames `json:"female"`
}

// NamesFile represents the names.json structure.
type NamesFile map[string]RaceNames

// TreasureFile represents the treasure.json structure.
type TreasureFile struct {
	TreasureTypes map[string]interface{} `json:"treasure_types"`
}

// SpellsFile represents the spells.json structure.
type SpellsFile struct {
	Spells    []Spell   `json:"spells"`
	SpellList SpellList `json:"spell_lists"`
}

type Spell struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type SpellList struct {
	Arcane map[string][]string `json:"arcane"`
	Divine map[string][]string `json:"divine"`
}

// LocationNamesFile represents the location-names.json structure.
type LocationNamesFile struct {
	Valdorine map[string]interface{} `json:"valdorine"`
	Karvath   map[string]interface{} `json:"karvath"`
	Lumenciel map[string]interface{} `json:"lumenciel"`
	Astrene   map[string]interface{} `json:"astrene"`
}

var validTreasureTypes = map[string]bool{
	"A": true, "B": true, "C": true, "D": true, "E": true,
	"F": true, "G": true, "H": true, "I": true, "J": true,
	"K": true, "L": true, "M": true, "N": true, "O": true,
	"P": true, "Q": true, "R": true, "S": true, "T": true,
	"U": true, "none": true,
}

func printHelp() {
	fmt.Println(`sw-validate - Validate SkillsWeaver game data

USAGE:
  sw-validate [OPTIONS]
  sw-validate help

OPTIONS:
  --json        Output validation results in JSON format
  --data PATH   Path to data directory (default: "data")
  --help, -h    Show this help message

VALIDATIONS:
  - species.json        : D&D 5e species data
  - equipment.json      : starting_equipment references valid items
  - monsters.json       : treasure_type is valid (A-U or 'none')
  - names.json          : all species have name entries
  - location-names.json : all kingdoms have required location types
  - spells.json         : spell_lists reference valid spell IDs
  - journal.json        : bilingual descriptions consistency and length

EXAMPLES:
  sw-validate              # Validate data in ./data directory
  sw-validate --json       # Output as JSON (for CI/CD integration)
  sw-validate --data /path # Validate data in custom directory

EXIT CODES:
  0  All validations passed (warnings are allowed)
  1  One or more errors found`)
}

func main() {
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	dataDir := flag.String("data", "data", "Path to data directory")
	showHelp := flag.Bool("help", false, "Show help message")
	flag.Bool("h", false, "Show help message (shorthand)")
	flag.Parse()

	if *showHelp || (len(os.Args) > 1 && (os.Args[1] == "help" || os.Args[1] == "-h")) {
		printHelp()
		return
	}

	var allErrors []data.ValidationError

	// Load and validate core game data (races, classes, equipment)
	gd, err := data.Load(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading game data: %v\n", err)
		os.Exit(1)
	}

	allErrors = append(allErrors, data.ValidateGameData(gd)...)

	// Validate monsters
	allErrors = append(allErrors, validateMonsters(*dataDir)...)

	// Validate names coverage
	allErrors = append(allErrors, validateNames(*dataDir, gd)...)

	// Validate location names
	allErrors = append(allErrors, validateLocationNames(*dataDir)...)

	// Validate spells
	allErrors = append(allErrors, validateSpells(*dataDir)...)

	// Validate journal descriptions
	allErrors = append(allErrors, validateJournalDescriptions(*dataDir)...)

	// Output results
	if *jsonOutput {
		outputJSON(allErrors)
	} else {
		outputText(allErrors)
	}

	// Exit with error code if there are errors
	hasErrors := false
	for _, e := range allErrors {
		if e.Severity == "error" {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func validateMonsters(dataDir string) []data.ValidationError {
	var errors []data.ValidationError

	filePath := filepath.Join(dataDir, "monsters.json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, data.ValidationError{
				File:     "monsters.json",
				Field:    "",
				Message:  "file not found",
				Severity: "warning",
			})
			return errors
		}
		errors = append(errors, data.ValidationError{
			File:     "monsters.json",
			Field:    "",
			Message:  fmt.Sprintf("error reading file: %v", err),
			Severity: "error",
		})
		return errors
	}

	var mf MonstersFile
	if err := json.Unmarshal(content, &mf); err != nil {
		errors = append(errors, data.ValidationError{
			File:     "monsters.json",
			Field:    "",
			Message:  fmt.Sprintf("JSON parse error: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Check each monster's treasure_type
	for _, monster := range mf.Monsters {
		if monster.TreasureType == "" {
			errors = append(errors, data.ValidationError{
				File:     "monsters.json",
				Field:    fmt.Sprintf("monsters[%s].treasure_type", monster.ID),
				Message:  "missing treasure_type",
				Severity: "warning",
			})
		} else if !validTreasureTypes[monster.TreasureType] {
			errors = append(errors, data.ValidationError{
				File:     "monsters.json",
				Field:    fmt.Sprintf("monsters[%s].treasure_type", monster.ID),
				Message:  fmt.Sprintf("invalid treasure_type '%s' (valid: A-U or 'none')", monster.TreasureType),
				Severity: "error",
			})
		}
	}

	return errors
}

func validateNames(dataDir string, gd *data.GameData) []data.ValidationError {
	var errors []data.ValidationError

	filePath := filepath.Join(dataDir, "names.json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, data.ValidationError{
				File:     "names.json",
				Field:    "",
				Message:  "file not found",
				Severity: "warning",
			})
			return errors
		}
		errors = append(errors, data.ValidationError{
			File:     "names.json",
			Field:    "",
			Message:  fmt.Sprintf("error reading file: %v", err),
			Severity: "error",
		})
		return errors
	}

	var nf NamesFile
	if err := json.Unmarshal(content, &nf); err != nil {
		errors = append(errors, data.ValidationError{
			File:     "names.json",
			Field:    "",
			Message:  fmt.Sprintf("JSON parse error: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Check that all species in species.json have names
	for speciesID := range gd.Species {
		speciesNames, ok := nf[speciesID]
		if !ok {
			errors = append(errors, data.ValidationError{
				File:     "names.json",
				Field:    speciesID,
				Message:  fmt.Sprintf("species '%s' from species.json has no name entries", speciesID),
				Severity: "warning",
			})
			continue
		}

		// Check for male names
		if len(speciesNames.Male.First) == 0 {
			errors = append(errors, data.ValidationError{
				File:     "names.json",
				Field:    fmt.Sprintf("%s.male.first", speciesID),
				Message:  "missing male first names",
				Severity: "warning",
			})
		}

		// Check for female names
		if len(speciesNames.Female.First) == 0 {
			errors = append(errors, data.ValidationError{
				File:     "names.json",
				Field:    fmt.Sprintf("%s.female.first", speciesID),
				Message:  "missing female first names",
				Severity: "warning",
			})
		}
	}

	return errors
}

func validateLocationNames(dataDir string) []data.ValidationError {
	var errors []data.ValidationError

	filePath := filepath.Join(dataDir, "location-names.json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, data.ValidationError{
				File:     "location-names.json",
				Field:    "",
				Message:  "file not found",
				Severity: "warning",
			})
			return errors
		}
		errors = append(errors, data.ValidationError{
			File:     "location-names.json",
			Field:    "",
			Message:  fmt.Sprintf("error reading file: %v", err),
			Severity: "error",
		})
		return errors
	}

	var lnf LocationNamesFile
	if err := json.Unmarshal(content, &lnf); err != nil {
		errors = append(errors, data.ValidationError{
			File:     "location-names.json",
			Field:    "",
			Message:  fmt.Sprintf("JSON parse error: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Check that all 4 kingdoms exist
	kingdoms := map[string]map[string]interface{}{
		"valdorine": lnf.Valdorine,
		"karvath":   lnf.Karvath,
		"lumenciel": lnf.Lumenciel,
		"astrene":   lnf.Astrene,
	}

	requiredSections := []string{"cities", "towns", "villages", "regions"}

	for kingdom, kingdomData := range kingdoms {
		if kingdomData == nil {
			errors = append(errors, data.ValidationError{
				File:     "location-names.json",
				Field:    kingdom,
				Message:  fmt.Sprintf("kingdom '%s' is missing", kingdom),
				Severity: "error",
			})
			continue
		}

		// Check that each kingdom has required sections
		for _, section := range requiredSections {
			if _, ok := kingdomData[section]; !ok {
				errors = append(errors, data.ValidationError{
					File:     "location-names.json",
					Field:    fmt.Sprintf("%s.%s", kingdom, section),
					Message:  fmt.Sprintf("missing required section '%s'", section),
					Severity: "warning",
				})
			}
		}
	}

	return errors
}

func validateSpells(dataDir string) []data.ValidationError {
	var errors []data.ValidationError

	filePath := filepath.Join(dataDir, "spells.json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, data.ValidationError{
				File:     "spells.json",
				Field:    "",
				Message:  "file not found",
				Severity: "warning",
			})
			return errors
		}
		errors = append(errors, data.ValidationError{
			File:     "spells.json",
			Field:    "",
			Message:  fmt.Sprintf("error reading file: %v", err),
			Severity: "error",
		})
		return errors
	}

	var sf SpellsFile
	if err := json.Unmarshal(content, &sf); err != nil {
		errors = append(errors, data.ValidationError{
			File:     "spells.json",
			Field:    "",
			Message:  fmt.Sprintf("JSON parse error: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Build spell ID index
	spellIndex := make(map[string]bool)
	for _, spell := range sf.Spells {
		spellIndex[spell.ID] = true
	}

	// Check spell_lists references
	for level, spellIDs := range sf.SpellList.Arcane {
		for _, spellID := range spellIDs {
			if !spellIndex[spellID] {
				errors = append(errors, data.ValidationError{
					File:     "spells.json",
					Field:    fmt.Sprintf("spell_lists.arcane.%s", level),
					Message:  fmt.Sprintf("references non-existent spell '%s'", spellID),
					Severity: "error",
				})
			}
		}
	}

	for level, spellIDs := range sf.SpellList.Divine {
		for _, spellID := range spellIDs {
			if !spellIndex[spellID] {
				errors = append(errors, data.ValidationError{
					File:     "spells.json",
					Field:    fmt.Sprintf("spell_lists.divine.%s", level),
					Message:  fmt.Sprintf("references non-existent spell '%s'", spellID),
					Severity: "error",
				})
			}
		}
	}

	return errors
}

func validateJournalDescriptions(dataDir string) []data.ValidationError {
	var errors []data.ValidationError

	advDir := filepath.Join(dataDir, "adventures")
	entries, err := os.ReadDir(advDir)
	if err != nil {
		// No adventures directory - skip validation
		return errors
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		advName := entry.Name()
		journalPath := filepath.Join(advDir, advName, "journal.json")

		journalData, err := os.ReadFile(journalPath)
		if err != nil {
			// No journal file - skip this adventure
			continue
		}

		var journal struct {
			Entries []struct {
				ID            int    `json:"id"`
				Description   string `json:"description"`
				DescriptionFr string `json:"description_fr"`
			} `json:"entries"`
		}

		if err := json.Unmarshal(journalData, &journal); err != nil {
			errors = append(errors, data.ValidationError{
				File:     fmt.Sprintf("adventures/%s/journal.json", advName),
				Field:    "",
				Message:  fmt.Sprintf("JSON parse error: %v", err),
				Severity: "error",
			})
			continue
		}

		// Validate each entry
		for _, e := range journal.Entries {
			hasEN := e.Description != ""
			hasFR := e.DescriptionFr != ""

			// Check for mismatched descriptions
			if hasEN != hasFR {
				errors = append(errors, data.ValidationError{
					File:     fmt.Sprintf("adventures/%s/journal.json", advName),
					Field:    fmt.Sprintf("entries[%d]", e.ID),
					Message:  "missing translation (one description present, other missing)",
					Severity: "warning",
				})
			}

			// Check English description length
			if hasEN {
				words := len(strings.Fields(e.Description))
				if words < 15 {
					errors = append(errors, data.ValidationError{
						File:     fmt.Sprintf("adventures/%s/journal.json", advName),
						Field:    fmt.Sprintf("entries[%d].description", e.ID),
						Message:  fmt.Sprintf("too short (%d words, recommended 30-50)", words),
						Severity: "warning",
					})
				} else if words > 80 {
					errors = append(errors, data.ValidationError{
						File:     fmt.Sprintf("adventures/%s/journal.json", advName),
						Field:    fmt.Sprintf("entries[%d].description", e.ID),
						Message:  fmt.Sprintf("too long (%d words, recommended 30-50)", words),
						Severity: "warning",
					})
				}
			}

			// Check French description length
			if hasFR {
				words := len(strings.Fields(e.DescriptionFr))
				if words < 15 {
					errors = append(errors, data.ValidationError{
						File:     fmt.Sprintf("adventures/%s/journal.json", advName),
						Field:    fmt.Sprintf("entries[%d].description_fr", e.ID),
						Message:  fmt.Sprintf("too short (%d words, recommended 30-50)", words),
						Severity: "warning",
					})
				} else if words > 80 {
					errors = append(errors, data.ValidationError{
						File:     fmt.Sprintf("adventures/%s/journal.json", advName),
						Field:    fmt.Sprintf("entries[%d].description_fr", e.ID),
						Message:  fmt.Sprintf("too long (%d words, recommended 30-50)", words),
						Severity: "warning",
					})
				}
			}
		}
	}

	return errors
}

func outputJSON(errors []data.ValidationError) {
	result := struct {
		Valid      bool                   `json:"valid"`
		ErrorCount int                    `json:"error_count"`
		WarnCount  int                    `json:"warning_count"`
		Errors     []data.ValidationError `json:"errors"`
	}{
		Valid:  true,
		Errors: errors,
	}

	for _, e := range errors {
		if e.Severity == "error" {
			result.ErrorCount++
			result.Valid = false
		} else {
			result.WarnCount++
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

func outputText(errors []data.ValidationError) {
	if len(errors) == 0 {
		fmt.Println("✓ All game data is valid")
		return
	}

	// Group by file
	byFile := make(map[string][]data.ValidationError)
	for _, e := range errors {
		byFile[e.File] = append(byFile[e.File], e)
	}

	errorCount := 0
	warnCount := 0

	for file, fileErrors := range byFile {
		fmt.Printf("\n%s:\n", file)
		fmt.Println(strings.Repeat("-", len(file)+1))

		for _, e := range fileErrors {
			icon := "⚠"
			if e.Severity == "error" {
				icon = "✗"
				errorCount++
			} else {
				warnCount++
			}

			if e.Field != "" {
				fmt.Printf("  %s [%s] %s: %s\n", icon, e.Severity, e.Field, e.Message)
			} else {
				fmt.Printf("  %s [%s] %s\n", icon, e.Severity, e.Message)
			}
		}
	}

	fmt.Printf("\nSummary: %d error(s), %d warning(s)\n", errorCount, warnCount)

	if errorCount == 0 {
		fmt.Println("✓ Data valid (warnings only)")
	} else {
		fmt.Println("✗ Data validation failed")
	}
}
