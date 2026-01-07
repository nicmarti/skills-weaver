---
name: equipment-browser
description: Consulte les armes, armures et équipement BFRPG. Dégâts, CA, coût, propriétés. Utilisez pour vérifier l'équipement des personnages.
allowed-tools: Bash
---

# Equipment Browser - Catalogue d'Équipement BFRPG

Skill pour consulter les armes, armures, équipement d'aventure et munitions. Indispensable pour équiper les personnages et vérifier les statistiques en jeu.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-equipment ./cmd/equipment

# Lister les armes
./sw-equipment weapons

# Détails d'un item
./sw-equipment show longsword

# Équipement de départ
./sw-equipment starting fighter
```

## Commandes Disponibles

### Lister les Armes

```bash
./sw-equipment weapons
./sw-equipment weapons --type=melee    # Armes de mêlée
./sw-equipment weapons --type=ranged   # Armes à distance
./sw-equipment weapons --format=json   # Format JSON
./sw-equipment weapons --format=md     # Fiches détaillées
```

### Lister les Armures

```bash
./sw-equipment armor
./sw-equipment armor --type=light      # Armures légères
./sw-equipment armor --type=medium     # Armures moyennes
./sw-equipment armor --type=heavy      # Armures lourdes
./sw-equipment armor --type=shield     # Boucliers
./sw-equipment armor --format=json     # Format JSON
```

### Équipement d'Aventure

```bash
./sw-equipment gear                    # Liste tout l'équipement
./sw-equipment gear --format=json      # Format JSON
```

### Munitions

```bash
./sw-equipment ammo                    # Liste les munitions
```

### Afficher un Item

```bash
./sw-equipment show <id>

# Exemples:
./sw-equipment show longsword          # Épée longue
./sw-equipment show chainmail          # Cotte de mailles
./sw-equipment show backpack           # Sac à dos
./sw-equipment show --format=json      # Format JSON
./sw-equipment show --format=short     # Une ligne
```

### Rechercher des Items

```bash
./sw-equipment search <terme>

# Exemples:
./sw-equipment search épée             # Par nom français
./sw-equipment search sword            # Par nom anglais
./sw-equipment search leather          # Armure de cuir
```

### Équipement de Départ

```bash
./sw-equipment starting <classe>

