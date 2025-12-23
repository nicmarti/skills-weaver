---
name: spell-reference
description: Consulte les sorts BFRPG par classe et niveau. Portée, durée, effets, formes inversées. Utilisez pour vérifier les sorts lancés.
allowed-tools: Bash
---

# Spell Reference - Grimoire des Sorts BFRPG

Skill pour consulter les sorts divins (Clerc) et arcaniques (Magicien). Indispensable pour vérifier les effets des sorts pendant le jeu.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-spell ./cmd/spell

# Lister les sorts par classe
./sw-spell list --class=cleric

# Détails d'un sort
./sw-spell show magic_missile

# Sorts réversibles
./sw-spell reversible
```

## Commandes Disponibles

### Lister les Sorts

```bash
./sw-spell list                              # Tous les sorts
./sw-spell list --class=cleric               # Sorts de Clerc
./sw-spell list --class=magic-user           # Sorts de Magicien
./sw-spell list --class=cleric --level=1     # Clerc niveau 1
./sw-spell list --level=2                    # Tous les sorts niveau 2
./sw-spell list --format=md                  # Fiches détaillées
./sw-spell list --format=json                # Format JSON
```

### Afficher un Sort

```bash
./sw-spell show <id>

# Exemples:
./sw-spell show magic_missile        # Projectile magique
./sw-spell show cure_light_wounds    # Soins légers
./sw-spell show sleep                # Sommeil
./sw-spell show --format=json        # Format JSON
./sw-spell show --format=short       # Une ligne
```

### Rechercher des Sorts

```bash
./sw-spell search <terme>

# Exemples:
./sw-spell search lumière            # Par nom français
./sw-spell search light              # Par nom anglais
./sw-spell search détection          # Sorts de détection
```

### Sorts Réversibles

```bash
./sw-spell reversible                # Liste les sorts avec forme inversée
./sw-spell reversible --format=md    # Fiches détaillées
```

## Types de Sorts

| Type | Classe | Description |
|------|--------|-------------|
| `divine` | Clerc | Sorts obtenus par la prière |
| `arcane` | Magicien | Sorts appris par étude |
| `both` | Les deux | Disponibles aux deux classes |

## Sorts de Clerc (Divins)

### Niveau 1 (8 sorts)

| Sort | Portée | Durée | Effet |
|------|--------|-------|-------|
| Soins légers* | Contact | Instantané | Soigne 1d6+1 PV |
| Détection du mal* | 60' | 1 round/niveau | Détecte le mal |
| Détection de la magie | 60' | 2 tours | Détecte la magie |
| Lumière* | 120' | 6 tours+1/niveau | Éclaire 30' rayon |
| Protection contre le mal* | Contact | 1 tour/niveau | +2 CA et JS contre mal |
| Purification nourriture | 10' | Instantané | Purifie vivres |
| Délivrance de la peur* | Contact | Instantané | Annule la peur |
| Résistance au froid | Contact | 1 round/niveau | +3 JS, -50% dégâts froid |

### Niveau 2 (8 sorts)

| Sort | Portée | Durée | Effet |
|------|--------|-------|-------|
| Bénédiction* | 50' rayon | 1 min/niveau | +1 attaque, moral, JS |
| Charme-animal | 60' | niveau+1d4 rounds | Charme animaux |
| Détection des pièges | 30' | 3 tours | Révèle les pièges |
| Immobilisation de personne | 180' | 2d8 tours | Paralyse humanoïdes |
| Résistance au feu | Contact | 1 round/niveau | +3 JS, -50% dégâts feu |
| Silence 15' de rayon | 360' | 2 rounds/niveau | Zone de silence |
| Communication avec animaux | Spécial | 1 tour/4 niveaux | Parle aux animaux |
| Marteau spirituel | 30' | 1 round/niveau | 1d6+1/3 niveaux dégâts |

## Sorts de Magicien (Arcaniques)

### Niveau 1 (13 sorts)

| Sort | Portée | Durée | Effet |
|------|--------|-------|-------|
| Charme-personne | 30' | Spécial | Charme humanoïde 4 DV max |
| Détection de la magie | 60' | 2 tours | Détecte la magie |
| Disque flottant | 0 | 5 tours+1/niveau | Porte 500 livres |
| Verrouillage | 100'+10'/niveau | 1 round/niveau | Verrouille porte |
| Lumière* | 120' | 6 tours+1/niveau | Éclaire 30' rayon |
| Projectile magique | 100'+10'/niveau | Instantané | 1d6+1 auto-touche |
| Bouche magique | 30' | Spécial | Message déclenché |
| Protection contre le mal* | Contact | 1 tour/niveau | +2 CA et JS |
| Lecture des langues | 0 | Spécial | Lit les langues |
| Lecture de la magie | 0 | Permanent | Lit textes magiques |
| Bouclier | Soi | 5 rounds+1/niveau | +3/+6 CA, annule projectile magique |
| Sommeil | 90' | 5 rounds/niveau | Endort créatures 3 DV max |
| Ventriloquie | 60' | 1 tour/niveau | Projette sa voix |

### Niveau 2 (12 sorts)

| Sort | Portée | Durée | Effet |
|------|--------|-------|-------|
| Lumière éternelle* | 360' | 1 an/niveau | Éclaire comme le jour |
| Détection du mal* | 60' | 1 round/niveau | Détecte le mal |
| Détection de l'invisible | 60' | 1 tour/niveau | Voit l'invisible |
| Invisibilité | Contact | Spécial | Rend invisible |
| Déblocage | 30' | Spécial | Ouvre portes verrouillées |
| Lévitation | Contact | 1 tour/niveau | Monte/descend 20'/round |
| Localisation d'objet | 360' | 1 round/niveau | Trouve un objet connu |
| Lecture des pensées | 60' | 1 tour/niveau | Lit pensées superficielles |
| Image miroir | Soi | 1 tour/niveau | 1d4+niveau/3 doubles |
| Force fantasmagorique | 180' | Concentration | Crée illusion visuelle |
| Toile d'araignée | 10'/niveau | 2 tours/niveau | Emprisonne créatures |
| Verrou magique | 20' | Permanent | Verrouille magiquement |

## Sorts Réversibles

Les sorts marqués d'un astérisque (*) ont une forme inversée :

| Sort | Forme inversée |
|------|----------------|
| Soins légers | Blessure légère |
| Détection du mal | Détection du bien |
| Lumière | Ténèbres |
| Protection contre le mal | Protection contre le bien |
| Délivrance de la peur | Cause de la peur |
| Bénédiction | Fléau |
| Lumière éternelle | Ténèbres éternelles |

**Règle** : Un sort réversible peut être préparé normalement et utilisé dans l'une ou l'autre forme.

## Format de Sortie

### Fiche Sort (Markdown)

```markdown
## Projectile magique (Magic Missile)

