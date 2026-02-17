---
name: map-generator
description: Génère des prompts enrichis pour cartes 2D fantasy avec validation world-keeper. Permet de créer des cartes de villes, régions, donjons et combats tactiques cohérents avec l'univers BFRPG.
allowed-tools:
  - Bash
---

# Map Generator - Générateur de Cartes

## Description
Génère des prompts enrichis pour cartes 2D fantasy avec validation world-keeper. Permet de créer des cartes de villes, régions, donjons et combats tactiques cohérents avec l'univers BFRPG.

## Usage
Cette skill permet à Claude de créer des prompts de cartes géographiques cohérentes avec l'univers établi, en validant les noms de lieux et en appliquant les styles architecturaux appropriés pour chaque royaume.

## Commands

```bash
# Générer un prompt de carte
sw-map generate <type> <name> [options]

# Valider un lieu
sw-map validate <name> [options]

# Lister les ressources
sw-map list [category]

# Afficher les types de cartes
sw-map types
```

## Types de Cartes

### City (Carte de Ville)
Vue aérienne détaillée d'une ville avec quartiers, POIs et infrastructure.

**Options**:
- `--scale`: small, medium (défaut), large
- `--features`: POIs additionnels (séparés par virgules)
- `--style`: illustrated (défaut), dark_fantasy

**Exemple**:
```bash
sw-map generate city Cordova --features="Villa de Valorian"
```

### Region (Carte Régionale)
Carte bird's eye view montrant multiple settlements, routes et terrain.

**Options**:
- `--scale`: small, medium, large (défaut)
- `--terrain`: Override du type de terrain

**Exemple**:
```bash
sw-map generate region "Côte Occidentale" --scale=large
```

### Dungeon (Plan de Donjon)
Plan top-down avec salles, couloirs, pièges et grille.

**Options**:
- `--level`: Niveau du donjon (1, 2, 3, etc.)
- `--features`: Salles spéciales (séparées par virgules)

**Exemple**:
```bash
sw-map generate dungeon "La Crypte des Ombres" --level=1 --features="Salle du trône,Crypte"
```

### Tactical (Carte Tactique)
Grille de combat avec terrain, couverture et élévation.

**Options**:
- `--terrain`: forêt, montagne, plaine, marais, etc.
- `--scene`: Description de la scène de combat
- `--features`: Éléments spéciaux (ruisseau, pont, etc.)

**Exemple**:
```bash
sw-map generate tactical "Embuscade" --terrain=forêt --scene="Combat en forêt dense" --features="Ruisseau,Pont"
```

## Options Communes

### --kingdom=<id>
Valide que le lieu appartient au royaume spécifié.

Royaumes valides: valdorine, karvath, lumenciel, astrene

### --style=<style>
Style visuel de la carte.

Styles: illustrated (défaut), dark_fantasy

### --output=<file>
Sauvegarde le prompt dans un fichier JSON spécifique.

### --dry-run
Prévisualise le prompt de base sans appeler l'API Claude.

### --generate-image
Génère aussi l'image via fal.ai flux-2 (nécessite FAL_KEY).

### --image-size=<size>
Taille de l'image générée.

Tailles: square_hd, landscape_16_9 (défaut), portrait_16_9

### --no-cache
Force la régénération du prompt en ignorant le cache.

## Validation de Lieux

```bash
# Valider un lieu existant
sw-map validate Cordova

# Valider avec royaume attendu
sw-map validate "Port-Nouveau" --kingdom=valdorine

# Obtenir des suggestions si le lieu n'existe pas
sw-map validate "Cordov" --suggest
```

## Lister les Ressources

```bash
# Types de cartes disponibles
sw-map list types

# Royaumes disponibles
sw-map list kingdoms

# Tous les lieux documentés
sw-map list locations

# Cités seulement
sw-map list cities

# Cités d'un royaume spécifique
sw-map list cities --kingdom=valdorine
```

## Styles Architecturaux par Royaume

### Valdorine
- Style: Maritime, influences italiennes
- Couleurs: Bleu et or
- Architecture: Ports, toits en tuiles colorées

### Karvath
- Style: Militariste, influences germaniques
- Couleurs: Rouge et noir
- Architecture: Forteresses, murailles épaisses

### Lumenciel
- Style: Religieux, influences latines
- Couleurs: Blanc et or
- Architecture: Cathédrales, monastères

