---
name: monster-manual
description: Bestiaire D&D 5e avec stats de combat, génération de rencontres et PV aléatoires. Indispensable pour le Maître du Jeu en combat. Contient 33 monstres classiques fantasy.
allowed-tools:
  - Bash
---

# Monster Manual - Bestiaire D&D 5e

Skill pour consulter les statistiques des monstres, générer des rencontres équilibrées et créer des groupes d'ennemis avec PV individuels.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-monster ./cmd/monster

# Consulter un monstre
./sw-monster show goblin

# Générer une rencontre (CR approprié)
./sw-monster encounter dungeon_level_1

# Générer par niveau de groupe
./sw-monster encounter --level=3

# Créer des ennemis avec PV
./sw-monster roll orc --count=4
```

## Commandes Disponibles

### Afficher un Monstre

```bash
./sw-monster show <id>

# Exemples:
./sw-monster show goblin           # Fiche complète du gobelin
./sw-monster show dragon_red_adult # Dragon rouge adulte
./sw-monster show --format=json    # Format JSON
./sw-monster show --format=short   # Une ligne
```

### Rechercher des Monstres

```bash
./sw-monster search <terme>

# Exemples:
./sw-monster search dragon         # Tous les dragons
./sw-monster search mort           # Morts-vivants (par nom FR)
./sw-monster search undead         # Morts-vivants (par type)
```

### Lister les Monstres

```bash
./sw-monster list                  # Tous les monstres
./sw-monster list --type=undead    # Par type
./sw-monster list --type=humanoid  # Humanoïdes seulement
./sw-monster list --cr=1/4         # Par Challenge Rating
./sw-monster list --cr=2           # CR 2
```

### Générer une Rencontre

```bash
./sw-monster encounter <table>
./sw-monster encounter --level=<N>

# Tables disponibles:
./sw-monster encounter dungeon_level_1    # Niveau 1-2 (CR 1/8 - 1/2)
./sw-monster encounter dungeon_level_2    # Niveau 3-4 (CR 1 - 2)
./sw-monster encounter dungeon_level_3    # Niveau 5-6 (CR 3 - 5)
./sw-monster encounter dungeon_level_4    # Niveau 7+ (CR 6+)
./sw-monster encounter forest             # Forêt
./sw-monster encounter undead_crypt       # Crypte

# Par niveau de groupe (utilise budget XP):
./sw-monster encounter --level=3          # Pour groupe niveau 3
./sw-monster encounter --level=5 --size=4 # 4 joueurs niveau 5
```

### Créer des Monstres avec PV

```bash
./sw-monster roll <id> --count=N

# Exemples:
./sw-monster roll goblin --count=6    # 6 gobelins
./sw-monster roll skeleton --count=4  # 4 squelettes
./sw-monster roll troll               # 1 troll
```

## Système D&D 5e

### Challenge Rating (CR)

Le **Challenge Rating** représente la difficulté d'un monstre pour un groupe de 4 aventuriers :

| CR | XP | Niveau Groupe | Exemples |
|----|-----|---------------|----------|
| 0 | 10 | 1 | Rat, Mille-pattes |
| 1/8 | 25 | 1 | Cultiste, Bandit |
| 1/4 | 50 | 1 | Gobelin, Squelette |
| 1/2 | 100 | 1 | Orc, Hobgobelin |
| 1 | 200 | 2 | Goule, Gnoll |
| 2 | 450 | 3-4 | Ogre, Zombie Ogre |
| 3 | 700 | 4-5 | Wight, Hibours |
| 4 | 1,100 | 5-6 | Cocatrice, Harpy |
| 5 | 1,800 | 6-7 | Troll, Méduse |
| 6+ | 2,300+ | 8+ | Dragon jeune, Vampire |

### Budget XP D&D 5e

Les rencontres générées suivent les budgets XP officiels :

| Niveau | Easy | Medium | Hard | Deadly |
|--------|------|--------|------|--------|
| 1 | 25 | 50 | 75 | 100 |
| 2 | 50 | 100 | 150 | 200 |
| 3 | 75 | 150 | 225 | 400 |
| 4 | 125 | 250 | 375 | 500 |
| 5 | 250 | 500 | 750 | 1,100 |
| 6+ | ... | ... | ... | ... |

**Note** : Les rencontres générées automatiquement visent la difficulté "Medium" à "Hard".

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
- `giant_rat` (CR 1/8), `giant_bat` (CR 1/4), `wolf` (CR 1/4), `dire_wolf` (CR 1), `bear` (CR 1)

### Humanoïdes
- `goblin` (CR 1/4), `hobgoblin` (CR 1/2), `kobold` (CR 1/8), `orc` (CR 1/2), `bugbear` (CR 1), `gnoll` (CR 1/2)

### Morts-vivants
- `skeleton` (CR 1/4), `zombie` (CR 1/4), `ghoul` (CR 1), `wight` (CR 3), `wraith` (CR 5), `vampire` (CR 13), `lich` (CR 21)

### Monstres
- `owlbear` (CR 3), `minotaur` (CR 3), `harpy` (CR 1), `cockatrice` (CR 1/2), `basilisk` (CR 3), `medusa` (CR 6), `rust_monster` (CR 1/2)

### Géants
- `ogre` (CR 2), `troll` (CR 5)

### Dragons
- `dragon_red_young` (CR 10), `dragon_red_adult` (CR 17)

### Vases
- `green_slime` (CR 2), `gelatinous_cube` (CR 2)

### Vermines
- `giant_spider` (CR 1), `giant_centipede` (CR 1/4)

## Format de Sortie

### Fiche Monstre (Markdown)

```markdown
## Gobelin (Goblin)

