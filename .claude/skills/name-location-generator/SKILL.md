---
name: name-location-generator
description: Génère des noms de lieux (cités, villes, villages, régions) cohérents avec les 4 factions. Utilise des styles distincts par royaume (valdorine maritime, karvath militaire, lumenciel religieux, astrène mélancolique). Intégré avec world-keeper pour validation.
allowed-tools: Bash
---

# Name Location Generator - Générateur de Noms de Lieux

Skill pour générer des noms de lieux cohérents avec l'univers de Basic Fantasy RPG et les 4 royaumes.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-location-names ./cmd/location-names

# Générer un nom
./sw-location-names city --kingdom=valdorine
./sw-location-names village --kingdom=karvath
./sw-location-names region --kingdom=lumenciel
```

## Commandes Disponibles

### Génération par Type

```bash
# Types disponibles: city, town, village, region
./sw-location-names <type> --kingdom=<royaume> [--count=N]

# Options:
#   --kingdom=<royaume>  Royaume (valdorine, karvath, lumenciel, astrene)
#   --count=N            Nombre de noms à générer
```

### Exemples par Royaume

#### Valdorine (Maritime, Marchand)

```bash
# Cité portuaire
./sw-location-names city --kingdom=valdorine
# Exemple: Cordova, Havremarpoint, Navrenaaven

# Village côtier
./sw-location-names village --kingdom=valdorine
# Exemple: Les Mouettes, La Marées, Les Voiles

# Région maritime
./sw-location-names region --kingdom=valdorine
# Exemple: Côte Occidentale, Golfe des Marchands, Îles Dorées
```

#### Karvath (Militariste, Défensif)

```bash
# Forteresse
./sw-location-names city --kingdom=karvath
# Exemple: Fer-de-Lance, Rocmurburg, Fortemarteauheim

# Bourg militaire
./sw-location-names town --kingdom=karvath
# Exemple: Hautgarde, Valbourg, Rocstein

# Région montagnarde
./sw-location-names region --kingdom=karvath
# Exemple: Montagnes de Fer, Plaines du Bouclier, Défilé de l'Aigle
```

#### Lumenciel (Théocratique, Hypocrite)

```bash
# Cité religieuse
./sw-location-names city --kingdom=lumenciel
# Exemple: Aurore-Sainte, Lumenciel, Saint-Aethel

# Village pieux
./sw-location-names village --kingdom=lumenciel
# Exemple: Saint-Lumière, Bonne-Grâce, Sainte-Foi

# Région sacrée
./sw-location-names region --kingdom=lumenciel
# Exemple: Terres Saintes, Forêt de la Grâce, Val de Lumière
```

#### Astrène (Décadent, Érudit)

```bash
# Cité impériale
./sw-location-names city --kingdom=astrene
# Exemple: Étoile-d'Automne, Lune-Crépusculaire, Albastra

# Village ancien
./sw-location-names village --kingdom=astrene
# Exemple: Vieux-brume, Ancien-ombre, Petit-oubli

