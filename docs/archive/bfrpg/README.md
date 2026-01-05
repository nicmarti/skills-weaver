# Basic Fantasy RPG - Règles de Référence

Ce répertoire contient les règles officielles de Basic Fantasy RPG (r142) extraites et organisées par sections pour consultation rapide par l'agent `rules-keeper`.

## Structure

Les règles sont organisées en 8 fichiers thématiques :

| Fichier | Contenu | Usage principal |
|---------|---------|-----------------|
| `01-character-creation.md` | Races, classes, caractéristiques, création de personnage | character-creator, rules-keeper |
| `02-combat.md` | Initiative, attaque, dégâts, CA, manœuvres de combat | dungeon-master, rules-keeper |
| `03-magic.md` | Sorts, incantation, préparation, listes de sorts par classe | dungeon-master, rules-keeper |
| `04-equipment.md` | Armes, armures, équipement d'aventure, encombrement | character-creator, rules-keeper |
| `05-monsters.md` | Règles créatures, types, DV, trésors, XP | dungeon-master |
| `06-treasure.md` | Tables de trésors, objets magiques, gemmes, bijoux | dungeon-master |
| `07-advancement.md` | Tables XP, progression des classes, capacités par niveau | rules-keeper |
| `08-conditions.md` | États (empoisonné, paralysé, etc.), guérison, repos | dungeon-master, rules-keeper |

## Consultation

L'agent `rules-keeper` utilise ces fichiers via ses tools (`Read`, `Grep`, `Glob`) pour répondre aux questions de règles.

**Exemple de consultation** :
```bash
# Recherche d'une règle spécifique
grep -i "initiative" data/rules/02-combat.md

# Lecture complète d'une section
cat data/rules/03-magic.md
```

## Source

Ces règles sont extraites de **Basic Fantasy RPG Core Rules Release 142** (docs/Basic-Fantasy-RPG-Rules-r142.pdf).

**Licence** : Open Game License (OGL)
**Site officiel** : https://basicfantasy.org/