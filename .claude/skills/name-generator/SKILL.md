---
name: name-generator
description: Génère des noms de personnages fantasy selon la race et le sexe. Supporte nains, elfes, halfelins, humains et PNJ (tavernier, marchand, garde, noble, mage, méchant). Utilisez pour nommer joueurs et PNJ.
allowed-tools: Bash
---

# Name Generator - Générateur de Noms Fantasy

Skill pour générer des noms de personnages pour Basic Fantasy RPG.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-names ./cmd/names

# Générer un nom
./sw-names generate dwarf              # Nom de nain
./sw-names generate elf --gender=f     # Nom d'elfe féminin
./sw-names npc innkeeper               # Nom de tavernier
```

## Commandes Disponibles

### Génération par Race

```bash
# Races disponibles: dwarf, elf, halfling, human
./sw-names generate <race> [options]

# Options:
#   --gender=m|f     Sexe (m=masculin, f=féminin, omis=aléatoire)
#   --count=N        Nombre de noms à générer
#   --first-only     Générer uniquement le prénom
```

### Exemples par Race

```bash
# Nains
./sw-names generate dwarf                    # Thorin Ironfoot
./sw-names generate dwarf --gender=f         # Disa Stoneheart

# Elfes
./sw-names generate elf                      # Legolas Moonwhisper
./sw-names generate elf --gender=f           # Arwen Starweaver

# Halfelins
./sw-names generate halfling                 # Bilbo Baggins
./sw-names generate halfling --gender=f      # Rose Greenhill

# Humains
./sw-names generate human                    # Aragorn Ironhand
./sw-names generate human --gender=f         # Eowyn Stormrider
```

### Génération Multiple

```bash
# Générer plusieurs noms pour une liste de choix
./sw-names generate dwarf --count=5
./sw-names generate elf --gender=f --count=3
```

### Prénoms Uniquement

```bash
# Pour les PNJ simples ou les alias
./sw-names generate human --first-only
./sw-names generate halfling --first-only --count=5
```

### Noms de PNJ

```bash
# Types disponibles: innkeeper, merchant, guard, noble, wizard, villain
./sw-names npc <type> [--count=N]

# Exemples
./sw-names npc innkeeper     # Barnabas (tavernier)
./sw-names npc merchant      # Cornelius (marchand)
./sw-names npc guard         # Bruno (garde)
./sw-names npc noble         # Casimir (noble)
./sw-names npc wizard        # Balthazar (mage)
./sw-names npc villain       # Malachar (méchant)
```

### Lister les Options

```bash
./sw-names list              # Tout lister
./sw-names list races        # Races disponibles
./sw-names list npc          # Types de PNJ
```

## Correspondances Français-Anglais

| Français | Anglais | Commande |
|----------|---------|----------|
| Nain | Dwarf | `dwarf` |
| Elfe | Elf | `elf` |
| Halfelin | Halfling | `halfling` |
| Humain | Human | `human` |
| Tavernier | Innkeeper | `innkeeper` |
| Marchand | Merchant | `merchant` |
| Garde | Guard | `guard` |
| Noble | Noble | `noble` |
| Mage | Wizard | `wizard` |
| Méchant | Villain | `villain` |

## Intégration avec Character Generator

Après avoir généré un nom, utilisez-le pour créer un personnage :

```bash
# 1. Générer un nom
./sw-names generate dwarf --gender=m
# Output: Thorin Ironfoot

# 2. Créer le personnage avec ce nom
./sw-character create "Thorin Ironfoot" --race=dwarf --class=fighter
```

## Intégration avec Adventure Manager

Pour nommer les PNJ rencontrés dans une aventure :

```bash
# Générer un nom de tavernier
./sw-names npc innkeeper
# Output: Barnabas

# Logger la rencontre
./sw-adventure log "Mon Aventure" npc "Rencontre avec Barnabas, le tavernier"
```

## Structure des Données

Les noms sont stockés dans `data/names.json` avec :

- **~100 prénoms** par race et par sexe
- **~100 noms de famille** par race
- **Noms de PNJ** par type (tavernier, marchand, etc.)

## Style des Noms

| Race | Style |
|------|-------|
| Nain | Nordique/germanique + noms composés (Ironfoot, Stoneheart) |
| Elfe | Tolkien/Sindarin + nature (Moonwhisper, Starweaver) |
| Halfelin | Anglais bucolique + nature (Baggins, Greenhill) |
| Humain | Médiéval européen + épique (Ironhand, Stormrider) |

## Conseils d'Utilisation

- Pour un **joueur** : `./sw-names generate <race> --gender=<m|f>`
- Pour un **PNJ récurrent** : `./sw-names generate <race>` (nom complet)
- Pour un **PNJ mineur** : `./sw-names npc <type>` ou `--first-only`
- Pour une **liste de choix** : `--count=5`

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Noms pour PNJ et lieux |
| `character-creator` | Suggestions de noms pour joueurs |

**Type** : Skill autonome, peut être invoqué directement via `/name-generator`