# Région mélancolique
./sw-location-names region --kingdom=astrene
# Exemple: Terres du Sud, Val de l'Oubli, Plaines Fanées
```

### Génération Multiple

```bash
# Générer plusieurs noms pour une liste de choix
./sw-location-names city --kingdom=valdorine --count=5
./sw-location-names village --kingdom=karvath --count=10
```

### Lister les Options

```bash
./sw-location-names list              # Tout lister
./sw-location-names list kingdoms     # Royaumes disponibles
./sw-location-names list types        # Types disponibles
```

## Styles de Noms par Faction

### Valdorine 🌊
**Style**: Maritime, cosmopolite, commercial

- **Cités**: Cor-, Port-, Havre-, Mar-, Nav- + racine maritime + -ia, -aven, -bay
- **Villages**: Le/La/Les + nom maritime (Mouettes, Marées, Voiles)
- **Régions**: Côte/Golfe/Îles + adjectif descriptif

**Exemples**: Cordova, Port-de-Lune, Havre-d'Argent, Les Sardines

### Karvath ⚔️
**Style**: Fort, martial, défensif

- **Cités**: Fer-, Roc-, Garde-, Forte- + arme/défense + -garde, -fort, -heim, -burg
- **Villages**: Préfixe + suffixe germanique (-bourg, -stein, -wald)
- **Régions**: Montagnes/Plaines/Défilé + nom martial

**Exemples**: Fer-de-Lance, Porte-de-Fer, Hautgarde, Montagnes de Fer

### Lumenciel ☀️
**Style**: Lumineux, pieux, céleste

- **Cités**: Aurore-, Lumière-, Saint-, Céleste- + racine religieuse + -sainte, -bénie
- **Villages**: Saint/Sainte/Bon/Bonne + vertu religieuse
- **Régions**: Terres/Forêt/Val + adjectif spirituel

**Exemples**: Aurore-Sainte, Saint-Aethel, Vallon-de-Prière, Terres Saintes

### Astrène 🌙
**Style**: Noble ancien, mélancolique, érudit

- **Cités**: Étoile-, Lune-, Astro-, Nyx- + racine temporelle + -Ancienne, -Impériale
- **Villages**: Vieux/Ancien/Petit + sentiment mélancolique
- **Régions**: Terres/Val/Plaines + adjectif nostalgique

**Exemples**: Étoile-d'Automne, Valombre, Brume-Ancienne, Terres du Sud

## Intégration avec World-Keeper

Le world-keeper peut utiliser ce skill pour créer des lieux cohérents :

```bash
# Workflow world-keeper
1. Générer un nom via sw-location-names
2. Vérifier unicité dans geography.json
3. Si existe, régénérer
4. Documenter dans geography.json
5. Retourner au DM
```

## Intégration avec Dungeon Master

Le dungeon-master peut appeler ce skill pour improvisation rapide :

```bash
# Exemple en session
DM: "Les PJ veulent aller dans une ville valdine non encore nommée"
DM: /name-location-generator city valdorine
Output: Marvelia
DM: [Utilise ce nom dans narration]
```

## Correspondances Français-Anglais

| Français | Anglais | Commande |
|----------|---------|----------|
| Cité | City | `city` |
| Bourg | Town | `town` |
| Village | Village | `village` |
| Région | Region | `region` |
| Valdorine | Valdorine | `valdorine` |
| Karvath | Karvath | `karvath` |
| Lumenciel | Lumenciel | `lumenciel` |
| Astrène | Astrene | `astrene` |

## Structure des Données

Les noms sont stockés dans `data/location-names.json` avec :

- **Prefixes**: ~9 préfixes par type par faction
- **Roots**: ~9 racines pour cités
- **Suffixes**: ~8 suffixes par type par faction
- **Templates**: Modèles de régions avec placeholders

Combinaisons possibles :
- **Cités**: 9 × 9 × 8 = **648 noms uniques** par faction
- **Villages**: 3 × 10 = **30 noms uniques** (Valdorine)
- **Régions**: 5 templates × 7 adjectifs = **35 noms uniques** par faction

## Cohérence Assurée

### Par le World-Keeper

✅ **Unicité**: Vérifie que le nom n'existe pas déjà
✅ **Style**: Respect du style de faction
✅ **Géographie**: Port sur côte, forteresse en montagne
✅ **Documentation**: Ajout automatique dans `geography.json`

### Exemples de Cohérence

- ❌ "Port-de-Fer" pour Valdorine (style Karvath)
- ✅ "Port-de-Lune" pour Valdorine (style maritime)
- ❌ "Vallon-de-Prière" dans Karvath (style Lumenciel)
- ✅ "Hautegarde" dans Karvath (style militaire)

## Conseils d'Utilisation

- Pour une **cité majeure** : `./sw-location-names city --kingdom=<faction>`
- Pour un **village mineur** : `./sw-location-names village --kingdom=<faction>`
- Pour une **région géographique** : `./sw-location-names region --kingdom=<faction>`
- Pour une **liste de choix** : `--count=5`

**Note** : Pour les ruines, Terres Brûlées et autres lieux spéciaux sans faction, laissez le dungeon-master créer des noms contextuels qui s'intègrent mieux à l'histoire.

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Improvisation rapide de noms de lieux |
| `world-keeper` | Création et documentation de lieux cohérents |

**Type** : Skill autonome, peut être invoqué directement via `/name-location-generator`