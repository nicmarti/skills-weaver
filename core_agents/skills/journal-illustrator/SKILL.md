---
name: journal-illustrator
description: Illustre automatiquement le journal d'une aventure BFRPG en générant des images pour les moments clés (combats, explorations, découvertes). Utilise la génération parallèle pour une performance optimale.
allowed-tools: Bash
---

# Journal Illustrator - Illustration automatique des aventures

Skill pour générer des illustrations basées sur le journal d'une aventure BFRPG. Analyse les entrées du journal et génère des images appropriées pour chaque type d'événement.

## Prérequis

**Variables d'environnement** :
```bash
# Requis pour la génération d'images
export FAL_KEY="votre_clé_api_fal"

# Optionnel pour l'enrichissement AI des descriptions
export ANTHROPIC_API_KEY="votre_clé_anthropic"
```

## Utilisation des Descriptions Enrichies (Recommandé)

Pour de meilleurs résultats, **enrichissez d'abord votre journal** avec des descriptions détaillées avant de générer les images :

```bash
# 1. Enrichir le journal avec des descriptions IA (30-50 mots)
./sw-adventure enrich "la-crypte-des-ombres"

# 2. Générer les images (utilise automatiquement les descriptions enrichies)
./sw-image journal "la-crypte-des-ombres"
```

**Avantages de l'enrichissement** :
- ✅ Descriptions détaillées (30-50 mots vs 5-10 mots)
- ✅ Contexte riche : personnages, lieux, atmosphère
- ✅ Images plus fidèles aux événements du jeu
- ✅ Automatique : utilise l'historique récent et la composition du groupe

**Sans enrichissement**, le système utilise le champ `content` (court, générique).

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-image ./cmd/image

# Voir les prompts qui seraient générés (sans générer d'images)
./sw-image journal "la-crypte-des-ombres" --dry-run

# Générer toutes les illustrations
./sw-image journal "la-crypte-des-ombres"

# Reprendre depuis un ID spécifique (utile pour éviter les régénérations)
./sw-image journal "la-crypte-des-ombres" --start-id=60

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
--start-id=<n>      # ID de départ pour reprendre depuis une entrée spécifique (optionnel)
--max=<n>           # Nombre maximum d'images à générer
--parallel=<n>      # Niveau de parallélisme (1-8, défaut: 4)
--model=<model>     # Modèle fal.ai (schnell, banana, pulid) défaut: schnell
--consistency       # Utiliser la cohérence de personnage (défaut: true avec pulid)
--dry-run           # Afficher les prompts sans générer d'images
```

### Modèles Disponibles

| Modèle | Vitesse | Coût/image | Cohérence de personnage |
|--------|---------|------------|------------------------|
| `schnell` | ~3s | ~$0.003 | Non |
| `banana` | ~5s | ~$0.039 | Non |
| `pulid` | ~4s | ~$0.003-0.04 | Oui (requiert images de référence) |

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

### Reprendre depuis un ID spécifique

```bash
# Première génération jusqu'à l'ID 59
./sw-image journal "la-crypte-des-ombres" --max=15

# Reprendre depuis l'ID 60 pour éviter les régénérations
./sw-image journal "la-crypte-des-ombres" --start-id=60
```

Utile après une nouvelle session pour illustrer uniquement les nouvelles entrées du journal.

## Sortie

Les images sont sauvegardées dans le répertoire de l'aventure :

```
data/adventures/la-crypte-des-ombres/images/
├── image_1703001234567890123.png  # Combat contre les squelettes
├── image_1703001234567890124.png  # Exploration du mausolée
├── image_1703001234567890125.png  # Découverte du parchemin
└── ...
```

## Comparaison : Avec vs Sans Enrichissement

### Exemple : Entrée de type "combat"

**Sans enrichissement (champ `content`)** :
```
Content: "Combat contre 3 gobelins"

