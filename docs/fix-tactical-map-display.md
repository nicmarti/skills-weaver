# Fix: Affichage des Cartes Tactiques dans l'Interface Web

**Date**: 2026-02-01
**Contexte**: Les cartes tactiques générées par generate_map n'apparaissaient pas dans l'interface web

## Problème

Lorsque le DM générait une carte tactique via `generate_map`, l'image était créée mais n'apparaissait pas dans l'interface web. Seule une description textuelle était visible :

```
CARTE TACTIQUE - GORGE DU PASSAGE

[IMAGE : Vue zénithale d'une gorge étroite entourée de deux falaises imposantes...]
```

**Symptômes** :
- ✅ L'image était générée (confirmé dans les logs)
- ✅ Le fichier PNG existait sur le disque
- ❌ L'image n'était pas affichée dans l'interface web
- ❌ L'image n'apparaissait pas dans la galerie

## Analyse

### Architecture de Stockage des Images

L'interface web (`internal/web/handlers.go`) charge les images depuis :
```go
adventureImagesDir := filepath.Join("data", "adventures", slug, "images", "session-N")
```

Mais `GenerateMapTool` sauvegardait les images dans :
```go
outputDir := filepath.Join(dataDir, "maps")  // Global, pas par aventure
```

### Comparaison avec generate_image

Le tool `generate_image` fonctionne correctement car il utilise :

```go
func (t *GenerateImageTool) getSessionImagesDir() (string, error) {
    sessionNum := 0
    if session, err := t.adventure.GetCurrentSession(); err == nil && session != nil {
        sessionNum = session.ID
    }
    return filepath.Join(t.adventure.BasePath(), "images", fmt.Sprintf("session-%d", sessionNum)), nil
}
```

**Résultat** : Les images générées par `generate_image` sont visibles, mais pas celles de `generate_map`.

## Solution Implémentée

### 1. Modifier GenerateMapTool pour Utiliser Adventure Instance

**Avant** :
```go
type GenerateMapTool struct {
    dataDir       string
    adventurePath string  // ❌ String path seulement
    enricher      *ai.Enricher
    geography     *world.Geography
    factions      *world.Factions
    notifier      MapGeneratedNotifier
}

func NewGenerateMapTool(dataDir string, adventurePath string, notifier MapGeneratedNotifier) {...}
```

**Après** :
```go
type GenerateMapTool struct {
    dataDir   string
    adventure *adventure.Adventure  // ✅ Adventure instance complète
    enricher  *ai.Enricher
    geography *world.Geography
    factions  *world.Factions
    notifier  MapGeneratedNotifier
}

func NewGenerateMapTool(dataDir string, adv *adventure.Adventure, notifier MapGeneratedNotifier) {...}
```

### 2. Ajouter getSessionImagesDir() à GenerateMapTool

```go
// getSessionImagesDir returns the images directory for the current session.
// Uses the active session number, or session-0 if no session is active.
func (t *GenerateMapTool) getSessionImagesDir() (string, error) {
    sessionNum := 0
    if session, err := t.adventure.GetCurrentSession(); err == nil && session != nil {
        sessionNum = session.ID
    }

    imagesDir := filepath.Join(t.adventure.BasePath(), "images", fmt.Sprintf("session-%d", sessionNum))
    if err := os.MkdirAll(imagesDir, 0755); err != nil {
        return "", fmt.Errorf("creating images directory: %w", err)
    }

    return imagesDir, nil
}
```

**Caractéristiques** :
- Récupère automatiquement le numéro de session active
- Crée le répertoire `data/adventures/<slug>/images/session-N/`
- Fallback à `session-0` si aucune session n'est active

### 3. Modifier generateImage() pour Utiliser Session Directory

**Avant** :
```go
func (t *GenerateMapTool) generateImage(prompt, name, mapType, scale string) (string, string, error) {
    // Create image generator
    outputDir := filepath.Join(t.dataDir, "maps")  // ❌ Global directory
    gen, err := image.NewGenerator(outputDir)
    ...
}
```

**Après** :
```go
func (t *GenerateMapTool) generateImage(prompt, name, mapType, scale string) (string, string, error) {
    // Get the images directory for the current session
    outputDir, err := t.getSessionImagesDir()  // ✅ Session-specific directory
    if err != nil {
        return "", "", fmt.Errorf("getting session images dir: %w", err)
    }

    // Create image generator
    gen, err := image.NewGenerator(outputDir)
    ...
}
```

### 4. Mettre à Jour register_tools.go

**Avant** :
```go
mapTool, err := dmtools.NewGenerateMapTool(dataDir, adv.BasePath(), mapNotifier)
```

