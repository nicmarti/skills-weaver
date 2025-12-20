---
name: image-generator
description: Génère des images heroic fantasy pour BFRPG via fal.ai FLUX.1. Portraits de personnages/PNJ, scènes d'aventure, monstres, objets magiques et lieux. Utilise des prompts optimisés pour le style fantasy médiéval.
allowed-tools: Bash
---

# Image Generator - Générateur d'images Heroic Fantasy

Skill pour générer des illustrations fantasy de haute qualité via l'API fal.ai avec le modèle FLUX.1 [schnell].

## Prérequis

**Variable d'environnement requise** :
```bash
export FAL_KEY="votre_clé_api_fal"
```

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-image ./cmd/image

# Portrait de personnage existant
./sw-image character "Aldric"

# Portrait de PNJ généré
./sw-image npc --race=elf --gender=f --occupation=skilled

# Scène d'aventure
./sw-image scene "Des aventuriers explorent une crypte"

# Monstre
./sw-image monster dragon --style=epic
```

## Commandes Disponibles

### Portrait de Personnage

```bash
./sw-image character <nom> [options]

# Exemples:
./sw-image character "Aldric" --style=realistic
./sw-image character "Lyra" --style=painted
```

### Portrait de PNJ

```bash
./sw-image npc [options]

# Options:
#   --race=<race>          Race (human, dwarf, elf, halfling)
#   --gender=<m|f>         Sexe
#   --occupation=<type>    Type d'occupation
#   --style=<style>        Style artistique

# Exemples:
./sw-image npc --race=dwarf --gender=m --occupation=authority
./sw-image npc --race=elf --occupation=religious --style=dark_fantasy
```

### Scène d'Aventure

```bash
./sw-image scene "<description>" [options]

# Options:
#   --type=<type>          Type de scène prédéfini
#   --style=<style>        Style artistique
#   --size=<size>          Taille d'image

# Types de scène:
#   tavern, dungeon, forest, castle, village,
#   cave, battle, treasure, camp, ruins

# Exemples:
./sw-image scene "Combat contre des gobelins" --type=battle --style=epic
./sw-image scene "Repos au coin du feu" --type=camp --style=painted
./sw-image scene "Une taverne animée" --type=tavern
```

### Illustration de Monstre

```bash
./sw-image monster <type> [options]

# Monstres disponibles:
#   goblin, orc, skeleton, zombie, dragon,
#   troll, ogre, wolf, spider, rat, bat, slime,
#   ghost, vampire, werewolf, minotaur, basilisk,
#   chimera, hydra, lich

# Exemples:
./sw-image monster dragon --style=epic
./sw-image monster lich --style=dark_fantasy
./sw-image monster goblin --style=illustrated
```

### Objet Magique

```bash
./sw-image item <type> [description] [options]

# Types d'objets:
#   weapon, armor, potion, scroll, ring,
#   amulet, staff, wand, book, artifact

# Exemples:
./sw-image item weapon "épée flamboyante ancienne"
./sw-image item potion "potion de guérison rouge brillante"
./sw-image item artifact "orbe de pouvoir mystérieux"
```

### Lieu / Carte

```bash
./sw-image location <type> [nom] [options]

# Types de lieux:
#   city, town, village, castle, dungeon,
#   forest, mountain, swamp, desert, coast,
#   island, underworld

# Exemples:
./sw-image location dungeon "Les Mines Abandonnées"
./sw-image location castle "Forteresse de Shadowkeep"
./sw-image location forest "La Forêt des Murmures"
```

### Prompt Personnalisé

```bash
./sw-image custom "<prompt>" [options]

# Pour des besoins spécifiques non couverts par les autres commandes

# Exemples:
./sw-image custom "Un groupe d'aventuriers traversant un pont de corde au-dessus d'un gouffre"
./sw-image custom "Une bibliothèque magique avec des livres volants"
```

## Styles Artistiques

| Style | Description | Utilisation recommandée |
|-------|-------------|------------------------|
| `realistic` | Photoréaliste, détaillé | Portraits immersifs |
| `painted` | Style peinture à l'huile | Scènes, lieux |
| `illustrated` | Illustration digitale | PNJ, personnages (défaut) |
| `dark_fantasy` | Sombre, atmosphérique | Monstres, donjons |
| `epic` | Cinématique, héroïque | Batailles, dragons |

## Tailles d'Image

| Taille | Dimensions | Utilisation |
|--------|------------|-------------|
| `square_hd` | 1024x1024 | Objets, portraits |
| `square` | 512x512 | Vignettes |
| `portrait_4_3` | 768x1024 | Portraits verticaux |
| `portrait_16_9` | 576x1024 | Portraits étroits |
| `landscape_4_3` | 1024x768 | Scènes |
| `landscape_16_9` | 1024x576 | Scènes panoramiques (défaut) |

## Options Communes

```bash
--style=<style>     # Style artistique (realistic, painted, illustrated, dark_fantasy, epic)
--size=<size>       # Taille d'image (square_hd, landscape_16_9, etc.)
--format=<format>   # Format de sortie (png, jpeg, webp)
```

## Exemples d'Utilisation en Session

### Illustrer un personnage créé

```bash
# Créer le personnage
./sw-character create "Thorin" --race=dwarf --class=fighter

# Générer son portrait
./sw-image character "Thorin" --style=epic
```

### Illustrer un PNJ rencontré

```bash
# Générer le PNJ
./sw-npc generate --race=human --occupation=authority --attitude=negative

# Générer son portrait dans la foulée
./sw-image npc --race=human --occupation=authority --style=dark_fantasy
```

### Illustrer une scène de combat

```bash
# Logger le combat
./adventure log "Mon Aventure" combat "Embuscade de gobelins dans la forêt"

# Générer l'illustration
./sw-image scene "Embuscade de gobelins dans une forêt sombre" --type=battle --style=epic
```

## Sortie

Les images sont sauvegardées dans `data/images/` avec un nom unique basé sur le timestamp.

```
data/images/
├── image_1703001234567890123.png
├── image_1703001234567890124.png
└── ...
```

## Coûts

FLUX.1 [schnell] via fal.ai coûte environ **$0.003 par image**, soit ~300 images par dollar.

## Dépannage

### Erreur "FAL_KEY environment variable not set"

```bash
export FAL_KEY="votre_clé_fal_ai"
```

### Erreur API 401

Vérifiez que votre clé API est valide sur [fal.ai/dashboard/keys](https://fal.ai/dashboard/keys).

### Images de mauvaise qualité

- Utilisez un style approprié au sujet
- Ajoutez plus de détails dans les descriptions
- Essayez `--style=realistic` pour plus de détails

## Lister les Options

```bash
./sw-image list              # Toutes les options
./sw-image list styles       # Styles disponibles
./sw-image list scenes       # Types de scènes
./sw-image list monsters     # Types de monstres
./sw-image list items        # Types d'objets
./sw-image list locations    # Types de lieux
./sw-image list sizes        # Tailles d'image
```