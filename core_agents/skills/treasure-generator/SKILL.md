---
name: treasure-generator
description: Génère des trésors BFRPG par type (A-U). Pièces, gemmes, bijoux et objets magiques. Utilise les tables officielles avec probabilités. Indispensable après un combat victorieux.
allowed-tools:
  - Bash
---

# Treasure Generator - Générateur de Trésors BFRPG

Skill pour générer des trésors aléatoires selon les tables officielles de Basic Fantasy RPG. Chaque monstre a un type de trésor assigné (visible dans le bestiaire).

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-treasure ./cmd/treasure

# Générer un trésor
./sw-treasure generate R           # Type R (Gobelin)
./sw-treasure generate A           # Type A (Dragon)
./sw-treasure generate B --count=3 # 3 trésors type B
```

## Commandes Disponibles

### Générer un Trésor

```bash
./sw-treasure generate <type>
./sw-treasure gen <type>           # Alias

# Exemples:
./sw-treasure generate R           # Trésor de gobelin
./sw-treasure generate A           # Trésor de dragon
./sw-treasure generate H           # Trésor de dragon ancien
./sw-treasure generate B --count=5 # 5 trésors d'ogre
./sw-treasure generate A --format=json
```

### Lister les Types

```bash
./sw-treasure types
```

### Détails d'un Type

```bash
./sw-treasure info <type>

# Exemples:
./sw-treasure info A               # Probabilités du type A
./sw-treasure info R               # Probabilités du type R
```

### Lister les Objets Magiques

```bash
./sw-treasure items                # Liste les catégories
./sw-treasure items potions        # Toutes les potions
./sw-treasure items weapons        # Armes magiques
./sw-treasure items armor          # Armures magiques
./sw-treasure items rings          # Anneaux
./sw-treasure items wands          # Baguettes
./sw-treasure items scrolls        # Parchemins
./sw-treasure items misc           # Objets divers
```

## Types de Trésors

### Trésors de Repaire (A-H)
Pour les repaires de groupes de monstres.

| Type | Description | Exemple |
|------|-------------|---------|
| A | Trésor majeur | Dragon |
| B | Trésor moyen | Ogre |
| C | Trésor petit | Orc |
| D | Trésor standard | Hobgobelin |
| E | Trésor standard | Gnoll |
| F | Trésor riche | Vampire |
| G | Trésor précieux | Nain |
| H | Trésor énorme | Dragon ancien |

### Trésors Individuels (I-U)
Portés par des créatures individuelles.

| Type | Description | Exemple |
|------|-------------|---------|
| I | Individuel riche | Garde élite |
| J-K | Individuel mineur | Bandit |
| L-M | Gemmes/bijoux | Noble |
| N | Potions | Alchimiste |
| O | Parchemins | Mage |
| P-Q | Pauvre | Paysan |
| R | Standard | Gobelin |
| S-T | Moyen | Mercenaire |
| U | Unique | Cas spécial |

## Correspondance Monstres → Types

| Monstre | Type Trésor |
|---------|-------------|
| Gobelin, Kobold | R |
| Orc, Hobgobelin | C, D |
| Gnoll, Bugbear | E |
| Ogre | B |
| Troll | D |
| Vampire | F |
| Dragon jeune | A |
| Dragon ancien | H |
| Squelette, Zombie | - (aucun) |
| Goule, Wight | B |

## Format de Sortie

### Trésor Généré (Markdown)

```markdown
## Trésor (Type R)

*Trésor individuel (ex: Gobelin)*

### Pièces

- **Pièces de cuivre** : 7 (0 po)
- **Pièces d'argent** : 4 (0 po)

**Valeur totale estimée** : 0 po
```

### Trésor Riche

```markdown
## Trésor (Type A)

*Trésor de repaire majeur (ex: Dragon)*

### Pièces

- **Pièces d'or** : 7000 (7000 po)
- **Pièces de platine** : 1000 (5000 po)

### Gemmes

- Rubis (1000 po)
- Émeraude (1000 po)

### Bijoux

- Couronne (3000 po)
- Sceptre orné (7200 po)

### Objets Magiques

- **Épée +2**
- **Potion de soins** - Restaure 1d6+1 PV
- **Anneau de protection +1** - +1 CA et sauvegardes

**Valeur totale estimée** : 59697 po
```

## Intégration avec Adventure Manager

```bash
# Après un combat victorieux
./sw-treasure generate R

# Logger le butin
./sw-adventure log "Mon Aventure" loot "Trésor du gobelin: 7 pc, 4 pa"

# Ajouter l'or au groupe
./sw-adventure add-gold "Mon Aventure" 5 "Trésor gobelin"
```

## Workflow Type en Session

```bash
# 1. Consulter le type de trésor du monstre
./sw-monster show goblin
# → Trésor: R

# 2. Générer le trésor après victoire
./sw-treasure generate R

# 3. Pour plusieurs monstres
./sw-treasure generate R --count=4

# 4. Pour un boss
./sw-treasure generate A
```

## Objets Magiques Disponibles

### Potions (10)
Soins, Vitesse, Invisibilité, Force de géant, Vol, Forme gazeuse, Croissance, Diminution, Poison

### Armes (11)
Épées +1/+2/+3, Épée ardente, Épée de givre, Dagues +1/+2, Hache +1, Masse +1, Arc +1, Flèches +1

### Armures (7)
Cuir +1, Cotte de mailles +1/+2, Plates +1/+2, Bouclier +1/+2

### Anneaux (5)
Protection +1/+2, Invisibilité, Résistance au feu, Stockage de sorts

### Baguettes (5)
Projectiles magiques, Paralysie, Froid, Éclairs, Peur

### Objets Divers (10)
Sac sans fond, Bottes de vitesse/lévitation, Cape de déplacement/protection, Gantelets de force, Casque de télépathie, etc.

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Génération de trésors après combats |

**Type** : Skill autonome, peut être invoqué directement via `/treasure-generator`

**Dépendances** : Complète `monster-manual` (type de trésor par monstre)