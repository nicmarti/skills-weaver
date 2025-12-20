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
go build -o names ./cmd/names

# Générer un nom
./names generate dwarf              # Nom de nain
./names generate elf --gender=f     # Nom d'elfe féminin
./names npc innkeeper               # Nom de tavernier
```

## Commandes Disponibles

### Génération par Race

```bash
# Races disponibles: dwarf, elf, halfling, human
./names generate <race> [options]

# Options:
#   --gender=m|f     Sexe (m=masculin, f=féminin, omis=aléatoire)
#   --count=N        Nombre de noms à générer
#   --first-only     Générer uniquement le prénom
```

### Exemples par Race

```bash
# Nains
./names generate dwarf                    # Thorin Ironfoot
./names generate dwarf --gender=f         # Disa Stoneheart

# Elfes
./names generate elf                      # Legolas Moonwhisper
./names generate elf --gender=f           # Arwen Starweaver

# Halfelins
./names generate halfling                 # Bilbo Baggins
./names generate halfling --gender=f      # Rose Greenhill

# Humains
./names generate human                    # Aragorn Ironhand
./names generate human --gender=f         # Eowyn Stormrider
```

### Génération Multiple

```bash
# Générer plusieurs noms pour une liste de choix
./names generate dwarf --count=5
./names generate elf --gender=f --count=3
```

### Prénoms Uniquement

```bash
# Pour les PNJ simples ou les alias
./names generate human --first-only
./names generate halfling --first-only --count=5
```

### Noms de PNJ

```bash
# Types disponibles: innkeeper, merchant, guard, noble, wizard, villain
./names npc <type> [--count=N]

# Exemples
./names npc innkeeper     # Barnabas (tavernier)
./names npc merchant      # Cornelius (marchand)
./names npc guard         # Bruno (garde)
./names npc noble         # Casimir (noble)
./names npc wizard        # Balthazar (mage)
./names npc villain       # Malachar (méchant)
```

### Lister les Options

```bash
./names list              # Tout lister
./names list races        # Races disponibles
./names list npc          # Types de PNJ
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
./names generate dwarf --gender=m
# Output: Thorin Ironfoot

# 2. Créer le personnage avec ce nom
./character create "Thorin Ironfoot" --race=dwarf --class=fighter
```

## Intégration avec Adventure Manager

Pour nommer les PNJ rencontrés dans une aventure :

```bash
# Générer un nom de tavernier
./names npc innkeeper
# Output: Barnabas

# Logger la rencontre
./adventure log "Mon Aventure" npc "Rencontre avec Barnabas, le tavernier"
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

- Pour un **joueur** : `./names generate <race> --gender=<m|f>`
- Pour un **PNJ récurrent** : `./names generate <race>` (nom complet)
- Pour un **PNJ mineur** : `./names npc <type>` ou `--first-only`
- Pour une **liste de choix** : `--count=5`