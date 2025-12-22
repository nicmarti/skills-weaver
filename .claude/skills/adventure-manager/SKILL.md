---
name: adventure-manager
description: Gère les aventures/campagnes BFRPG. Crée et charge des aventures, gère le groupe de personnages, l'inventaire partagé, les sessions de jeu et le journal automatique. Utilisez pour toute gestion de campagne.
allowed-tools: Bash, Read
---

# Adventure Manager - Gestionnaire d'Aventures BFRPG

Skill pour créer et gérer des aventures/campagnes dans Basic Fantasy RPG.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-adventure ./cmd/adventure

# Créer une aventure
./sw-adventure create "La Mine Perdue" "Une aventure dans les montagnes"
```

## Commandes Disponibles

### Gestion des Aventures

```bash
# Créer une aventure
./sw-adventure create "La Mine Perdue" "Description optionnelle"

# Lister les aventures
./sw-adventure list

# Afficher une aventure
./sw-adventure show "La Mine Perdue"

# Statut complet
./sw-adventure status "La Mine Perdue"

# Supprimer
./sw-adventure delete "La Mine Perdue"
```

### Gestion du Groupe

```bash
# Ajouter un personnage à l'aventure
./sw-adventure add-character "La Mine Perdue" "Aldric"
./sw-adventure add-character "La Mine Perdue" "Lyra"

# Retirer un personnage
./sw-adventure remove-character "La Mine Perdue" "Aldric"

# Afficher le groupe
./sw-adventure party "La Mine Perdue"
```

### Inventaire Partagé

```bash
# Voir l'inventaire
./sw-adventure inventory "La Mine Perdue"

# Ajouter de l'or
./sw-adventure add-gold "La Mine Perdue" 50 "Trésor gobelin"
./sw-adventure add-gold "La Mine Perdue" -10 "Achat de rations"

# Ajouter des objets
./sw-adventure add-item "La Mine Perdue" "Potion de soin" 3
./sw-adventure add-item "La Mine Perdue" "Corde 50 pieds"

# Retirer des objets
./sw-adventure remove-item "La Mine Perdue" "Potion de soin" 1
```

### Sessions de Jeu

```bash
# Démarrer une session
./sw-adventure start-session "La Mine Perdue"

# Terminer une session
./sw-adventure end-session "La Mine Perdue" "Le groupe a exploré le premier niveau"

# Lister les sessions
./sw-adventure sessions "La Mine Perdue"
```

### Journal Automatique

```bash
# Ajouter une entrée au journal
./sw-adventure log "La Mine Perdue" combat "Le groupe affronte 3 gobelins"
./sw-adventure log "La Mine Perdue" loot "Trouvé 20 po et une dague +1"
./sw-adventure log "La Mine Perdue" story "Les aventuriers arrivent à Valdris"
./sw-adventure log "La Mine Perdue" quest "Nouvelle quête: Retrouver le marchand"

# Voir le journal
./sw-adventure journal "La Mine Perdue"

# Journal d'une session spécifique
./sw-adventure journal "La Mine Perdue" --session=1

# Dernières entrées
./sw-adventure journal "La Mine Perdue" --recent=10
```

## Types d'Entrées Journal

| Type | Icône | Usage |
|------|-------|-------|
| `combat` | ⚔️ | Rencontres et combats |
| `loot` | 💰 | Trésors trouvés |
| `story` | 📖 | Progression narrative |
| `note` | 📝 | Notes diverses |
| `quest` | 🎯 | Quêtes et objectifs |
| `npc` | 👤 | Interactions PNJ |
| `location` | 📍 | Nouveaux lieux |
| `rest` | 🏕️ | Repos |
| `death` | 💀 | Morts de personnages |
| `levelup` | ⬆️ | Montées de niveau |

## Structure des Fichiers

Une aventure crée le répertoire suivant :

```
data/adventures/la-mine-perdue/
├── adventure.json         # Métadonnées de l'aventure
├── party.json             # Groupe et formation
├── inventory.json         # Inventaire partagé
├── sessions.json          # Historique des sessions
├── journal-meta.json      # Métadonnées journal (NextID, Categories)
├── journal-session-0.json # Journal hors session
├── journal-session-1.json # Journal session 1
├── journal-session-N.json # Journal session N
├── state.json             # État du jeu
├── images/
│   ├── session-0/         # Images hors session
│   ├── session-1/         # Images session 1
│   └── session-N/         # Images session N
└── characters/            # Copies des personnages
    ├── aldric.json
    └── lyra.json
```

**Note** : Le journal est organisé par session pour optimiser les performances. Les commandes CLI fonctionnent de manière transparente avec cette structure.

## Workflow Typique

### 1. Créer l'aventure
```bash
./sw-adventure create "La Mine Perdue" "Les aventuriers explorent une mine abandonnée"
```

### 2. Ajouter les personnages
```bash
./sw-adventure add-character "La Mine Perdue" "Aldric"
./sw-adventure add-character "La Mine Perdue" "Lyra"
./sw-adventure add-character "La Mine Perdue" "Gorim"
```

### 3. Démarrer une session
```bash
./sw-adventure start-session "La Mine Perdue"
```

### 4. Pendant la partie
```bash
# Noter les événements importants
./sw-adventure log "La Mine Perdue" story "Les aventuriers arrivent à l'entrée de la mine"
./sw-adventure log "La Mine Perdue" combat "Combat contre 4 gobelins - victoire"
./sw-adventure add-gold "La Mine Perdue" 35 "Butin gobelins"
./sw-adventure log "La Mine Perdue" loot "Trouvé: épée courte, 35 po"
```

### 5. Terminer la session
```bash
./sw-adventure end-session "La Mine Perdue" "Premier niveau de la mine exploré"
```

### 6. Consulter le statut
```bash
./sw-adventure status "La Mine Perdue"
```

## Intégration avec autres Skills

- **dice-roller** : Pour les jets de dés pendant la partie
- **character-generator** : Pour créer les personnages avant de les ajouter

## Exemple de Sortie

### Commande `status`
```markdown
# La Mine Perdue

*Les aventuriers explorent une mine abandonnée*

## Informations
- **Statut** : active
- **Sessions** : 3
- **Dernière partie** : 15/12/2024 20:30

## Groupe
**Formation** : travel
- Aldric (human fighter N1) - PV: 9/9
- Lyra (elf magic-user N1) - PV: 5/5
- Gorim (dwarf cleric N1) - PV: 7/7

## Inventaire
**Or** : 185 po
**Objets** : 5

## Derniers événements
- `15/12 20:15` 📖 Découverte d'une salle secrète
- `15/12 20:00` ⚔️ Combat contre le chef gobelin
- `15/12 19:45` 💰 Trouvé coffre: 50 po, potion
```

## Conseils d'Utilisation

- Démarrez toujours une session avant de jouer pour tracker le temps
- Utilisez `log` régulièrement pour maintenir un historique
- Les événements sont automatiquement horodatés
- L'or peut être négatif pour les dépenses (utilisez un nombre négatif)
- Le journal génère automatiquement un résumé par session

## Utilisé par

Ce skill est utilisé par les agents suivants :

| Agent | Usage |
|-------|-------|
| `dungeon-master` | Gestion sessions, journal, inventaire |

**Type** : Skill autonome, peut être invoqué directement via `/adventure-manager`

**Dépendances** : Utilise `dice-roller` et `character-generator` en complément