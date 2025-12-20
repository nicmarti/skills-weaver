package adventure

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"dungeons/internal/character"
)

// Party represents the group of characters in an adventure.
type Party struct {
	Characters    []string `json:"characters"`     // Character names/slugs
	MarchingOrder []string `json:"marching_order"` // Order for exploration
	Formation     string   `json:"formation"`      // combat, travel, stealth
}

// SharedInventory represents items shared by the party.
type SharedInventory struct {
	Gold  int             `json:"gold"`
	Items []InventoryItem `json:"items"`
}

// InventoryItem represents an item in the shared inventory.
type InventoryItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description,omitempty"`
	AddedAt     string `json:"added_at"`
	AddedBy     string `json:"added_by,omitempty"` // Character who added it
}

// LoadParty loads the party configuration.
func (a *Adventure) LoadParty() (*Party, error) {
	path := filepath.Join(a.basePath, "party.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Party{
			Characters:    []string{},
			MarchingOrder: []string{},
			Formation:     "travel",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading party.json: %w", err)
	}

	var party Party
	if err := json.Unmarshal(data, &party); err != nil {
		return nil, fmt.Errorf("parsing party.json: %w", err)
	}

	return &party, nil
}

// SaveParty saves the party configuration.
func (a *Adventure) SaveParty(party *Party) error {
	path := filepath.Join(a.basePath, "party.json")
	return a.saveJSON(path, party)
}

// AddCharacter adds a character to the adventure.
func (a *Adventure) AddCharacter(charDir, charName string) error {
	// Load character from main characters directory
	charPath := filepath.Join(charDir, slugify(charName)+".json")
	srcChar, err := character.Load(charPath)
	if err != nil {
		return fmt.Errorf("loading character %s: %w", charName, err)
	}

	// Copy character to adventure's characters directory
	destPath := filepath.Join(a.basePath, "characters", slugify(charName)+".json")
	if err := copyFile(charPath, destPath); err != nil {
		return fmt.Errorf("copying character: %w", err)
	}

	// Update party
	party, err := a.LoadParty()
	if err != nil {
		return err
	}

	// Check if already in party
	for _, c := range party.Characters {
		if c == srcChar.Name {
			return fmt.Errorf("character %s is already in the party", charName)
		}
	}

	party.Characters = append(party.Characters, srcChar.Name)
	party.MarchingOrder = append(party.MarchingOrder, srcChar.Name)

	return a.SaveParty(party)
}

// RemoveCharacter removes a character from the adventure.
func (a *Adventure) RemoveCharacter(charName string) error {
	party, err := a.LoadParty()
	if err != nil {
		return err
	}

	// Remove from characters list
	newChars := []string{}
	found := false
	for _, c := range party.Characters {
		if c != charName {
			newChars = append(newChars, c)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("character %s not in party", charName)
	}

	party.Characters = newChars

	// Remove from marching order
	newOrder := []string{}
	for _, c := range party.MarchingOrder {
		if c != charName {
			newOrder = append(newOrder, c)
		}
	}
	party.MarchingOrder = newOrder

	// Remove character file
	charPath := filepath.Join(a.basePath, "characters", slugify(charName)+".json")
	os.Remove(charPath) // Ignore error if file doesn't exist

	return a.SaveParty(party)
}

// GetCharacters loads all characters in the party.
func (a *Adventure) GetCharacters() ([]*character.Character, error) {
	party, err := a.LoadParty()
	if err != nil {
		return nil, err
	}

	var characters []*character.Character
	for _, name := range party.Characters {
		charPath := filepath.Join(a.basePath, "characters", slugify(name)+".json")
		c, err := character.Load(charPath)
		if err != nil {
			continue // Skip missing characters
		}
		characters = append(characters, c)
	}

	return characters, nil
}

// LoadInventory loads the shared inventory.
func (a *Adventure) LoadInventory() (*SharedInventory, error) {
	path := filepath.Join(a.basePath, "inventory.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &SharedInventory{
			Gold:  0,
			Items: []InventoryItem{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading inventory.json: %w", err)
	}

	var inv SharedInventory
	if err := json.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parsing inventory.json: %w", err)
	}

	return &inv, nil
}

// SaveInventory saves the shared inventory.
func (a *Adventure) SaveInventory(inv *SharedInventory) error {
	path := filepath.Join(a.basePath, "inventory.json")
	return a.saveJSON(path, inv)
}

// AddGold adds gold to the shared inventory.
func (a *Adventure) AddGold(amount int, source string) error {
	inv, err := a.LoadInventory()
	if err != nil {
		return err
	}

	inv.Gold += amount

	// Log to journal
	if amount > 0 {
		a.LogEvent("loot", fmt.Sprintf("Le groupe gagne %d po (%s)", amount, source))
	} else {
		a.LogEvent("expense", fmt.Sprintf("Le groupe dépense %d po (%s)", -amount, source))
	}

	return a.SaveInventory(inv)
}

// AddItem adds an item to the shared inventory.
func (a *Adventure) AddItem(name string, quantity int, description, addedBy string) error {
	inv, err := a.LoadInventory()
	if err != nil {
		return err
	}

	// Check if item already exists
	for i, item := range inv.Items {
		if item.Name == name {
			inv.Items[i].Quantity += quantity
			a.LogEvent("loot", fmt.Sprintf("%s ajoute %d× %s à l'inventaire", addedBy, quantity, name))
			return a.SaveInventory(inv)
		}
	}

	// Add new item
	inv.Items = append(inv.Items, InventoryItem{
		ID:          slugify(name),
		Name:        name,
		Quantity:    quantity,
		Description: description,
		AddedAt:     time.Now().Format(time.RFC3339),
		AddedBy:     addedBy,
	})

	a.LogEvent("loot", fmt.Sprintf("%s ajoute %d× %s à l'inventaire", addedBy, quantity, name))

	return a.SaveInventory(inv)
}

// RemoveItem removes items from the shared inventory.
func (a *Adventure) RemoveItem(name string, quantity int) error {
	inv, err := a.LoadInventory()
	if err != nil {
		return err
	}

	for i, item := range inv.Items {
		if item.Name == name || item.ID == slugify(name) {
			if item.Quantity < quantity {
				return fmt.Errorf("not enough %s (have %d, need %d)", name, item.Quantity, quantity)
			}

			inv.Items[i].Quantity -= quantity
			if inv.Items[i].Quantity <= 0 {
				// Remove item entirely
				inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			}

			a.LogEvent("use", fmt.Sprintf("Le groupe utilise/retire %d× %s", quantity, name))
			return a.SaveInventory(inv)
		}
	}

	return fmt.Errorf("item not found: %s", name)
}

// SetFormation sets the party formation.
func (a *Adventure) SetFormation(formation string) error {
	party, err := a.LoadParty()
	if err != nil {
		return err
	}

	party.Formation = formation
	a.LogEvent("party", fmt.Sprintf("Formation changée: %s", formation))

	return a.SaveParty(party)
}

// SetMarchingOrder sets the marching order.
func (a *Adventure) SetMarchingOrder(order []string) error {
	party, err := a.LoadParty()
	if err != nil {
		return err
	}

	party.MarchingOrder = order
	a.LogEvent("party", fmt.Sprintf("Ordre de marche: %v", order))

	return a.SaveParty(party)
}

// Helper function to copy files
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
