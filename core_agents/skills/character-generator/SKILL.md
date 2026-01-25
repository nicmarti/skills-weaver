---
name: character-generator
description: Crée des personnages D&D 5e. Génère caractéristiques (4d6kh3 ou répartition standard), applique modificateurs d'espèce, calcule bonus maîtrise, points de vie et or de départ. 9 espèces, 12 classes. Sauvegarde dans data/characters/.
allowed-tools:
  - Bash
  - Read
---

# Character Generator - Générateur de Personnages D&D 5e

Skill pour créer et gérer des personnages dans D&D 5e (5ème édition).

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-character ./cmd/character

# Créer un personnage
./sw-character create "Nom" --species=human --class=fighter
```

## Commandes Disponibles

### Créer un personnage

```bash
./sw-character create "Aldric" --species=human --class=fighter
./sw-character create "Lyra" --species=elf --class=wizard
./sw-character create "Gorim" --species=dwarf --class=cleric
./sw-character create "Zara" --species=dragonborn --class=paladin
./sw-character create "Finnian" --species=halfling --class=rogue

# Méthode classique (3d6, plus difficile)
./sw-character create "Sage" --species=human --class=wizard --method=classic

# Avec background
./sw-character create "Marcus" --species=human --class=fighter --background=soldier
```

### Gérer les personnages

```bash
./sw-character list              # Liste tous les personnages
./sw-character show "Aldric"     # Affiche la fiche complète
./sw-character delete "Aldric"   # Supprime un personnage
```

### Exporter

```bash
./sw-character export "Aldric" --format=json    # Export JSON
./sw-character export "Aldric" --format=md      # Export Markdown
```

## Espèces Disponibles (9)

| Espèce | ID | Modificateurs | Particularités |
|--------|-----|---------------|----------------|
| Humain | `human` | +1 toutes ou variante | Polyvalent |
| Drakéide | `dragonborn` | +2 FOR, +1 CHA | Souffle, résistance élémentaire |
| Elfe | `elf` | +2 DEX | Vision dans le noir, transe |
| Gnome | `gnome` | +2 INT | Vision dans le noir, résistance magie |
| Goliath | `goliath` | +2 FOR, +1 CON | Endurance de pierre |
| Halfelin | `halfling` | +2 DEX | Chanceux, brave |
| Nain | `dwarf` | +2 CON | Vision dans le noir, résistance |
| Orc | `orc` | +2 FOR, +1 CON | Vision dans le noir, endurance |
| Tieffelin | `tiefling` | +2 CHA, +1 INT | Vision dans le noir, magie infernale |

## Classes Disponibles (12)

| Classe | ID | Dé de Vie | Rôle | Sauvegardes |
|--------|-----|-----------|------|-------------|
| Barbare | `barbarian` | d12 | Tank, rage | FOR, CON |
| Barde | `bard` | d8 | Support, magie | DEX, CHA |
| Clerc | `cleric` | d8 | Soins, divine | WIS, CHA |
| Druide | `druid` | d8 | Magie nature | INT, WIS |
| Ensorceleur | `sorcerer` | d6 | Magie innée | CON, CHA |
| Guerrier | `fighter` | d10 | Combat | FOR, CON |
| Magicien | `wizard` | d6 | Magie arcanique | INT, WIS |
| Moine | `monk` | d8 | Arts martiaux | FOR, DEX |
| Occultiste | `warlock` | d8 | Magie pacte | WIS, CHA |
| Paladin | `paladin` | d10 | Combat sacré | WIS, CHA |
| Rôdeur | `ranger` | d10 | Pistage, nature | FOR, DEX |
| Roublard | `rogue` | d8 | Discrétion | DEX, INT |

## Backgrounds Disponibles

| Background | ID | Compétences |
|------------|-----|-------------|
| Acolyte | `acolyte` | Intuition, Religion |
| Criminel | `criminal` | Tromperie, Discrétion |
| Érudit | `sage` | Arcanes, Histoire |
| Héros du Peuple | `folk-hero` | Dressage, Survie |
| Noble | `noble` | Histoire, Persuasion |
| Artisan | `guild-artisan` | Intuition, Persuasion |
| Soldat | `soldier` | Athlétisme, Intimidation |

## Pas de Restrictions Espèce/Classe

**D&D 5e** : Toutes les combinaisons espèce/classe sont valides. Pas de limite de niveau par espèce.

## Processus de Création

1. **Génération des caractéristiques** :
   - Standard : 4d6kh3 (×6)
   - Classic : 3d6 (×6, plus difficile)
   - Répartition standard : 15, 14, 13, 12, 10, 8
2. **Application des modificateurs d'espèce** : Bonus selon l'espèce
3. **Calcul des modificateurs** : `(Score - 10) ÷ 2` (arrondi vers le bas)
4. **Bonus de maîtrise** : +2 au niveau 1
5. **Points de vie** : Dé de classe max + modificateur CON
6. **Compétences** : Classe (2-4) + Background (2)
7. **Or de départ** : Selon classe
8. **Sauvegarde** : Fichier JSON dans `data/characters/`

## Formule des Modificateurs D&D 5e

```
Modificateur = (Score - 10) ÷ 2 (arrondi vers le bas)
```

| Score | Modificateur | Score | Modificateur |
|-------|-------------|-------|-------------|
| 1 | -5 | 10-11 | 0 |
| 2-3 | -4 | 12-13 | +1 |
| 4-5 | -3 | 14-15 | +2 |
| 6-7 | -2 | 16-17 | +3 |
| 8-9 | -1 | 18-19 | +4 |

## Bonus de Maîtrise par Niveau

| Niveau | Bonus | Niveau | Bonus |
|--------|-------|--------|-------|
| 1-4 | +2 | 13-16 | +5 |
| 5-8 | +3 | 17-20 | +6 |
| 9-12 | +4 | | |

## Exemples de Résultats

### Création d'un guerrier humain

```
## Création de Aldric

