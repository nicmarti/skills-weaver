package adventure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dungeons/internal/character" // Import character package
)

func TestLoadPartyNonExistent(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Non-existent party.json should return empty party
	party, err := adv.LoadParty()
	if err != nil {
		t.Fatalf("LoadParty() error = %v, want nil", err)
	}
	if party == nil {
		t.Errorf("LoadParty() returned nil, want empty party")
	}
	if len(party.Characters) != 0 {
		t.Errorf("LoadParty() characters = %v, want []", party.Characters)
	}
	if party.Formation != "travel" {
		t.Errorf("LoadParty() formation = %q, want 'travel'", party.Formation)
	}
}

func TestSaveAndLoadParty(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	original := &Party{
		Characters:    []string{"Aldric", "Lyra", "Thorin"},
		MarchingOrder: []string{"Aldric", "Lyra", "Thorin"},
		Formation:     "combat",
	}

	err := adv.SaveParty(original)
	if err != nil {
		t.Fatalf("SaveParty() error = %v", err)
	}

	loaded, err := adv.LoadParty()
	if err != nil {
		t.Fatalf("LoadParty() error = %v", err)
	}

	if !equal(loaded.Characters, original.Characters) {
		t.Errorf("Characters mismatch: got %v, want %v", loaded.Characters, original.Characters)
	}
	if !equal(loaded.MarchingOrder, original.MarchingOrder) {
		t.Errorf("MarchingOrder mismatch: got %v, want %v", loaded.MarchingOrder, original.MarchingOrder)
	}
	if loaded.Formation != original.Formation {
		t.Errorf("Formation mismatch: got %q, want %q", loaded.Formation, original.Formation)
	}
}

func TestAddCharacterBasic(t *testing.T) {
	// Test the party structure itself without file I/O
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	// Manually create party with characters
	party := &Party{
		Characters:    []string{"Aldric"},
		MarchingOrder: []string{"Aldric"},
		Formation:     "travel",
	}

	err := adv.SaveParty(party)
	if err != nil {
		t.Fatalf("SaveParty() error = %v", err)
	}

	// Load and verify
	loaded, _ := adv.LoadParty()
	if len(loaded.Characters) != 1 || loaded.Characters[0] != "Aldric" {
		t.Errorf("Party save/load failed")
	}
}

func TestAddCharacterDuplicate(t *testing.T) {
	baseDir := t.TempDir()
	charDir := t.TempDir()

	testChar := character.New("Aldric", "human", "fighter")
	testChar.Save(filepath.Join(charDir, "aldric.json"))

	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)
	os.MkdirAll(filepath.Join(baseDir, "characters"), 0755)

	adv.AddCharacter(charDir, "Aldric")

	// Try adding again
	err := adv.AddCharacter(charDir, "Aldric")
	if err == nil {
		t.Errorf("AddCharacter() should fail for duplicate, got nil error")
	}
}

func TestRemoveCharacter(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	party := &Party{
		Characters:    []string{"Aldric", "Lyra", "Thorin"},
		MarchingOrder: []string{"Aldric", "Lyra", "Thorin"},
		Formation:     "travel",
	}
	adv.SaveParty(party)

	err := adv.RemoveCharacter("Lyra")
	if err != nil {
		t.Fatalf("RemoveCharacter() error = %v", err)
	}

	loaded, _ := adv.LoadParty()
	if len(loaded.Characters) != 2 {
		t.Errorf("RemoveCharacter() resulted in %d characters, want 2", len(loaded.Characters))
	}

	for _, name := range loaded.Characters {
		if name == "Lyra" {
			t.Errorf("RemoveCharacter() didn't remove Lyra")
		}
	}
}

func TestRemoveCharacterNotFound(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.RemoveCharacter("NonExistent")
	if err == nil {
		t.Errorf("RemoveCharacter() should fail for non-existent character")
	}
}

func TestLoadInventoryNonExistent(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	inv, err := adv.LoadInventory()
	if err != nil {
		t.Fatalf("LoadInventory() error = %v, want nil", err)
	}
	if inv.Gold != 0 {
		t.Errorf("LoadInventory() gold = %d, want 0", inv.Gold)
	}
	if len(inv.Items) != 0 {
		t.Errorf("LoadInventory() items = %v, want []", inv.Items)
	}
}

func TestAddGold(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.AddGold(100, "Treasure")
	if err != nil {
		t.Fatalf("AddGold() error = %v", err)
	}

	inv, _ := adv.LoadInventory()
	if inv.Gold != 100 {
		t.Errorf("AddGold(100) resulted in %d gold, want 100", inv.Gold)
	}

	// Add more gold
	adv.AddGold(50, "More Treasure")
	inv, _ = adv.LoadInventory()
	if inv.Gold != 150 {
		t.Errorf("Adding 50 to 100 resulted in %d, want 150", inv.Gold)
	}
}