**Type** : Arcanique (Magicien) | **Niveau** : 1

| Caractéristique | Valeur |
|-----------------|--------|
| **Portée** | 100'+10'/niveau |
| **Durée** | instantaneous |
| **Dégâts** | 1d6+1 par projectile |

### Description

Ce sort fait jaillir un projectile d'énergie magique du bout du doigt du lanceur
pour frapper sa cible, infligeant 1d6+1 points de dégâts. Le projectile frappe
immanquablement. Pour chaque trois niveaux au-delà du 1er, un projectile
supplémentaire est tiré.
```

### Format Court

```
Projectile magique [N1 Arc] - 100'+10'/niveau, instantaneous
Soins légers * [N1 Div] - touch, instantaneous
```

## Emplacements de Sorts par Niveau

### Magicien

| Niveau perso | 1er | 2e | 3e |
|--------------|-----|----|----|
| 1 | 1 | - | - |
| 2 | 2 | - | - |
| 3 | 2 | 1 | - |
| 4 | 2 | 2 | - |
| 5 | 2 | 2 | 1 |

### Clerc

| Niveau perso | 1er | 2e | 3e |
|--------------|-----|----|----|
| 1 | - | - | - |
| 2 | 1 | - | - |
| 3 | 2 | - | - |
| 4 | 2 | 1 | - |
| 5 | 2 | 2 | - |

## Règles de Lancement

1. **Main libre** et **parole** nécessaires
2. **Durée** : comme une attaque
3. **Interruption** : si attaqué ou JS requis pendant incantation → sort perdu
4. **Préparation** : Magicien après repos (1 tour/3 niveaux), Clerc après prière (3 tours)

## Intégration avec Adventure Manager

```bash
# Vérifier un sort avant de l'utiliser
./sw-spell show sleep

# Logger le sort dans le journal
./sw-adventure log "Mon Aventure" combat "Lyra lance Sommeil sur 4 gobelins"
```

## Conseils d'Utilisation

### Pour le Dungeon Master

```bash
# Vérifier rapidement un sort ennemi
./sw-spell show hold_person

# Trouver des sorts de zone
./sw-spell search radius
```

### Pour vérifier un personnage

```bash
# Sorts disponibles pour un magicien niveau 1
./sw-spell list --class=magic-user --level=1

# Sorts de clerc niveau 2
./sw-spell list --class=cleric --level=2
```

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Vérification effets des sorts lancés |
| `rules-keeper` | Référence des sorts pour arbitrage |
| `character-creator` | Choix des sorts initiaux |

**Type** : Skill autonome, peut être invoqué directement via `/spell-reference`