### Génération des caractéristiques (4d6kh3)

| Caractéristique | Jets | Total | Modificateur |
|-----------------|------|-------|--------------|
| Force           | 6, ~~1~~, 5, 4 | **15** | +2 |
| Intelligence    | ~~2~~, 3, 6, 4 | **13** | +1 |
| Sagesse         | 5, ~~2~~, 6, 3 | **14** | +2 |
| Dextérité       | 4, 4, ~~1~~, 6 | **14** | +2 |
| Constitution    | 6, 5, ~~3~~, 4 | **15** | +2 |
| Charisme        | 3, 4, ~~2~~, 5 | **12** | +1 |

### Modificateurs d'espèce (Humain)

+1 à toutes les caractéristiques → FOR 16 (+3), INT 14 (+2), etc.

### Points de vie (niveau 1, d10 max)

PV = 10 (dé max) + 3 (CON) = **13**

### Bonus de maîtrise

**+2** au niveau 1

### Or de départ

**150 po** (5d4×10 pour guerrier)
```

## Fichiers de Sortie

Les personnages sont sauvegardés en JSON dans `data/characters/` :

```json
{
  "id": "uuid",
  "name": "Aldric",
  "species": "human",
  "class": "fighter",
  "level": 1,
  "proficiency_bonus": 2,
  "background": "soldier",
  "abilities": {
    "strength": 16,
    "dexterity": 14,
    "constitution": 15,
    "intelligence": 14,
    "wisdom": 14,
    "charisma": 12
  },
  "modifiers": {
    "strength": 3,
    "dexterity": 2,
    "constitution": 2,
    "intelligence": 2,
    "wisdom": 2,
    "charisma": 1
  },
  "hit_points": 13,
  "max_hit_points": 13,
  "armor_class": 10,
  "gold": 150,
  "skills": {
    "athletics": true,
    "intimidation": true
  },
  "saving_throw_profs": {
    "strength": true,
    "constitution": true
  }
}
```

## Conseils d'Utilisation

- **Aucune restriction** espèce/classe en D&D 5e
- Utilisez `--method=classic` pour génération 3d6 (plus difficile)
- La skill `dice-roller` peut être utilisée pour jets supplémentaires
- Les compétences sont automatiquement assignées selon classe + background
- Bonus de maîtrise augmente avec le niveau (+2 → +6)

## Différences vs BFRPG

| Aspect | BFRPG | D&D 5e |
|--------|-------|--------|
| Espèces | 4 races | 9 espèces |
| Classes | 4 classes | 12 classes |
| Restrictions | Limites niveau/race | Aucune |
| Modificateurs | Table lookup | Formule (Score-10)÷2 |
| Compétences | Aucune | 18 formelles |
| Bonus maîtrise | Par caractéristique | +2 à +6 (niveau) |

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `character-creator` | Création guidée de personnages |

**Type** : Skill autonome, peut être invoqué directement via `/character-generator`

**Dépendances** : Utilise `dice-roller` pour la génération des caractéristiques