func TestAddItem(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.AddItem("Longsword", 1, "A magical blade", "Aldric")
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	inv, _ := adv.LoadInventory()
	if len(inv.Items) != 1 {
		t.Errorf("AddItem() resulted in %d items, want 1", len(inv.Items))
	}

	item := inv.Items[0]
	if item.Name != "Longsword" {
		t.Errorf("Item name = %q, want 'Longsword'", item.Name)
	}
	if item.Quantity != 1 {
		t.Errorf("Item quantity = %d, want 1", item.Quantity)
	}
	if item.AddedBy != "Aldric" {
		t.Errorf("Item added_by = %q, want 'Aldric'", item.AddedBy)
	}
}

func TestAddItemStackable(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.AddItem("Potion", 2, "Healing potion", "Lyra")
	adv.AddItem("Potion", 3, "Healing potion", "Lyra")

	inv, _ := adv.LoadInventory()
	if len(inv.Items) != 1 {
		t.Errorf("Stackable items resulted in %d items, want 1", len(inv.Items))
	}

	if inv.Items[0].Quantity != 5 {
		t.Errorf("Item quantity = %d, want 5", inv.Items[0].Quantity)
	}
}

func TestRemoveItem(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.AddItem("Rope", 5, "50ft rope", "Thorin")

	err := adv.RemoveItem("Rope", 2)
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}

	inv, _ := adv.LoadInventory()
	if inv.Items[0].Quantity != 3 {
		t.Errorf("RemoveItem(2) from 5 resulted in %d, want 3", inv.Items[0].Quantity)
	}
}

func TestRemoveItemCompletely(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.AddItem("Key", 1, "Mysterious key", "Aldric")
	adv.RemoveItem("Key", 1)

	inv, _ := adv.LoadInventory()
	if len(inv.Items) != 0 {
		t.Errorf("RemoveItem(all) still has items: %v", inv.Items)
	}
}

func TestRemoveItemNotFound(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	err := adv.RemoveItem("NonExistent", 1)
	if err == nil {
		t.Errorf("RemoveItem() should fail for non-existent item")
	}
}

func TestRemoveItemNotEnough(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	adv.AddItem("Arrow", 5, "Arrows", "Lyra")

	err := adv.RemoveItem("Arrow", 10)
	if err == nil {
		t.Errorf("RemoveItem(10) from 5 should fail")
	}
}

func TestSetFormation(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	formations := []string{"combat", "travel", "stealth"}

	for _, f := range formations {
		err := adv.SetFormation(f)
		if err != nil {
			t.Fatalf("SetFormation(%q) error = %v", f, err)
		}

		party, _ := adv.LoadParty()
		if party.Formation != f {
			t.Errorf("SetFormation(%q) resulted in %q", f, party.Formation)
		}
	}
}

func TestSetMarchingOrder(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test Adventure", "Test")
	adv.SetBasePath(baseDir)

	order := []string{"Thorin", "Aldric", "Lyra"}
	err := adv.SetMarchingOrder(order)
	if err != nil {
		t.Fatalf("SetMarchingOrder() error = %v", err)
	}

	party, _ := adv.LoadParty()
	if !equal(party.MarchingOrder, order) {
		t.Errorf("MarchingOrder mismatch: got %v, want %v", party.MarchingOrder, order)
	}
}

func TestCopyFile(t *testing.T) {
	baseDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(baseDir, "source.txt")
	srcContent := []byte("test content")
	os.WriteFile(srcPath, srcContent, 0644)

	// Copy file
	dstPath := filepath.Join(baseDir, "destination.txt")
	err := copyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify destination exists and has same content
	dstContent, _ := os.ReadFile(dstPath)
	if string(dstContent) != string(srcContent) {
		t.Errorf("Copied file content mismatch")
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	baseDir := t.TempDir()

	err := copyFile(
		filepath.Join(baseDir, "nonexistent.txt"),
		filepath.Join(baseDir, "dest.txt"),
	)
	if err == nil {
		t.Errorf("copyFile() should fail for non-existent source")
	}
}

// Helper function to compare string slices
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper to verify JSON marshaling of party
func TestPartyJSON(t *testing.T) {
	party := &Party{
		Characters:    []string{"A", "B"},
		MarchingOrder: []string{"B", "A"},
		Formation:     "combat",
	}

	data, err := json.Marshal(party)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}

	var loaded Party
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if loaded.Formation != party.Formation {
		t.Errorf("JSON round-trip failed for formation")
	}
}

// Helper to verify inventory item has valid ID
func TestInventoryItemID(t *testing.T) {
	baseDir := t.TempDir()
	adv := New("Test", "Test")
	adv.SetBasePath(baseDir)

	adv.AddItem("Dragon's Tooth", 1, "Rare item", "Aldric")

	inv, _ := adv.LoadInventory()
	item := inv.Items[0]

	if item.ID == "" {
		t.Errorf("InventoryItem.ID is empty")
	}
	if item.ID != slugify("Dragon's Tooth") {
		t.Errorf("InventoryItem.ID = %q, want %q", item.ID, slugify("Dragon's Tooth"))
	}
}