# Classes disponibles:
./sw-equipment starting fighter        # Guerrier
./sw-equipment starting cleric         # Clerc
./sw-equipment starting magic-user     # Magicien
./sw-equipment starting thief          # Voleur
```

## Types d'Armes

| Type | Description | Exemples |
|------|-------------|----------|
| `melee` | Armes de corps à corps | Épée, Hache, Masse |
| `ranged` | Armes à distance | Arc, Arbalète, Fronde |

### Propriétés d'Armes

| Propriété | Description |
|-----------|-------------|
| `light` | Légère, utilisable en main secondaire |
| `two-handed` | Nécessite deux mains |
| `thrown` | Peut être lancée |
| `versatile` | Utilisable à une ou deux mains |
| `reach` | Allonge (2m au lieu de 1.5m) |
| `loading` | Nécessite rechargement |

## Types d'Armures

| Type | Bonus CA | Exemples |
|------|----------|----------|
| `light` | +2 | Armure de cuir |
| `medium` | +4 | Cotte de mailles |
| `heavy` | +6 | Armure de plates |
| `shield` | +1 | Bouclier |

### Système de Classe d'Armure

SkillsWeaver utilise **l'AC montante** (plus c'est haut, mieux c'est) :

```
AC = 11 (base) + modificateur DEX + bonus armure + bonus bouclier
```

**Exemples** :
- Sans armure, DEX 10 : AC = 11
- Cuir, DEX 10 : AC = 11 + 2 = 13
- Cotte de mailles, DEX 10 : AC = 11 + 4 = 15
- Plates + bouclier, DEX 10 : AC = 11 + 6 + 1 = 18

## Armes Disponibles (17)

### Armes de Mêlée
| ID | Nom FR | Dégâts | Coût | Propriétés |
|----|--------|--------|------|------------|
| `dagger` | Dague | 1d4 | 2 po | light, thrown |
| `short_sword` | Épée courte | 1d6 | 6 po | light |
| `longsword` | Épée longue | 1d8 | 10 po | - |
| `two_handed_sword` | Épée à deux mains | 1d10 | 18 po | two-handed |
| `battle_axe` | Hache de bataille | 1d8 | 7 po | - |
| `hand_axe` | Hachette | 1d6 | 4 po | light, thrown |
| `mace` | Masse d'armes | 1d8 | 6 po | - |
| `club` | Gourdin | 1d4 | 0 po | - |
| `staff` | Bâton | 1d6 | 2 po | two-handed |
| `hammer` | Marteau de guerre | 1d6 | 4 po | - |
| `spear` | Lance | 1d6 | 3 po | thrown, versatile |
| `polearm` | Arme d'hast | 1d10 | 9 po | two-handed, reach |

### Armes à Distance
| ID | Nom FR | Dégâts | Coût | Portée |
|----|--------|--------|------|--------|
| `shortbow` | Arc court | 1d6 | 25 po | 50/100 |
| `longbow` | Arc long | 1d8 | 60 po | 70/140 |
| `crossbow` | Arbalète | 1d8 | 30 po | 60/120 |
| `sling` | Fronde | 1d4 | 1 po | 30/60 |
| `dart` | Dard | 1d4 | 0.5 po | 15/30 |

## Armures Disponibles (4)

| ID | Nom FR | Bonus CA | Type | Coût | Poids |
|----|--------|----------|------|------|-------|
| `leather` | Armure de cuir | +2 | light | 20 po | 7.5 po |
| `chainmail` | Cotte de mailles | +4 | medium | 60 po | 25 po |
| `plate` | Armure de plates | +6 | heavy | 300 po | 35 po |
| `shield` | Bouclier | +1 | shield | 7 po | 5 po |

## Équipement d'Aventure (22 items)

Items essentiels : `backpack`, `bedroll`, `rope_50ft`, `torch`, `lantern`, `oil_flask`, `rations_iron`, `waterskin`, `thieves_tools`, `holy_symbol`, `holy_water`, `spellbook`, etc.

## Munitions (3)

| ID | Nom FR | Coût | Pour |
|----|--------|------|------|
| `arrows_20` | Flèches (20) | 2 po | Arc |
| `bolts_20` | Carreaux (20) | 2 po | Arbalète |
| `stones_20` | Pierres de fronde (20) | 0 po | Fronde |

## Format de Sortie

### Fiche Arme (Markdown)

```markdown
## Épée longue (Longsword)

| Caractéristique | Valeur |
|-----------------|--------|
| **Dégâts** | 1d8 |
| **Type** | melee |
| **Poids** | 2.0 po |
| **Coût** | 10 po |
```

### Format Court

```
Épée longue - 1d8, 10 po
Armure de cuir - CA +2, 20 po (light)
```

## Équipement de Départ par Classe

### Guerrier (Fighter)
- **Obligatoire** : Sac à dos, sac de couchage, rations, outre, corde, torche
- **Armes** : Épée longue + bouclier OU Épée à deux mains OU Hache + bouclier
- **Armure** : Cuir ou Cotte de mailles

### Clerc (Cleric)
- **Obligatoire** : Sac à dos, sac de couchage, rations, outre, symbole sacré, eau bénite
- **Armes** : Masse + bouclier OU Marteau + bouclier OU Bâton
- **Armure** : Cuir ou Cotte de mailles

### Magicien (Magic-user)
- **Obligatoire** : Sac à dos, sac de couchage, rations, outre, grimoire, plume et encre
- **Armes** : Dague OU Bâton OU 3 dards
- **Armure** : Aucune

### Voleur (Thief)
- **Obligatoire** : Sac à dos, sac de couchage, rations, outre, outils de voleur, corde
- **Armes** : Épée courte + dague OU Arc court + dague OU 2 dagues + fronde
- **Armure** : Cuir uniquement

## Intégration avec Character Generator

```bash
# Lors de la création d'un personnage
./sw-equipment starting fighter    # Voir les options d'équipement

# Vérifier un item spécifique
./sw-equipment show chainmail      # Stats de la cotte de mailles
```

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Vérification dégâts armes, CA armures |
| `rules-keeper` | Référence équipement pour arbitrage |
| `character-creator` | Équipement de départ |

**Type** : Skill autonome, peut être invoqué directement via `/equipment-browser`
