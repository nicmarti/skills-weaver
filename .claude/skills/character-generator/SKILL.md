---
name: character-generator
description: Crée des personnages Basic Fantasy RPG. Génère caractéristiques (4d6kh3 ou 3d6), applique bonus raciaux, calcule modificateurs, points de vie et or de départ. Sauvegarde dans data/characters/. Utilisez pour créer un nouveau personnage joueur.
allowed-tools: Bash, Read
---

# Character Generator - Générateur de Personnages BFRPG

Skill pour créer et gérer des personnages dans Basic Fantasy RPG.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o character ./cmd/character

# Créer un personnage
./character create "Nom" --race=human --class=fighter
```

## Commandes Disponibles

### Créer un personnage

```bash
./character create "Aldric" --race=human --class=fighter
./character create "Lyra" --race=elf --class=magic-user
./character create "Gorim" --race=dwarf --class=cleric
./character create "Pip" --race=halfling --class=thief

# Méthode classique (3d6)
./character create "Vieux Sage" --race=human --class=magic-user --method=classic
```

### Gérer les personnages

```bash
./character list              # Liste tous les personnages
./character show "Aldric"     # Affiche la fiche complète
./character delete "Aldric"   # Supprime un personnage
```

### Exporter

```bash
./character export "Aldric" --format=json    # Export JSON
./character export "Aldric" --format=md      # Export Markdown
```

## Races Disponibles

| Race | Modificateurs | Classes Autorisées | Niveau Max |
|------|--------------|-------------------|------------|
| `human` | Aucun | Toutes | Illimité |
| `elf` | +1 DEX, -1 CON | Guerrier, Magicien, Voleur | 6/9/∞ |
| `dwarf` | +1 CON, -1 CHA | Guerrier, Clerc, Voleur | 7/6/∞ |
| `halfling` | +1 DEX, -1 FOR | Guerrier, Voleur | 4/∞ |

## Classes Disponibles

| Classe | ID | Dé de Vie | Armes | Armures |
|--------|-----|-----------|-------|---------|
| Guerrier | `fighter` | d8 | Toutes | Toutes |
| Clerc | `cleric` | d6 | Contondantes | Toutes |
| Magicien | `magic-user` | d4 | Dague, bâton | Aucune |
| Voleur | `thief` | d4 | Toutes | Cuir |

## Combinaisons Valides

### Humain
- fighter, cleric, magic-user, thief (toutes)

### Elfe
- fighter (max niveau 6)
- magic-user (max niveau 9)
- thief (illimité)

### Nain
- fighter (max niveau 7)
- cleric (max niveau 6)
- thief (illimité)

### Halfelin
- fighter (max niveau 4)
- thief (illimité)

## Processus de Création

1. **Génération des caractéristiques** : 4d6kh3 (standard) ou 3d6 (classic)
2. **Application des modificateurs raciaux** : bonus/malus selon la race
3. **Calcul des modificateurs** : -3 à +3 selon le score
4. **Points de vie** : Dé de classe max + modificateur CON
5. **Or de départ** : 3d6×10 po (ou 2d6×10 pour voleur)
6. **Sauvegarde** : Fichier JSON dans `data/characters/`

## Table des Modificateurs

| Score | Modificateur |
|-------|-------------|
| 3 | -3 |
| 4-5 | -2 |
| 6-8 | -1 |
| 9-12 | 0 |
| 13-15 | +1 |
| 16-17 | +2 |
| 18 | +3 |

## Exemples de Résultats

### Création d'un guerrier humain

```
## Création de Aldric

### Génération des caractéristiques

| Caractéristique | Jets | Total |
|-----------------|------|-------|
| Force           | 6, ~~1~~, 5, 4 | **15** |
| Intelligence    | ~~2~~, 3, 6, 4 | **13** |
...

### Points de vie (niveau 1, d8 max)

PV = 8 (dé max) + 1 (CON) = **9**

### Or de départ

**120 po**
```

## Fichiers de Sortie

Les personnages sont sauvegardés en JSON dans `data/characters/` :

```json
{
  "id": "uuid",
  "name": "Aldric",
  "race": "human",
  "class": "fighter",
  "level": 1,
  "abilities": {
    "strength": 15,
    "intelligence": 13,
    ...
  },
  "hit_points": 9,
  "gold": 120
}
```

## Conseils d'Utilisation

- Utilisez `--method=classic` pour une génération old-school plus difficile
- La skill `dice-roller` peut être utilisée pour des jets supplémentaires
- Vérifiez les combinaisons race/classe avant de créer
- Les personnages sont automatiquement sauvegardés