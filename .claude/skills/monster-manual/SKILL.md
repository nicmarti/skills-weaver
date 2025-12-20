---
name: monster-manual
description: Bestiaire BFRPG avec stats de combat, génération de rencontres et PV aléatoires. Indispensable pour le Maître du Jeu en combat. Contient 33 monstres classiques fantasy.
allowed-tools: Bash
---

# Monster Manual - Bestiaire BFRPG

Skill pour consulter les statistiques des monstres, générer des rencontres aléatoires et créer des groupes d'ennemis avec PV individuels.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o monster ./cmd/monster

# Consulter un monstre
./monster show goblin

# Générer une rencontre
./monster encounter dungeon_level_1

# Créer des ennemis avec PV
./monster roll orc --count=4
```

## Commandes Disponibles

### Afficher un Monstre

```bash
./monster show <id>

# Exemples:
./monster show goblin           # Fiche complète du gobelin
./monster show dragon_red_adult # Dragon rouge adulte
./monster show --format=json    # Format JSON
./monster show --format=short   # Une ligne
```

### Rechercher des Monstres

```bash
./monster search <terme>

# Exemples:
./monster search dragon         # Tous les dragons
./monster search mort           # Morts-vivants (par nom FR)
./monster search undead         # Morts-vivants (par type)
```

### Lister les Monstres

```bash
./monster list                  # Tous les monstres
./monster list --type=undead    # Par type
./monster list --type=humanoid  # Humanoïdes seulement
```

### Générer une Rencontre

```bash
./monster encounter <table>
./monster encounter --level=<N>

# Tables disponibles:
./monster encounter dungeon_level_1    # Niveau 1 (faible)
./monster encounter dungeon_level_2    # Niveau 2 (modéré)
./monster encounter dungeon_level_3    # Niveau 3 (élevé)
./monster encounter dungeon_level_4    # Niveau 4+ (très élevé)
./monster encounter forest             # Forêt
./monster encounter undead_crypt       # Crypte

# Par niveau de groupe:
./monster encounter --level=3          # Pour groupe niveau 3
```

### Créer des Monstres avec PV

```bash
./monster roll <id> --count=N

# Exemples:
./monster roll goblin --count=6    # 6 gobelins
./monster roll skeleton --count=4  # 4 squelettes
./monster roll troll               # 1 troll
```

## Types de Monstres

| Type | Description | Exemples |
|------|-------------|----------|
| `animal` | Animaux naturels | Loup, Ours, Rat géant |
| `dragon` | Dragons | Dragon rouge jeune/adulte |
| `giant` | Géants | Ogre, Troll |
| `humanoid` | Humanoïdes | Gobelin, Orc, Gnoll |
| `monstrosity` | Monstres magiques | Hibours, Méduse, Minotaure |
| `ooze` | Vases | Cube gélatineux, Gelée verte |
| `undead` | Morts-vivants | Squelette, Zombie, Vampire |
| `vermin` | Vermines | Araignée géante, Mille-pattes |

## Monstres Disponibles (33)

### Animaux
- `giant_rat`, `giant_bat`, `wolf`, `dire_wolf`, `bear`

### Humanoïdes
- `goblin`, `hobgoblin`, `kobold`, `orc`, `bugbear`, `gnoll`

### Morts-vivants
- `skeleton`, `zombie`, `ghoul`, `wight`, `wraith`, `vampire`, `lich`

### Monstres
- `owlbear`, `minotaur`, `harpy`, `cockatrice`, `basilisk`, `medusa`, `rust_monster`

### Géants
- `ogre`, `troll`

### Dragons
- `dragon_red_young`, `dragon_red_adult`

### Vases
- `green_slime`, `gelatinous_cube`

### Vermines
- `giant_spider`, `giant_centipede`

## Format de Sortie

### Fiche Monstre (Markdown)

```markdown
## Gobelin (Goblin)

**Type** : humanoid | **Taille** : small

### Statistiques de Combat

| Stat | Valeur |
|------|--------|
| **Dés de Vie** | 1d8-1 (moy. 3 PV) |
| **Classe d'Armure** | 14 |
| **Mouvement** | 20 |
| **Sauvegarde** | Normal Human |
| **Moral** | 7 |
| **Trésor** | R |
| **XP** | 10 |

### Attaques
- **arme** : +0, 1d6

### Capacités Spéciales
- darkvision 60ft
- -1 attack in daylight
```

### Rencontre Générée

```markdown
## Rencontre : Niveau 1 de donjon - Faible difficulté

### Monstres

**Gobelin** x4
- CA 14, PV : 3, 2, 5, 1
- arme : +0, 1d6

**XP Total** : 40
```

## Intégration avec Adventure Manager

```bash
# Générer une rencontre
./monster encounter forest

# Logger le combat
./adventure log "Mon Aventure" combat "Embuscade de 3 loups"

# Après victoire, ajouter l'XP et le butin
./adventure add-gold "Mon Aventure" 25 "Trésor des loups"
```

## Conseils d'Utilisation

### Pour préparer un combat
```bash
# 1. Générer la rencontre
./monster encounter dungeon_level_2

# 2. Ou créer des monstres spécifiques
./monster roll orc --count=3
```

### Pour consulter rapidement
```bash
# Stats en une ligne
./monster show goblin --format=short
# Gobelin (humanoid) - CA 14, DV 1d8-1 (3 PV), XP 10
```

### Pour un boss
```bash
./monster show troll
./monster show dragon_red_adult
./monster show lich
```

## Tables de Rencontres

### dungeon_level_1 (Niveau 1-2)
Rats géants, Gobelins, Kobolds, Squelettes, Araignées, Chauves-souris

### dungeon_level_2 (Niveau 3-4)
Orcs, Hobgobelins, Zombies, Goules, Loups, Bugbears, Gnolls

### dungeon_level_3 (Niveau 5-6)
Ogres, Wights, Hibours, Harpies, Cocatrices, Minotaures

### dungeon_level_4 (Niveau 7+)
Trolls, Vampires, Méduses, Dragons, Basilics, Liches