Prompt généré:
"Epic fantasy battle scene: Combat contre 3 gobelins. Dynamic action, dramatic
lighting, cinematic fantasy art, heroic pose"
```

**Avec enrichissement (champ `description`)** :
```
Description EN: "Aldric and Lyra battle three goblins in a torch-lit stone
corridor, steel clashing against crude blades as shadows dance on moss-covered
walls"

Prompt généré:
"Epic fantasy battle scene: Aldric and Lyra battle three goblins in a torch-lit
stone corridor, steel clashing against crude blades as shadows dance on moss-
covered walls. Dynamic action, dramatic lighting, cinematic fantasy art,
heroic pose"
```

**Résultat** : L'image enrichie montrera les personnages nommés (Aldric, Lyra),
l'environnement spécifique (couloir de pierre, torches, mousse), et l'atmosphère
exacte du combat !

## Cohérence de Personnage (Character Consistency)

Le journal illustrator supporte la cohérence des personnages via FLUX PuLID :

### Prérequis
1. **Apparences des personnages** définies via `sw-character appearance`
2. **Images de référence** définies via `sw-character set-reference`

### Utilisation
```bash
# Configuration de l'apparence d'un personnage
./sw-character appearance "Aldric" \
  --age=34 \
  --build=muscular \
  --armor="plate armor" \
  --weapon=longsword

# Définir l'image de référence (portrait frontal recommandé)
./sw-character set-reference "Aldric" path/to/aldric_portrait.png

# Génération standard (sans cohérence)
./sw-image journal "adventure" --model=schnell

# Avec cohérence de personnage (requiert images de référence)
./sw-image journal "adventure" --model=pulid --consistency
```

### Comment Ça Marche
1. Charge les personnages du groupe d'aventure
2. Identifie le personnage principal par entrée du journal (mention du nom ou défaut au leader)
3. Utilise l'image de référence du personnage avec FLUX PuLID
4. Injecte des descriptions courtes de personnages au lieu de répéter le texte complet

### Exemple de Différence de Prompt

**Sans cohérence** :
```
Epic fantasy battle scene: Aldric the human fighter (34 years old, muscular build,
black hair, bearded, wearing plate armor) and Lyra the elf magic-user (young,
slender, long silver hair) battle three goblins...
```

**Avec cohérence** :
```
Epic fantasy battle scene featuring Aldric (human fighter, plate armor, longsword),
Lyra (elf magic-user, staff): The heroes battle three goblins in a torch-lit corridor...
```

L'image de référence garantit que le visage d'Aldric correspond à travers toutes les images.

### Avantages
- ✅ Visages cohérents des personnages à travers toutes les images du journal
- ✅ Réduction des prompts (pas besoin de répéter les descriptions complètes)
- ✅ Coût similaire à schnell (~$0.003-0.04/image)
- ✅ Fallback automatique si l'image de référence est manquante

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

### Option A : Avec enrichissement AI (recommandé)

1. **Enrichir le journal** : Générer des descriptions détaillées
   ```bash
   ./sw-adventure enrich "mon-aventure" --dry-run  # Prévisualiser
   ./sw-adventure enrich "mon-aventure"             # Enrichir
   ```

2. **Dry-run d'abord** : Vérifier les prompts enrichis
   ```bash
   ./sw-image journal "mon-aventure" --dry-run
   ```

3. **Test limité** : Générer quelques images pour valider
   ```bash
   ./sw-image journal "mon-aventure" --max=3
   ```

4. **Génération complète** : Si satisfait, générer tout
   ```bash
   ./sw-image journal "mon-aventure"
   ```

### Option B : Sans enrichissement (rapide, mais moins précis)

1. **Dry-run d'abord** : Vérifier les prompts
   ```bash
   ./sw-image journal "mon-aventure" --dry-run
   ```

2. **Génération complète** : Utilise le champ `content` (court)
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

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Illustration automatique des journaux d'aventure |

**Type** : Skill autonome, peut être invoqué directement via `/journal-illustrator`

**Dépendances** : Utilise `adventure-manager` (lecture du journal) et `image-generator` (génération d'images)