### Astrène
- Style: Mélancolique, influences nordiques
- Couleurs: Gris et argent
- Architecture: Pierre météorisée, simplicité

## Cache et Performance

Les prompts enrichis sont automatiquement mis en cache dans:
```
data/maps/<nom>_<type>_<scale>_prompt.json
```

Le cache réduit significativement les appels API. Utilisez `--no-cache` pour forcer la régénération.

## Génération d'Images

Avec `--generate-image`, la skill génère aussi l'image via fal.ai:

```bash
sw-map generate city Cordova --generate-image --image-size=landscape_16_9
```

Images sauvegardées dans:
```
data/maps/<nom>_<type>_<scale>_<model>.png
```

Modèle utilisé: `fal-ai/flux-2` (haute qualité pour cartes détaillées)

## Prérequis

- **ANTHROPIC_API_KEY**: Requis pour enrichissement AI (Claude Haiku 3.5)
- **FAL_KEY**: Requis pour génération d'images (optionnel)

## Exemples d'Utilisation

### Workflow Typique: Carte de Ville

```bash
# 1. Valider que le lieu existe
sw-map validate Cordova

# 2. Générer le prompt (avec cache)
sw-map generate city Cordova --features="Taverne du Voile Écarlate,Docks"

# 3. Générer l'image
sw-map generate city Cordova --generate-image
```

### Workflow: Plan de Donjon

```bash
# 1. Générer prompt niveau 1
sw-map generate dungeon "La Crypte des Ombres" --level=1 --dry-run

# 2. Générer avec image
sw-map generate dungeon "La Crypte des Ombres" --level=1 --generate-image
```

### Workflow: Carte Tactique

```bash
# 1. Générer avec scène
sw-map generate tactical "Embuscade" \
  --terrain=forêt \
  --scene="Combat contre des bandits en forêt dense" \
  --features="Ruisseau,Pont de bois,Rochers" \
  --generate-image --image-size=square_hd
```

## Intégration avec Agents

### dungeon-master
Le dungeon-master peut invoquer cette skill pour:
- Créer des cartes de lieux visités
- Illustrer des donjons explorés
- Générer des cartes tactiques pour combats importants

### world-keeper
Le world-keeper valide automatiquement:
- Existence des lieux dans geography.json
- Cohérence des noms avec les conventions du royaume
- Styles architecturaux appropriés

## Formats de Sortie

### Prompt JSON
```json
{
  "prompt": "Cette carte montre la ville portuaire de Cordova...",
  "map_type": "city",
  "location_name": "Cordova",
  "kingdom": "valdorine",
  "features": ["Taverne du Voile Écarlate"],
  "style_hints": "aerial view, maritime Italian style, blue/gold colors",
  "enriched_at": "2025-01-15T10:30:00Z"
}
```

### Métadonnées Image
- URL: Lien temporaire fal.ai
- LocalPath: Chemin fichier local
- Dimensions: Largeur x hauteur
- Prompt: Prompt utilisé

## Notes d'Implémentation

- **Validation automatique**: Tous les noms de lieux sont vérifiés contre geography.json
- **Fuzzy matching**: Suggestions basées sur similarité Levenshtein
- **Enrichissement AI**: Claude Haiku 3.5 enrichit les prompts de base avec détails visuels
- **Guidelines**: 400+ lignes de directives pour prompts optimaux
- **Longueur cible**: 100-200 mots (sweet spot: 150)

## Troubleshooting

### Lieu non trouvé
```bash
✗ Lieu "Cordov" non trouvé dans geography.json

Vouliez-vous dire ?
  - Cordova (Valdorine)
  - Port-de-Lune (Valdorine)
```

**Solution**: Utilisez `--suggest` pour voir les suggestions complètes.

### API Key manquante
```
Error: creating enricher: ANTHROPIC_API_KEY environment variable not set
```

**Solution**: Définissez `export ANTHROPIC_API_KEY="votre_clé"`

### Prompt trop court/long
Le système valide automatiquement que les prompts font 80-250 mots et régénère si nécessaire.

## Voir Aussi

- **image-generator**: Génération d'illustrations fantasy
- **name-location-generator**: Génération de noms de lieux cohérents
- **world-keeper**: Agent de cohérence géographique
