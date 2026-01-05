---
name: spell-reference
description: Consulte les sorts D&D 5e par classe et niveau (0-9). Cantrips, écoles, concentration, rituels. Utilisez pour vérifier les sorts lancés.
allowed-tools: Bash
---

# Spell Reference - Grimoire des Sorts D&D 5e

Skill pour consulter les sorts D&D 5e (257 sorts, 12 classes, 8 écoles de magie). Indispensable pour vérifier les effets des sorts pendant le jeu.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-spell ./cmd/spell

# Lister les sorts par classe
./sw-spell list --class=wizard

# Cantrips d'une classe
./sw-spell cantrips wizard

# Détails d'un sort
./sw-spell show projectile_magique

# Sorts de concentration
./sw-spell concentration

# Slots de sorts
./sw-spell slots wizard --level=5
```

## Commandes Disponibles

### Lister les Sorts

```bash
./sw-spell list                              # Tous les sorts (257)
./sw-spell list --class=wizard               # Sorts de magicien
./sw-spell list --class=cleric --level=1     # Clerc niveau 1
./sw-spell list --level=0                    # Cantrips
./sw-spell list --school=evocation           # École d'évocation
./sw-spell list --format=md                  # Fiches détaillées
./sw-spell list --format=json                # Format JSON
```

### Cantrips (Level 0)

```bash
./sw-spell cantrips <classe>

# Exemples:
./sw-spell cantrips wizard        # Cantrips de magicien (13)
./sw-spell cantrips cleric        # Cantrips de clerc
./sw-spell cantrips warlock       # Cantrips d'occultiste
```

### Écoles de Magie

```bash
./sw-spell schools                # Liste les 8 écoles
./sw-spell schools --format=detail   # Avec exemples de sorts
```

**8 Écoles D&D 5e** :
- **Abjuration** : Protection (Bouclier, Protection contre le mal)
- **Invocation** : Création/téléportation (Invoquer familier)
- **Divination** : Connaissance (Détection de la magie)
- **Enchantement** : Contrôle mental (Charme-personne)
- **Évocation** : Énergie/dégâts (Projectile magique, Boule de feu)
- **Illusion** : Tromperie (Image silencieuse)
- **Nécromancie** : Mort/non-mort (Animation des morts)
- **Transmutation** : Transformation (Métamorphose)

### Sorts de Concentration

```bash
./sw-spell concentration          # Liste 69 sorts concentration
./sw-spell concentration --format=md   # Fiches détaillées
```

**Règle concentration** : Seul 1 sort de concentration peut être actif à la fois. Brisée si : dégâts (JdS CON DC 10 ou ½ dégâts), incapacité, mort, nouveau sort concentration.

### Sorts Rituels

```bash
./sw-spell rituals                # Liste 22 sorts rituels
```

**Règle rituels** : Sorts rituels prennent +10 minutes mais ne consomment pas de slot.

### Afficher un Sort

```bash
./sw-spell show <id>

# Exemples:
./sw-spell show projectile_magique    # Projectile magique
./sw-spell show soin_des_blessures    # Soins des blessures
./sw-spell show sommeil               # Sommeil
./sw-spell show --format=json         # Format JSON
./sw-spell show --format=short        # Une ligne
```

### Rechercher des Sorts

```bash
./sw-spell search <terme>

# Exemples:
./sw-spell search lumière            # Par nom français
./sw-spell search light              # Par nom anglais
./sw-spell search feu                # Sorts de feu
```

### Slots de Sorts

```bash
./sw-spell slots <classe> --level=<niveau>

# Exemples:
./sw-spell slots wizard --level=5    # Magicien niveau 5
./sw-spell slots paladin --level=3   # Paladin niveau 3
./sw-spell slots warlock --level=11  # Occultiste niveau 11
```

## Classes de Lanceurs D&D 5e

| Classe | Type | Début | Niveaux max |
|--------|------|-------|-------------|
| **Magicien** (wizard) | Full caster | 1 | 9 |
| **Ensorceleur** (sorcerer) | Full caster | 1 | 9 |
| **Clerc** (cleric) | Full caster | 1 | 9 |
| **Druide** (druid) | Full caster | 1 | 9 |
| **Barde** (bard) | Full caster | 1 | 9 |
| **Occultiste** (warlock) | Pact caster | 1 | 5 (pact slots) |
| **Paladin** | Half caster | 2 | 5 |
| **Rôdeur** (ranger) | Half caster | 2 | 5 |
| **Guerrier** (fighter) | 1/3 caster | 3 | 4 (Eldritch Knight) |
| **Roublard** (rogue) | 1/3 caster | 3 | 4 (Arcane Trickster) |

**Aliases FR/EN acceptés** : magicien/wizard, clerc/cleric, ensorceleur/sorcerer, occultiste/warlock, rôdeur/ranger, guerrier/fighter, roublard/rogue.

## Format de Sortie

### Fiche Sort (Markdown)

```markdown
## Projectile magique

**Niveau 1** | **École** : Évocation

| Caractéristique | Valeur |
|-----------------|--------|
| **Temps d'incantation** | action |
| **Portée** | 36 m |
| **Composantes** | V, S |
| **Durée** | instantanée |
| **Classes** | Magicien, Ensorceleur |
| **Dégâts** | 1d4+1 |

### Description

Vous créez trois fléchettes lumineuses d'énergie magique...

### Aux niveaux supérieurs