**Type** : humanoid | **Taille** : Small

### Statistiques de Combat

| Stat | Valeur |
|------|--------|
| **Challenge Rating** | 1/4 (50 XP) |
| **Classe d'Armure** | 15 (armure de cuir, bouclier) |
| **Points de Vie** | 2d6 (moy. 7 PV) |
| **Vitesse** | 30 pieds |
| **Type de Trésor** | R (personnel) |

### Caractéristiques

| FOR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| 8 (-1) | 14 (+2) | 10 (+0) | 10 (+0) | 8 (-1) | 8 (-1) |

### Attaques
- **Cimeterre** : +4 au toucher, 1d6+2 dégâts tranchants
- **Arc court** : +4 au toucher, 1d6+2 dégâts perforants (portée 80/320)

### Capacités Spéciales
- **Vision dans le noir** : 60 pieds
- **Fuite agile** : Peut utiliser l'action Esquiver en action bonus
```

### Rencontre Générée

```markdown
## Rencontre : Niveau 1 de donjon - Difficulté Medium

Budget XP : 200 (4 joueurs niveau 2)

### Monstres

**Gobelin** x4 (CR 1/4 chacun, 50 XP)
- CA 15, PV individuels : 7, 5, 9, 6
- Cimeterre : +4, 1d6+2 tranchant
- Arc court : +4, 1d6+2 perforant (80/320)

**XP Total** : 200 XP
```

## Intégration avec Adventure Manager

```bash
# Générer une rencontre
./sw-monster encounter forest

# Logger le combat
./sw-adventure log "Mon Aventure" combat "Embuscade de 3 loups (CR 1/4 chacun)"

# Après victoire, ajouter l'XP et le butin
./sw-adventure add-gold "Mon Aventure" 25 "Trésor des loups"
```

## Conseils d'Utilisation

### Pour préparer un combat

```bash
# 1. Générer la rencontre par niveau de groupe
./sw-monster encounter --level=3

# 2. Ou utiliser une table prédéfinie
./sw-monster encounter dungeon_level_2

# 3. Ou créer des monstres spécifiques
./sw-monster roll orc --count=3
```

### Pour consulter rapidement

```bash
# Stats en une ligne
./sw-monster show goblin --format=short
# Gobelin (humanoid) - CA 15, CR 1/4 (50 XP), 2d6 PV

# Par Challenge Rating
./sw-monster list --cr=1/4  # Tous les monstres CR 1/4
./sw-monster list --cr=2    # Tous les monstres CR 2
```

### Pour un boss

```bash
./sw-monster show troll           # CR 5 (1,800 XP)
./sw-monster show dragon_red_adult # CR 17 (18,000 XP)
./sw-monster show lich            # CR 21 (33,000 XP)
```

## Tables de Rencontres

### dungeon_level_1 (Niveau 1-2, CR 1/8 - 1/2)
Rats géants, Gobelins, Kobolds, Squelettes, Araignées, Chauves-souris

**Budget XP** : 50-100 par joueur (Easy à Medium)

### dungeon_level_2 (Niveau 3-4, CR 1/2 - 2)
Orcs, Hobgobelins, Zombies, Goules, Loups, Bugbears, Gnolls

**Budget XP** : 150-250 par joueur (Medium à Hard)

### dungeon_level_3 (Niveau 5-6, CR 3 - 5)
Ogres, Wights, Hibours, Harpies, Cocatrices, Minotaures

**Budget XP** : 500-750 par joueur (Medium à Hard)

### dungeon_level_4 (Niveau 7+, CR 6+)
Trolls, Vampires, Méduses, Dragons, Basilics, Liches

**Budget XP** : 1,100+ par joueur (Hard à Deadly)

## Différences D&D 5e vs BFRPG

| Aspect | BFRPG | D&D 5e |
|--------|-------|--------|
| Difficulté | Hit Dice (HD) | Challenge Rating (CR) |
| Valeurs | Entiers (1-20) | Fractionnaires (1/8, 1/4, 1/2, 1-30) |
| XP | Par HD | Par CR (table officielle) |
| Rencontres | Basé sur HD | Budget XP (Easy/Medium/Hard/Deadly) |
| Sauvegardes | 5 catégories fixes | 6 sauvegardes (FOR, DEX, CON, INT, WIS, CHA) |
| Moral | Oui (2-12) | Non (décision du MJ) |

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Stats monstres, génération de rencontres équilibrées |
| `rules-keeper` | Consultation des statistiques de monstres |

**Type** : Skill autonome, peut être invoqué directement via `/monster-manual`