# Fix: Assouplissement de la Validation des Cartes Régionales

**Date**: 2026-01-31
**Ticket**: User feedback - Session 5 log analysis
**Fichiers modifiés**: `internal/dmtools/map_tool.go`

## Problème Identifié

Lors de la session 5 de l'aventure "Les naufragés du Pierre Lune", le DM a tenté de générer une carte régionale pour "Route entre Greystone et Portus Lunaris" (lignes 2638-2749 du log).

**Tentatives échouées** :
1. `"Lumarios - Côte entre Greystone et Portus Lunaris"` ❌
2. `"Région côtière de Lumarios"` ❌
3. `"Greystone"` ❌
4. `"route côtière"` ❌

**Erreur systématique** :
```
"error": "Location 'X' not found in geography.json"
```

**Conséquence** : Le DM a dû contourner le problème en utilisant `generate_image` directement (ligne 2749) au lieu du tool `generate_map` dédié.

## Cause Racine

Le code validait **strictement** les cartes `city` ET `region` contre `data/world/geography.json` :

```go
// AVANT (ligne 199)
if mapType == "city" || mapType == "region" {
    exists, loc, _, _ := world.ValidateLocationExists(name, t.geography)
    if !exists {
        return error // Bloque la génération
    }
}
```

**Problème** : Les lieux spécifiques aux aventures (ex: Greystone, Portus Lunaris) existent dans `campaign-plan.json` mais PAS dans `geography.json` global.

## Solution Implémentée (Option 1)

**Retirer les `region` maps de la validation stricte** :

```go
// APRÈS (ligne 199)
if mapType == "city" {  // Seulement city maintenant
    exists, loc, _, _ := world.ValidateLocationExists(name, t.geography)
    if !exists {
        return error
    }
}
// Region, dungeon, tactical: aucune validation requise
```

### Changements Détaillés

#### 1. Validation (ligne 199)
- **Avant** : `if mapType == "city" || mapType == "region"`
- **Après** : `if mapType == "city"`

#### 2. Commentaire (ligne 195)
- **Avant** : `// Validate and get location data (for city/region types)`
- **Après** : `// Validate and get location data (for city type only)`

#### 3. Hint message (ligne 213)
- **Avant** : `"hint": "For dungeons and tactical maps, location validation is not required."`
- **Après** : `"hint": "For region, dungeon and tactical maps, location validation is not required."`

#### 4. Description du tool (ligne 67)
- **Avant** : "Validates locations against world-keeper data and applies kingdom-specific architectural styles."
- **Après** : "City maps are validated against world-keeper data for architectural consistency. Region, dungeon, and tactical maps can use any location name."

#### 5. MAP TYPES documentation (lignes 76-79)
```diff
- city: Aerial view of a city with districts, POIs, and infrastructure
+ city: Aerial view of a city with districts, POIs, and infrastructure (requires location in geography.json)
- region: Bird's eye view of multiple settlements, routes, and terrain
+ region: Bird's eye view of multiple settlements, routes, and terrain (no validation required)
- dungeon: Top-down floor plan with rooms, corridors, traps, and grid
+ dungeon: Top-down floor plan with rooms, corridors, traps, and grid (no validation required)
- tactical: Combat grid with terrain, cover, obstacles, and elevation
+ tactical: Combat grid with terrain, cover, obstacles, and elevation (no validation required)
```

#### 6. Paramètre `name` description (ligne 96)
- **Avant** : "For city/region: must exist in geography.json"
- **Après** : "For city: must exist in geography.json. For region/dungeon/tactical: any descriptive name (e.g., 'Route entre Greystone et Portus Lunaris')"

## Comportement Après le Fix

| Type de Carte | Validation | Exemple de Nom Accepté |
|--------------|------------|------------------------|
| **city** | ✅ Stricte (geography.json) | `"Cordova"`, `"Port-Royal"` |
| **region** | ❌ Aucune | `"Route entre Greystone et Portus Lunaris"`, `"Lumarios - Côte nord"` |
| **dungeon** | ❌ Aucune | `"La Crypte des Ombres"`, `"Temple souterrain"` |
| **tactical** | ❌ Aucune | `"Embuscade en forêt"`, `"Combat dans la carrière"` |

## Justification de la Solution

### Pourquoi valider uniquement les city maps ?

1. **City maps** :
   - Représentent des lieux fixes du monde
   - Nécessitent cohérence architecturale (styles Valdorine, Karvath, etc.)
   - Bénéficient des données kingdom (factions, styles, descriptions)
   - Peu nombreuses et bien documentées dans geography.json

2. **Region maps** :
   - Représentent souvent des zones entre lieux (routes, forêts, côtes)
   - Ces zones intermédiaires sont rarement dans geography.json
   - Souvent créées dynamiquement pendant les aventures
   - Validation stricte bloque la créativité du DM

3. **Dungeon et tactical maps** :
   - Déjà sans validation (comportement existant)
   - Purement situationnels et temporaires

## Test de Régression

Pour vérifier que le fix fonctionne :

```bash
# Compiler
go build -o sw-dm ./cmd/dm

# Test dans une session sw-dm
# Commande qui échouait avant :
generate_map(map_type="region", name="Route entre Greystone et Portus Lunaris", features=["Auberge", "Pont"], generate_image=true)

# Résultat attendu : ✅ SUCCESS (génération du prompt et de l'image)
```

## Impact

### ✅ Avantages
- Les DM peuvent générer des region maps pour n'importe quelle zone
- Résout le cas d'usage "Route entre Greystone et Portus Lunaris"
- Pas besoin de contourner avec `generate_image`
- Cohérent avec le hint existant ("dungeons and tactical maps don't need validation")

### ⚠️ Limitations
- Les region maps perdent les style hints liés aux kingdoms
- Pas de suggestions automatiques de lieux similaires
- Le DM doit manuellement assurer la cohérence géographique

### 🔒 Maintien de la Qualité
- Les **city maps** gardent la validation stricte
- La cohérence architecturale des villes est préservée
- Les styles par royaume (valdorine, karvath, etc.) restent appliqués aux cities

## Alternatives Considérées mais Non Retenues

### Option 2 : Flag `skip_validation`
```go
"skip_validation": {
    "type": "boolean",
    "description": "Skip location validation for adventure-specific locations"
}
```
❌ **Rejetée** : Ajoute complexité inutile (un paramètre de plus à gérer)

### Option 3 : Warning au lieu d'erreur
```go
if !exists {
    response["warning"] = "Location not found in geography.json"
    // Continue anyway
}
```
❌ **Rejetée** : Perd les suggestions de lieux similaires, moins clair pour le DM

## Commit

```bash
git add internal/dmtools/map_tool.go docs/map-generation-validation-fix.md
git commit -m "fix: relax validation for region maps to allow adventure-specific locations"
```

## Monitoring

- Surveiller les logs pour voir si les DM utilisent plus de region maps
- Vérifier que les city maps continuent d'avoir une bonne cohérence
- Potentiellement ajouter télémétrie sur les types de maps générées