**Après** :
```go
mapTool, err := dmtools.NewGenerateMapTool(dataDir, adv, mapNotifier)
```

### 5. Corriger les Tests (map_tool_test.go)

Les tests utilisaient un simple string path. Ils doivent maintenant créer une instance d'Adventure :

```go
// Avant
tempAdventure := filepath.Join(dataDir, "adventures", "test-map-validation")
tool, err := NewGenerateMapTool(dataDir, tempAdventure, nil)

// Après
adventuresDir := filepath.Join(dataDir, "adventures")
advData := adventure.New("Test Map Validation", "Test adventure for map generation")
advData.Save(adventuresDir)
tempAdventurePath := filepath.Join(adventuresDir, advData.Slug)
tempAdventure, err := adventure.Load(tempAdventurePath)
tool, err := NewGenerateMapTool(dataDir, tempAdventure, nil)
```

## Fichiers Modifiés

```
M  internal/dmtools/map_tool.go
   - Ligne 10: Ajout import "dungeons/internal/adventure"
   - Ligne 24: adventure *adventure.Adventure (au lieu de adventurePath string)
   - Ligne 32: NewGenerateMapTool prend *adventure.Adventure
   - Ligne 61-73: Nouvelle méthode getSessionImagesDir()
   - Ligne 322-329: generateImage() utilise getSessionImagesDir()

M  internal/agent/register_tools.go
   - Ligne 97: Passe adv au lieu de adv.BasePath()

M  internal/dmtools/map_tool_test.go
   - Ligne 3-7: Ajout import adventure
   - Ligne 13-33: Création Adventure instance pour test 1
   - Ligne 131-151: Création Adventure instance pour test 2
```

## Résultat

### Avant

```
Génération carte tactique:
data/maps/gorge-du-passage---embuscade_tactical_medium_flux-pro-11.png
                    ↓
    ❌ Interface web ne trouve pas l'image
    ❌ Galerie vide
```

### Après

```
Génération carte tactique (Session 5):
data/adventures/les-naufrages-du-pierre-lune/images/session-5/gorge-du-passage---embuscade_tactical_medium_flux-pro-11.png
                    ↓
    ✅ Interface web trouve l'image
    ✅ Galerie affiche l'image
    ✅ Carte visible inline dans le chat
```

## Tests

Tous les tests passent après les modifications :

```bash
go test ./internal/dmtools/... -v -run TestMapGeneration
✅ TestMapGenerationValidation (16.27s)
   ✅ Region_map_with_adventure-specific_location
   ✅ Region_map_with_any_name
   ✅ Dungeon_map_with_any_name
   ✅ Tactical_map_with_any_name
   ✅ City_map_with_invalid_location
✅ TestMapGenerationHintMessage (0.00s)

go build -o sw-dm ./cmd/dm
✅ SUCCESS
```

## Impact

### Avantages

1. **Cohérence** : generate_map et generate_image utilisent la même logique
2. **Organisation** : Images par session pour faciliter la navigation
3. **Visibilité** : Les cartes tactiques apparaissent maintenant dans l'interface web
4. **Galerie** : Toutes les images de session sont automatiquement disponibles

### Pas de Breaking Changes

- Les anciennes images dans `data/maps/` restent accessibles manuellement
- Les nouvelles générations utilisent le répertoire par session
- Les prompts en cache continuent d'utiliser `data/maps/` (intentionnel)

## Migration des Images Existantes (Optionnel)

Si des aventures ont des images dans `data/maps/` qui doivent être migrées :

```bash
# Trouver les images d'aventure
find data/maps/ -name "*gorge-du-passage*.png"

# Les copier dans la session appropriée
cp data/maps/gorge-du-passage---embuscade_tactical_medium_flux-pro-11.png \
   data/adventures/les-naufrages-du-pierre-lune/images/session-5/
```

**Note** : Cette migration n'est pas nécessaire si les images n'ont jamais été affichées. Le problème est corrigé pour toutes les futures générations.

## Logs d'Exemple (Session 5)

**Avant fix** :
```
[2026-02-01 00:16:02] TOOL RESULT: generate_map
  image_path: data/maps/gorge-du-passage---embuscade_tactical_medium_flux-pro-11.png
```

**Après fix** :
```
[2026-02-01 XX:XX:XX] TOOL RESULT: generate_map
  image_path: data/adventures/les-naufrages-du-pierre-lune/images/session-5/gorge-du-passage---embuscade_tactical_medium_flux-pro-11.png
```

## Conclusion

Le fix aligne `generate_map` avec `generate_image`, garantissant que toutes les images générées pendant une session sont visibles dans l'interface web. Cela améliore significativement l'expérience utilisateur lors des combats tactiques.
