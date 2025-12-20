---
name: journal-illustrator
description: Illustre automatiquement le journal d'une aventure BFRPG en générant des images pour les moments clés (combats, explorations, découvertes). Utilise la génération parallèle pour une performance optimale.
allowed-tools: Bash
---

# Journal Illustrator - Illustration automatique des aventures

Skill pour générer des illustrations basées sur le journal d'une aventure BFRPG. Analyse les entrées du journal et génère des images appropriées pour chaque type d'événement.

## Prérequis

**Variable d'environnement requise** :
```bash
export FAL_KEY="votre_clé_api_fal"
```

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-image ./cmd/image

# Voir les prompts qui seraient générés (sans générer d'images)
./sw-image journal "la-crypte-des-ombres" --dry-run

# Générer toutes les illustrations
./sw-image journal "la-crypte-des-ombres"

# Limiter le nombre d'images
./sw-image journal "la-crypte-des-ombres" --max=5

# Filtrer par type d'événement
./sw-image journal "la-crypte-des-ombres" --types=combat,discovery
```

## Types d'Événements Illustrables

| Type | Description | Style | Taille |
|------|-------------|-------|--------|
| `combat` | Scènes de combat et batailles | Epic | landscape_16_9 |
| `exploration` | Exploration de lieux | Painted | landscape_16_9 |
| `discovery` | Découvertes et révélations | Dark Fantasy | landscape_4_3 |
| `loot` | Trésors trouvés | Painted | square_hd |
| `session` | Résumés de fin de session | Epic | landscape_16_9 |

## Options

```bash
--types=<types>     # Types à illustrer (combat,exploration,discovery,loot,session)
--max=<n>           # Nombre maximum d'images à générer
--parallel=<n>      # Niveau de parallélisme (1-8, défaut: 4)
--dry-run           # Afficher les prompts sans générer d'images
```

## Exemples d'Utilisation

### Prévisualiser les illustrations

```bash
./sw-image journal "la-crypte-des-ombres" --dry-run
```

Affiche tous les prompts qui seraient utilisés, avec le style et la taille de chaque image.

### Illustrer uniquement les combats

```bash
./sw-image journal "la-crypte-des-ombres" --types=combat
```

### Générer les 10 premières illustrations

```bash
./sw-image journal "la-crypte-des-ombres" --max=10
```

### Génération plus rapide (8 images en parallèle)

```bash
./sw-image journal "la-crypte-des-ombres" --parallel=8
```

## Sortie

Les images sont sauvegardées dans le répertoire de l'aventure :

```
data/adventures/la-crypte-des-ombres/images/
├── image_1703001234567890123.png  # Combat contre les squelettes
├── image_1703001234567890124.png  # Exploration du mausolée
├── image_1703001234567890125.png  # Découverte du parchemin
└── ...
```

## Adaptation des Prompts

Le générateur adapte automatiquement les prompts selon le type d'événement :

### Combat
- Préfixe : "Epic fantasy battle scene"
- Style : Epic (cinematic, dramatic lighting, heroic pose)
- Ajouts : "dynamic action, dramatic lighting"

### Exploration
- Préfixe : "Fantasy adventurers exploring"
- Style : Painted (oil painting, rich colors)
- Ajouts : "atmospheric, mysterious"

### Discovery
- Préfixe : "Moment of discovery in a dungeon"
- Style : Dark Fantasy (moody lighting, shadows)
- Ajouts : "revealing light, magical glow"

### Loot
- Préfixe : "Fantasy treasure"
- Style : Painted
- Ajouts : "glittering gold, magical items"

### Session (fin)
- Détecte automatiquement victoire/défaite
- Ajuste l'ambiance en conséquence
- Victoire : "triumphant heroes, celebration"
- Défaite : "somber scene, aftermath of battle"

## Coûts

Avec FLUX.1 [schnell] via fal.ai :
- ~$0.003 par image
- 22 entrées de journal = ~$0.07
- ~300 images par dollar

## Performance

Grâce à la génération parallèle :
- **Séquentiel** : 22 images × 3 sec = ~66 secondes
- **Parallèle (4)** : 22 images / 4 × 3 sec = ~17 secondes
- **Parallèle (8)** : 22 images / 8 × 3 sec = ~9 secondes

## Workflow Recommandé

1. **Dry-run d'abord** : Vérifier les prompts générés
   ```bash
   ./sw-image journal "mon-aventure" --dry-run
   ```

2. **Test limité** : Générer quelques images pour valider
   ```bash
   ./sw-image journal "mon-aventure" --max=3
   ```

3. **Génération complète** : Si satisfait, générer tout
   ```bash
   ./sw-image journal "mon-aventure"
   ```

## Intégration avec Adventure Manager

Après une session de jeu, utilisez cette skill pour créer un résumé visuel :

```bash
# Terminer la session
./sw-adventure end-session "Mon Aventure" "Victoire contre le boss !"

# Illustrer le journal
./sw-image journal "mon-aventure" --types=combat,discovery
```