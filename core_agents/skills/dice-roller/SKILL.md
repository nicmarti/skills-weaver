---
name: dice-roller
description: Lance des dés pour jeux de rôle (D&D 5e). Supporte d4, d6, d8, d10, d12, d20, d100. Notation standard comme 2d6+3, 4d6kh3 (keep highest). Avantage et désavantage. Utilisez pour tout jet de dé en session de JdR.
allowed-tools: Bash
---

# Dice Roller - Lanceur de Dés

Skill pour lancer des dés dans le cadre de sessions de jeu de rôle D&D 5e.

## Utilisation Rapide

Exécutez la CLI `dice` depuis le répertoire du projet :

```bash
# Compiler si nécessaire
go build -o sw-dice ./cmd/dice

# Lancer des dés
./sw-dice roll d20
./sw-dice roll 2d6+3
./sw-dice roll 4d6kh3
```

## Commandes Disponibles

### Lancer un dé simple

```bash
./sw-dice roll d20           # Lance 1d20
./sw-dice roll 2d6           # Lance 2d6
./sw-dice roll d100          # Lance 1d100 (percentile)
```

### Avec modificateur

```bash
./sw-dice roll 2d6+3         # Lance 2d6 et ajoute 3
./sw-dice roll d20-2         # Lance d20 et soustrait 2
./sw-dice roll d20+5         # Attaque avec bonus +5
```

### Keep Highest / Keep Lowest

```bash
./sw-dice roll 4d6kh3        # Lance 4d6, garde les 3 plus hauts (génération de stats)
./sw-dice roll 2d20kl1       # Lance 2d20, garde le plus bas (désavantage)
```

### Avantage et Désavantage (D&D 5e)

```bash
./sw-dice roll d20 --advantage      # ou -a : Lance 2d20, garde le plus haut
./sw-dice roll d20 --disadvantage   # ou -d : Lance 2d20, garde le plus bas

# Exemples d'utilisation:
./sw-dice roll d20 -a               # Jet d'attaque avec avantage
./sw-dice roll d20 -d               # Jet de sauvegarde avec désavantage
```

### Génération de Caractéristiques

```bash
./sw-dice stats              # Méthode standard : 4d6kh3 × 6
./sw-dice stats --classic    # Méthode classique : 3d6 × 6
```

## Exemples de Résultats

### Jet simple
```
2d6+3: [4, 5] + 3 = 12
```

### Jet avec dés éliminés
```
4d6kh3: [~2~, 5, 4, 6] = 15
```
Les dés barrés (`~2~`) sont ceux qui n'ont pas été gardés.

### Avantage (D&D 5e)
```
d20 (avantage): [15, ~8~] = 15
```

### Désavantage (D&D 5e)
```
d20 (désavantage): [~15~, 8] = 8
```

### Génération de stats
```
## Génération de caractéristiques (4d6kh3 (standard))

| Caractéristique | Jets | Total |
|-----------------|------|-------|
| Force           | 6, ~~1~~, 5, 2 | **13** |
| Intelligence    | ~~3~~, 4, 4, 5 | **13** |
| Sagesse         | 5, 4, ~~2~~, 6 | **15** |
| Dextérité       | ~~3~~, 5, 6, 4 | **15** |
| Constitution    | 4, 5, 6, ~~2~~ | **15** |
| Charisme        | 4, 4, ~~1~~, 3 | **11** |
```

## Types de Dés Supportés

| Dé | Usage D&D 5e |
|----|--------------|
| d4 | Dégâts dague, dard ; PV Magicien (d6 en D&D 5e) |
| d6 | Dégâts épée courte, arc court ; PV Ensorceleur/Magicien |
| d8 | Dégâts épée longue, rapière ; PV Barde/Clerc/Druide/Moine/Roublard/Occultiste |
| d10 | Dégâts pique, arbalète lourde ; PV Guerrier/Paladin/Rôdeur |
| d12 | Dégâts hache à deux mains, hache d'armes ; PV Barbare |
| d20 | Jets d'attaque, sauvegardes, tests de caractéristique, initiative |
| d100 | Tables aléatoires, objets magiques, effets de magie sauvage |

## Cas d'Usage D&D 5e

### Initiative (d20 + modificateur DEX)
```bash
# Pour un personnage avec DEX +2
./sw-dice roll d20+2
```

### Jets d'Attaque
```bash
# Attaque au corps à corps : d20 + bonus de maîtrise + mod FOR/DEX
./sw-dice roll d20+5         # Niveau 1-4, FOR +3, prof +2

# Avec avantage
./sw-dice roll d20+5 -a

# Avec désavantage
./sw-dice roll d20+5 -d
```

### Dégâts
```bash
# Épée longue : 1d8 + mod FOR
./sw-dice roll d8+3

# Attaque critique (double les dés)
./sw-dice roll 2d8+3

# Attaque sournoise du Roublard (1d6 supplémentaire niveau 1)
./sw-dice roll d8+3          # Dégâts de l'arme
./sw-dice roll d6            # Dégâts sournois
```

### Tests de Caractéristique
```bash
# Test de Force (soulever une porte) : d20 + mod FOR + prof (si compétence)
./sw-dice roll d20+1         # FOR 12 (+1)
./sw-dice roll d20+5         # FOR 16 (+3) + Athlétisme (+2)
```

### Jets de Sauvegarde
```bash
# Sauvegarde de Dextérité : d20 + mod DEX + prof (si maîtrisé)
./sw-dice roll d20+2         # DEX +2, pas de maîtrise
./sw-dice roll d20+4         # DEX +2, maîtrise +2
```

### DD de Sorts (Difficulté fixe, pas de jet)
```
DD = 8 + bonus de maîtrise + modificateur de caractéristique d'incantation
```

## Conseils

### Combat
- **Jets d'attaque** : `./sw-dice roll d20+<bonus attaque>`
- **Dégâts** : Utilisez le dé approprié à l'arme + modificateur de caractéristique
- **Initiative** : `./sw-dice roll d20+<mod DEX>` (pas d6 comme en BFRPG)
- **Critiques** : Sur un 20 naturel, doublez les dés de dégâts (pas les modificateurs)

### Exploration
- **Tests de caractéristique** : `./sw-dice roll d20+<mod>+<prof si applicable>`
- **Perception passive** : 10 + mod WIS + prof Perception (pas de jet)

### Création de Personnage
- **Caractéristiques** : `./sw-dice stats` (méthode standard 4d6kh3)
- **Points de vie** : Dé de classe max au niveau 1, puis moyenne ou jet

### Avantage/Désavantage
- **Avantage** : Circonstances favorables (action Aide, surprise, etc.)
- **Désavantage** : Circonstances défavorables (attaquer une cible invisible, terrain difficile, etc.)
- **Important** : Avantage et désavantage s'annulent, même si multiples

## Différences D&D 5e vs BFRPG

| Aspect | BFRPG | D&D 5e |
|--------|-------|--------|
| Initiative | 1d6 + DEX | 1d20 + DEX |
| Avantage/Désavantage | Modificateurs | 2d20 (garde meilleur/pire) |
| Tests caractéristique | d20 | d20 + mod + prof (si compétence) |
| DD par défaut | 15 | 10 (facile) à 30 (quasi impossible) |
| Critiques | ×2 au total | ×2 aux dés seulement |

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Jets de combat, initiative, sauvegardes, tests |
| `character-creator` | Génération des caractéristiques |
| `rules-keeper` | Vérification et exécution des jets |

**Type** : Skill autonome, peut être invoqué directement via `/dice-roller`