Chaque emplacement de sort de niveau supérieur crée une fléchette supplémentaire.
```

### Format Court

```
Projectile magique [N1 Évocation] - 36 m, instantanée
Lumière [Cantrip Évocation] - contact, 1 heure
Détection de la magie (C) (R) [N1 Divination] - personnelle, Concentration, jusqu'à 10 minutes
```

**(C)** = Concentration | **(R)** = Rituel

## Slots de Sorts par Niveau

### Full Casters (Wizard, Sorcerer, Cleric, Druid, Bard)

| Niv | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | Cantrips |
|-----|---|---|---|---|---|---|---|---|---|----------|
| 1   | 2 | - | - | - | - | - | - | - | - | 3-4 |
| 2   | 3 | - | - | - | - | - | - | - | - | 3-4 |
| 3   | 4 | 2 | - | - | - | - | - | - | - | 3-4 |
| 5   | 4 | 3 | 2 | - | - | - | - | - | - | 4-5 |
| 9   | 4 | 3 | 3 | 3 | 1 | - | - | - | - | 4-6 |
| 17  | 4 | 3 | 3 | 3 | 2 | 1 | 1 | 1 | 1 | 5-6 |
| 20  | 4 | 3 | 3 | 3 | 3 | 2 | 2 | 1 | 1 | 5-6 |

### Half Casters (Paladin, Ranger)

| Niv | 1 | 2 | 3 | 4 | 5 | Cantrips |
|-----|---|---|---|---|---|----------|
| 2   | 2 | - | - | - | - | - |
| 5   | 4 | 2 | - | - | - | - |
| 9   | 4 | 3 | 2 | - | - | - |
| 13  | 4 | 3 | 3 | 1 | - | - |
| 17  | 4 | 3 | 3 | 3 | 1 | - |
| 20  | 4 | 3 | 3 | 3 | 2 | - |

### Warlock (Pact Magic)

| Niv | Slot Level | Slots | Cantrips |
|-----|-----------|-------|----------|
| 1   | 1 | 1 | 2 |
| 2   | 1 | 2 | 2 |
| 3   | 2 | 2 | 2 |
| 5   | 3 | 2 | 3 |
| 11  | 5 | 3 | 4 |
| 17  | 5 | 4 | 4 |

**Note Occultiste** : Tous les slots sont du même niveau. Restaurés au repos court.

## Mécaniques D&D 5e

### Concentration

- **Maximum** : 1 sort de concentration actif à la fois
- **Brisée si** :
  - Dégâts : JdS CON DC 10 ou ½ dégâts (le plus élevé)
  - Incapacité ou mort
  - Nouveau sort de concentration lancé
- **69 sorts** avec concentration (vérifier avec `./sw-spell concentration`)

### Rituels

- **+10 minutes** de temps d'incantation
- **Pas de slot** consommé
- **22 sorts** rituels (vérifier avec `./sw-spell rituals`)
- Exemples : Détection de la magie, Identification, Alarme

### Cantrips

- **Illimités** par jour
- **Scale avec niveau perso** (pas niveau sort)
- Exemples : Trait de feu 1d10 → 2d10 (niv 5) → 3d10 (niv 11) → 4d10 (niv 17)

### Upcasting

- Lancer sort avec slot **niveau supérieur** pour effet amélioré
- Exemple : Projectile magique avec slot niveau 3 = 5 fléchettes au lieu de 3
- Champ `upcast` dans la fiche sort décrit l'effet

### Composantes

- **V** (Verbal) : Parole nécessaire
- **S** (Somatique) : Main libre nécessaire
- **M** (Matériel) : Matériau ou focalisateur arcanique/divin
- Si **M** avec coût spécifique : matériau consommé

### Spell Save DC

**Formule** : `8 + bonus maîtrise + modificateur caractéristique`

Exemple : Magicien niveau 5, INT 16 (+3)
- Bonus maîtrise : +3
- DD sauvegarde : 8 + 3 + 3 = **14**

### Spell Attack Bonus

**Formule** : `bonus maîtrise + modificateur caractéristique`

Exemple : Magicien niveau 5, INT 16 (+3)
- Bonus attaque : 3 + 3 = **+6**

## Intégration avec Adventure Manager

```bash
# Vérifier un sort avant de l'utiliser
./sw-spell show sommeil

# Logger le sort dans le journal (via sw-dm)
# L'agent dungeon-master peut appeler get_spell tool
```

## Conseils d'Utilisation

### Pour le Dungeon Master

```bash
# Vérifier rapidement un sort ennemi
./sw-spell show charme_personne

# Trouver sorts de zone
./sw-spell search "rayon"

# Vérifier sorts concentration actifs
./sw-spell concentration
```

### Pour vérifier un personnage

```bash
# Sorts disponibles pour magicien niveau 1
./sw-spell list --class=wizard --level=1

# Cantrips de clerc
./sw-spell cantrips cleric

# Slots disponibles
./sw-spell slots wizard --level=5
```

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Vérification effets sorts, concentration, slots |
| `rules-keeper` | Référence sorts pour arbitrage |
| `character-creator` | Choix sorts initiaux (cantrips + level 1) |

**Type** : Skill autonome, peut être invoqué directement via `/spell-reference`

## Données

- **257 sorts** D&D 5e dans `data/5e/spells.json`
- **Source** : docs/markdown-new/sorts_et_magie.md (376 KB)
- **Niveaux** : 0-9 (0 = cantrips, 25 sorts)
- **Écoles** : 8 (abjuration, conjuration, divination, enchantment, evocation, illusion, necromancy, transmutation)
- **Concentration** : 69 sorts
- **Rituels** : 22 sorts