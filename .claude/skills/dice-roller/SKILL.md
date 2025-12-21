---
name: dice-roller
description: Lance des dés pour jeux de rôle (Basic Fantasy RPG). Supporte d4, d6, d8, d10, d12, d20, d100. Notation standard comme 2d6+3, 4d6kh3 (keep highest). Avantage et désavantage. Utilisez pour tout jet de dé en session de JdR.
allowed-tools: Bash
---

# Dice Roller - Lanceur de Dés

Skill pour lancer des dés dans le cadre de sessions de jeu de rôle Basic Fantasy RPG.

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
```

### Keep Highest / Keep Lowest

```bash
./sw-dice roll 4d6kh3        # Lance 4d6, garde les 3 plus hauts (génération de stats)
./sw-dice roll 2d20kl1       # Lance 2d20, garde le plus bas (désavantage)
```

### Avantage et Désavantage

```bash
./sw-dice roll d20 --advantage      # ou -a : Lance 2d20, garde le plus haut
./sw-dice roll d20 --disadvantage   # ou -d : Lance 2d20, garde le plus bas
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

### Génération de stats
```
## Génération de caractéristiques (4d6kh3 (standard))

| Caractéristique | Jets | Total |
|-----------------|------|-------|
| Force           | 6, ~~1~~, 5, 2 | **13** |
| Intelligence    | ~~3~~, 4, 4, 5 | **13** |
...
```

## Types de Dés Supportés

| Dé | Usage BFRPG |
|----|-------------|
| d4 | Dégâts dague, bâton |
| d6 | Dégâts épée courte, PV Clerc |
| d8 | Dégâts épée longue, PV Guerrier |
| d10 | Dégâts hallebarde |
| d12 | Dégâts hache à deux mains |
| d20 | Jets d'attaque, sauvegardes, tests |
| d100 | Tables aléatoires, rencontres |

## Conseils

- Pour les jets de combat: `./sw-dice roll d20+<bonus>`
- Pour les dégâts: utilisez le dé approprié à l'arme
- Pour les sauvegardes: `./sw-dice roll d20` et comparez au seuil
- Pour créer un personnage: `./sw-dice stats`

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Jets de combat, initiative, sauvegardes |
| `character-creator` | Génération des caractéristiques |
| `rules-keeper` | Vérification et exécution des jets |

**Type** : Skill autonome, peut être invoqué directement via `/dice-roller`